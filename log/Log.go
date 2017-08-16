package log

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type Worker struct {
	Prefix  string
	Level   LogLevel
	Writers []LogWriter

	showLevel    bool
	showCaller   bool
	timeFormater func(time.Time) string
}

func (wrk *Worker) ShowLevel(show bool) {
	wrk.showLevel = show
}

func (wrk *Worker) ShowCaller(show bool) {
	wrk.showCaller = show
}

type LogHandler struct {
	Message string
	Worker  *Worker
}

type LogMaster struct {
	workers []*Worker
	loggers []*Logger
}

var (
	logMaster = &LogMaster{
		workers: make([]*Worker, 0),
		loggers: make([]*Logger, 0),
	}
)

func init() {
	// root logger
	Register("/", DEBUG, NewConsoleAppender(false))
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

func Register(namespace string, level LogLevel, writers ...LogWriter) *Worker {
	namespace = normalizeNamespace(namespace)

	// if there are no supplied writers use the ones from the parent
	worker := &Worker{
		Prefix:    namespace,
		Level:     level,
		Writers:   writers,
		showLevel: true,
	}
	if len(writers) == 0 {
		wrk := logMaster.fetchWorker(namespace)
		worker.Writers = wrk.Writers
	}

	if len(logMaster.workers) == 0 {
		logMaster.workers = append(logMaster.workers, worker)
	} else {
		for k, v := range logMaster.workers {
			if namespace > v.Prefix {
				// insert. Worker are inserted in descending order by Prefix
				var s = append(logMaster.workers, nil)
				copy(s[k+1:], s[k:])
				s[k] = worker
				logMaster.workers = s
				break
			} else if v.Prefix == namespace {
				// replace on match
				logMaster.workers[k] = worker
				break
			}
		}
	}

	// default timestamp
	worker.SetTimeFormat("%Y-%02M-%02D %02h:%02m:%02s.%03x")

	logMaster.fireWorkerListeners(worker)

	return worker
}

func RootLogger() *Logger {
	return LoggerFor("/")
}

func LoggerFor(namespace string) *Logger {
	namespace = normalizeNamespace(namespace)
	logger := new(Logger)
	logger.tag = namespace
	return logger
}

func (this *LogMaster) fetchWorker(tag string) *Worker {
	namespace := normalizeNamespace(tag)

	for _, v := range logMaster.workers {
		if strings.HasPrefix(namespace, v.Prefix) {
			return v
		}
	}

	panic(fmt.Sprintf("No Worker was found for %s", namespace))
}

func (this *LogMaster) fireWorkerListeners(worker *Worker) {
	for _, v := range logMaster.loggers {
		v.workerListener(worker)
	}
}

func Shutdown() {
	for _, v := range logMaster.workers {
		for _, w := range v.Writers {
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
func (wrk *Worker) SetTimeFormat(format string) {
	if format == "" {
		wrk.timeFormater = nil
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
	wrk.timeFormater = func(t time.Time) string {
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

var logLevels = [...]string{"ALL", "TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL", "NONE"}

var logLevelColors = [...]func(a ...interface{}) string{
	nil,
	nil, // TRACE
	color.New(color.FgMagenta).SprintFunc(),  // DEBUG
	color.New(color.FgCyan).SprintFunc(),     // INFO
	color.New(color.FgHiYellow).SprintFunc(), // WARN
	color.New(color.FgHiRed).SprintFunc(),    // ERROR
	color.New(color.FgHiRed).SprintFunc(),    // FATAL
	nil,
}

func (this LogLevel) String() string {
	var level = int(this)
	if level >= 0 && level <= len(logLevels) {
		return logLevels[level]
	} else {
		return ""
	}
}

const (
	ALL LogLevel = iota
	TRACE
	DEBUG
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

type ILogger interface {
	IsActive(LogLevel) bool
	Tracef(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
}

type Logger struct {
	sync.Mutex

	tag       string
	worker    *Worker
	calldepth int
}

var _ ILogger = &Logger{}

// workerListener is called when a worker is Registered.
// this way we keep all worker loggers updated when they are later redefined
func (this *Logger) workerListener(worker *Worker) {
	this.Lock()
	if strings.HasPrefix(this.tag, worker.Prefix) && worker.Prefix > this.worker.Prefix {
		this.worker = worker
	}
	this.Unlock()
}

func (this *Logger) loadWorker() {
	this.Lock()
	defer this.Unlock()

	if this.worker == nil {
		this.worker = logMaster.fetchWorker(this.tag)
	}
}

func (this *Logger) getWorker() *Worker {
	if this.worker == nil {
		this.loadWorker()
	}
	return this.worker
}

func (this *Logger) Level() LogLevel {
	return this.getWorker().Level
}

func (this *Logger) Namespace() string {
	return this.tag
}

func (this *Logger) SetCallerAt(depth int) *Logger {
	this.calldepth = depth
	return this
}

func (this *Logger) CallerAt(depth int) *Logger {
	// creates a temporary logger
	tmp := LoggerFor(this.tag)
	tmp.calldepth = depth
	return tmp
}

func (this *Logger) logStamp(level LogLevel) string {
	t := time.Now()
	var wrk = logMaster.fetchWorker(this.tag)

	var result bytes.Buffer
	if wrk.timeFormater != nil {
		result.WriteString(wrk.timeFormater(t))
	}

	if wrk.showLevel {
		if result.Len() > 0 {
			result.WriteString(" ")
		}
		// left padding level
		var s = level.String()
		/*
			s = strings.Repeat(" ", 5-len(s)) + s
			result.WriteString(s)
		*/
		var colorFunc = logLevelColors[level]
		if colorFunc != nil {
			s = colorFunc(s)
		}
		result.WriteString(fmt.Sprintf("%6s", s))
	}

	if wrk.showCaller {
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
		flush(level, str, this.getWorker().Writers)
	}
}

func (this *Logger) log(level LogLevel, a ...interface{}) {
	if this.IsActive(level) {
		var arr = []interface{}{this.logStamp(level)}
		arr = append(arr, a...)
		arr = append(arr, "\n")
		flush(level, fmt.Sprint(arr...), this.getWorker().Writers)
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

func (this *Logger) Tracef(format string, what ...interface{}) {
	this.logf(TRACE, format, what...)
}

func (this *Logger) Debugf(format string, what ...interface{}) {
	this.logf(DEBUG, format, what...)
}

func (this *Logger) Infof(format string, what ...interface{}) {
	this.logf(INFO, format, what...)
}

func (this *Logger) Warnf(format string, what ...interface{}) {
	this.logf(WARN, format, what...)
}

func (this *Logger) Errorf(format string, what ...interface{}) {
	this.logf(ERROR, format, what...)
}

func (this *Logger) Fatalf(format string, what ...interface{}) {
	this.logf(FATAL, format, what...)
}

func (this *Logger) Trace(a ...interface{}) {
	this.log(DEBUG, a...)
}

func (this *Logger) Debug(a ...interface{}) {
	this.log(DEBUG, a...)
}

func (this *Logger) Info(a ...interface{}) {
	this.log(INFO, a...)
}

func (this *Logger) Warn(a ...interface{}) {
	this.log(WARN, a...)
}

func (this *Logger) Error(a ...interface{}) {
	this.log(ERROR, a...)
}

func (this *Logger) Fatal(a ...interface{}) {
	this.log(FATAL, a...)
}

type Wrap struct {
	Logger ILogger
	Tag    string
}

var _ ILogger = Wrap{}

func (this Wrap) IsActive(level LogLevel) bool {
	return this.Logger.IsActive(level)
}

func (this Wrap) Tracef(format string, what ...interface{}) {
	if this.IsActive(TRACE) {
		this.Logger.Tracef(this.Tag+format, what...)
	}
}

func (this Wrap) Debugf(format string, what ...interface{}) {
	if this.IsActive(DEBUG) {
		this.Logger.Debugf(this.Tag+format, what...)
	}
}

func (this Wrap) Infof(format string, what ...interface{}) {
	if this.IsActive(INFO) {
		this.Logger.Infof(this.Tag+format, what...)
	}
}

func (this Wrap) Warnf(format string, what ...interface{}) {
	if this.IsActive(WARN) {
		this.Logger.Warnf(this.Tag+format, what...)
	}
}

func (this Wrap) Errorf(format string, what ...interface{}) {
	if this.IsActive(ERROR) {
		this.Logger.Errorf(this.Tag+format, what...)
	}
}

func (this Wrap) Fatalf(format string, what ...interface{}) {
	if this.IsActive(FATAL) {
		this.Logger.Fatalf(this.Tag+format, what...)
	}
}
