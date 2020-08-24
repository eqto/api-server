package apims

import "errors"

type funcAction struct {
	Action
	prop string
	f    ActionFunc
}

func (f *funcAction) execute(ctx *context) (interface{}, error) {
	if f.f == nil {
		return nil, errors.New(`nil func`)
	}
	return f.f(ctx)
}
func (f *funcAction) property() string {
	return f.prop
}
func (f *funcAction) params() []string {
	return nil
}

func newFuncAction(f ActionFunc, property string) (*funcAction, error) {
	return &funcAction{f: f, prop: property}, nil
}
