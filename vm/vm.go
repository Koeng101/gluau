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

var luaVmTab = objectTab{
	dtor: func(ptr *C.void) {
		C.freeluavm((*C.struct_Lua)(unsafe.Pointer(ptr)))
	},
}

// A handle to the Lua VM.
type Lua struct {
	obj *object
}

func (l *Lua) lua() (*C.struct_Lua, error) {
	ptr, err := l.obj.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the object is closed
	}
	return (*C.struct_Lua)(unsafe.Pointer(ptr)), nil
}

// SetCompilerOpts sets the default compiler options for the Lua VM.
//
// This is a Luau-specific feature
func (l *Lua) SetCompilerOpts(opts CompilerOpts) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return // No-op if the Lua VM is closed
	}

	cOpts := opts.toC()
	C.luavm_setcompileropts(lua, cOpts)
}

// SetMemoryLimit sets the memory limit for the Lua VM.
//
// Upon exceeding this limit, Luau will return a memory error
// back to the caller (which may either be in Luau still or in Go
// as a error value).
func (l *Lua) SetMemoryLimit(limit int) error {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return err
	}
	res := C.luavm_setmemorylimit(lua, C.size_t(limit))
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return err
	}
	return nil
}

// Sandbox enables or disables the sandbox mode for the Luau VM.
//
// This method, in particular:
//
// - Set all libraries to read-only
// - Set all builtin metatables to read-only
// - Set globals to read-only (and activates safeenv)
// - Setup local environment table that performs writes locally and proxies reads to the global environment.
// - Allow only count mode in collectgarbage function.
//
// Note that this is a Luau-specific feature.
func (l *Lua) Sandbox(enabled bool) error {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return err
	}
	res := C.luavm_sandbox(lua, C.bool(enabled))
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return err
	}
	return nil
}

// Globals returns the global environment table of the Lua VM.
func (l *Lua) Globals(enabled bool) *LuaTable {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil
	}
	globals := C.luago_globals(lua)
	if globals == nil {
		return nil // Return nil if the globals table is not available
	}
	return &LuaTable{object: newObject((*C.void)(unsafe.Pointer(globals)), tableTab), lua: l}
}

// SetGlobals sets the global environment table of the Lua VM.
//
// Note that any existing Lua functions have cached global environment and will not see the changes made by this method.
//
// To update the environment for existing Lua functions, use LuaFunction.SetEnvironment
func (l *Lua) SetGlobals(tab *LuaTable) error {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil
	}
	if tab == nil {
		return errors.New("globals table cannot be nil")
	}
	defer tab.object.RUnlock()
	tab.object.RLock()
	tabPtr, err := tab.innerPtr()
	if err != nil {
		return err // Return error if the table is closed
	}
	res := C.luago_setglobals(lua, tabPtr)
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return err
	}
	return nil
}

type VmState int

const (
	VmStateContinue VmState = iota
	VmStateYield            // Yield the VM execution / stop execution
)

type InterruptFn = func(funcVm *Lua) (VmState, error)

// Sets an interrupt function that will periodically be called by Luau VM.
//
// Any Luau code is guaranteed to call this handler “eventually” (in practice this can happen at any function call or at any loop iteration).
//
// The provided interrupt function can error, and this error will be propagated through the Luau code that was executing at the time the interrupt was triggered.
//
// Also this can be used to implement continuous execution limits by instructing Luau VM to yield by returning VmState::Yield.
func (l *Lua) SetInterrupt(callback InterruptFn) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return
	}

	cbWrapper := newGoCallback(func(val unsafe.Pointer) {
		cval := (*C.struct_InterruptData)(val)

		// Safety: it is undefined behavior for the callback to unwind into
		// Rust (or even C!) frames from Go, so we must recover() any panic
		// that occurs in the callback to prevent a crash.
		defer func() {
			if r := recover(); r != nil {
				// Deallocate any existing error
				if cval.error != nil {
					C.luago_error_free(cval.error)
				}

				// Replace
				errBytes := []byte(fmt.Sprintf("panic in CreateFunction callback: %v", r))
				errv := C.luago_error_new((*C.char)(unsafe.Pointer(&errBytes[0])), C.size_t(len(errBytes)))
				cval.error = errv // Rust side will deallocate it for us
			}
		}()

		callbackVm := &Lua{obj: newObject((*C.void)(unsafe.Pointer(cval.lua)), luaVmTab)}
		defer callbackVm.Close() // Free the memory associated with the callback VM. TODO: Maybe switch to using a Deref API instead of Close?

		vmState, err := callback(callbackVm)

		if err != nil {
			errBytes := []byte(err.Error())
			errv := C.luago_error_new((*C.char)(unsafe.Pointer(&errBytes[0])), C.size_t(len(errBytes)))
			cval.error = errv // Rust side will deallocate it for us
			return
		}

		cval.vm_state = C.uint8_t(vmState)
	}, func() {
		fmt.Println("interrupt callback is being dropped")
	})

	C.luago_set_interrupt(lua, cbWrapper.ToC())
}

