package api

const (
	actionTypeList = iota
	actionTypeGet
	actionTypeInsert
	actionTypePHP
)

//Action ...
type Action interface {
	execute(ctx *context) (interface{}, error)
	property() string
	params() []string
}
