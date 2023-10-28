package json

import (
	"bytes"
	"sync"
)

var (
	JSON_START_TOKEN     = []byte("{")
	JSON_SEPARATOR_TOKEN = []byte(",")
	JSON_END_TOKEN       = []byte("}")
)

var bufferPool sync.Pool

func GetBuffer() *bytes.Buffer {
	if v := bufferPool.Get(); v != nil {
		e := v.(*bytes.Buffer)
		e.Reset()
		return e
	}
	return &bytes.Buffer{}
}

func PutBuffer(buf *bytes.Buffer) {
	bufferPool.Put(buf)
}
