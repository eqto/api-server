package api

type Render func(Context) bool

func render(ctx Context) bool {
	resp := ctx.Response()
	data := resp.Data()
	if data != nil {
		pmsg := resp.statusMessage()
		if pmsg == nil {
			msg := `Success`
			pmsg = &msg
		}
		data.Put(`status`, resp.StatusCode()).Put(`message`, *pmsg)
		resp.setBody(data.Bytes())
	}
	return true
}
