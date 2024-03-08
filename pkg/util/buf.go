package util

import (
	"bytes"
	"sync"
	"unsafe"
)

var (
	bufpool = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

func GetBuf() *bytes.Buffer {
	buf := bufpool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func PutBuf(b *bytes.Buffer) {
	bufpool.Put(b)
}

func Str2bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

func Bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
