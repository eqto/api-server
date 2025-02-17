package api

// Action ...
type Action interface {
	AssignTo(prop string) Action

	execute(*Context) error
	property() string
	params() []string
}
