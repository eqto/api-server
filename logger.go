package api

import (
	"fmt"
	"log"
	"os"
)

//Logger ...
type Logger interface {
	D(v ...interface{})
	I(v ...interface{})
	W(v ...interface{})
	E(v ...interface{})
	F(v ...interface{})
	SetCallDepth(depth int)
}

type stdLogger struct {
	Logger
}

func (s *stdLogger) D(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}
func (s *stdLogger) I(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}
func (s *stdLogger) W(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}
func (s *stdLogger) E(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
}
func (s *stdLogger) F(v ...interface{}) {
	log.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}
func (s *stdLogger) SetCallDepth(depth int) {
}
