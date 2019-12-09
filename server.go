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

	//supported auth: jwt
	authType  string
	authQuery string

	configFile string
	config     json.Object

	port int
	cn   *db.Connection
}

//Start ...
func (s *Server) Start() error {
	cfg, e := json.ParseFile(s.configFile)
	if e != nil {
		if errors.Is(e, os.ErrNotExist) {
			return fmt.Errorf(`configuration file %s not found, SetAPIConfig() first`, s.configFile)
		}
	}
	logInfo(fmt.Sprintf(`open config file %s`, s.configFile))
	s.config = cfg

	dbHostname := cfg.GetStringOr(`database.hostname`, `localhost`)
	dbPort := cfg.GetIntOr(`database.port`, 3306)
	dbUsername := cfg.GetString(`database.username`)
	dbPassword := cfg.GetString(`database.password`)
	dbName := cfg.GetString(`database.name`)
	cn, e := db.NewConnection(dbHostname, dbPort, dbUsername, dbPassword, dbName)
	if e != nil {
		return fmt.Errorf(`unable to open database connection %s@%s:%d/%s`, dbUsername, dbHostname, dbPort, dbName)
	}
	s.cn = cn
	logInfo(fmt.Sprintf(`database connection %s@%s:%d/%s`, dbUsername, dbHostname, dbPort, dbName))

	addr := `:` + strconv.Itoa(s.Port())
	if auth := cfg.GetJSONObject(`auth`); auth != nil {
		s.authType = auth.GetStringOr(`type`, `jwt`)
		s.authQuery = auth.GetString(`query`)
		if s.authQuery == `` {
			return fmt.Errorf(`incorrect configuration: %s`, `auth.query`)
		}
	}

	if e := s.parseAPI(); e != nil {
		return e
	}
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

func (s *Server) parseAPI() error {
	apis := s.config.GetJSONObject(`api`)

	regex := regexp.MustCompile(`^(GET|POST)(?:,(GET|POST)|) (/\S*)$`)
	s.getMap = make(map[string]*Parameter)
	s.postMap = make(map[string]*Parameter)
	for key, val := range apis {
		m := regex.FindStringSubmatch(key)
		if m == nil {
			return fmt.Errorf(`invalid API: %s`, key)
		}

		var p *Parameter
		if j, ok := val.(map[string]interface{}); ok {
			p = newParameter(json.Object(j))
		} else {
			arr, ok := val.([]interface{})
			if !ok {
				return fmt.Errorf(`invalid API value for %s`, key)
			}
			params := []Parameter{}
			for _, val := range arr {
				if j, ok := val.(map[string]interface{}); ok {
					p := newParameter(json.Object(j))
					params = append(params, *p)
				} else {
					return fmt.Errorf(`invalid API value for %s`, key)
				}
			}
			if len(params) > 0 {
				p = &Parameter{children: params, secure: params[0].secure}
			}
		}
		if p != nil {
			s.registerAPI(m[1], m[3], p)
			s.registerAPI(m[2], m[3], p)
		}
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
		w.Header().Set(`Content-Type`, `application/json`)
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
		logDebug(path, p.secure)

		if p.secure {
			var e error
			if s.authType == `jwt` {
				e = jwtAuthorize(s, r, &resp, p)
			}
			if e != nil {
				logError(e)
				panic(fmt.Sprintf(`authorization for %s failed`, path))
			}
		}

		if e := s.process(parseRequest(r), &resp, p); e != nil {
			logError(e)
			panic(fmt.Sprintf(`unable to process resource %s`, path))
		}
	} //TODO if path not prefix with /
}

func (s *Server) process(req *Request, resp *json.Object, p *Parameter) error {
	if p.isArray() {
		tx, e := s.cn.Begin()
		if e != nil {
			return e
		}
		defer tx.Commit()

		for _, val := range p.children {
			e := s.processParameter(tx, req, resp, &val)
			if e != nil {
				tx.Rollback()
				return e
			}
		}
	} else {
		return s.processParameter(s.cn.Tx(nil), req, resp, p)
	}
	return nil
}

func (s *Server) processParameter(tx *db.Tx, req *Request, resp *json.Object, p *Parameter) error {
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
		rs, e := tx.Exec(p.query, values...)
		if e != nil {
			return e
		}
		id, e := rs.LastInsertID()
		if e != nil {
			return e
		}
		p.putOutput(req, resp, id)

	case `UPDATE`:

	case `GET`:
		fallthrough
	case `SELECT`:
		page := req.GetInt(`page`)

		maxRows := req.GetInt(`max_rows`)
		if maxRows == 0 {
			maxRows = qb.LimitLength()
			if maxRows == 0 {
				maxRows = 100
			}
		}
		if page >= 1 {
			qb.Limit((page-1)*maxRows, maxRows)
		}
		active, direction := req.GetString(`sort.active`), req.GetString(`sort.direction`)

		if active != `` && direction != `` {
			qb.Order(active, direction)
		}

		if p.queryType == `GET` {
			qb.Limit(qb.LimitStart(), 1)
			rs, e := tx.Get(qb.ToSQL(), values...)
			if e != nil {
				return e
			}
			p.putOutput(req, resp, rs)
		} else {
			rs, e := tx.Select(qb.ToSQL(), values...)
			if e != nil {
				return e
			}
			p.putOutput(req, resp, rs)
		}
	}
	return nil
}

//SetConfig ...
func (s *Server) SetConfig(configFile string) {
	s.configFile = configFile
}
