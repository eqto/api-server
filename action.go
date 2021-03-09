package api

const (
	actionTypeList = iota
	actionTypeGet
	actionTypeInsert
	actionTypePHP
)

//Action ...
type Action interface {
	execute(*context) (interface{}, error)
	property() string
	params() []string
}
