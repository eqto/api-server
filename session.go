package apims

type Session interface {
	Get(key string) interface{}
	Put(key string, val interface{})
}
type session struct {
	Session
	val map[string]interface{}
}

func (s *session) Put(key string, val interface{}) {
	if s.val == nil {
		s.val = make(map[string]interface{})
	}
	s.val[key] = val
}

func (s *session) Get(key string) interface{} {
	if s.val == nil {
		s.val = make(map[string]interface{})
	}
	if val, ok := s.val[key]; ok {
		return val
	}
	return nil
}
