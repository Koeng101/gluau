package vm

/*
#include "../rustlib/rustlib.h"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

var bufferTab = objectTab{
	dtor: func(ptr *C.void) {
		C.luago_free_buffer((*C.struct_LuaBuffer)(unsafe.Pointer(ptr)))
	},
}

// A LuaBuffer is an abstraction over a Lua buffer object.
type LuaBuffer struct {
	lua    *Lua // The Lua VM wrapper that owns this buffer
	object *object
}

func (l *LuaBuffer) innerPtr() (*C.struct_LuaBuffer, error) {
	ptr, err := l.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_LuaBuffer)(unsafe.Pointer(ptr)), nil
}

// Returns a 'pointer' to a Lua-owned buffer
//
// This pointer is only useful for hashing/debugging
// and cannot be converted back to the original Lua buffer object.
func (l *LuaBuffer) Pointer() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return nil if the Lua VM is closed
	}

	l.object.RLock()
	defer l.object.RUnlock()
	lptr, err := l.innerPtr()
	if err != nil {
		return 0 // Return 0 if the object is closed
	}

	ptr := C.luago_buffer_to_pointer(lptr)
	return uint64(ptr)
}

// Returns the LuaBuffer as a byte slice
func (l *LuaBuffer) Bytes() []byte {
	if l.lua.object.IsClosed() {
		return nil // Return nil if the Lua VM is closed
	}
	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return nil // Return nil if the object is closed
	}
	data := C.luago_buffer_to_bytes(ptr)
	bytes := moveBytesToGo(data)
	C.luago_buffer_free_bytes(data) // Free the buffer's bytes in Rust
	return bytes
}

// Returns the bytes from the LuaBuffer starting at the given offset
// with the specified length.
func (l *LuaBuffer) ReadBytes(offset, len uint64) []byte {
	if l.lua.object.IsClosed() {
		return nil // Return nil if the Lua VM is closed
	}
	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return nil // Return nil if the object is closed
	}
	data := C.luago_buffer_read_bytes(ptr, C.size_t(offset), C.size_t(len))
	bytes := moveBytesToGo(data)
	C.luago_buffer_free_bytes(data) // Free the buffer's bytes in Rust
	return bytes
}

// Writes data into the LuaBuffer starting at the given offset
func (l *LuaBuffer) WriteBytes(offset uint64, data []byte) error {
	if len(data) == 0 {
		return nil // No data to write, return early
	}

	if l.lua.object.IsClosed() {
		return fmt.Errorf("Lua VM is closed, cannot write to LuaBuffer")
	}

	if offset > l.Len() {
		return fmt.Errorf("offset %d is out of bounds for LuaBuffer of length %d", offset, l.Len())
	}
	if offset+uint64(len(data)) > l.Len() {
		return fmt.Errorf("data length %d exceeds LuaBuffer length %d at offset %d", len(data), l.Len(), offset)
	}

	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return err
	}

	C.luago_buffer_write_bytes(ptr, C.size_t(offset), (*C.char)(unsafe.Pointer(&data[0])), C.size_t(len(data)))
	return nil
}

// Returns the LuaBuffer's length
func (l *LuaBuffer) Len() uint64 {
	if l.lua.object.IsClosed() {
		return 0 // Return nil if the Lua VM is closed
	}
	l.object.RLock()
	defer l.object.RUnlock()
	ptr, err := l.innerPtr()
	if err != nil {
		return 0 // Return nil if the object is closed
	}
	size := C.luago_buffer_len(ptr)
	return uint64(size)
}

// Returns if the LuaBuffer is empty
func (l *LuaBuffer) IsEmpty() bool {
	return l == nil || l.Len() == 0
}

// String returns a string representation of the LuaBuffer.
//
// This is currently just the pointer address of the buffer.
func (l *LuaBuffer) String() string {
	ptr := l.Pointer()
	if ptr == 0 {
		return "<closed LuaBuffer>"
	}
	return fmt.Sprintf("LuaBuffer 0x%x", ptr)
}

// Equals checks if the LuaBuffer equals another LuaBuffer by lua value reference
func (l *LuaBuffer) Equals(other *LuaBuffer) bool {
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

	return bool(C.luago_buffer_equals(ptr, ptr2))
}

// ToValue converts the LuaBuffer to a Value.
func (l *LuaBuffer) ToValue() Value {
	return &ValueBuffer{value: l}
}

func (l *LuaBuffer) Close() error {
	if l == nil || l.object == nil {
		return nil // Nothing to close
	}
	// Close the LuaBuffer object
	return l.object.Close()
}
