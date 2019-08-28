package logging

import (
	"fmt"
	l "log"
	"os"
)

type Logger struct {
	i *l.Logger
	e *l.Logger
	t *l.Logger
}

// NewLog : Instantiate a new logger
func NewLog() Logger {
	return Logger{
		l.New(os.Stdout, "INFO  - ", l.Ldate|l.Ltime),
		l.New(os.Stdout, "ERROR - ", l.Ldate|l.Ltime),
		l.New(os.Stdout, "TRACE - ", l.Ldate|l.Ltime),
	}
}

func (l Logger) Error(f string, s ...interface{}) {
	l.e.Print(f, " ", fmt.Sprintln(s...))
}

func (l Logger) Info(f string, s ...interface{}) {
	l.i.Print(f, " ", fmt.Sprintln(s...))
}

func (l Logger) Trace(f string, s ...interface{}) {
	l.t.Print(f, " ", fmt.Sprintln(s...))
}
