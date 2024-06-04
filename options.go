package api

import "time"

func OptionTimeout(timeout time.Duration) ServerOptions {
	return func(s *Server) {
		if s != nil {
			s.timeout = timeout
		}
	}
}
