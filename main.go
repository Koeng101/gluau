package main

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	// Import to ensure callback package is initialized
	vmlib "github.com/gluau/gluau/vm"
)

// #include <stdlib.h>
// #include "./rustlib/rustlib.h"
import "C"

func main() {
	vm, err := vmlib.CreateLuaVm()
	if err != nil {
		fmt.Println("Error creating Lua VM:", err)
		return
	}
	defer vm.Close() // Ensure we close the VM when done
	fmt.Println("Lua VM created successfully", vm)
	// Example of creating a Lua string
	luaString, err := vm.CreateString("Hello, Lua!")
	if err != nil {
		fmt.Println("Error creating Lua string:", err)
		return
	}
	fmt.Println("Lua string created successfully:", luaString)
	fmt.Println("Lua string as bytes:", string(luaString.Bytes()))
	fmt.Println("Lua string as bytes without nil:", luaString.Bytes())
	fmt.Println("Lua string as bytes with nil:", luaString.BytesWithNul())
	fmt.Printf("Lua string pointer: 0x%x\n", luaString.Pointer())
	luaString.Close() // Clean up the Lua string when done
	fmt.Println("Lua string as bytes after free (should be empty/nil):", luaString.Bytes())

	// Example of creating a Lua table
	tab, err := vm.CreateTable()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua table: %v", err))
	}

	// Insert some values into the table
	err = tab.Set(vmlib.GoString("key1"), vmlib.GoString("value1"))
	if err != nil {
		panic(fmt.Sprintf("Failed to set value in Lua table: %v", err))
	}

	err = tab.Set(vmlib.GoString("key2"), vmlib.NewValueInteger(42))
	if err != nil {
		panic(fmt.Sprintf("Failed to set value in Lua table: %v", err))
	}

	err = tab.Set(vmlib.GoString("key3"), vmlib.NewValueVector(1, 2, 3))
	if err != nil {
		panic(fmt.Sprintf("Failed to set value in Lua table: %v", err))
	}

	var testKey vmlib.Value
	tab.ForEach(func(key, value vmlib.Value) error {
		if key.Type() == vmlib.LuaValueString {
			fmt.Println("Key is a LuaString:", key.(*vmlib.ValueString).Value().String())
			testKey = key
		}
		if value.Type() == vmlib.LuaValueString {
			fmt.Println("Value is a LuaString:", value.(*vmlib.ValueString).Value().String())
		} else if value.Type() == vmlib.LuaValueInteger {
			fmt.Println("Value is a LuaInteger:", value.(*vmlib.ValueInteger).Value())
		} else if value.Type() == vmlib.LuaValueVector {
			vec := value.(*vmlib.ValueVector).Value()
			fmt.Println("Value is a LuaVector:", vec[0], vec[1], vec[2])
		} else {
			return fmt.Errorf("unexpected value type: %s", value.Type().String())
		}
		fmt.Println("Key:", key, "Value:", value)
		//time.Sleep(time.Second * 20) // Simulate some processing time
		go func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println("Recovered from panic in goroutine:", r)
				}
			}()
			fmt.Println("Processing key-value pair in a goroutine:", key, value)
			// Simulate some processing
			fmt.Println("Finished processing key-value pair in goroutine:", key, value)
			panic("whee")
		}()
		return nil
	})

	fmt.Println("Key is a LuaString:", testKey.(*vmlib.ValueString).Value().String())

	err = tab.ForEach(func(key, value vmlib.Value) error {
		panic("test panic")
	})
	if err == nil {
		panic("Expected error from ForEach, got nil")
	} else if err.Error() != "panic in ForEach callback: test panic" {
		panic("Expected 'panic in ForEach callback: test panic' error, got: " + err.Error())
	}
	fmt.Println("ForEach callback error:", err)

	isEmpty := tab.IsEmpty()
	if isEmpty {
		panic("Non-empty Lua table is empty")
	}
	tablen, err := tab.Len()
	if err != nil {
		panic(fmt.Sprintf("Failed to get Lua table length: %v", err))
	}
	if tablen != 0 {
		panic("Lua table length should be 0 (as all key-value pairs so no array indices), got " + fmt.Sprint(tablen))
	}
	mt := tab.Metatable()
	if mt != nil {
		panic("Lua table should not have a metatable")
	}
	poppedValue, err := tab.Pop()
	if err != nil {
		panic(fmt.Sprintf("Failed to pop value from Lua table: %v", err))
	}
	if poppedValue.Type() != vmlib.LuaValueNil {
		panic(fmt.Sprintf("Expected LuaValueNil, got %d", poppedValue.Type()))
	}
	err = tab.Push(vmlib.GoString("test"))
	if err != nil {
		panic(fmt.Sprintf("Failed to push value to Lua table: %v", err))
	}
	tablen, err = tab.Len()
	if err != nil {
		panic(fmt.Sprintf("Failed to get Lua table length after push: %v", err))
	}
	if tablen != 1 {
		panic("Lua table length should be 1 after push, got " + fmt.Sprint(tablen))
	}
	fmt.Printf("Lua table string %s with ptr 0x%x\n", tab, tab.Pointer())

	// Create a new Lua table to act as this table's metatable
	myNewMt, err := vm.CreateTable()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua table for metatable: %v", err))
	}
	// Set the metatable for the Lua table
	err = tab.SetMetatable(myNewMt)
	if err != nil {
		panic(fmt.Sprintf("Failed to set metatable for Lua table: %v", err))
	}
	mt = tab.Metatable()
	if mt == nil {
		panic("Lua table should have a metatable after setting it")
	}
	doesItEqual, err := mt.Equals(myNewMt)
	if err != nil {
		panic(fmt.Sprintf("Failed to check if Lua table metatable equals another: %v", err))
	}
	if !doesItEqual {
		panic("Lua table metatable does not match the one we set")
	}
	err = tab.SetMetatable(nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to unset metatable for Lua table: %v", err))
	}
	mt = tab.Metatable()
	if mt != nil {
		panic("Lua table should not have a metatable after unsetting it")
	}

	// Clean up the Lua table when done
	err = tab.Close()
	if err != nil {
		fmt.Println("Error closing Lua table:", err)
		return
	}

	time.Sleep(time.Millisecond)

	luaEmptyString, err := vm.CreateString("")
	if err != nil {
		fmt.Println("Error creating Lua string:", err)
		return
	}
	fmt.Println("Lua empty string created successfully:", luaEmptyString)
	fmt.Println("Lua empty string as bytes:", luaEmptyString.Bytes())
	fmt.Println("Lua empty string as bytes with nil:", luaEmptyString.BytesWithNul())
	fmt.Printf("Lua empty string pointer: 0x%x\n", luaEmptyString.Pointer())
	luaEmptyString.Close() // Clean up the Lua empty string when done
	fmt.Println("Lua empty string as bytes after free (should be empty/nil):", luaEmptyString.Bytes())

	// Create a Lua table
	if err := vm.SetMemoryLimit(100000000000000); err != nil {
		panic(fmt.Sprintf("Failed to set memory limit: %v", err))
	}
	luaTable, err := vm.CreateTableWithCapacity(100000000, 10)
	if err != nil {
		fmt.Println("Error creating Lua table:", err)
	} else {
		fmt.Println("Lua table created successfully:", luaTable)
		panic("this should never happen (table overflow expected)")
	}
	defer luaTable.Close() // Ensure we close the Lua table when done

	luaTable2, err := vm.CreateTableWithCapacity(10000, 10)
	if err != nil {
		panic(err)
	}
	fmt.Println("Lua table created successfully:", luaTable2)
	defer luaTable2.Close() // Ensure we close the Lua table when done
	if err := luaTable2.Clear(); err != nil {
		panic(fmt.Sprintf("Failed to clear Lua table: %v", err))
	}
	fooStr, err := vm.CreateString("foo")
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua string: %v", err))
	}
	defer fooStr.Close() // Ensure we close the Lua string when done
	containsKey, err := luaTable2.ContainsKey(fooStr.ToValue())
	if err != nil {
		panic(fmt.Sprintf("Failed to check if Lua table contains key: %v", err))
	}
	if containsKey {
		panic("Lua table should not contain 'foo' key")
	}
	fmt.Println("empty table contains 'foo'", containsKey)
	equals, err := luaTable2.Equals(luaTable2)
	if err != nil {
		panic(fmt.Sprintf("Failed to check if Lua table equals another: %v", err))
	}
	if !equals {
		panic("Lua table should equal itself")
	}
	fmt.Println("empty table equals itself", equals)

	myFunc, err := vm.CreateFunction(func(lua *vmlib.Lua, args []vmlib.Value) ([]vmlib.Value, error) {
		return []vmlib.Value{
			vmlib.GoString("Hello world"),
		}, nil
	})
	if err != nil {
		panic(err)
	}

	res, err := myFunc.Call([]vmlib.Value{vmlib.GoString("foo")})
	if err != nil {
		panic(err)
	}
	fmt.Println("Function call response", res[0].(*vmlib.ValueString).Value().String())
	defer res[0].Close()

	res, err = myFunc.Call([]vmlib.Value{vmlib.GoString("foo")})
	if err != nil {
		panic(err)
	}
	fmt.Println("Function call response", res[0].(*vmlib.ValueString).Value().String())
	defer res[0].Close()

	myFunc, err = vm.CreateFunction(func(lua *vmlib.Lua, args []vmlib.Value) ([]vmlib.Value, error) {
		return nil, errors.New(args[0].(*vmlib.ValueString).Value().String())
	})
	if err != nil {
		panic(err)
	}

	_, err = myFunc.Call([]vmlib.Value{vmlib.GoString("foo")})
	if err != nil {
		fmt.Println("function error", err)
	}
	_, err = myFunc.Call([]vmlib.Value{vmlib.NewValueVector(1, 2, 3)})
	if err != nil {
		fmt.Println("function error", err)
	}

	runtime.GC()
	runtime.GC()

	tab, err = vm.CreateTable()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua table: %v", err))
	}
	defer tab.Close() // Ensure we close the Lua table when done
	err = tab.Set(vmlib.GoString("test"), myFunc.ToValue())
	if err != nil {
		panic(fmt.Sprintf("Failed to set value in Lua table: %v", err))
	}

	testFn, err := tab.Get(vmlib.GoString("test"))
	if err != nil {
		panic(fmt.Sprintf("Failed to get value from Lua table: %v", err))
	}
	if testFn.Type() != vmlib.LuaValueFunction {
		panic(fmt.Sprintf("Expected LuaValueFunction, got %d", testFn.Type()))
	}

	// Compiler API
	vm.SetCompilerOpts(vmlib.CompilerOpts{
		OptimizationLevel: vmlib.OptimizationLevelFull,
	})
	globTab, err := vm.CreateTable()
	if err != nil {
		panic("failed to make global table")
	}
	err = globTab.Set(vmlib.GoString("a"), vmlib.NewValueInteger(5))
	if err != nil {
		panic("failed to set a")
	}
	clonedGlobTab, err := vmlib.CloneValue(vm, globTab.ToValue())
	if err != nil {
		panic("failed to clone global table: " + err.Error())
	}
	err = globTab.Set(vmlib.GoString("_G"), clonedGlobTab)
	if err != nil {
		panic("failed to set _G")
	}
	runtime.GC() // testing

	luaFunc, err := vm.LoadChunk(vmlib.ChunkOpts{
		Code: "_G.a = _G.a + 1; return _G.a",
		Env:  globTab,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Lua func: ", luaFunc)

	// Lets clone this function
	clonedFunc, err := luaFunc.DeepClone()
	if err != nil {
		panic(fmt.Sprintf("Failed to clone Lua function: %v", err))
	}
	fmt.Println("Cloned Lua function:", clonedFunc)
	myNewTab, err := vm.CreateTable()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua table for cloned function: %v", err))
	}
	ok, err := clonedFunc.SetEnvironment(myNewTab)
	if err != nil {
		panic(fmt.Sprintf("Failed to set environment for cloned Lua function: %v", err))
	}
	if !ok {
		panic("Failed to set environment for cloned Lua function")
	}
	env, err := clonedFunc.Environment() // This will be myNewTab
	if err != nil {
		panic(fmt.Sprintf("Failed to get environment for cloned Lua function: %v", err))
	}
	if env.Pointer() != myNewTab.Pointer() {
		panic("Cloned Lua function environment does not match the one we set")
	}
	env, err = luaFunc.Environment()
	if err != nil {
		panic(fmt.Sprintf("Failed to get environment for original Lua function: %v", err))
	}
	if env.Pointer() != globTab.Pointer() {
		panic("Original Lua function environment does not match the one we set")
	}

	ret, err := luaFunc.Call([]vmlib.Value{})
	if err != nil {
		panic(err)
	}
	if len(ret) != 1 || ret[0].Type() != vmlib.LuaValueInteger {
		panic("ret is not a single integer")
	}
	if ret[0].(*vmlib.ValueInteger).Value() != 6 {
		panic("ret[0] must be 6")
	} else {
		fmt.Println("got ret:", ret[0].(*vmlib.ValueInteger).Value())
	}
	luaFunc.Close()

	udMt, err := vm.CreateTable()
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua table for userdata metatable: %v", err))
	}
	// Set the __type
	err = udMt.Set(vmlib.GoString("__type"), vmlib.GoString("MyUserDataType"))
	if err != nil {
		panic(fmt.Sprintf("Failed to set __type in Lua userdata metatable: %v", err))
	}

	ud, err := vm.CreateUserData("test data", udMt)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Lua user data: %v", err))
	}
	fmt.Println("Lua user data created successfully:", ud)
	associatedData, err := ud.AssociatedData()
	if err != nil {
		panic(fmt.Sprintf("Failed to get associated data from Lua user data: %v", err))
	}
	fmt.Println("Associated data from Lua user data:", associatedData)
	if associatedData != "test data" {
		panic(fmt.Sprintf("Expected associated data 'test data', got '%v'", associatedData))
	}

	// Interrupt API
	vm.SetInterrupt(func(funcVm *vmlib.Lua) (vmlib.VmState, error) {
		return vmlib.VmStateContinue, errors.New("test interrupt error")
	})

	// Create a Lua function that will trigger the interrupt
	luaFunc.Close() // Close the previous function to avoid memory leaks
	luaFunc, err = vm.LoadChunk(vmlib.ChunkOpts{
		Name: "test_interrupt",
		Code: "while true do end", // Infinite loop to trigger the interrupt
		Env:  globTab,
	})
	if err != nil {
		panic(err)
	}

	// Call the Lua function to trigger the interrupt
	_, err = luaFunc.Call([]vmlib.Value{})
	if err != nil {
		if !strings.Contains(err.Error(), "test interrupt error") {
			panic(fmt.Sprintf("Expected interrupt error, got: %v", err))
		}
		fmt.Println("Lua function call error (expected due to interrupt):", err)
	} else {
		panic("Expected an error from the Lua function call due to interrupt")
	}

	// Set a new interrupt which will yield the execution
	// after 100 milliseconds
	timeNow := time.Now()
	vm.SetInterrupt(func(funcVm *vmlib.Lua) (vmlib.VmState, error) {
		if time.Since(timeNow) > 10*time.Millisecond {
			fmt.Println("Interrupt triggered after 1 second")
			return vmlib.VmStateYield, nil // Yield the execution after 100 milliseconds
		}
		return vmlib.VmStateContinue, nil // Continue execution
	})

	// Call the Lua function again to trigger the interrupt
	//
	// Currently, as gluau does not implement LuaThread yet, this will yield a attempt to yield
	// across metamethod/C-call boundary
	_, err = luaFunc.Call([]vmlib.Value{})
	if err != nil {
		fmt.Println("Lua function call error (expected due to interrupt):", err)
	} else {
		panic("Expected an error from the Lua function call due to interrupt")
	}
}
