package log

// check if it implements LogWriter interface
var _ LogWriter = &NullAppender{}

func NewNullAppender() *NullAppender {
	this := new(NullAppender)
	return this
}

type NullAppender struct{}

func (this *NullAppender) Log(msgLevel LogLevel, msg string) {
}

func (this *NullAppender) Discard() {
}
