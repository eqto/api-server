package api

type Render func(req Request, resp Response) bool

func render(req Request, resp Response) bool {
	data := resp.Data()
	if data != nil {
		data.Put(`status`, resp.StatusCode()).Put(`message`, resp.StatusMessage())
		resp.SetBody(data.ToBytes())
	}
	return true
}
