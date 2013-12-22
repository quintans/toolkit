package toolkit

type Fault interface {
	error
	GetCode() string
	GetMessage() string
}

type Fail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
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
