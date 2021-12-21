package api

import "bufio"

type StreamWriter struct {
	dataCh chan []byte
	doneCh chan bool
}

func (s *StreamWriter) write(w *bufio.Writer) {
	s.dataCh = make(chan []byte)
	s.doneCh = make(chan bool)
	for {
		select {
		case data := <-s.dataCh:
			w.Write(data)
		case done := <-s.doneCh:
			if done {
				return
			} else {
				w.Flush()
			}
		}

	}
}

func (s *StreamWriter) Write(data []byte) {
	s.dataCh <- data
}

func (s *StreamWriter) Flush() {
	s.doneCh <- false
}

func (s *StreamWriter) Close() {
	s.doneCh <- true
}