// Removes the interrupt function set by SetInterrupt.
func (l *Lua) RemoveInterrupt() {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return // No-op if the Lua VM is closed
	}

	C.luago_remove_interrupt(lua)
}

// CreateString creates a Lua string from a Go string.
func (l *Lua) CreateString(s string) (*LuaString, error) {
	return l.createString([]byte(s))
}

// CreateStringBytes creates a Lua string from a byte slice.
// This is useful for creating strings from raw byte data.
func (l *Lua) CreateStringBytes(s []byte) (*LuaString, error) {
	return l.createString(s)
}

func (l *Lua) createString(s []byte) (*LuaString, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	if len(s) == 0 {
		// Passing nil to luago_create_string creates an empty string.
		res := C.luago_create_string(lua, (*C.char)(nil), C.size_t(len(s)))
		if res.error != nil {
			return nil, moveErrorToGoError(res.error)
		}
		return &LuaString{object: newObject((*C.void)(unsafe.Pointer(res.value)), stringTab)}, nil
	}

	res := C.luago_create_string(lua, (*C.char)(unsafe.Pointer(&s[0])), C.size_t(len(s)))
	if res.error != nil {
		return nil, moveErrorToGoError(res.error)
	}
	return &LuaString{object: newObject((*C.void)(unsafe.Pointer(res.value)), stringTab)}, nil
}

// Create string as pointer (without any finalizer)
func (l *Lua) createStringAsPtr(s []byte) (*C.struct_LuaString, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	if len(s) == 0 {
		// Passing nil to luago_create_string creates an empty string.
		res := C.luago_create_string(lua, (*C.char)(nil), C.size_t(len(s)))
		if res.error != nil {
			return nil, moveErrorToGoError(res.error)
		}
		return res.value, nil
	}

	res := C.luago_create_string(lua, (*C.char)(unsafe.Pointer(&s[0])), C.size_t(len(s)))
	if res.error != nil {
		return nil, moveErrorToGoError(res.error)
	}
	return res.value, nil
}

// CreateTable creates a new Lua table.
func (l *Lua) CreateTable() (*LuaTable, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	res := C.luago_create_table(lua)
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}
	return &LuaTable{object: newObject((*C.void)(unsafe.Pointer(res.value)), tableTab), lua: l}, nil
}

// CreateTableWithCapacity creates a new Lua table with specified capacity for array and record parts.
// with narr as the number of array elements and nrec as the number of record elements.
func (l *Lua) CreateTableWithCapacity(narr, nrec int) (*LuaTable, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	res := C.luago_create_table_with_capacity(lua, C.size_t(narr), C.size_t(nrec))
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}
	return &LuaTable{object: newObject((*C.void)(unsafe.Pointer(res.value)), tableTab), lua: l}, nil
}

// CreateErrorVariant creates a new ErrorVariant from a byte slice.
func CreateErrorVariant(s []byte) *ErrorVariant {
	if len(s) == 0 {
		// Passing nil to luago_create_string creates an empty string.
		res := C.luago_error_new((*C.char)(nil), C.size_t(len(s)))
		return &ErrorVariant{object: newObject((*C.void)(unsafe.Pointer(res)), errorVariantTab)}
	}

	res := C.luago_error_new((*C.char)(unsafe.Pointer(&s[0])), C.size_t(len(s)))
	return &ErrorVariant{object: newObject((*C.void)(unsafe.Pointer(res)), errorVariantTab)}
}

type FunctionFn = func(funcVm *Lua, args []Value) ([]Value, error)

