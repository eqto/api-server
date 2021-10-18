package api

type Render func(Context) bool

func render(ctx Context) bool {
	resp := ctx.Response()
	data := resp.Data()
	if data != nil {
		data.Put(`status`, resp.StatusCode()).Put(`message`, resp.StatusMessage())
		resp.setBody(data.ToBytes())
	}
	return true
}
