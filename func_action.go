package apims

import "errors"

type funcAction struct {
	Action
	f ActionFunc
}

func (f *funcAction) execute(ctx *context) (interface{}, error) {
	if f.f == nil {
		return nil, errors.New(`nil func`)
	}
	f.f(ctx)

	return nil, nil
}
func (f *funcAction) property() string {
	return ``
}
func (f *funcAction) params() []string {
	return nil
}

func newFuncAction(f ActionFunc, property string) (*funcAction, error) {
	return &funcAction{f: f}, nil
}
