package vm

/*
#include "../rustlib/rustlib.h"
*/
import "C"
import (
	"errors"
	"strings"
	"unsafe"
)

func moveStringToRust(s string) *C.char {
	if len(s) == 0 {
		return C.luago_string_new(nil, 0) // Return empty char array if the string is empty
	}
	s = strings.ReplaceAll(s, "\x00", "") // Remove any null characters
	return moveBytesToRust([]byte(s))
}

func moveBytesToRust(b []byte) *C.char {
	if b == nil {
		return C.luago_string_new(nil, 0) // Return empty char array if the byte slice is nil
	}
	return C.luago_string_new((*C.char)(unsafe.Pointer(&b[0])), C.size_t(len(b)))
}

func freeRustString(s *C.char) {
	if s == nil {
		return // Nothing to free
	}
	C.luago_string_free(s) // Free the Rust string
}

func moveStringToGo(err *C.char) string {
	if err == nil {
		return ""
	}
	errStr := C.GoString(err)
	C.luago_string_free(err) // Free the error string
	return errStr
}

func moveErrorToGo(err *C.char) error {
	if err == nil {
		return nil
	}
	errStr := C.GoString(err)
	C.luago_string_free(err) // Free the error string
	return errors.New(errStr)
}

func moveBytesToGo(data C.struct_LuaStringBytes) []byte {
	if data.data == nil {
		return nil
	}

	goSlice := C.GoBytes(unsafe.Pointer(data.data), C.int(data.len))
	return goSlice
}
