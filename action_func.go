package api

import "errors"

type actionFunc struct {
	Action
	prop string
	f    ActionFunc
}

func (f *actionFunc) execute(ctx *context) (interface{}, error) {
	if f.f == nil {
		return nil, errors.New(`nil func`)
	}
	return f.f(ctx)
}
func (f *actionFunc) property() string {
	return f.prop
}
func (f *actionFunc) params() []string {
	return nil
}

func newFuncAction(f ActionFunc, property string) (*actionFunc, error) {
	return &actionFunc{f: f, prop: property}, nil
}
