package log

import (
	"io"
	"os"
)

func NewConsoleAppender(async bool) *Console {
	this := new(Console)
	this.Writer = io.Writer(os.Stdout)
	this.async = async
	return this
}

type Console struct {
	io.Writer
	async bool
}

func (this *Console) IsAsync() bool {
	return this.async
}
