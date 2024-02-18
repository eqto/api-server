package api

import (
	"github.com/dgrr/websocket"
	"github.com/valyala/fasthttp"
)

type wsServer struct {
	ws        *websocket.Server
	onMessage func(*WsContext, bool, []byte)
}

func (s *wsServer) handleData(c *websocket.Conn, isBinary bool, data []byte) {
	if s.onMessage == nil {
		return
	}

	s.onMessage(&WsContext{conn: c}, isBinary, data)
}

func (s *wsServer) OnMessage(fn func(ctx *WsContext, isBinary bool, data []byte)) {
	s.onMessage = fn
}

func (s *wsServer) Upgrade(fastCtx *fasthttp.RequestCtx) {
	s.ws.Upgrade(fastCtx)
}

func newWsServer() *wsServer {
	svr := new(wsServer)

	ws := new(websocket.Server)
	ws.HandleData(svr.handleData)
	svr.ws = ws
	return svr
}