// CreateFunction creates a new Function
//
// # Note that funcVm will only be open until the callback function returns
//
// Locking behavior: All values returned by the callback function
// will be write-locked (taken ownership of). Having any sort of read-lock
// during a return will cause a error to be returned to Luau
func (l *Lua) CreateFunction(callback FunctionFn) (*LuaFunction, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	cbWrapper := newGoCallback(func(val unsafe.Pointer) {
		cval := (*C.struct_FunctionCallbackData)(val)

		// Safety: it is undefined behavior for the callback to unwind into
		// Rust (or even C!) frames from Go, so we must recover() any panic
		// that occurs in the callback to prevent a crash.
		defer func() {
			if r := recover(); r != nil {
				// Deallocate any existing error
				if cval.error != nil {
					C.luago_error_free(cval.error)
				}

				// Replace
				errBytes := []byte(fmt.Sprintf("panic in CreateFunction callback: %v", r))
				errv := C.luago_error_new((*C.char)(unsafe.Pointer(&errBytes[0])), C.size_t(len(errBytes)))
				cval.error = errv // Rust side will deallocate it for us
			}
		}()

		// Take out args
		// mw as a object will be deallocated by the Rust side
		mw := &luaMultiValue{ptr: cval.args, lua: l}
		args := mw.take()

		callbackVm := &Lua{obj: newObject((*C.void)(unsafe.Pointer(cval.lua)), luaVmTab)}
		defer callbackVm.Close() // Free the memory associated with the callback VM. TODO: Maybe switch to using a Deref API instead of Close?

		values, err := callback(callbackVm, args)

		if err != nil {
			errBytes := []byte(err.Error())
			errv := C.luago_error_new((*C.char)(unsafe.Pointer(&errBytes[0])), C.size_t(len(errBytes)))
			cval.error = errv // Rust side will deallocate it for us
			return
		}

		outMw, err := l.multiValueFromValues(values)
		if err != nil {
			errBytes := []byte(err.Error())
			errv := C.luago_error_new((*C.char)(unsafe.Pointer(&errBytes[0])), C.size_t(len(errBytes)))
			cval.error = errv // Rust side will deallocate it for us
			return
		}

		cval.values = outMw.ptr // Rust will deallocate values as well
	}, func() {
		fmt.Println("function callback is being dropped")
	})

	res := C.luago_create_function(lua, cbWrapper.ToC())
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}

	return &LuaFunction{object: newObject((*C.void)(unsafe.Pointer(res.value)), functionTab), lua: l}, nil
}

// LoadChunk loads a Lua chunk from the given options.
func (l *Lua) LoadChunk(opts ChunkOpts) (*LuaFunction, error) {
	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	var env *C.struct_LuaTable
	if opts.Env != nil {
		defer opts.Env.object.RUnlock()
		opts.Env.object.RLock()
		envPtr, err := opts.Env.object.PointerNoLock()
		if err == nil {
			env = (*C.struct_LuaTable)(unsafe.Pointer(envPtr))
		}
	}

	var compilerOpts *C.struct_CompilerOpts = nil
	if opts.CompilerOpts != nil {
		compilerOptsC := opts.CompilerOpts.toC()
		compilerOpts = &compilerOptsC
	}

	var name = newChunkString([]byte(opts.Name))
	var code = newChunkString([]byte(opts.Code))

	res := C.luago_load_chunk(
		lua,
		C.struct_ChunkOpts{
			name:          name,
			env:           env,
			mode:          C.uint8_t(opts.Mode),
			compiler_opts: compilerOpts,
			code:          code,
		},
	)

	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}
	return &LuaFunction{object: newObject((*C.void)(unsafe.Pointer(res.value)), functionTab), lua: l}, nil
}

// CreateUserData creates a LuaUserData with associated data and a metatable.
func (l *Lua) CreateUserData(associatedData any, mt *LuaTable) (*LuaUserData, error) {
	if mt == nil {
		return nil, fmt.Errorf("metatable cannot be nil")
	}

	l.obj.RLock()
	defer l.obj.RUnlock()

	lua, err := l.lua()
	if err != nil {
		return nil, err
	}

	defer mt.object.RUnlock()
	mt.object.RLock()
	mtPtr, err := mt.object.PointerNoLock()
	if err != nil {
		return nil, err // Return error if the metatable is closed
	}

	dynData := newDynamicData(associatedData, func() {
		fmt.Println("dynamic data is being dropped")
	})
	cDynData := dynData.ToC()
	res := C.luago_create_userdata(lua, cDynData, (*C.struct_LuaTable)(unsafe.Pointer(mtPtr)))
	if res.error != nil {
		err := moveErrorToGoError(res.error)
		return nil, err
	}
	return &LuaUserData{
		lua:    l,
		object: newObject((*C.void)(unsafe.Pointer(res.value)), userdataTab),
	}, nil
}

func (l *Lua) Close() error {
	if l == nil || l.obj == nil {
		return nil // Nothing to close
	}

	// Close the Lua VM object
	return l.obj.Close()
}

func CreateLuaVm() (*Lua, error) {
	ptr := C.newluavm()
	if ptr == nil {
		return nil, fmt.Errorf("failed to create Lua VM")
	}
	vm := &Lua{obj: newObject((*C.void)(unsafe.Pointer(ptr)), luaVmTab)}
	return vm, nil
}
