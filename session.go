package api

import (
	"fmt"
	"strconv"
)

//Session ...
type Session interface {
	Put(key string, value interface{})
	Get(key string) interface{}
	Remove(key string)
	GetInt(key string) int
	GetString(key string) string
}

type session struct {
	Session
	logger *logger
	val    map[string]interface{}
}

func (s *session) init() {
	if s.val == nil {
		s.val = make(map[string]interface{})
	}
}

func (s *session) Remove(key string) {
	s.init()
	delete(s.val, key)
}

func (s *session) Put(key string, value interface{}) {
	s.init()
	s.val[key] = value
}

func (s *session) Get(key string) interface{} {
	s.init()
	return s.val[key]
}

func (s *session) GetString(key string) string {
	s.init()
	val, ok := s.val[key]
	if !ok || val == nil {
		return ``
	}
	switch val := val.(type) {
	case int:
		return strconv.Itoa(val)
	case string:
		return val
	}
	s.logger.W(fmt.Sprintf(`unable convert to string key:%s val: %v`, key, val))
	return ``
}

func (s *session) GetInt(key string) int {
	s.init()
	val, ok := s.val[key]
	if !ok || val == nil {
		return 0
	}
	switch val := val.(type) {
	case int:
		return val
	case string:
		i, e := strconv.Atoi(val)
		if e != nil {
			return 0
		}
		return i
	}
	s.logger.W(fmt.Sprintf(`unable convert to string int:%s val: %v`, key, val))
	return 0
}
