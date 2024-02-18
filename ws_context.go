package api

import "github.com/dgrr/websocket"

type WsContext struct {
	conn *websocket.Conn
}

func (w *WsContext) Write(data []byte) {
	w.conn.Write(data)
}
