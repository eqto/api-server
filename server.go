package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gitlab.com/tuxer/go-json"

	"gitlab.com/tuxer/go-db"
)

//Server ...
type Server struct {
	getMap  map[string]*Parameter
	postMap map[string]*Parameter

	database struct {
		hostname string
		port     int
		username string
		password string
		name     string
	}
	apiFile string

	port int
	cn   *db.Connection
}

//Start ...
func (s *Server) Start() error {
	if e := s.parseAPIFile(); e != nil {
		return e
	}
	c := s.database
	cn, e := db.NewConnection(c.hostname, c.port, c.username, c.password, c.name)
	if e != nil {
		return fmt.Errorf(`unable to open database connection %s@%s:%d/%s`, c.username, c.hostname, c.port, c.name)
	}
	logDebug(fmt.Sprintf(`database connection %s@%s:%d/%s`, c.username, c.hostname, c.port, c.name))
	s.cn = cn

	addr := `:` + strconv.Itoa(s.Port())

	svr := &http.Server{
		Addr:           addr,
		Handler:        s,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ch := make(chan error)
	go func() {
		logDebug(`starting server at`, addr)
		if e := svr.ListenAndServe(); e != nil {
			if ch != nil {
				ch <- e
			}
		}
	}()
	select {
	case <-time.After(2 * time.Second):
		logInfo(`server listening at`, addr)
	case e := <-ch:
		return e
	}

	ch = nil
	return nil
}

func (s *Server) parseAPIFile() error {
	js, e := json.ParseFile(s.apiFile)
	if e != nil {
		if errors.Is(e, os.ErrNotExist) {
			return fmt.Errorf(`configuration file %s not found, SetAPIConfig() first`, s.apiFile)
		}
	}

	regex := regexp.MustCompile(`^(GET|POST)(?:,(GET|POST)|) (/\S*)$`)
	s.getMap = make(map[string]*Parameter)
	s.postMap = make(map[string]*Parameter)
	for key, val := range js {
		m := regex.FindStringSubmatch(key)
		if m == nil {
			return fmt.Errorf(`invalid API: %s`, key)
		}
		j, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf(`invalid API value for %s`, key)
		}
		p := newParameter(json.Object(j))

		s.registerAPI(m[1], m[3], p)
		s.registerAPI(m[2], m[3], p)
	}
	return nil
}

func (s *Server) registerAPI(method, path string, p *Parameter) {
	switch method {
	case `GET`:
		s.getMap[path] = p
	case `POST`:
		s.postMap[path] = p
	}

}

//SetPort ...
func (s *Server) SetPort(port int) {
	s.port = port
}

//Port ...
func (s *Server) Port() int {
	if s.port == 0 {
		return 8000
	}
	return s.port
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resp := json.Object{
		`status`:  0,
		`message`: `success`,
	}

	defer func() {
		if r := recover(); r != nil {
			resp.Put(`status`, 99).Put(`message`, `Error`)
			switch r := r.(type) {
			case error:
				logDebug(r)
				// resp.Put(`message`, r.Error())
			case string:
				resp.Put(`message`, r)
			}
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Header().Set(`Content-Type`, `application/json`)
		w.Write(resp.ToBytes())
	}()

	m := s.postMap
	if r.Method == http.MethodGet {
		m = s.getMap
	}

	path := r.URL.Path
	if strings.HasPrefix(path, `/`) {
		p, ok := m[path]
		if !ok {
			panic(`resource not found`)
		}
		s.processParam(parseRequest(r), &resp, p)
	}
}

func (s *Server) processParam(req *Request, resp *json.Object, p *Parameter) {
	values := []interface{}{}

	for _, val := range p.params {
		values = append(values, req.MustString(val))
	}
	qb := *p.qb

	if filter := req.GetJSONObject(`filter`); filter != nil {
		for keyFilter := range filter {
			valFilter := filter.GetJSONObject(keyFilter)
			value := valFilter.GetString(`value`)

			switch valFilter.GetString(`type`) {
			case `input`:
				qb.WhereOp(keyFilter, ` LIKE `)
				values = append(values, value+`%`)
			case `number`:
				value = strings.TrimSpace(value)
				if strings.HasPrefix(value, `<`) {
					value = strings.TrimSpace(value[1:])
					qb.WhereOp(keyFilter, ` < `)
				} else if strings.HasPrefix(value, `>`) {
					value = strings.TrimSpace(value[1:])
					qb.WhereOp(keyFilter, ` > `)
				}
				values = append(values, value)
			default:
				qb.Where(keyFilter)
				values = append(values, value)
			}
		}
	}
	switch p.queryType {
	case `INSERT`:

	case `UPDATE`:

	case `GET`:
		fallthrough
	case `SELECT`:
		page := req.GetInt(`page`)
		maxRows := req.GetInt(`max_rows`)
		if maxRows == 0 {
			maxRows = 100
		}
		if page >= 1 {
			qb.Limit((page-1)*maxRows, maxRows)
		}
		if p.queryType == `GET` {
			qb.Limit(qb.LimitStart(), 1)
		}
		active, direction := req.GetString(`sort.active`), req.GetString(`sort.direction`)

		if active != `` && direction != `` {
			qb.Order(active, direction)
		}
		if p.queryType == `GET` {
			resp.Put(`data`, s.cn.MustGet(qb.ToSQL(), values...))
		} else {
			resp.Put(`data`, s.cn.MustSelect(qb.ToSQL(), values...))
		}
	}
}

//SetDatabase ...
func (s *Server) SetDatabase(hostname string, port int, username string, password string, name string) {
	s.database.hostname = hostname
	s.database.port = port
	s.database.username = username
	s.database.password = password
	s.database.name = name
}

//SetAPIFile ...
func (s *Server) SetAPIFile(apiFile string) {
	s.apiFile = apiFile
}
