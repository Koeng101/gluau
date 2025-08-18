//go:build cgo

package vm

/*
#include "../rustlib/rustlib.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var functionTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_free_function((*C.struct_LuaFunction)(unsafe.Pointer(ptr)))
	},
}

// A LuaFunction is an wrapper around a function
//
// API's to be implemented as of now: coverage, info (more complex to implement),
type LuaFunction struct {
	object *object
	lua    *Lua
}

func (l *LuaFunction) innerPtr() (*C.struct_LuaFunction, error) {
	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_LuaFunction)(unsafe.Pointer(ptr)), nil
}

// Call calls a function `f` returning either the returned arguments
// or the error
//
// Locking behavior: This function acquires a read lock on the LuaFunction object
// and a write lock on all arguments passed to the function.
func (l *LuaFunction) Call(args ...Value) ([]Value, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot call function on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	mw, err := l.lua.multiValueFromValues(args)
	if err != nil {
		return nil, err // Return error if the value cannot be converted
	}

	res := C.luago_function_call(ptr, mw.ptr)
	if res.error != nil {
		return nil, moveErrorToGoError(res.error)
	}
	rets := &luaMultiValue{ptr: res.value, lua: l.lua}
	retsMw := rets.take()
	rets.close()
	return retsMw, nil
}

// Returns a deep clone to a Lua-owned function
//
// If called on a Luau function, this method copies the function prototype and all its upvalues to the
// newly created function
//
// If called on a Go function, this method merely clones the function's handle
func (l *LuaFunction) DeepClone() (*LuaFunction, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot deep clone function on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}

	res := C.luago_function_deepclone(ptr)
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}

	return &LuaFunction{object: newObject((*C.void)(unsafe.Pointer(res.value)), functionTab), lua: l.lua}, nil
}

// Returns the environment table of the LuaFunction.
//
// If the function has no environment, it returns nil and Go functions will never have
// an environment table either.
func (l *LuaFunction) Environment() (*LuaTable, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot get environment of function on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}

	tab := C.luago_function_environment(ptr)
	if tab == nil {
		return nil, nil // No environment table
	}

	return &LuaTable{object: newObject((*C.void)(unsafe.Pointer(tab)), tableTab), lua: l.lua}, nil
}

// Sets the environment table of the LuaFunction returning true if the environment was set
func (l *LuaFunction) SetEnvironment(env *LuaTable) (bool, error) {
	if l.lua.object.IsClosed() {
		return false, fmt.Errorf("cannot set environment of function on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return false, err // Return error if the object is closed
	}

	if env == nil {
		return false, nil // No environment to set
	}
	if env.lua != l.lua {
		return false, fmt.Errorf("cannot set environment table from different Lua instance")
	}
	env.object.RLock()
	defer env.object.RUnlock()
	envPtr, err := env.object.PointerNoLock()
	if err != nil {
		return false, err // Return error if the environment table is closed
	}

	res := C.luago_function_set_environment(ptr, (*C.struct_LuaTable)(unsafe.Pointer(envPtr)))
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return false, err
	}

	return bool(res.value), nil
}

// Returns a 'pointer' to a Lua-owned function
//
// This pointer is only useful for hashing/debugging
// and cannot be converted back to the original Lua function object.
func (l *LuaFunction) Pointer() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return 0 if the Lua VM is closed
	}
	l.object.RLock()
	defer l.object.RUnlock()
	lptr, err := l.innerPtr()
	if err != nil {
		return 0 // Return error if the object is closed
	}

	ptr := C.luago_function_to_pointer(lptr)
	return uint64(ptr)
}

// Equals checks if the LuaFunction equals another LuaFunction by lua value reference
func (l *LuaFunction) Equals(other *LuaFunction) bool {
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

	return bool(C.luago_function_equals(ptr, ptr2))
}

// ToValue converts the LuaFunction to a Value.
func (l *LuaFunction) ToValue() Value {
	return &ValueFunction{value: l}
}

// String returns a string representation of the LuaFunction.
//
// This is currently just the pointer address of the function.
func (l *LuaFunction) String() string {
	ptr := l.Pointer()
	if ptr == 0 {
		return "<closed LuaFunction>"
	}
	return fmt.Sprintf("LuaFunction 0x%x", ptr)
}

func (l *LuaFunction) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaFunction object
	return l.object.Close()
}
