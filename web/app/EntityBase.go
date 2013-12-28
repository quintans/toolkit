package app

import (
	. "github.com/quintans/toolkit/ext"
)

type IEntity interface {
	GetId() *int64
	SetId(*int64)
	GetVersion() *int64
	SetVersion(*int64)
	GetCreation() *Date
	SetCreation(*Date)
	GetModification() *Date
	SetModification(*Date)
}

type EntityBase struct {
	Id           *int64 `json:"id"`
	Version      *int64 `json:"version"`
	Creation     *Date  `json:"creation"`
	Modification *Date  `json:"modification"`
}

func (this *EntityBase) GetId() *int64 {
	return this.Id
}

func (this *EntityBase) SetId(id *int64) {
	this.Id = id
}

func (this *EntityBase) GetVersion() *int64 {
	return this.Version
}

func (this *EntityBase) SetVersion(version *int64) {
	this.Version = version
}

func (this *EntityBase) GetCreation() *Date {
	return this.Creation
}

func (this *EntityBase) SetCreation(creation *Date) {
	this.Creation = creation
}

func (this *EntityBase) GetModification() *Date {
	return this.Modification
}

func (this *EntityBase) SetModification(modification *Date) {
	this.Modification = modification
}

func (this *EntityBase) Copy(entity EntityBase) {
	this.Id = CloneInt64(entity.Id)
	this.Version = CloneInt64(entity.Version)
	this.Creation = CloneDate(entity.Creation)
	this.Modification = CloneDate(entity.Modification)
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
