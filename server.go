package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"gitlab.com/tuxer/go-api/auth"

	"gitlab.com/tuxer/go-json"

	"gitlab.com/tuxer/go-db"
)

//Func ...
type Func func(ctx Context, params ...Parameter) (interface{}, error)

//Server ...
type Server struct {
	port    int
	healthy int32

	authManager *auth.Manager

	configFile string
	config     json.Object
	cn         *db.Connection

	logger Logger

	funcMap map[string]Func

	serveMux *ServeMux
}

//OpenDatabase ...
func (s *Server) OpenDatabase(hostname string, port int, username, password, name string) error {
	cn, e := db.NewConnection(hostname, port, username, password, name)
	if e != nil {
		return fmt.Errorf(`unable to open database connection %s@%s:%d/%s`, username, hostname, port, name)
	}
	s.cn = cn
	s.logger.D(fmt.Sprintf(`Database connection %s@%s:%d/%s`, username, hostname, port, name))

	return nil
}

//Connection ...
func (s *Server) Connection() *db.Connection {
	return s.cn
}

//AddFunc ...
func (s *Server) AddFunc(name string, f Func) {
	if s.funcMap == nil {
		s.funcMap = make(map[string]Func)
	}
	s.funcMap[name] = f
	s.logger.D(fmt.Sprintf(`Add function: %s`, name))
}

//AddPaths ...
func (s *Server) AddPaths(paths json.Object) error {
	regex := regexp.MustCompile(`^(GET|POST)(?:,(GET|POST)|) (/\S*)$`)

	for key, path := range paths {
		m := regex.FindStringSubmatch(key)
		if m == nil {
			return fmt.Errorf(`invalid API: %s`, key)
		}
		routePath := new(RoutePath)

		if j, ok := path.(map[string]interface{}); ok {
			routePath.AddRoute(RouteFromJSON(json.Object(j)))
		} else {
			arr, ok := path.([]interface{})
			if !ok {
				return fmt.Errorf(`invalid API value for %s`, key)
			}
			for _, val := range arr {
				if j, ok := val.(map[string]interface{}); ok {
					routePath.AddRoute(RouteFromJSON(json.Object(j)))
				} else {
					return fmt.Errorf(`invalid API value for %s`, key)
				}
			}
		}
		if len(routePath.Routes()) > 0 {
			if e := s.SetRoutePath(m[1], m[3], routePath); e != nil {
				s.logger.W(e)
			}
			if m[2] != `` {
				if e := s.SetRoutePath(m[2], m[3], routePath); e != nil {
					s.logger.W(e)
				}
			}
		}
	}
	return nil
}

//SetRoutePath ...
func (s *Server) SetRoutePath(method, path string, routePath *RoutePath) error {
	return s.SetModuleRoutePath(``, method, path, routePath)
}

//SetModuleRoutePath ...
func (s *Server) SetModuleRoutePath(module, method, path string, routePath *RoutePath) error {
	method = strings.ToUpper(method)

	if s.serveMux == nil {
		s.serveMux = newServeMux(s)
	}
	s.serveMux.setRoutePath(module, method, path, routePath)
	return nil
}

//AddMiddleware ...
func (s *Server) AddMiddleware(middleware Middleware) {
	if s.serveMux == nil {
		s.serveMux = newServeMux(s)
	}
	s.serveMux.AddMiddleware(middleware)
}

//Start ...
func (s *Server) Start() error {
	if s.configFile != `` {
		cfg, e := json.ParseFile(s.configFile)
		if e != nil {
			if errors.Is(e, os.ErrNotExist) {
				return fmt.Errorf(`configuration file %s not found, SetConfig() first`, s.configFile)
			}
		}
		s.logger.D(fmt.Sprintf(`open config file %s`, s.configFile))
		s.config = cfg
		if s.cn == nil {
			dbHostname := cfg.GetStringOr(`database.hostname`, `localhost`)
			dbPort := cfg.GetIntOr(`database.port`, 3306)
			dbUsername := cfg.GetString(`database.username`)
			dbPassword := cfg.GetString(`database.password`)
			dbName := cfg.GetString(`database.name`)
			if e := s.OpenDatabase(dbHostname, dbPort, dbUsername, dbPassword, dbName); e != nil {
				return e
			}
		}
		if jsAuth := cfg.GetJSONObject(`auth`); jsAuth != nil {
			for key := range jsAuth {
				val := jsAuth.GetJSONObject(key)
				switch val.GetString(`type`) {
				case auth.TypeJWT:
					s.authManager.Set(key, &auth.JWTAuth{Query: val.GetString(`query`)})
				default:
					s.logger.W(fmt.Sprintf(`authentication type %s not supported or not loaded yet`, key))
				}
			}
		}
		if e := s.AddPaths(cfg.GetJSONObject(`paths`)); e != nil {
			return e
		}
	}

	if s.serveMux == nil {
		s.serveMux = newServeMux(s)
	}

	httpServer := &http.Server{
		Addr:           fmt.Sprintf(`:%d`, s.port),
		Handler:        s.serveMux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	go func() {
		<-quit
		s.logger.I(`Server shutting down...`)
		atomic.StoreInt32(&s.healthy, 0)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		httpServer.SetKeepAlivesEnabled(false)
		if err := httpServer.Shutdown(ctx); err != nil {
			s.logger.W(fmt.Sprintf(`Gracefully shutdown failed. %v`, err))
		}
		close(done)
	}()
	s.logger.I(fmt.Sprintf(`Server listening at %d`, s.port))

	atomic.StoreInt32(&s.healthy, 1)
	if e := httpServer.ListenAndServe(); e != nil && e != http.ErrServerClosed {
		s.logger.W(fmt.Sprintf(`Error starting server. %v`, e))
	}

	<-done
	s.logger.I(`Server stopped`)
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

//Route ...
func (s *Server) Route(method []string, path string, routeFunc RouteFunc) error {
	routePath := new(RoutePath)
	routePath.AddRoute(&Route{routeFunc: routeFunc})

	for _, val := range method {
		if e := s.SetRoutePath(val, path, routePath); e != nil {
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
	s.logger.SetCallDepth(0)
}

//New ...
func New(port int) *Server {
	return &Server{
		port:   port,
		logger: new(stdLogger),
	}
}
