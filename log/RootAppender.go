package log

import (
	"io"
)

type RootAppender struct {
	io.Writer
	Channel chan string
}

func AsyncWriter(ch chan string, writer io.Writer) {
	for msg := range ch {
		writer.Write([]byte(msg))
	}
}

func (this *RootAppender) Log(msgLevel LogLevel, msg string) {
	if this.Channel != nil {
		if msgLevel == FATAL {
			this.DrainChannel()
		} else {
			this.Channel <- msg
			return
		}
	}

	this.Write([]byte(msg))

}

func (this *RootAppender) Discard() {
	if this.Channel != nil {
		close(this.Channel)
	}
}

func (this *RootAppender) DrainChannel() {
	for {
		select {
		case msg := <-this.Channel:
			this.Write([]byte(msg))
		default:
			return
		}
	}
}
