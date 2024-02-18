package api

type Websocket struct {
	wsServ *wsServer
}

func (w *Websocket) OnMessage(fn func(ctx *WsContext, isBinary bool, data []byte)) {
	w.wsServ.OnMessage(fn)
}
