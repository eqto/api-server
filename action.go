package api

const (
	actionTypeList = iota
	actionTypeGet
	actionTypeInsert
	actionTypePHP
)

//Action ...
type Action interface {
	execute(ctx *ctx) (interface{}, error)
	property() string
	params() []string
}
