package log

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const COUNTER_SEP = "-"

var _ LogWriter = &RollingFileAppender{}

func NewRollingFileAppender(file string, size int64, count int, async bool) *RollingFileAppender {
	this := new(RollingFileAppender)
	this.name, this.ext = splitNameExt(file)

	this.maxsize = size
	this.count = count

	// finds the last modified file to determine the current backup number
	log, err := filepath.Abs(file)
	if err != nil {
		panic(err)
	}
	folder := filepath.Dir(log)
	var maxtime time.Time
	var lastName string
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasPrefix(info.Name(), this.name) && strings.HasSuffix(info.Name(), "."+this.ext) && info.ModTime().After(maxtime) {
			maxtime = info.ModTime()
			lastName = info.Name()
		}
		return nil
	})
	if lastName != "" {
		name, _ := splitNameExt(lastName)
		parts := strings.Split(name, COUNTER_SEP)
		this.currentCount, err = strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		// check size of the current log
		finfo, err := os.Stat(lastName)
		if err == nil {
			this.written = finfo.Size()
		}
	}
	this.currentFilename = this.fullName()

	this.async = async
	return this
}

// split name and extension
func splitNameExt(file string) (string, string) {
	idx := strings.LastIndex(file, ".")
	if idx > 0 && idx < len(file)-1 {
		return file[:idx], file[idx+1:]
	} else {
		return file, ""
	}
}

// Roll the log file over a range of files once they go over the maxsize.
// If maxsize == 0, then the log file will never rool.
// If count == 0, then the backup log files will be infinite.
// The format of the backup log files will be <name>-<counter>.<extension>
type RollingFileAppender struct {
	name            string
	ext             string
	currentFilename string
	maxsize         int64
	written         int64
	count           int
	currentCount    int
	async           bool
}

// mark as AsyncLog
func (this *RollingFileAppender) IsAsync() bool {
	return this.async
}

func (this *RollingFileAppender) rollFile() {
	this.written = 0
	this.currentCount++
	if this.currentCount == this.count {
		this.currentCount = 0
	}
	this.currentFilename = this.fullName()
	os.Remove(this.currentFilename)
}

func (this *RollingFileAppender) fullName() string {
	name := this.name + COUNTER_SEP + strconv.Itoa(this.currentCount)
	if this.ext != "" {
		name += "." + this.ext
	}
	return name
}

func (this *RollingFileAppender) Write(p []byte) (n int, err error) {
	if this.written > this.maxsize {
		this.rollFile()
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
