package api

import "bufio"

type StreamWriter struct {
	doneCh chan bool
	writer *bufio.Writer
}

func (s *StreamWriter) write(w *bufio.Writer) {
	s.doneCh = make(chan bool)
	s.writer = w
	<-s.doneCh
}

func (s *StreamWriter) Write(data []byte) (int, error) {
	return s.writer.Write(data)
}

func (s *StreamWriter) Flush() error {
	return s.writer.Flush()
}

func (s *StreamWriter) Close() {
	s.doneCh <- true
}
