package gotools

import (
	"reflect"
	"unsafe"
)

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func StringToBytes(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

const maxStartEndStringLen = 80

func StartEndString(s string, startLength int) string {
	if len(s) <= startLength || len(s) <= maxStartEndStringLen {
		return s
	}

	return s[:startLength] + "..." + s[len(s)-startLength:]
}
