package websocket

import (
	"sync"

	"github.com/dgrr/websocket"
	"github.com/valyala/fasthttp"
)

type Server struct {
	ws         *websocket.Server
	clients    map[int]*Client
	clientLock sync.RWMutex
	onAccept   func(client *Client)
}

func (s *Server) Upgrade(ctx *fasthttp.RequestCtx) {
	s.ws.Upgrade(ctx)
}

func (s *Server) OnAccept(fn func(client *Client)) {
	s.onAccept = fn
}

func (s *Server) handleOpen(c *websocket.Conn) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	client := newClient(c)
	s.clients[int(c.ID())] = client
	if (s.onAccept) != nil {
		s.onAccept(client)
	}
}
func (s *Server) handleClose(c *websocket.Conn, err error) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	delete(s.clients, int(c.ID()))
}
func (s *Server) handleError(c *websocket.Conn, err error) {
	s.clientLock.Lock()
	defer s.clientLock.Unlock()
	delete(s.clients, int(c.ID()))
}
func (s *Server) handleData(c *websocket.Conn, isBinary bool, data []byte) {
	s.clientLock.RLock()
	defer s.clientLock.RUnlock()
	if client, ok := s.clients[int(c.ID())]; ok {
		client.receiveMessage(isBinary, data)
	}
}

func (s *Server) initWebsocket() {
	ws := new(websocket.Server)
	ws.HandleOpen(s.handleOpen)
	ws.HandleClose(s.handleClose)
	ws.HandleError(s.handleError)
	ws.HandleData(s.handleData)
	s.ws = ws
}

func NewServer() *Server {
	svr := &Server{
		clients: make(map[int]*Client),
	}
	svr.initWebsocket()
	return svr
}
