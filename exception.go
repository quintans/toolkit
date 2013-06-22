package toolkit

type Fault interface {
	error
	GetCode() string
	GetMessage() string
}

type Fail struct {
	Code    string
	Message string
}

func (this *Fail) GetCode() string {
	return this.Code
}

func (this *Fail) GetMessage() string {
	return this.Message
}

func (this *Fail) Error() string {
	sb := NewStrBuffer()
	if this.Code != "" {
		sb.Add("[", this.Code, "] ")
	}
	sb.Add(this.Message)
	return sb.String()
}
