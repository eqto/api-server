package api

import "errors"

type actionFunc struct {
	Action
	prop string
	f    func(*Context) error
}

func (f *actionFunc) AssignTo(prop string) Action {
	f.prop = prop
	return f
}

func (f *actionFunc) execute(ctx *Context) error {
	if f.f == nil {
		return errors.New(`nil func`)
	}
	return f.f(ctx)
}
func (f *actionFunc) property() string {
	return f.prop
}
func (f *actionFunc) params() []string {
	return nil
}

func newFuncAction(f func(*Context) error) *actionFunc {
	return &actionFunc{f: f}
}
