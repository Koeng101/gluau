use std::ffi::{c_char, c_void, CString};

use crate::{multivalue::GoMultiValue, result::{wrap_failable, GoBoolResult, GoFunctionResult, GoMultiValueResult}, IGoCallback, IGoCallbackWrapper};

#[repr(C)]
// NOTE: Aside from the Lua, Rust will deallocate everything
pub struct FunctionCallbackData {
    // Lua representing the Lua State
    // as called from Lua.
    //
    // This means that (future) API's like Lua.CurrentThread will return
    // the correct thread when using this Lua.
    pub lua: *mut mluau::Lua,
    // Arguments passed to the function by Lua
    pub args: *mut GoMultiValue,

    // Go side may set this to set a response
    pub values: *mut GoMultiValue,
    pub error: *mut c_char,
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_create_function(ptr: *mut mluau::Lua, cb: IGoCallback) -> GoFunctionResult  {
    wrap_failable(|| {
        // Safety: Assume ptr is a valid, non-null pointer to a Lua
        if ptr.is_null() {
            return GoFunctionResult::err("Lua pointer is null".to_string());
        }

        let cb_wrapper = IGoCallbackWrapper::new(cb);

        let lua = unsafe { &*ptr };
        let func = lua.create_function(move |lua, args: mluau::MultiValue| {
            let wrapper = Box::new(lua.clone());
            let lua_ptr = Box::into_raw(wrapper);
            
            let data = FunctionCallbackData {
                lua: lua_ptr,
                args: GoMultiValue::inst(args),
                values: std::ptr::null_mut(),
                error: std::ptr::null_mut(),
            };

            let ptr = Box::into_raw(Box::new(data));
            cb_wrapper.callback(ptr as *mut c_void);
            let data = unsafe { Box::from_raw(ptr) };
            unsafe { drop(Box::from_raw(data.args)) }
            
            if !data.error.is_null() {
                if !data.values.is_null() {
                    // Avoid a memory leak by deallocating it
                    unsafe { drop(Box::from_raw(data.values)) };
                }

                // Safety: Go must not use the error after this point
                // as it is deallocated here.
                let error = unsafe { CString::from_raw(data.error) };
                return Err(mluau::Error::external(error.to_string_lossy()));
            } else {
                // If values is set, return them to Lua.
                if !data.values.is_null() {
                    // Safety: Go side must ensure values cannot be used after it is set
                    // here as a return value
                    let values = unsafe { Box::from_raw(data.values) };
                    let values_mv = values.values.into_inner().unwrap();
                    return Ok(values_mv);
                } else {
                    // If no values are set, return an empty MultiValue.
                    return Ok(mluau::MultiValue::new());
                }
            }
        });

        match func {
            Ok(f) => GoFunctionResult::ok(Box::into_raw(Box::new(f))),
            Err(err) => GoFunctionResult::err(format!("{err}")),
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_call(ptr: *mut mluau::Function, args: *mut GoMultiValue) -> GoMultiValueResult  {
    wrap_failable(|| {
        if ptr.is_null() {
            return GoMultiValueResult::err("Function pointer is null".to_string());
        }

        let func = unsafe { &*ptr };
        
        // Safety: Go side must ensure values cannot be used after it is set
        // here as a return value
        let values = unsafe { Box::from_raw(args) };
        let values_mv = values.values.into_inner().unwrap();
        let res = func.call::<mluau::MultiValue>(values_mv);
        match res {
            Ok(mv) => GoMultiValueResult::ok(GoMultiValue::inst(mv)),
            Err(e) => GoMultiValueResult::err(format!("{e}"))
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_deepclone(f: *mut mluau::Function) -> GoFunctionResult {
    wrap_failable(|| {
        if f.is_null() {
            return GoFunctionResult::err("LuaFunction pointer is null".to_string());
        }

        let lua_f = unsafe { &*f };

        let cloned_fn = lua_f.deep_clone();

        match cloned_fn {
            Ok(func) => GoFunctionResult::ok(Box::into_raw(Box::new(func))),
            Err(e) => GoFunctionResult::err(format!("{e}")),
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_environment(f: *mut mluau::Function) -> *mut mluau::Table {
    wrap_failable(|| {
        if f.is_null() {
            return std::ptr::null_mut();
        }

        let lua_f = unsafe { &*f };

        let env = lua_f.environment();

        match env {
            Some(table) => Box::into_raw(Box::new(table)),
            None => std::ptr::null_mut(),
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_set_environment(f: *mut mluau::Function, table: *mut mluau::Table) -> GoBoolResult {
    wrap_failable(|| {
        if f.is_null() {
            return GoBoolResult::err("LuaFunction pointer is null".to_string());
        }

        let lua_f = unsafe { &*f };
        let table = unsafe { &*table };

        let res = lua_f.set_environment(table.clone());

        match res {
            Ok(res) => GoBoolResult::ok(res),
            Err(e) => GoBoolResult::err(format!("{e}")),
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_to_pointer(f: *mut mluau::Function) -> usize {
    wrap_failable(|| {
        // Safety: Assume function is a valid, non-null pointer to a Lua function
        if f.is_null() {
            return 0;
        }

        let lua_f = unsafe { &*f };

        let ptr = lua_f.to_pointer();

        ptr as usize
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_function_equals(f: *mut mluau::Function, f2: *mut mluau::Function) -> bool {
    wrap_failable(|| {
        // Safety: Assume table is a valid, non-null pointer to a Lua Table
        if f.is_null() || f2.is_null() {
            return f.is_null() && f2.is_null(); // If both are null, they are equal
        }

        let f1 = unsafe { &*f };
        let f2 = unsafe { &*f2 };

        f1 == f2
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_free_function(f: *mut mluau::Function) {
    wrap_failable(|| {
        // Safety: Assume function is a valid, non-null pointer to a Lua function
        if f.is_null() {
            return;
        }

        // Re-box the Lua function pointer to manage its memory automatically.
        unsafe { drop(Box::from_raw(f)) };
    })
}