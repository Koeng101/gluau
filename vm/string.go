package vm

import (
	"unsafe"
)

/*
#include "../rustlib/rustlib.h"
*/
import "C"

var stringTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_free_string((*C.struct_LuaString)(unsafe.Pointer(ptr)))
	},
}

// A LuaString is an abstraction over a Lua string object.
type LuaString struct {
	lua    *Lua // The Lua VM wrapper that owns this string
	object *object
}

func (l *LuaString) innerPtr() (*C.struct_LuaString, error) {
	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_LuaString)(unsafe.Pointer(ptr)), nil
}

// Returns the LuaString as a byte slice
func (l *LuaString) Bytes() []byte {
	if l.lua.object.IsClosed() {
		return nil // Return nil if the Lua VM is closed
	}
	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return nil // Return nil if the object is closed
	}
	data := C.luago_string_as_bytes(ptr)

	return moveBytesToGo(data)
}

// Returns the LuaString as a byte slice with nul terminator
func (l *LuaString) BytesWithNUL() []byte {
	if l.lua.object.IsClosed() {
		return nil // Return nil if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return nil // Return nil if the object is closed
	}
	data := C.luago_string_as_bytes_with_nul(ptr)

	return moveBytesToGo(data)
}

// Returns a 'pointer' to a Lua-owned string
//
// This pointer is only useful for hashing/debugging
// and cannot be converted back to the original Lua string object.
func (l *LuaString) Pointer() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return nil if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()
	lptr, err := l.innerPtr()
	if err != nil {
		return 0 // Return 0 if the object is closed
	}

	ptr := C.luago_string_to_pointer(lptr)
	return uint64(ptr)
}

// Equals checks if the LuaString equals another LuaString
// in terms of string content.
func (l *LuaString) Equals(other *LuaString) bool {
	if l.lua.object.IsClosed() {
		return false // Return false if the Lua VM is closed
	}

	if other == nil || l.lua != other.lua {
		return false // Return false if the Lua instances are different
	}

	l.object.RLock()
	defer l.object.RUnlock()
	other.object.RLock()
	defer other.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return false // Return error if the object is closed
	}
	ptr2, err := other.innerPtr()
	if err != nil {
		return false // Return error if the other object is closed
	}

	return bool(C.luago_string_equals(ptr, ptr2))
}

// String returns the LuaString as a Go string.
func (l *LuaString) String() string {
	// Convert the LuaString to a Go string
	return string(l.Bytes())
}

// ToValue converts the LuaString to a Value.
func (l *LuaString) ToValue() Value {
	return &ValueString{value: l}
}

func (l *LuaString) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaString object
	return l.object.Close()
}
