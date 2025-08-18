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

var threadTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_free_thread((*C.struct_LuaThread)(unsafe.Pointer(ptr)))
	},
}

// A LuaThread is an abstraction over a Lua thread object.
type LuaThread struct {
	lua    *Lua // The Lua VM wrapper that owns this thread
	object *object
}

func (l *LuaThread) innerPtr() (*C.struct_LuaThread, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot access thread on closed Lua VM")
	}

	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_LuaThread)(unsafe.Pointer(ptr)), nil
}

type ThreadStatus int

const (
	ThreadStatusResumable ThreadStatus = iota
	ThreadStatusRunning
	ThreadStatusFinished
	ThreadStatusError
)

func (ts ThreadStatus) String() string {
	switch ts {
	case ThreadStatusResumable:
		return "resumable"
	case ThreadStatusRunning:
		return "running"
	case ThreadStatusFinished:
		return "finished"
	case ThreadStatusError:
		return "error"
	default:
		return "unknown"
	}
}

// Returns the current status of the LuaThread or ThreadStatusFinished if the thread has been closed
//
// Locking behavior: This function acquires a read lock on the LuaThread object.
func (l *LuaThread) Status() ThreadStatus {
	if l.lua.object.IsClosed() {
		return ThreadStatusFinished // Return finished if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return ThreadStatusFinished // Return finished if the object is closed
	}

	status := C.luago_thread_status(ptr)
	switch uint8(status) {
	case 0:
		return ThreadStatusResumable // Resumable
	case 1:
		return ThreadStatusRunning // Running
	case 2:
		return ThreadStatusFinished // Finished
	case 3:
		return ThreadStatusError // Error
	default:
		return ThreadStatusFinished // Default to finished for unknown statuses
	}
}

// Resume resumes a thread `th`
//
// Passes args as arguments to the thread. If the coroutine has called coroutine.yield, it will return these arguments. Otherwise, the coroutine wasn’t yet started, so the arguments are passed to its main function.
//
// If the thread is no longer resumable (meaning it has finished execution or encountered an error), this will return a coroutine unresumable error, otherwise will return as follows:
// If the thread is yielded via coroutine.yield or CallbackLua.YieldWith, returns the values passed to yield. If the thread returns values from its main function, returns those.
func (l *LuaThread) Resume(args ...Value) ([]Value, error) {
	if l.lua.object.IsClosed() {
		return nil, fmt.Errorf("cannot resume thread on closed Lua VM")
	}

	l.object.RLock()
	defer l.object.RUnlock()

	ptr, err := l.innerPtr()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	mw, err := l.lua.multiValueFromValues(args)
	if err != nil {
		return nil, err // Return error if the value cannot be converted (diff lua state, closed object, etc)
	}

	res := C.luago_thread_resume(ptr, mw.ptr)
	if res.error != nil {
		return nil, moveErrorToGoError(res.error)
	}
	rets := &luaMultiValue{ptr: res.value, lua: l.lua}
	retsMw := rets.take()
	rets.close()
	return retsMw, nil
}

// Returns a 'pointer' to a Lua-owned thread
//
// This pointer is only useful for hashing/debugging
// and cannot be converted back to the original Lua thread object.
func (l *LuaThread) Pointer() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return 0 if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()
	lptr, err := l.object.PointerNoLock()
	if err != nil {
		return 0 // Return 0 if the object is closed
	}

	ptr := C.luago_thread_to_pointer((*C.struct_LuaThread)(unsafe.Pointer(lptr)))
	return uint64(ptr)
}

// Equals checks if the LuaThread equals another LuaThread by lua value reference
func (l *LuaThread) Equals(other *LuaThread) bool {
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

	return bool(C.luago_thread_equals(ptr, ptr2))
}

// String returns a string representation of the LuaThread.
func (l *LuaThread) String() string {
	status := l.Status()
	return "LuaThread(status: " + status.String() + ", pointer: " + fmt.Sprintf("%#x", l.Pointer()) + ")"
}

// ToValue converts the LuaThread to a Value.
func (l *LuaThread) ToValue() Value {
	return &ValueThread{value: l}
}

func (l *LuaThread) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaThread object
	return l.object.Close()
}
