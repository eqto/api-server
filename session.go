package api

import (
	"fmt"
	"reflect"
	"strconv"
)

type Session struct {
	logger *logger
	val    map[string]interface{}
}

func (s *Session) init() {
	if s.val == nil {
		s.val = make(map[string]interface{})
	}
}

func (s *Session) Remove(key string) {
	s.init()
	delete(s.val, key)
}

func (s *Session) Put(key string, value interface{}) {
	s.init()
	s.val[key] = value
}

func (s *Session) Get(key string) interface{} {
	s.init()
	return s.val[key]
}

func (s *Session) GetString(key string) string {
	s.init()
	val, ok := s.val[key]
	if !ok || val == nil {
		return ``
	}
	switch val := val.(type) {
	case float32:
		return fmt.Sprintf(`%f`, val)
	case float64:
		return fmt.Sprintf(`%f`, val)
	case int:
		return strconv.Itoa(val)
	case string:
		return val
	}
	s.logger.W(fmt.Sprintf(`unable convert to string key:%s val: %v type: %s`, key, val, reflect.TypeOf(val).String()))
	return ``
}

func (s *Session) GetInt(key string) int {
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

	s.logger.W(fmt.Sprintf(`unable convert to string int:%s val: %v type: %s`, key, val, reflect.TypeOf(val).String()))
	return 0
}
