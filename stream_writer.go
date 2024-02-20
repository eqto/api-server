package api

import "bufio"

type streamWriter struct {
	Writer
	doneCh chan bool
	writer *bufio.Writer
}

func (s *streamWriter) write(w *bufio.Writer) {
	s.doneCh = make(chan bool)
	s.writer = w
	<-s.doneCh
}

func (s *streamWriter) Write(data []byte) (int, error) {
	return s.writer.Write(data)
}

func (s *streamWriter) Flush() error {
	return s.writer.Flush()
}

func (s *streamWriter) Close() error {
	s.doneCh <- true
	return nil
}
