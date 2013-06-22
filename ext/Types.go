package ext

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	tk "github.com/quintans/toolkit"
	"strconv"
	"time"
)

const json_null = "null"

// == Any ==

type Any struct {
	Val interface{}
}

func ANY(val interface{}) Any {
	return Any{val}
}

// Scan implements the Scanner interface.
func (this *Any) Scan(value interface{}) error {
	this.Val = value
	return nil
}

// Value implements the driver Valuer interface.
func (this Any) Value() (driver.Value, error) {
	if this.Val == nil {
		return nil, nil
	}
	return this.Val, nil
}

func (this Any) IsNil() bool {
	return this.Val == nil
}

// LAZY STRING
type LazyString struct {
	Str func() string
}

func (this LazyString) String() string {
	return this.String()
}

// STRING
type Str string

var _ sql.Scanner = new(Str)
var _ driver.Valuer = new(Str)

func NewStr(val Str) *Str {
	return &val
}

func (this Str) HashCode() int {
	return tk.HashString(tk.HASH_SEED, string(this))
}

func (this Str) Equals(other interface{}) bool {
	s, _ := other.(Str)
	return this == s
}

func (this *Str) String() string {
	if this == nil {
		return "<nil>"
	} else {
		return string(*this)
	}
}

func (this *Str) Scan(value interface{}) error {
	var s sql.NullString
	err := s.Scan(value)
	if !s.Valid || err != nil {
		return err
	}

	*this = Str(s.String)
	return nil
}

// Value implements the driver Valuer interface.
func (this Str) Value() (driver.Value, error) {
	return string(this), nil
}

// LONG

type Int64 int64

func NewInt64(val Int64) *Int64 {
	return &val
}

func (this Int64) HashCode() int {
	return tk.HashLong(tk.HASH_SEED, int64(this))
}

func (this Int64) Equals(other interface{}) bool {
	s, _ := other.(Int64)
	return this == s
}

// Float

type Float64 float64

func NewFloat64(val Float64) *Float64 {
	return &val
}

func (this Float64) HashCode() int {
	return tk.HashDouble(tk.HASH_SEED, float64(this))
}

func (this Float64) Equals(other interface{}) bool {
	s, _ := other.(Float64)
	return this == s
}

// BOOL

type Bool bool

func NewBool(val Bool) *Bool {
	return &val
}

func (this Bool) HashCode() int {
	return tk.HashBool(tk.HASH_SEED, bool(this))
}

func (this Bool) Equals(other interface{}) bool {
	s, _ := other.(Bool)
	return this == s
}

// Date
// I Created this type because I wanted to control the generated/parsed JSON
type Date time.Time

var _ sql.Scanner = new(Date)
var _ driver.Valuer = new(Date)

func NOW() *Date {
	date := Date(time.Now())
	return &date
}

func NewDate(val Date) *Date {
	return &val
}

func (this Date) HashCode() int {
	bs, err := time.Time(this).GobEncode()
	if err != nil {
		return tk.HashBytes(tk.HASH_SEED, bs)
	}
	return 0
}

func (this Date) Equals(other interface{}) bool {
	var tick int64
	switch t := other.(type) { //type switch
	case Date:
		tick = time.Time(t).UnixNano()
	case time.Time:
		tick = t.UnixNano()
	case *Date:
		tick = time.Time(*t).UnixNano()
	case *time.Time:
		tick = t.UnixNano()
	}
	return time.Time(this).UnixNano() == tick
}

// Scan implements the Scanner interface.
func (this *Date) Scan(value interface{}) error {
	switch t := value.(type) {
	case time.Time:
		*this = Date(t)
		return nil
	case *time.Time:
		*this = Date(*t)
		return nil
	}

	return errors.New(fmt.Sprintf("[pqp] Value (%s) not a time.Time", value))
}

// Value implements the driver Valuer interface.
func (this Date) Value() (driver.Value, error) {
	return (time.Time)(this), nil
}

func (this Date) MarshalJSON() ([]byte, error) {
	v := time.Time(this).UnixNano() / int64(time.Millisecond)
	return []byte(strconv.FormatInt(v, 10)), nil
}

func (this *Date) UnmarshalJSON(b []byte) error {
	s := string(b)
	if s != json_null {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*this = Date(time.Unix(0, v*int64(time.Millisecond)))
	}
	return nil
}

func (this Date) String() string {
	t := time.Time(this)
	return fmt.Sprintf("%d/%02d/%02d-%02d:%02d:%02d",
		t.Year(),
		t.Month(),
		t.Day(),
		t.Hour(),
		t.Minute(),
		t.Second())
}
