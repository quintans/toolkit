package log

import (
	"bytes"
	"container/list"
	"fmt"
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

type LogMaster struct {
	workers *list.List
}

var (
	logMaster    = &LogMaster{}
	quit         = make(chan struct{})
	showLevel    = true
	showCaller   bool
	timeFormater func(time.Time) string
)

func init() {
	logMaster.workers = list.New()
	// root logger
	Register("/", DEBUG, NewConsoleAppender(false))
	SetTimeFormat("%Y/%02M/%02D %02h:%02m:%02s.%03x")
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

	// if there are no supplied writers use the ones from the parent
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
			// stop the old writers if new ones are defined
			if writers != nil {
				for _, w := range wrk.Writers {
					w.Discard()
				}
			}
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
	namespace = normalizeNamespace(namespace)
	logger := new(Logger)
	logger.tag = namespace
	return logger
}

func ShowLevel(show bool) {
	showLevel = true
}

func ShowCaller(show bool) {
	showCaller = true
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

func Shutdown() {
	for e := logMaster.workers.Front(); e != nil; e = e.Next() {
		wrk := e.Value.(*Worker)
		for _, w := range wrk.Writers {
			w.Discard()
		}
	}
}

const (
	mark   = '%'
	tokens = "YMDhmsx"
)

// available formats: Y,M,D,h,m,s,x.
// these format will be replaced by 'd' and used normaly with fmt.Sprintf
func SetTimeFormat(format string) {
	if format == "" {
		timeFormater = nil
		return
	}

	newFormat := format
	keys := make([]rune, 0)
	guard := false
	last := false
	for k, v := range format {
		if v == mark {
			// check if previous was %
			if last {
				last = false
				guard = false
			} else {
				last = true
				guard = true
			}
		} else if x := isToken(v); guard && x != 0 {
			keys = append(keys, x)
			newFormat = newFormat[:k] + "d" + newFormat[k+1:]
			last = false
			guard = false
		} else {
			last = false
		}
	}
	timeFormater = func(t time.Time) string {
		params := make([]interface{}, 0)
		for _, v := range keys {
			switch v {
			case 'Y':
				params = append(params, t.Year())
			case 'M':
				params = append(params, t.Month())
			case 'D':
				params = append(params, t.Day())
			case 'h':
				params = append(params, t.Hour())
			case 'm':
				params = append(params, t.Minute())
			case 's':
				params = append(params, t.Second())
			case 'x':
				params = append(params, t.Nanosecond()/1e6)

			}
		}
		return fmt.Sprintf(newFormat, params...)
	}
}

func isToken(t rune) rune {
	for _, v := range tokens {
		if t == v {
			return v
		}
	}
	return 0
}

type LogLevel int

var logLevels = [...]string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

func (this LogLevel) String() string {
	var level = int(this)
	if level >= 0 && level <= len(logLevels) {
		return logLevels[level]
	} else {
		return ""
	}
}

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	NONE
)

func ParseLevel(name string, optional LogLevel) LogLevel {
	name = strings.ToUpper(name)
	for k, v := range logLevels {
		if v == name {
			return LogLevel(k)
		}
	}
	return optional
}

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

func (this *Logger) SetCallerAt(depth int) {
	this.calldepth = depth
}

func (this *Logger) CallerAt(depth int) *Logger {
	// creates a temporary logger
	tmp := LoggerFor(this.tag)
	tmp.calldepth = depth
	return tmp
}

func (this *Logger) logStamp(level LogLevel) string {
	t := time.Now()
	var result bytes.Buffer
	if timeFormater != nil {
		result.WriteString(timeFormater(t))
	}

	if showLevel {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		result.WriteString(level.String())
	}

	if showCaller {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		_, file, line, ok := runtime.Caller(this.calldepth + 3)
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
			file = "???"
			line = 0
		}
		if result.Len() > 0 {
			result.WriteString(fmt.Sprintf("[%s:%d]", file, line))
		}
	}
	result.WriteString(": ")
	return result.String()
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
		flush(level, str, this.worker.Writers)
	}
}

func (this *Logger) logF(level LogLevel, handler func() string) {
	if this.IsActive(level) {
		str := this.logStamp(level) + handler() + "\n"
		flush(level, str, this.worker.Writers)
	}
}

type LogWriter interface {
	Discard()
	Log(LogLevel, string)
}

func flush(msgLevel LogLevel, msg string, workers []LogWriter) {
	for _, v := range workers {
		v.Log(msgLevel, msg)
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
