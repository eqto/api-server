package api

//Action ...
type Action interface {
	execute(*context) (interface{}, error)
	property() string
	params() []string
}
