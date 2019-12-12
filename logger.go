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

// type goLogger struct {
// 	Logger
// }

// func (g *g) Debug(v ...interface{}) {
// 	log.Output(2, fmt.Sprintln(v...))
// }
// func (s *stdLogger) Info(v ...interface{}) {
// 	log.Output(2, fmt.Sprintln(v...))
// }
// func (s *stdLogger) Warning(v ...interface{}) {
// 	log.Output(2, fmt.Sprintln(v...))
// }
// func (s *stdLogger) Error(v ...interface{}) {
// 	log.Output(2, fmt.Sprintln(v...))
// }
// func (s *stdLogger) Fatal(v ...interface{}) {
// 	log.Output(2, fmt.Sprintln(v...))
// 	os.Exit(1)
// }

// func logDebug(v ...interface{}) {
// 	// checkLogger()
// 	// if stdL != nil {
// 	// 	stdL.Output(2, fmt.Sprintln(v...))
// 	// } else if goL != nil {
// 	// 	goL.D(v...)
// 	// }
// }
// func logInfo(v ...interface{}) {
// 	checkLogger()
// 	if stdL != nil {
// 		stdL.Output(2, fmt.Sprintln(v...))
// 	} else if goL != nil {
// 		goL.I(v...)
// 	}
// }
// func logError(v ...interface{}) {
// 	checkLogger()
// 	if stdL != nil {
// 		stdL.Output(2, fmt.Sprintln(v...))
// 	} else if goL != nil {
// 		goL.E(v...)
// 	}
// }
// func logFatal(v ...interface{}) {
// 	checkLogger()
// 	if stdL != nil {
// 		stdL.Output(2, fmt.Sprintln(v...))
// 		os.Exit(1)
// 	} else if goL != nil {
// 		goL.F(v...)
// 	}
// }

// func checkLogger() {
// 	if stdL == nil && goL == nil {
// 		stdL = log.New(os.Stderr, ``, log.LstdFlags|log.Lshortfile)
// 	}
// }

// //SetLogger ...
// func SetLogger(logger interface{}) {
// 	if reflect.TypeOf(logger).Kind() != reflect.Ptr {
// 		logger = &logger
// 	}

// 	t := reflect.TypeOf(logger)
// 	if _, ok := t.MethodByName(`Output`); ok {
// 		stdL = logger.(stdLogger)
// 	} else if _, ok := t.MethodByName(`D`); ok {
// 		goL = logger.(goLogger)
// 	}
// }
