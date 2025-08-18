use std::ffi::c_void;

use mluau::Lua;

use crate::{compiler::CompilerOpts, multivalue::GoMultiValue, result::GoNoneResult, value::ErrorVariant, IGoCallback, IGoCallbackWrapper};

// Base functions

#[unsafe(no_mangle)]
pub extern "C-unwind" fn newluavm() -> *mut mluau::Lua {
    let lua = Lua::new_with(
        mluau::StdLib::ALL_SAFE, // TODO: Allow configuring this
        mluau::LuaOptions::new()
        .catch_rust_panics(false)
        .disable_error_userdata(true)
    ).unwrap(); // Will never error, as we are using safe libraries only.

    lua.set_on_close(|| {
        println!("Lua VM is being closed");
    });

    let wrapper = Box::new(lua);
    Box::into_raw(wrapper)
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luavm_setcompileropts(ptr: *mut mluau::Lua, opts: CompilerOpts) {
    if ptr.is_null() {
        return; // no-op if pointer is null
    }
    let lua = unsafe { &*ptr };
    lua.set_compiler(opts.to_compiler());
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luavm_setmemorylimit(ptr: *mut mluau::Lua, limit: usize) -> GoNoneResult {
    // Safety: Assume the Lua VM is valid and we can set its memory limit.
    if ptr.is_null() {
        return GoNoneResult::err("Lua pointer is null".to_string());
    }
    let lua = unsafe { &*ptr };
    match lua.set_memory_limit(limit) {
        Ok(_) => GoNoneResult::ok(),
        Err(err) => GoNoneResult::err(format!("{err}")),
    }
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luavm_sandbox(ptr: *mut mluau::Lua, enabled: bool) -> GoNoneResult {
    // Safety: Assume the Lua VM is valid and we can set its sandbox mode.
    if ptr.is_null() {
        return GoNoneResult::err("Lua pointer is null".to_string());
    }
    let lua = unsafe { &*ptr };
    match lua.sandbox(enabled) {
        Ok(_) => GoNoneResult::ok(),
        Err(err) => GoNoneResult::err(format!("{err}")),
    }
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_globals(ptr: *mut mluau::Lua) -> *mut mluau::Table {
    // Safety: Assume the Lua VM is valid and we can access its globals.
    if ptr.is_null() {
        return std::ptr::null_mut();
    }
    let lua = unsafe { &*ptr };
    Box::into_raw(Box::new(lua.globals()))
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_setglobals(ptr: *mut mluau::Lua, tab: *mut mluau::Table) -> GoNoneResult {
    // Safety: Assume the Lua VM is valid and we can access its globals.
    if ptr.is_null() {
        return GoNoneResult::err("Lua pointer is null".to_string());
    }
    let lua = unsafe { &*ptr };
    let tab = unsafe { &*tab };
    match lua.set_globals(tab.clone()) {
        Ok(_) => GoNoneResult::ok(),
        Err(err) => GoNoneResult::err(format!("{err}")),
    }
}

#[repr(C)]
pub struct InterruptData {
    // mluau::Lua representing the Lua State
    // as called from Lua.
    //
    // This means that (future) API's like Lua.CurrentThread will return
    // the correct thread when using this Lua.
    pub lua: *mut mluau::Lua,

    // Go side may set this to set a response
    pub vm_state: u8,
    pub error: *mut ErrorVariant,
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_set_interrupt(ptr: *mut mluau::Lua, cb: IGoCallback)  {
    // Safety: Assume ptr is a valid, non-null pointer to a mluau::Lua
    if ptr.is_null() {
        return;
    }

    let cb_wrapper = IGoCallbackWrapper::new(cb);

    let lua = unsafe { &*ptr };
    lua.set_interrupt(move |lua| {
        let wrapper = Box::new(lua.clone());
        let lua_ptr = Box::into_raw(wrapper);
        
        let data = InterruptData {
            lua: lua_ptr,
            vm_state: 0, // Default state (Continue)
            error: std::ptr::null_mut(), // No error by default
        };

        let ptr = Box::into_raw(Box::new(data));
        cb_wrapper.callback(ptr as *mut c_void);
        let data = unsafe { Box::from_raw(ptr) };

        if !data.error.is_null() {
            let error = unsafe { Box::from_raw(data.error) };
            return Err(mluau::Error::external(error.error.to_string_lossy()));
        }
        
        match data.vm_state {
            0 => Ok(mluau::VmState::Continue), // Continue
            1 => Ok(mluau::VmState::Yield),    // Yield
            _ => Err(mluau::Error::external("Invalid VM state".to_string())),
        }
    });
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_remove_interrupt(ptr: *mut mluau::Lua)  {
    // Safety: Assume ptr is a valid, non-null pointer to a Lua
    if ptr.is_null() {
        return;
    }

    let lua = unsafe { &*ptr };
    lua.remove_interrupt();
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_current_thread(ptr: *mut mluau::Lua) -> *mut mluau::Thread {
    // Safety: Assume ptr is a valid, non-null pointer to a Lua
    if ptr.is_null() {
        return std::ptr::null_mut();
    }

    let lua = unsafe { &*ptr };
    Box::into_raw(Box::new(lua.current_thread()))
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_yield_with(ptr: *mut mluau::Lua, args: *mut GoMultiValue) -> GoNoneResult {
    // Safety: Assume ptr is a valid, non-null pointer to a Lua
    if ptr.is_null() {
        return GoNoneResult::err("Lua pointer is null".to_string());
    }

    let lua = unsafe { &*ptr };
    // Safety: Go side must ensure values cannot be used after it is set
    // here as a return value
    let values = unsafe { Box::from_raw(args) };
    let values_mv = values.values.into_inner().unwrap();

    match lua.yield_with(values_mv) {
        Ok(_) => GoNoneResult::ok(),
        Err(err) => GoNoneResult::err(format!("{err}")),
    }
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn freeluavm(ptr: *mut mluau::Lua) {
    // Safety: Assume ptr is a valid, non-null pointer to a mluau::Lua
    // and that ownership is being transferred back to Rust to be dropped.
    unsafe {
        drop(Box::from_raw(ptr));
    }
}