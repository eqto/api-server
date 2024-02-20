package websocket

import (
	"sync"

	"github.com/dgrr/websocket"
)

type Client struct {
	conn      *websocket.Conn
	writer    *Writer
	onMessage func(bool, []byte, *Writer)

	preBuffer  []bufferMsg
	bufferLock sync.Mutex
}

func (c *Client) receiveMessage(isBinary bool, data []byte) {
	if c.onMessage != nil {
		c.onMessage(isBinary, data, c.writer)
	} else {
		c.bufferLock.Lock()
		defer c.bufferLock.Unlock()
		c.preBuffer = append(c.preBuffer, bufferMsg{isBinary, data})
	}
}

func (c *Client) ID() uint64 {
	return c.conn.ID()
}

func (c *Client) OnMessage(fn func(isBinary bool, data []byte, w *Writer)) {
	if c.onMessage == nil {
		c.bufferLock.Lock()
		defer c.bufferLock.Unlock()
		for _, msg := range c.preBuffer {
			fn(msg.isBinary, msg.data, c.writer)
		}
		c.preBuffer = []bufferMsg{}
	}
	c.onMessage = fn
}

func (c *Client) Write(data []byte) (int, error) {
	return c.writer.Write(data)
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{conn: conn, writer: &Writer{conn: conn}}
}

type bufferMsg struct {
	isBinary bool
	data     []byte
}
