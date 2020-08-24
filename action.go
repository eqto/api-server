package apims

const (
	actionTypeList = iota
	actionTypeGet
	actionTypeInsert
	actionTypePHP
)

//Action ...
type Action interface {
	execute(ctx *actionCtx) (interface{}, error)
	property() string
	params() []string
}
