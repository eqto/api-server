package api

import (
	"fmt"
	"log"
	"os"
	"reflect"
)

type stdLogger interface {
	Output(depth int, msg string) error
}

type goLogger interface {
	D(v ...interface{})
	I(v ...interface{})
	W(v ...interface{})
	E(v ...interface{})
	F(v ...interface{})
}

var (
	stdL stdLogger
	goL  goLogger
)

func logDebug(v ...interface{}) {
	checkLogger()
	if stdL != nil {
		stdL.Output(2, fmt.Sprintln(v...))
	} else if goL != nil {
		goL.D(v...)
	}
}
func logInfo(v ...interface{}) {
	checkLogger()
	if stdL != nil {
		stdL.Output(2, fmt.Sprintln(v...))
	} else if goL != nil {
		goL.I(v...)
	}
}
func logError(v ...interface{}) {
	checkLogger()
	if stdL != nil {
		stdL.Output(2, fmt.Sprintln(v...))
	} else if goL != nil {
		goL.E(v...)
	}
}
func logFatal(v ...interface{}) {
	checkLogger()
	if stdL != nil {
		stdL.Output(2, fmt.Sprintln(v...))
		os.Exit(1)
	} else if goL != nil {
		goL.F(v...)
	}
}

func checkLogger() {
	if stdL == nil && goL == nil {
		stdL = log.New(os.Stderr, ``, log.LstdFlags|log.Lshortfile)
	}
}

//SetLogger ...
func SetLogger(logger interface{}) {
	if reflect.TypeOf(logger).Kind() != reflect.Ptr {
		logger = &logger
	}

	t := reflect.TypeOf(logger)
	if _, ok := t.MethodByName(`Output`); ok {
		stdL = logger.(stdLogger)
	} else if _, ok := t.MethodByName(`D`); ok {
		goL = logger.(goLogger)
	}
}
