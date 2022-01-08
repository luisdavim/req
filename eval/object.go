package eval

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

type Type uint

//go:generate stringer -type Type -linecomment
const (
	String   Type = iota + 1 // string
	Int                      // int
	Bool                     // bool
	Array                    // array
	Hash                     // hash
	File                     // file
	Request                  // request
	Response                 // response
	Stream                   // stream
	Name                     // name
	Key                      // key
	Nil                      // nil
)

type Selector interface {
	Select(obj Object) (Object, error)
}

type Object interface {
	Type() Type

	String() string
}

type nilObj struct {}

func (e nilObj) String() string { return "<nil>" }

func (e nilObj) Type() Type { return Nil }

type stringObj struct {
	value string
}

func (s stringObj) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.value)
}

func (s stringObj) String() string { return s.value }

func (s stringObj) Type() Type { return String }

type intObj struct {
	value int64
}

func (i intObj) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.value)
}

func (i intObj) String() string { return strconv.FormatInt(i.value, 10) }

func (i intObj) Type() Type { return Int }

type boolObj struct {
	value bool
}

func (b boolObj) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.value)
}

func (b boolObj) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b boolObj) Type() Type { return Bool }

type arrayObj struct {
	items []Object
}

func (a arrayObj) String() string {
	b, err := json.Marshal(a.items)

	if err != nil {
		return "[]"
	}
	return string(b)
}

func (a arrayObj) Type() Type { return Array }

type keyObj struct {
	name string
}

func (k keyObj) String() string { return k.name }

func (k keyObj) Type() Type { return Key }

type hashObj struct {
	pairs map[string]Object
}

func (h hashObj) String() string {
	b, err := json.Marshal(h.pairs)

	if err != nil {
		return "{}"
	}
	return string(b)
}

func (h hashObj) Type() Type { return Hash }

type fileObj struct {
	*os.File
}

func (f fileObj) String() string {
	return fmt.Sprintf("File<addr=%p, name=%q>", f.File, f.Name())
}

func (f fileObj) Type() Type { return File }

type streamObj struct {
	rs io.ReadSeeker
}

func (s streamObj) String() string {
	if s.rs == nil {
		return ""
	}

	var buf bytes.Buffer

	buf.ReadFrom(s.rs)

	return buf.String()
}

func (s streamObj) Type() Type { return Stream }

type reqObj struct {
	*http.Request
}

func (r reqObj) String() string {
	var buf bytes.Buffer

	buf.WriteString(r.Method + " " + r.Proto + "\n")

	r.Header.Write(&buf)

	if r.Body != nil {
		buf.WriteString("\n")

		rc, rc2 := copyrc(r.Body)

		r.Body = rc
		io.Copy(&buf, rc2)
	}
	return buf.String()
}

func (r reqObj) Type() Type { return Request }

func (r reqObj) Select(obj Object) (Object, error) {
	typ := obj.Type()

	if typ != Name {
		return nil, errors.New("cannot use type " + typ.String() + " as selector")
	}

	val := obj.(nameObj).value

	switch val {
	case "Method":
		return stringObj{value: r.Method}, nil
	case "URL":
		return stringObj(value: r.URL.String()}, nil
	case "Header":
		hash := hashObj{
			pairs: make(map[string]Object),
		}

		for k, v := range r.Header {
			hash.pairs[k] = stringObj{value: v[0]}
		}
		return hash, nil
	case "Body":
		if r.Body == nil {
			return streamObj{}, nil
		}

		rc, rc2 := copyrc(r.Body)
		r.Body = rc

		b, _ := io.ReadAll(rc2)

		return streamObj{rs: bytes.NewReader(b)}, nil
	default:
		return nil, errors.New("type " + r.Type().String() + " has no field " + val)
	}
}

type respObj struct {
	*http.Response
}

func copyrc(rc io.ReadCloser) (io.ReadCloser, io.ReadCloser) {
	var buf bytes.Buffer
	buf.ReadFrom(rc)

	return io.NopCloser(&buf), io.NopCloser(bytes.NewBuffer(buf.Bytes()))
}

func (r respObj) String() string {
	var buf bytes.Buffer

	buf.WriteString(r.Proto + " " + r.Status + "\n")

	rc, rc2 := copyrc(r.Body)

	r.Body = rc

	r.Header.Write(&buf)
	buf.WriteString("\n")
	io.Copy(&buf, rc2)

	return buf.String()
}

func (r respObj) Type() Type { return Response }

func (r respObj) Select(obj Object) (Object, error) {
	typ := obj.Type()

	if typ != Name {
		return nil, errors.New("cannot use type " + typ.String() + " as selector")
	}

	val := obj.(nameObj).value

	switch val {
	case "Status":
		return stringObj{value: r.Status}, nil
	case "StatusCode":
		return intObj{value: int64(r.StatusCode)}, nil
	case "Header":
		hash := hashObj{
			pairs: make(map[string]Object),
		}

		for k, v := range r.Header {
			hash.pairs[k] = stringObj{value: v[0]}
		}
		return hash, nil
	case "Body":
		if r.Body == nil {
			return streamObj{}, nil
		}

		rc, rc2 := copyrc(r.Body)
		r.Body = rc

		b, _ := io.ReadAll(rc2)

		return streamObj{rs: bytes.NewReader(b)}, nil
	default:
		return nil, errors.New("type " + r.Type().String() + " has no field " + val)
	}
}

type nameObj struct {
	value string
}

func (n nameObj) String() string {
	return n.value
}

func (n nameObj) Type() Type { return Name }
