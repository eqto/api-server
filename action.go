package api

//Action ...
type Action interface {
	AssignTo(prop string) Action

	execute(*context) error
	property() string
	params() []string
}
