package log

import (
	"container/list"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Worker struct {
	Prefix  string
	Level   LogLevel
	Writers []LogWriter
}

type LogHandler struct {
	Message string
	Worker  *Worker
}

type LogMsg struct {
	Message string
	Out     LogWriter
}

type LogMaster struct {
	workers *list.List
}

var (
	logMaster  = &LogMaster{}
	msgChannel chan LogMsg
)

func init() {
	logMaster.workers = list.New()
	// root logger
	Register("/", DEBUG, NewConsoleAppender(false))
	// buffered channels are asynchronous
	msgChannel = make(chan LogMsg, 100)
	go log()
}

/*
func SetLogMaster(master LogMaster) {
	if logMaster != nil {
		logMaster.Close()
	}
	logMaster = master
}
*/

// normalize namespace
func normalizeNamespace(namespace string) string {
	if !strings.HasPrefix(namespace, "/") {
		namespace = "/" + namespace
	}
	if !strings.HasSuffix(namespace, "/") {
		namespace += "/"
	}
	return namespace
}

func Register(namespace string, level LogLevel, writers ...LogWriter) {
	namespace = normalizeNamespace(namespace)

	// if there is no supplied writers use the ones from the parent
	worker := &Worker{Prefix: namespace, Level: level, Writers: writers}
	if writers == nil {
		wrk := logMaster.fetchWorker(namespace)
		worker.Writers = wrk.Writers
	}

	// Iterate through list of workers
	var processed bool
	for e := logMaster.workers.Front(); e != nil; e = e.Next() {
		wrk := e.Value.(*Worker)
		if namespace == wrk.Prefix {
			// replace
			e.Value = worker
			processed = true
		} else if namespace > wrk.Prefix {
			logMaster.workers.InsertBefore(worker, e)
			processed = true
			break
		}
	}

	// namespaces are positioned in descending order
	if !processed {
		logMaster.workers.PushBack(worker)
	}
}

func LoggerFor(namespace string) *Logger {
	logger := new(Logger)
	logger.tag = namespace
	return logger
}

func (this *LogMaster) fetchWorker(tag string) *Worker {
	namespace := normalizeNamespace(tag)
	for e := this.workers.Front(); e != nil; e = e.Next() {
		wrk := e.Value.(*Worker)
		if strings.HasPrefix(namespace, wrk.Prefix) {
			return wrk
		}
	}
	panic(fmt.Sprintf("No Worker was found for %s", namespace))
}

func shutdown() {
	// signal shutdown to the go routine
	close(msgChannel)
}

func log() {
	for {
		msg, ok := <-msgChannel
		if !ok {
			// EXIT
			return
		}

		msg.Out.Write([]byte(msg.Message))
	}
}

type LogLevel int

func (this LogLevel) String() string {
	switch this {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return ""
}

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	NONE
)

type Logger struct {
	sync.Mutex

	tag       string
	worker    *Worker
	calldepth int
}

func (this *Logger) loadWorker() {
	this.Lock()
	defer this.Unlock()

	if this.worker == nil {
		this.worker = logMaster.fetchWorker(this.tag)
	}
}

func (this *Logger) Level() LogLevel {
	if this.worker == nil {
		this.loadWorker()
	}
	return this.worker.Level
}

func (this *Logger) Namespace() string {
	return this.tag
}

func (this *Logger) CallDepth(depth int) {
	this.calldepth = depth
}

func (this *Logger) logStamp(level LogLevel) string {
	t := time.Now()
	var fl string
	if this.calldepth > 0 {
		_, file, line, ok := runtime.Caller(this.calldepth + 2)
		if ok {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		} else {
			file = "<Undetermined Caller>"
			line = 0
		}
		fl = fmt.Sprintf(" %s:%d", file, line)
	}
	return fmt.Sprintf("%d/%02d/%02d-%02d:%02d:%02d %s%s: ",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		level,
		fl)
}

func (this *Logger) IsActive(level LogLevel) bool {
	return level >= this.Level()
}

func (this *Logger) logf(level LogLevel, format string, what ...interface{}) {
	if this.IsActive(level) {
		str := this.logStamp(level)
		if len(what) > 0 {
			str += fmt.Sprintf(format+"\n", what...)
		} else {
			str += format + "\n"
		}
		//fmt.Printf("=====> %s", str)
		flush(str, this.worker.Writers)
	}
}

func (this *Logger) logF(level LogLevel, handler func() string) {
	if this.IsActive(level) {
		str := this.logStamp(level) + handler() + "\n"
		flush(str, this.worker.Writers)
	}
}

type LogWriter interface {
	io.Writer
	IsAsync() bool
}

func flush(msg string, workers []LogWriter) {
	for _, v := range workers {
		if v.IsAsync() {
			msgChannel <- LogMsg{msg, v}
		} else {
			v.Write([]byte(msg))
		}
	}
}

func (this *Logger) Debugf(format string, what ...interface{}) {
	this.logf(DEBUG, format, what...)
}

func (this *Logger) DebugF(handler func() string) {
	this.logF(DEBUG, handler)
}

func (this *Logger) Infof(format string, what ...interface{}) {
	this.logf(INFO, format, what...)
}

func (this *Logger) InfoF(handler func() string) {
	this.logF(INFO, handler)
}

func (this *Logger) Warnf(format string, what ...interface{}) {
	this.logf(WARN, format, what...)
}

func (this *Logger) WarnF(handler func() string) {
	this.logF(WARN, handler)
}

func (this *Logger) Errorf(format string, what ...interface{}) {
	this.logf(ERROR, format, what...)
}

func (this *Logger) ErrorF(handler func() string) {
	this.logF(ERROR, handler)
}

func (this *Logger) Fatalf(format string, what ...interface{}) {
	this.logf(FATAL, format, what...)
}

func (this *Logger) FatalF(handler func() string) {
	this.logF(FATAL, handler)
}
