//go:build cgo

package vm

/*
#include "../rustlib/rustlib.h"
*/
import "C"

// luaMultiValue is an abstraction over multiple Lua values
//
// These values are initially owned by the Rust/Lua layer with
// the luaMultiValue acting as a barrier between the Rust/Lua layer
// and the Go side.
//
// Internal API
type luaMultiValue struct {
	lua *Lua
	ptr *C.struct_GoMultiValue
}

// Creates a new empty LuaMultiValue object.
func (l *Lua) newMultiValueWithCapacity(cap uint64) *luaMultiValue {
	ptr := C.luago_create_multivalue_with_capacity(C.size_t(cap))
	if ptr == nil {
		return nil // Handle error if needed
	}
	mv := &luaMultiValue{
		ptr: ptr,
		lua: l,
	}
	return mv
}

// Pop a Lua value from the luaMultiValue.
//
// This pops the first value in the multivalue, not the last one.
func (l *luaMultiValue) pop() Value {
	cValue := C.luago_multivalue_pop(l.ptr)
	return l.lua.valueFromC(cValue)
}

// Returns the number of values in the luaMultiValue.
func (l *luaMultiValue) len() uint64 {
	return uint64(C.luago_multivalue_len(l.ptr))
}

// fromValues takes a []Value and makes a MultiValue
func (l *Lua) multiValueFromValues(values []Value) (*luaMultiValue, error) {
	created := []C.struct_GoLuaValue{}
	for _, value := range values {
		luaValue, err := l.valueToC(value)
		if err != nil {
			// Destroy the values created so far
			// to avoid a memory leak
			for _, createdValue := range created {
				destroyValue(createdValue)
			}

			return nil, err // Return error if the value cannot be converted
		}
		created = append(created, luaValue)
	}

	mw := l.newMultiValueWithCapacity(uint64(len(values)))

	for _, v := range created {
		C.luago_multivalue_push(mw.ptr, v)
	}

	return mw, nil
}

// takes a MultiValue and outputs a []Value
func (l *luaMultiValue) take() []Value {
	len := l.len()
	values := make([]Value, 0, len)
	var i uint64
	for i = 0; i < len; i++ {
		values = append(values, l.pop())
	}
	return values
}

func (l *luaMultiValue) close() {
	C.luago_free_multivalue(l.ptr)
	l.ptr = nil
}
