package apims

const (
	actionTypeList = iota
	actionTypeGet
	actionTypeInsert
	actionTypePHP
)

//Action ...
type Action struct {
	command  string
	typ      int8
	params   string
	property string
}
