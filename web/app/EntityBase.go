package app

import (
	"github.com/quintans/goSQL/db"
	. "github.com/quintans/toolkit/ext"
)

type IEntity interface {
	GetId() *int64
	SetId(*int64)
	GetVersion() *int64
	SetVersion(*int64)
}

type EntityBase struct {
	db.Marker

	Id      *int64 `json:"id"`
	Version *int64 `json:"version"`
}

func (this *EntityBase) GetId() *int64 {
	return this.Id
}

func (this *EntityBase) SetId(id *int64) {
	this.Id = id
	this.Mark("Id")
}

func (this *EntityBase) GetVersion() *int64 {
	return this.Version
}

func (this *EntityBase) SetVersion(version *int64) {
	this.Version = version
	this.Mark("Version")
}

func (this *EntityBase) Copy(entity EntityBase) {
	this.Id = CloneInt64(entity.Id)
	this.Version = CloneInt64(entity.Version)
}

func CopyCurrency(value *float64) *float64 {
	return CloneFloat64(value)
}

func CopyString(value *string) *string {
	return CloneStr(value)
}

func CopyInteger(value *int64) *int64 {
	return CloneInt64(value)
}

func CopyDate(value *Date) *Date {
	return CloneDate(value)
}

func CopyBin(value []byte) []byte {
	v := make([]byte, len(value), len(value))
	copy(v, value)
	return v
}
