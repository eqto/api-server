package api

//Action ...
type Action interface {
	execute(*context) error
	property() string
	params() []string
}
