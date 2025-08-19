package vm

/*
#include "../rustlib/rustlib.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

type LuaValueType int

const (
	LuaValueNil           LuaValueType = 0
	LuaValueBoolean       LuaValueType = 1
	LuaValueLightUserData LuaValueType = 2
	LuaValueInteger       LuaValueType = 3
	LuaValueNumber        LuaValueType = 4
	LuaValueVector        LuaValueType = 5
	LuaValueString        LuaValueType = 6
	LuaValueTable         LuaValueType = 7
	LuaValueFunction      LuaValueType = 8
	LuaValueThread        LuaValueType = 9
	LuaValueUserData      LuaValueType = 10
	LuaValueBuffer        LuaValueType = 11
	LuaValueOther         LuaValueType = 12

	// Custom types not sent by Rust ever
	// to make the library more ergonomic
	LuaValueCustom_GoString LuaValueType = 14
)

func (l LuaValueType) String() string {
	switch l {
	case LuaValueNil:
		return "nil"
	case LuaValueBoolean:
		return "boolean"
	case LuaValueLightUserData:
		return "lightuserdata"
	case LuaValueInteger:
		return "integer"
	case LuaValueNumber:
		return "number"
	case LuaValueVector:
		return "vector"
	case LuaValueString:
		return "string"
	case LuaValueTable:
		return "table"
	case LuaValueFunction:
		return "function"
	case LuaValueThread:
		return "thread"
	case LuaValueUserData:
		return "userdata"
	case LuaValueBuffer:
		return "buffer"
	case LuaValueOther:
		return "other"
	case LuaValueCustom_GoString:
		return "string"
	default:
		return "unknown"
	}
}

type Value interface {
	Type() LuaValueType
	Close() error
	object() *object // Returns the underlying object for this value
	String() string  // Returns a string representation of the value
	Equals(other Value) (bool, error)
}

// ValueNil represents a Lua nil value.
type ValueNil struct{}

func NewValueNil() *ValueNil {
	return &ValueNil{}
}
func (v *ValueNil) Type() LuaValueType {
	return LuaValueNil
}
func (v *ValueNil) Close() error { return nil }
func (v *ValueNil) object() *object {
	return nil // Nil has no underlying object
}
func (v *ValueNil) String() string {
	return "ValueNil"
}
func (v *ValueNil) Equals(other Value) (bool, error) {
	if other == nil {
		return true, nil
	}
	return other.Type() == LuaValueNil, nil // Only equal to other nil values
}

type ValueBoolean struct {
	value bool
}

func NewValueBoolean(value bool) *ValueBoolean {
	return &ValueBoolean{value: value}
}

// Value returns the boolean value of the ValueBoolean.
func (v *ValueBoolean) Value() bool {
	return v.value
}
func (v *ValueBoolean) Type() LuaValueType {
	return LuaValueBoolean
}
func (v *ValueBoolean) Close() error { return nil }
func (v *ValueBoolean) object() *object {
	return nil // Boolean has no underlying object
}
func (v *ValueBoolean) String() string {
	if v.value {
		return "ValueBoolean: true"
	}
	return "ValueBoolean: false"
}
func (v *ValueBoolean) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any boolean
	}
	if other.Type() != LuaValueBoolean {
		return false, nil // Only equal to other booleans
	}
	otherBool := other.(*ValueBoolean)
	return v.value == otherBool.value, nil // Compare boolean values
}

// ValueLightUserData is a pointer to an arbitrary C data type.
type ValueLightUserData struct {
	value unsafe.Pointer
}

func NewValueLightUserData(value unsafe.Pointer) *ValueLightUserData {
	return &ValueLightUserData{value: value}
}

// Value returns the pointer to the light user data.
// This pointer is not managed by Lua and should be used with caution.
// It is typically used for passing pointers to C code or for storing arbitrary data.
func (v *ValueLightUserData) Value() unsafe.Pointer {
	return v.value
}
func (v *ValueLightUserData) Type() LuaValueType {
	return LuaValueLightUserData
}
func (v *ValueLightUserData) Close() error { return nil }
func (v *ValueLightUserData) object() *object {
	return nil // LightUserData has no underlying object
}
func (v *ValueLightUserData) String() string {
	if v.value == nil {
		return "ValueLightUserData: <nil>"
	}
	return fmt.Sprintf("ValueLightUserData: %p", v.value)
}
func (v *ValueLightUserData) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any light user data
	}
	if other.Type() != LuaValueLightUserData {
		return false, nil // Only equal to other light user data
	}
	otherData := other.(*ValueLightUserData)
	return v.value == otherData.value, nil // Compare pointers
}

// ValueInteger represents a Lua integer value.
type ValueInteger struct {
	value int64
}

func NewValueInteger(value int64) *ValueInteger {
	return &ValueInteger{value: value}
}

func (v *ValueInteger) Value() int64 {
	return v.value
}
func (v *ValueInteger) Type() LuaValueType {
	return LuaValueInteger
}
func (v *ValueInteger) Close() error { return nil }
func (v *ValueInteger) object() *object {
	return nil // Integer has no underlying object
}
func (v *ValueInteger) String() string {
	return fmt.Sprintf("ValueInteger: %d", v.value)
}
func (v *ValueInteger) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any integer
	}
	switch other.Type() {
	case LuaValueInteger:
		otherInt := other.(*ValueInteger)
		return v.value == otherInt.value, nil // Compare integer values
	case LuaValueNumber:
		otherNum := other.(*ValueNumber)
		return float64(v.value) == otherNum.value, nil // Compare integer with number
	default:
		return false, nil // Only equal to other integers or numbers
	}
}

// ValueNumber represents a Lua number value.
type ValueNumber struct {
	value float64
}

func NewValueNumber(value float64) *ValueNumber {
	return &ValueNumber{value: value}
}

// Value returns the float64 value of the ValueNumber.
func (v *ValueNumber) Value() float64 {
	return v.value
}
func (v *ValueNumber) Type() LuaValueType {
	return LuaValueNumber
}
func (v *ValueNumber) Close() error { return nil }
func (v *ValueNumber) object() *object {
	return nil // Number has no underlying object
}
func (v *ValueNumber) String() string {
	return fmt.Sprintf("ValueNumber: %f", v.value)
}
func (v *ValueNumber) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any number
	}
	switch other.Type() {
	case LuaValueNumber:
		otherNum := other.(*ValueNumber)
		return v.value == otherNum.value, nil // Compare number values
	case LuaValueInteger:
		otherInt := other.(*ValueInteger)
		return v.value == float64(otherInt.value), nil // Compare number with integer
	default:
		return false, nil // Only equal to other numbers or integers
	}
}

// ValueVector represents a Luau vector value (3D vector).
//
// This is Luau-specific
type ValueVector struct {
	value [3]float32
}

func NewValueVector(x, y, z float32) *ValueVector {
	return &ValueVector{value: [3]float32{x, y, z}}
}

// Value returns the [3]float32 value of the ValueVector.
func (v *ValueVector) Value() [3]float32 {
	return v.value
}
func (v *ValueVector) Type() LuaValueType {
	return LuaValueVector
}
func (v *ValueVector) Close() error { return nil }
func (v *ValueVector) object() *object {
	return nil // Vector has no underlying object
}
func (v *ValueVector) String() string {
	return fmt.Sprintf("ValueVector: [%f, %f, %f]", v.value[0], v.value[1], v.value[2])
}
func (v *ValueVector) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any vector
	}
	if other.Type() != LuaValueVector {
		return false, nil // Only equal to other vectors
	}
	otherVec := other.(*ValueVector)
	return v.value == otherVec.value, nil // Compare vector values
}

// ValueString represents a Lua string value.
type ValueString struct {
	value *LuaString
}

// NewValueString creates a new ValueString from a LuaString.
func (v *ValueString) Value() *LuaString {
	return v.value
}
func (v *ValueString) Type() LuaValueType {
	return LuaValueString
}
func (v *ValueString) Close() error {
	return v.value.Close()
}
func (v *ValueString) object() *object {
	return v.value.object
}
func (v *ValueString) String() string {
	if v.value == nil {
		return "<nil ValueString>"
	}
	return "ValueString: " + v.value.String()
}
func (v *ValueString) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any string
	}
	switch other.Type() {
	case LuaValueString:
		otherStr := other.(*ValueString)
		if v.value == nil || otherStr.value == nil {
			return v.value == nil && otherStr.value == nil, nil // Both nil are equal
		}
		return v.value.Equals(otherStr.value), nil // Compare string content
	case LuaValueCustom_GoString:
		otherGoStr := other.(GoString)
		if v.value == nil {
			return false, nil // Nil string is not equal to GoString
		}
		return v.value.String() == string(otherGoStr), nil // Compare LuaString with GoString
	default:
		return false, nil // Only equal to other strings or GoStrings
	}
}

// ValueTable represents a Lua table value.
type ValueTable struct {
	value *LuaTable
}

func (v *ValueTable) Value() *LuaTable {
	return v.value
}
func (v *ValueTable) Type() LuaValueType {
	return LuaValueTable
}
func (v *ValueTable) Close() error {
	return v.value.Close()
}
func (v *ValueTable) object() *object {
	if v.value == nil {
		return nil // Table has no underlying object if nil
	}
	return v.value.object
}
func (v *ValueTable) String() string {
	if v.value == nil {
		return "<nil ValueTable>"
	}
	return "ValueTable: " + v.value.String()
}
func (v *ValueTable) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any table
	}
	if other.Type() != LuaValueTable {
		return false, nil // Only equal to other tables
	}
	otherTable := other.(*ValueTable)
	if v.value == nil || otherTable.value == nil {
		return v.value == nil && otherTable.value == nil, nil // Both nil are equal
	}
	return v.value.Equals(otherTable.value) // Compare table content
}

type ValueFunction struct {
	value *LuaFunction
}

func (v *ValueFunction) Value() *LuaFunction {
	return v.value
}
func (v *ValueFunction) Type() LuaValueType {
	return LuaValueFunction
}
func (v *ValueFunction) Close() error {
	return v.value.Close()
}
func (v *ValueFunction) object() *object {
	return v.value.object
}
func (v *ValueFunction) String() string {
	if v.value == nil {
		return "<nil ValueFunction>"
	}
	return "ValueFunction: " + v.value.String()
}
func (v *ValueFunction) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any function
	}
	if other.Type() != LuaValueFunction {
		return false, nil // Only equal to other functions
	}
	otherFunc := other.(*ValueFunction)
	if v.value == nil || otherFunc.value == nil {
		return v.value == nil && otherFunc.value == nil, nil // Both nil are equal
	}
	return v.value.Equals(otherFunc.value), nil // Compare function content
}

type ValueThread struct {
	value *LuaThread
}

// Value returns the LuaThread value of the ValueThread.
func (v *ValueThread) Value() *LuaThread {
	return v.value
}
func (v *ValueThread) Type() LuaValueType {
	return LuaValueThread
}
func (v *ValueThread) Close() error {
	return v.value.Close()
}
func (v *ValueThread) object() *object {
	return v.value.object
}
func (v *ValueThread) String() string {
	if v.value == nil {
		return "<nil ValueThread>"
	}
	return "ValueThread: " + v.value.String()
}
func (v *ValueThread) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any thread
	}
	if other.Type() != LuaValueThread {
		return false, nil // Only equal to other threads
	}
	otherThread := other.(*ValueThread)
	if v.value == nil || otherThread.value == nil {
		return v.value == nil && otherThread.value == nil, nil // Both nil are equal
	}
	return v.value.Equals(otherThread.value), nil // Compare thread content
}

type ValueUserData struct {
	value *LuaUserData
}

// Value returns the LuaUserData value of the ValueUserData.
func (v *ValueUserData) Value() *LuaUserData {
	return v.value
}
func (v *ValueUserData) Type() LuaValueType {
	return LuaValueUserData
}
func (v *ValueUserData) Close() error {
	return v.value.Close() // Close the LuaUserData
}
func (v *ValueUserData) object() *object {
	return v.value.object // Return the underlying object of the LuaUserData
}
func (v *ValueUserData) String() string {
	if v.value == nil {
		return "<nil ValueUserData>"
	}
	return "ValueUserData: " + v.value.String()
}
func (v *ValueUserData) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any userdata
	}
	if other.Type() != LuaValueUserData {
		return false, nil // Only equal to other userdata
	}
	otherUserData := other.(*ValueUserData)
	if v.value == nil || otherUserData.value == nil {
		return v.value == nil && otherUserData.value == nil, nil // Both nil are equal
	}
	return v.value.Equals(otherUserData.value), nil // Compare userdata content
}

type ValueBuffer struct {
	value *LuaBuffer
}

func (v *ValueBuffer) Value() *LuaBuffer {
	return v.value
}
func (v *ValueBuffer) Type() LuaValueType {
	return LuaValueBuffer
}
func (v *ValueBuffer) Close() error {
	return v.value.Close() // Close the LuaBuffer
}
func (v *ValueBuffer) object() *object {
	return v.value.object // Return the underlying object of the LuaBuffer
}
func (v *ValueBuffer) String() string {
	if v.value == nil {
		return "<nil ValueBuffer>"
	}
	return "ValueBuffer: " + v.value.String()
}
func (v *ValueBuffer) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any buffer
	}
	if other.Type() != LuaValueBuffer {
		return false, nil // Only equal to other buffers
	}
	otherBuf := other.(*ValueBuffer)
	if v.value == nil || otherBuf.value == nil {
		return v.value == nil && otherBuf.value == nil, nil // Both nil are equal
	}
	return v.value.Equals(otherBuf.value), nil // Compare buffer content
}

type ValueOther struct {
	value *C.void // TODO
}

func (v *ValueOther) Type() LuaValueType {
	return LuaValueOther
}
func (v *ValueOther) Close() error { return nil }
func (v *ValueOther) object() *object {
	return nil // Other has no underlying object
}
func (v *ValueOther) String() string {
	return "ValueOther: <not implemented>" // TODO: Implement other value string representation
}
func (v *ValueOther) Equals(other Value) (bool, error) {
	return false, fmt.Errorf("ValueOther does not support equality comparison") // TODO: Implement other value equality
}

type GoString string

func (v GoString) Type() LuaValueType {
	return LuaValueCustom_GoString
}
func (v GoString) Close() error { return nil }
func (v GoString) object() *object {
	return nil // GoString has no underlying object
}
func (v GoString) String() string {
	return string(v) // Convert GoString to string
}
func (v GoString) Equals(other Value) (bool, error) {
	if other == nil {
		return false, nil // Nil is not equal to any GoString
	}
	switch other.Type() {
	case LuaValueCustom_GoString:
		otherGoStr := other.(GoString)
		return string(v) == string(otherGoStr), nil // Compare GoString content
	case LuaValueString:
		otherStr := other.(*ValueString)
		if otherStr.value == nil {
			return false, nil // Nil LuaString is not equal to GoString
		}
		return string(v) == otherStr.value.String(), nil // Compare GoString with LuaString
	default:
		return false, nil // Only equal to other GoStrings or LuaStrings
	}
}

// CloneValue clones a vmlib.Value to a new vmlib.Value
//
// Locking behavior: CloneValue acquires a read lock on the object
// being cloned (value) to ensure it is not closed while cloning.
func CloneValue(l *Lua, value Value) (Value, error) {
	if value == nil {
		return nil, errors.New("cannot clone nil value")
	}

	// Acquire read lock to ensure the object is not closed while cloning
	obj := value.object()
	if obj != nil {
		obj.RLock()
		defer obj.RUnlock()
	}

	cVal, err := l._directValueToC(value)
	if err != nil {
		return nil, err
	}
	clonedCVal := cloneValue(cVal)

	clonedValue := l.valueFromC(clonedCVal)

	return clonedValue, nil
}

// cloneValue clones a C struct_GoLuaValue to a new C struct_GoLuaValue.
func cloneValue(item C.struct_GoLuaValue) C.struct_GoLuaValue {
	return C.luago_value_clone(item)
}

// destroyValue destroys a C struct_GoLuaValue.
func destroyValue(item C.struct_GoLuaValue) {
	C.luago_value_destroy(item)
}

// ValueFromC converts a C struct_GoLuaValue to a Go Value interface.
// Note: this does not clone the value, it simply converts it.
//
// Internal API: do not use unless you know what you're doing
func (l *Lua) valueFromC(item C.struct_GoLuaValue) Value {
	if l == nil {
		panic("internal safety check failure: Lua instance is nil, cannot perform valueFromC (this is a bug, please report it)")
	}

	switch item.tag {
	case C.LuaValueTypeNil:
		return &ValueNil{}
	case C.LuaValueTypeBoolean:
		val := *(*bool)(unsafe.Pointer(&item.data))
		return &ValueBoolean{value: val}
	case C.LuaValueTypeLightUserData:
		valPtr := *(**unsafe.Pointer)(unsafe.Pointer(&item.data))
		val := *valPtr
		return &ValueLightUserData{value: val}
	case C.LuaValueTypeInteger:
		val := *(*int64)(unsafe.Pointer(&item.data))
		return &ValueInteger{value: val}
	case C.LuaValueTypeNumber:
		val := *(*float64)(unsafe.Pointer(&item.data))
		return &ValueNumber{value: val}
	case C.LuaValueTypeVector:
		vec := *(*[3]C.float)(unsafe.Pointer(&item.data))
		return &ValueVector{value: [3]float32{float32(vec[0]), float32(vec[1]), float32(vec[2])}}
	case C.LuaValueTypeString:
		ptrToPtr := (**C.struct_LuaString)(unsafe.Pointer(&item.data))
		strPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		str := &LuaString{object: newObject(strPtr, stringTab), lua: l}
		return &ValueString{value: str}
	case C.LuaValueTypeTable:
		ptrToPtr := (**C.struct_LuaTable)(unsafe.Pointer(&item.data))
		tabPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		tab := &LuaTable{object: newObject(tabPtr, tableTab), lua: l}
		return &ValueTable{value: tab}
	case C.LuaValueTypeFunction:
		ptrToPtr := (**C.struct_LuaFunction)(unsafe.Pointer(&item.data))
		funcPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		funct := &LuaFunction{object: newObject(funcPtr, functionTab), lua: l}
		return &ValueFunction{value: funct}
	case C.LuaValueTypeThread:
		ptrToPtr := (**C.struct_LuaThread)(unsafe.Pointer(&item.data))
		threadPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		thread := &LuaThread{object: newObject(threadPtr, threadTab), lua: l}
		return &ValueThread{value: thread}
	case C.LuaValueTypeUserData:
		ptrToPtr := (**C.struct_LuaUserData)(unsafe.Pointer(&item.data))
		udPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		udt := &LuaUserData{object: newObject(udPtr, userdataTab), lua: l}
		return &ValueUserData{value: udt}
	case C.LuaValueTypeBuffer:
		ptrToPtr := (**C.struct_LuaBuffer)(unsafe.Pointer(&item.data))
		bufPtr := (*C.void)(unsafe.Pointer(*ptrToPtr))
		buf := &LuaBuffer{object: newObject(bufPtr, bufferTab), lua: l}
		return &ValueBuffer{value: buf}
	case C.LuaValueTypeOther:
		// Currently, always nil
		return &ValueOther{value: nil} // TODO: Support other types
	default:
		// Unknown type, return as Other
		return &ValueOther{value: nil} // Return nil for unknown types (as we cannot safely handle them)
	}
}

// DirectValueToC converts a Go Value interface to a C struct_GoLuaValue
// with the intent that the value will be passed to Rust code.
// Note: this does not clone the value or performing any locking, it simply converts it.
//
// Internal API: do not use unless you know what you're doing
//
// # WARNING
//
// You probably want to use ValueToC instead of this function.
//
// In particular, ValueFromC should *never* be called directly on the result of this function,
// as it may lead to memory corruption or undefined behavior.
func (l *Lua) _directValueToC(value Value) (C.struct_GoLuaValue, error) {
	if l == nil {
		panic("internal safety check failure: Lua instance is nil, cannot perform _directValueToC (this is a bug, please report it)")
	}

	var cVal C.struct_GoLuaValue
	switch value.Type() {
	case LuaValueNil:
		break
	case LuaValueBoolean:
		boolVal := value.(*ValueBoolean)
		cVal.tag = C.LuaValueTypeBoolean
		*(*C.bool)(unsafe.Pointer(&cVal.data)) = C.bool(boolVal.value)
	case LuaValueLightUserData:
		lightUserDataVal := value.(*ValueLightUserData)
		cVal.tag = C.LuaValueTypeLightUserData
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = lightUserDataVal.value
	case LuaValueInteger:
		intVal := value.(*ValueInteger)
		cVal.tag = C.LuaValueTypeInteger
		*(*int64)(unsafe.Pointer(&cVal.data)) = intVal.value
	case LuaValueNumber:
		numVal := value.(*ValueNumber)
		cVal.tag = C.LuaValueTypeNumber
		*(*float64)(unsafe.Pointer(&cVal.data)) = numVal.value
	case LuaValueVector:
		cVal.tag = C.LuaValueTypeVector
		vecVal := value.(*ValueVector)
		*(*[3]float32)(unsafe.Pointer(&cVal.data)) = vecVal.value
	case LuaValueString:
		strVal := value.(*ValueString)
		if strVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaString from different Lua instance")
		}
		ptr, err := strVal.value.object.UnsafePointer()
		if err != nil {
			return cVal, errors.New("cannot convert closed LuaString to C value")
		}
		cVal.tag = C.LuaValueTypeString
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(ptr)
	case LuaValueTable:
		tabVal := value.(*ValueTable)
		if tabVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaTable from different Lua instance")
		}
		ptr, err := tabVal.value.object.UnsafePointer()
		if err != nil {
			return cVal, errors.New("cannot convert closed LuaTable to C value")
		}
		cVal.tag = C.LuaValueTypeTable
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(ptr)
	case LuaValueFunction:
		funcVal := value.(*ValueFunction)
		if funcVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaFunction from different Lua instance")
		}
		ptr, err := funcVal.value.object.UnsafePointer()
		if err != nil {
			return cVal, errors.New("cannot convert closed LuaFunction to C value")
		}
		cVal.tag = C.LuaValueTypeFunction
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(ptr)
	case LuaValueThread:
		threadVal := value.(*ValueThread)
		if threadVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaThread from different Lua instance")
		}
		if threadVal.value == nil {
			return cVal, errors.New("cannot convert nil LuaThread to C value")
		}
		cVal.tag = C.LuaValueTypeThread
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(threadVal.value)
	case LuaValueUserData:
		udVal := value.(*ValueUserData)
		if udVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaUserData from different Lua instance")
		}
		ptr, err := udVal.value.object.UnsafePointer()
		if err != nil {
			return cVal, errors.New("cannot convert closed LuaUserData to C value")
		}
		cVal.tag = C.LuaValueTypeUserData
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(ptr)
	case LuaValueBuffer:
		bufVal := value.(*ValueBuffer)
		if bufVal.value.lua != l {
			return cVal, errors.New("cannot convert LuaBuffer from different Lua instance")
		}
		ptr, err := bufVal.value.object.UnsafePointer()
		if err != nil {
			return cVal, errors.New("cannot convert closed LuaBuffer to C value")
		}
		cVal.tag = C.LuaValueTypeBuffer
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(ptr)
	case LuaValueOther:
		// Currently, always nil
		cVal.tag = C.LuaValueTypeOther
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = nil // Return nil
	case LuaValueCustom_GoString:
		// This is a temporary string that should not have a finalizer attached to it
		goStrVal := value.(GoString)
		// Create a LuaString from the Go string
		luaString, err := l.createStringAsPtr([]byte(goStrVal))
		if err != nil {
			return cVal, err // Return error if the string cannot be created
		}
		cVal.tag = C.LuaValueTypeString
		*(*unsafe.Pointer)(unsafe.Pointer(&cVal.data)) = unsafe.Pointer(luaString)
	default:
		return cVal, errors.New("unknown Lua value type")
	}

	return cVal, nil
}

// ValueToC converts a Go Value interface to a C struct_GoLuaValue
// with the intent that the value will be passed to Rust code.
// It disarms the value ref pointer to ensure it cannot be used after conversion.
//
// Internal API: do not use unless you know what you're doing
func (l *Lua) valueToC(value Value) (C.struct_GoLuaValue, error) {
	if value == nil {
		return C.struct_GoLuaValue{}, errors.New("cannot convert nil value to C")
	}

	obj := value.object()
	if obj != nil {
		// Disarm the object to prevent it from being used after conversion
		// Now disarm it
		//
		// Disarming the object won't affect the pointer but will stop normal function
		// access to it.
		err := obj.Disarm()
		if err != nil {
			return C.struct_GoLuaValue{}, fmt.Errorf("failed to disarm object: %w", err)
		}
	}

	cptr, err := l._directValueToC(value)
	if err != nil {
		return cptr, err
	}

	if obj != nil {
		// The object has already been disarmed (so only UnsafePointer can be used)
		// and Disarm will error if the object is already disarmed,
		// so this should be (??) safe in concurrent code
		obj.ptr = nil
	}

	return cptr, nil
}
