package vm

/*
#include "../rustlib/rustlib.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

var userdataTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_free_userdata((*C.struct_LuaUserData)(unsafe.Pointer(ptr)))
	},
}

// A LuaUserData is an abstraction over a Lua userdata object.
type LuaUserData struct {
	lua    *Lua // The Lua VM wrapper that owns this userdata
	object *object
}

func (l *LuaUserData) innerPtr() (*C.struct_LuaUserData, error) {
	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_LuaUserData)(unsafe.Pointer(ptr)), nil
}

// Returns the associated data within the LuaUserData.
//
// Errors if there is no associated data or if the userdata is closed.
func (l *LuaUserData) AssociatedData() (any, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot access userdata on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}

	res := C.luago_get_userdata_handle(ptr)
	if res.error != nil {
		err := moveErrorToGo(res.error)
		return nil, err
	}

	value := uintptr(res.value)
	if value == 0 {
		return nil, nil // No associated data
	}
	data := getDynamicData(value)
	if data == nil {
		return nil, errors.New("internal error: handle is invalid")
	}
	return data, nil
}

// Returns a 'pointer' to a Lua-owned userdata
//
// This pointer is only useful for hashing/debugging
// and cannot be converted back to the original Lua userdata object.
func (l *LuaUserData) Pointer() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return 0 if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()
	lptr, err := l.object.PointerNoLock()
	if err != nil {
		return 0 // Return 0 if the object is closed
	}

	ptr := C.luago_userdata_to_pointer((*C.struct_LuaUserData)(unsafe.Pointer(lptr)))
	return uint64(ptr)
}

// Equals checks if the LuaUserData equals another LuaUserData by lua value reference
func (l *LuaUserData) Equals(other *LuaUserData) bool {
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

	return bool(C.luago_userdata_equals(ptr, ptr2))
}

// Metatable returns the metatable of the LuaUserData.
func (l *LuaUserData) Metatable() (*LuaTable, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot access metatable of userdata on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}

	res := C.luago_userdata_metatable(ptr)
	if res.error != nil {
		err := moveErrorToGo(res.error)
		return nil, err
	}

	return &LuaTable{object: newObject((*C.void)(unsafe.Pointer(res.value)), tableTab), lua: l.lua}, nil
}

// ToValue converts the LuaUserData to a Value.
func (l *LuaUserData) ToValue() Value {
	return &ValueUserData{value: l}
}

// String returns a string representation of the LuaUserData.
func (l *LuaUserData) String() string {
	if l == nil || l.object == nil {
		return "<nil LuaUserData>"
	}
	return "LuaUserData(pointer: " + fmt.Sprintf("%#x", l.Pointer()) + ")"
}

func (l *LuaUserData) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaUserData object
	return l.object.Close()
}
