# gluau

gluau provides Go bindings for the Luau (dialect of Lua) programming language

**Heavy WIP**: This library is currently in development and not yet ready for production use. The API may change frequently, and there are many features that are not yet implemented.

## Implemented APIs so far

- VM initialization and shutdown
- Basic Lua value API to abstract over Lua values via Go interfaces
- Lua Strings (along with API's)
- Lua Tables (along with API's)
- Lua Functions (API's are WIP, but basic creating from both Luau and Go and calling functions is implemented)
- Lua userdata (along with API's)
- Lua Threads (API's mostly implemented, but not fully. Both resume and yield are supported)

## Value Ownership Semantics

There are two cases in which gluau will take ownership of a value:
- When a ``vmlib.Value`` is passed to a function that takes a ``vmlib.Value`` as an argument.
- When a ``vmlib.Value`` is returned from a Luau function implemented in Go.

By default, any function in gluau's vm library that takes in a ``vmlib.Value`` will take ownership of the value (and its internals). This means that the value will be closed/disarmed so Go code cannot use it. As such, attempting to recursively disarm a value will return a error.

As an example:

```go
// This is bad and will hit the panic: failed to set _G
err = globTab.Set(vmlib.GoString("_G"), globTab.ToValue())
if err != nil {
    panic("failed to set _G")
}
```

In the above code, ``globTab.ToValue()`` returns a ``vmlib.Value``. ``Set`` however takes ownership of the value which requires a write-lock on the value and a read-lock on the table object. This is obviously a conflict which gluau detects and returns an error for.

To fix this, you can use the ``CloneValue`` function to create a new value that is a clone of the original value. This will allow you to pass the cloned value to the ``Set`` function without taking ownership of the original value.

```go
// The correct way
clonedGlobTab, err := vmlib.CloneValue(vm, globTab.ToValue())
if err != nil {
    panic("failed to clone global table: " + err.Error())
}
// The set now takes ownership of the cloned reference, not the original
err = globTab.Set(vmlib.GoString("_G"), clonedGlobTab)
if err != nil {
    panic("failed to set _G")
}
```

## Benefits over other libraries

### Exception Handling Support

Unlike prior attempts at this such as [golua](https://github.com/aarzilli/golua), gluau has full support for Luau exception handling. This means that you can use Luau's `pcall` and `xpcall` functions to handle errors in your Lua code, and they will work seamlessly with Go's error handling.

gluau achieves this feat by using a Rust proxy layer to actually manage the Lua VM, which allows it to handle exceptions in a way that is compatible with Go's error handling.

## Example Usage

### Simple Table/String Handling Example

```go
    vm, err := vmlib.CreateLuaVm()
    if err != nil {
        fmt.Println("Error creating Lua VM:", err)
        return
    }
    defer vm.Close() // Ensure we close the VM when done

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

    // Assuming tab is a LuaTable created via vm.CreateTable/vm.CreateTableWithCapacity
    tab.ForEach(func(key, value vmlib.Value) error {
        if key.Type() == vmlib.LuaValueString {
            fmt.Println("Key is a LuaString:", key.(*vmlib.ValueString).Value().String())
        }
        if value.Type() == vmlib.LuaValueString {
            fmt.Println("Value is a LuaString:", value.(*vmlib.ValueString).Value().String())
        } else if value.Type() == vmlib.LuaValueInteger {

            fmt.Println("Value is a LuaInteger:", value.(*vmlib.ValueInteger).Value())
        }
        return nil
    })
```

## Windows Note

On Windows, you will need to install MinGW. Also, you may need to copy `libstdc++-6.dll` and `libgcc_s_seh-1.dll` from your MinGW installation to the same directory as your Go executable, or ensure it is in your PATH.