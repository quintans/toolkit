package log

import (
	"io"
	"os"
)

// check if it implements LogWriter interface
var _ LogWriter = &Console{}

func NewConsoleAppender(async bool) *Console {
	this := new(Console)
	this.Writer = io.Writer(os.Stdout)
	if async {
		this.Channel = make(chan string, 10)
		go AsyncWriter(this.Channel, this)
	}
	return this
}

type Console struct {
	RootAppender
}
