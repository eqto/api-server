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
	// key: METHOD_path => GET_users POST_users POST_command_users
	// value: *RouteConfig or []*RouteConfig
	routeMap map[string][]RouteConfig

	//supported auth: jwt
	authType  string
	authQuery string

	configFile string
	config     json.Object

	port int
	cn   *db.Connection

	logger Logger
}

//Start ...
func (s *Server) Start() error {
	cfg, e := json.ParseFile(s.configFile)
	if e != nil {
		if errors.Is(e, os.ErrNotExist) {
			return fmt.Errorf(`configuration file %s not found, SetAPIConfig() first`, s.configFile)
		}
	}
	s.logger.D(fmt.Sprintf(`open config file %s`, s.configFile))
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
	s.logger.D(fmt.Sprintf(`database connection %s@%s:%d/%s`, dbUsername, dbHostname, dbPort, dbName))

	addr := `:` + strconv.Itoa(s.Port())
	if auth := cfg.GetJSONObject(`auth`); auth != nil {
		s.authType = auth.GetStringOr(`type`, `jwt`)
		s.authQuery = auth.GetString(`query`)
		if s.authQuery == `` {
			return fmt.Errorf(`incorrect configuration: %s`, `auth.query`)
		}
	}

	if e := s.parseJSONRoute(cfg.GetJSONObject(`route`)); e != nil {
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
		s.logger.D(`starting server at`, addr)
		if e := svr.ListenAndServe(); e != nil {
			if ch != nil {
				ch <- e
			}
		}
	}()
	select {
	case <-time.After(2 * time.Second):
		s.logger.I(`server listening at`, addr)
	case e := <-ch:
		return e
	}

	ch = nil
	return nil
}

func (s *Server) addRouteMap(method, path string, routeCfg []RouteConfig) error {
	if s.routeMap == nil {
		s.routeMap = make(map[string][]RouteConfig)
	}
	method = strings.ToUpper(method)
	if method != `GET` && method != `POST` {
		return fmt.Errorf(`unable to register route %s with method %s`, path, method)
	}
	s.routeMap[method+` `+path] = routeCfg
	return nil
}

func (s *Server) getRouteMap(method, path string) ([]RouteConfig, error) {
	method = strings.ToUpper(method)
	if routeCfg, ok := s.routeMap[method+` `+path]; ok {
		return routeCfg, nil
	}
	return nil, fmt.Errorf(`unable to get route %s with method %s`, path, method)
}

func (s *Server) parseJSONRoute(routes json.Object) error {
	regex := regexp.MustCompile(`^(GET|POST)(?:,(GET|POST)|) (/\S*)$`)

	for key, val := range routes {
		m := regex.FindStringSubmatch(key)
		if m == nil {
			return fmt.Errorf(`invalid API: %s`, key)
		}
		configs := []RouteConfig{}

		if j, ok := val.(map[string]interface{}); ok {
			configs = append(configs, *newRouteConfig(json.Object(j)))
		} else {
			arr, ok := val.([]interface{})
			if !ok {
				return fmt.Errorf(`invalid API value for %s`, key)
			}
			for _, val := range arr {
				if j, ok := val.(map[string]interface{}); ok {
					configs = append(configs, *newRouteConfig(json.Object(j)))
				} else {
					return fmt.Errorf(`invalid API value for %s`, key)
				}
			}
		}
		if len(configs) > 0 {
			if e := s.addRouteMap(m[1], m[3], configs); e != nil {
				s.logger.W(e)
			}
			if e := s.addRouteMap(m[2], m[3], configs); e != nil {
				s.logger.W(e)
			}
		}
	}
	return nil
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
	req := parseRequest(s, r)
	resp := &Response{Object: json.Object{`status`: 0, `message`: `success`}}

	defer func() {
		w.Header().Set(`Content-Type`, `application/json`)
		if r := recover(); r != nil {
			resp.Put(`status`, 99).Put(`message`, `Error`)
			switch r := r.(type) {
			case error:
				s.logger.D(r)
			case string:
				resp.Put(`message`, r)
			}
			w.WriteHeader(http.StatusBadRequest)
		}
		w.Write(resp.ToBytes())
	}()

	path := r.URL.Path
	if strings.HasPrefix(path, `/`) {
		configs, ok := s.routeMap[r.Method+` `+r.URL.Path]
		if !ok {
			panic(`resource not found`)
		}

		tx, e := s.cn.Begin()
		if e != nil {
			panic(e)
		}
		defer tx.MustRecover()

		for _, val := range configs {
			if val.secure {
				if e := req.Authenticate(); e != nil {
					panic(e)
				}
			}
			ctx := Context{req: req, resp: resp, tx: tx}
			if val.routeFunc != nil {
				if e := val.routeFunc(ctx); e != nil {
					panic(e)
				}
			} else {
				if e := val.process(ctx); e != nil {
					panic(e)
				}
			}
		}
	} else {

	}
}

//Route ...
func (s *Server) Route(method []string, path string, routeFunc RouteFunc) error {
	routes := []RouteConfig{RouteConfig{routeFunc: routeFunc}}
	for _, val := range method {
		if e := s.addRouteMap(val, path, routes); e != nil {
			return e
		}
	}
	return nil
}

//GET ...
func (s *Server) GET(path string, routeFunc RouteFunc) error {
	return s.Route([]string{MethodGet}, path, routeFunc)
}

//POST ...
func (s *Server) POST(path string, routeFunc RouteFunc) error {
	return s.Route([]string{MethodPost}, path, routeFunc)
}

//SetConfig ...
func (s *Server) SetConfig(configFile string) {
	s.configFile = configFile
}

//SetLogger ...
func (s *Server) SetLogger(logger Logger) {
	s.logger = logger
}

//New ...
func New() *Server {
	return &Server{
		logger: new(stdLogger),
	}
}
