package api

import "github.com/eqto/go-json"

//Session ...
type Session interface {
	Put(key string, value interface{})
	Get(key string) interface{}
	GetString(key string) string
}

type session struct {
	Session
	val json.Object
}

func (s *session) init() {
	if s.val == nil {
		s.val = make(json.Object)
	}
}

func (s *session) Put(key string, value interface{}) {
	s.init()
	s.val.Put(key, value)
}

func (s *session) Get(key string) interface{} {
	s.init()
	return s.val.Get(key)
}

func (s *session) GetString(key string) string {
	s.init()
	return s.val.GetString(key)
}
