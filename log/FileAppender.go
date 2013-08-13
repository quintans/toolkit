package log

import (
	"os"
	"path/filepath"
)

const COUNTER_SEP = "-"

var _ LogWriter = &FileAppender{}

func NewFileAppender(file string, size int64, resetOnStartup bool, async bool) *FileAppender {
	this := new(FileAppender)

	this.maxsize = size

	log, err := filepath.Abs(file)
	if err != nil {
		panic(err)
	}

	this.currentFilename = log

	if resetOnStartup {
		this.resetFile()
	}

	this.async = async
	return this
}

/*
Resets the log file as soon it goe over the maxsize.
If maxsize == 0, then the log file will never reset.
*/
type FileAppender struct {
	currentFilename string
	maxsize         int64
	written         int64
	async           bool
}

// mark as AsyncLog
func (this *FileAppender) IsAsync() bool {
	return this.async
}

func (this *FileAppender) resetFile() error {
	this.written = 0
	err := os.Remove(this.currentFilename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func (this *FileAppender) Write(p []byte) (n int, err error) {
	if this.maxsize > 0 && this.written > this.maxsize {
		if err = this.resetFile(); err != nil {
			return
		}
	}

	fo, err := os.OpenFile(this.currentFilename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	n, err = fo.Write(p)
	if this.maxsize > 0 {
		this.written += int64(n)
	}
	if err != nil {
		panic(err)
	}

	return n, nil
}
