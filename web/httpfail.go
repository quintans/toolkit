package web

import "github.com/quintans/toolkit"

type HttpFail struct {
	*toolkit.Fail

	Status int
}

func NewHttpFail(status int, code string, message string) *HttpFail {
	this := new(HttpFail)
	this.Status = status
	this.Fail = new(toolkit.Fail)
	this.Code = code
	this.Fail.Message = message
	return this
}
