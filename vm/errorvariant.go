package vm

import (
	"unsafe"
)

/*
#include "../rustlib/rustlib.h"
*/
import "C"

var errorVariantTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_error_free((*C.struct_ErrorVariant)(unsafe.Pointer(ptr)))
	},
}

// A ErrorVariant is an wrapper around a Rust Arc<String> that holds an error string.
type ErrorVariant struct {
	*object
}

// Returns the ErrorVariant as a byte slice
func (l *ErrorVariant) Bytes() []byte {
	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil // Return nil if the object is closed
	}

	data := C.luago_error_get_string((*C.struct_ErrorVariant)(unsafe.Pointer(ptr)))
	goSlice := C.GoBytes(unsafe.Pointer(data.data), C.int(data.len))
	return goSlice
}

// Returns the ErrorVariant as a string
func (l *ErrorVariant) String() string {
	bytes := l.Bytes()
	if bytes == nil {
		return "" // Return empty string if the object is closed
	}
	return string(bytes)
}

// Equals checks if the ErrorVariant equals another ErrorVariant
//
// Equivalent to l.String() == other.String().
func (l *ErrorVariant) Equals(other *ErrorVariant) bool {
	if l == nil && other == nil {
		return true // Both are nil
	}
	if l == nil || other == nil {
		return false // One is nil, the other is not
	}
	return l.String() == other.String() // Compare string representations
}

func (l *ErrorVariant) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaTable object
	return l.object.Close()
}
