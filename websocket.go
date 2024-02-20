package api

import "github.com/eqto/api-server/websocket"

type Websocket struct {
	wsServ *websocket.Server
}

func (w *Websocket) OnAccept(fn func(client *websocket.Client)) *Websocket {
	w.wsServ.OnAccept(fn)
	return w
}
