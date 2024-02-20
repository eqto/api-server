package websocket

import (
	"errors"

	"github.com/dgrr/websocket"
)

type Writer struct {
	conn *websocket.Conn
}

func (w *Writer) Write(data []byte) (int, error) {
	if w.conn != nil {
		return w.conn.Write(data)
	}
	return 0, errors.New(`invalid nil socket connection`)
}

func (w *Writer) Close() error {
	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}
