package gocache

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"time"
)

var (
	ErrInterfaceAssertion error = errors.New("interface type assertion failed")
)

const (
	Greater byte = iota + 1
	Less
	GreaterEq
	LessEq
	Eq
	Neq
	Regex
)

type Matcher func(v Object) bool

func NewTimeQuery(v time.Time, c byte, rex string) TimeQuery {
	return TimeQuery{
		V:     v,
		C:     c,
		S:     rex,
		Check: true,
	}
}

type TimeQuery struct {
	V     time.Time
	C     byte
	S     string
	Check bool `msgpack:"chk"`
}

func (q TimeQuery) Match(i interface{}) bool {
	v, ok := i.(time.Time)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v.UnixNano() > q.V.UnixNano()
	case GreaterEq:
		return v.UnixNano() >= q.V.UnixNano()
	case LessEq:
		return v.UnixNano() <= q.V.UnixNano()
	case Regex:
		if len(q.S) == 0 {
			return q.V == v
		}
		rex, err := regexp.Compile(q.S)
		if err != nil {
			return false
		}
		return len(rex.FindString(v.Format("2006-01-02 15:04:05.999"))) > 0
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewIntQuery(v int, c byte, rex string) IntQuery {
	return IntQuery{
		V:     v,
		C:     c,
		S:     rex,
		Check: true,
	}
}

type IntQuery struct {
	V     int
	C     byte
	S     string
	Check bool `msgpack:"chk"`
}

func (q IntQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}

	switch val := i.(type) {
	case int:
		return q.cmp(val)
	case []int:
		for _, v := range val {
			if q.cmp(v) {
				return true
			}
		}
		return false
	default:
		return false
	}

}

func (q IntQuery) cmp(v int) bool {
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Regex:
		if len(q.S) == 0 {
			return q.V == v
		}
		rex, err := regexp.Compile(q.S)
		if err != nil {
			return false
		}
		return len(rex.FindString(fmt.Sprintf("%d", v))) > 0
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewFloatQuery(v float64, c byte, rex string) FloatQuery {
	return FloatQuery{
		V:     v,
		C:     c,
		S:     rex,
		Check: true,
	}
}

type FloatQuery struct {
	V     float64
	C     byte
	S     string
	Check bool `msgpack:"chk"`
}

func (q FloatQuery) Match(i interface{}) bool {
	v, ok := i.(float64)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Regex:
		if len(q.S) == 0 {
			return q.V == v
		}
		rex, err := regexp.Compile(q.S)
		if err != nil {
			return false
		}
		return len(rex.FindString(fmt.Sprintf("%.4f", v))) > 0
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewStringQuery(v string, c byte, rex string) StringQuery {
	return StringQuery{
		V:     v,
		C:     c,
		S:     rex,
		Check: true,
	}
}

type StringQuery struct {
	V     string
	C     byte
	S     string
	Check bool `msgpack:"chk"`
}

func (q StringQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}
	switch v := i.(type) {
	case string:
		return q.cmp(v)
	case []string:
		for _, val := range v {
			if q.cmp(val) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (q StringQuery) cmp(v string) bool {
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Regex:
		if len(q.S) == 0 {
			return q.V == v
		}
		rex, err := regexp.Compile(q.S)
		if err != nil {
			return false
		}
		return len(rex.FindString(v)) > 0
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewByteQuery(v byte, c byte) ByteQuery {
	return ByteQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type ByteQuery struct {
	V     byte
	C     byte
	Check bool `msgpack:"chk"`
}

func (q ByteQuery) Match(i interface{}) bool {
	v, ok := i.(byte)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v > q.V
	case GreaterEq:
		return v >= q.V
	case LessEq:
		return v <= q.V
	case Regex:
		return q.V == v
	case Neq:
		return q.V != v
	default:
		return q.V == v
	}
}

func NewByteSliceQuery(v []byte, c byte, rex string) ByteSliceQuery {
	return ByteSliceQuery{
		V:     v,
		C:     c,
		S:     rex,
		Check: true,
	}
}

type ByteSliceQuery struct {
	V     []byte
	C     byte
	S     string
	Check bool `msgpack:"chk"`
}

func (q ByteSliceQuery) Match(i interface{}) bool {
	if !q.Check {
		return true
	}
	switch val := i.(type) {
	case []byte:
		return q.cmp(val)
	case [][]byte:
		for _, v := range val {
			if q.cmp(v) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func (q ByteSliceQuery) cmp(v []byte) bool {
	switch q.C {
	case Greater:
		return string(v) > string(q.V)
	case GreaterEq:
		return string(v) >= string(q.V)
	case LessEq:
		return string(v) <= string(q.V)
	case Regex:
		if len(q.S) == 0 {
			return bytes.Equal(q.V, v)
		}
		rex, err := regexp.Compile(q.S)
		if err != nil {
			return false
		}
		return len(rex.Find(v)) > 0
	case Neq:
		return !bytes.Equal(q.V, v)
	default:
		return bytes.Equal(q.V, v)
	}
}

func NewBoolQuery(v bool, c byte) BoolQuery {
	return BoolQuery{
		V:     v,
		C:     c,
		Check: true,
	}
}

type BoolQuery struct {
	V     bool
	C     byte
	Check bool `msgpack:"chk"`
}

func (q BoolQuery) Match(i interface{}) bool {
	v, ok := i.(bool)
	if !ok {
		return false
	}
	if !q.Check {
		return true
	}
	switch q.C {
	case Greater:
		return v && !q.V
	case GreaterEq:
		return v
	case LessEq:
		return !v
	case Regex:
		return v
	case Neq:
		return v != q.V
	case Eq:
		return v == q.V
	default:
		return v
	}
}

func And(v ...bool) bool {
	for _, val := range v {
		if !val {
			return false
		}
	}
	return true
}

func Or(v ...bool) bool {
	for _, val := range v {
		if val {
			return true
		}
	}
	return false
}
