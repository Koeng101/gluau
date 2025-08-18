use crate::{multivalue::GoMultiValue, result::{GoMultiValueResult, GoThreadResult}};

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_create_thread(ptr: *mut mluau::Lua, func: *mut mluau::Function) -> GoThreadResult {
    if ptr.is_null() || func.is_null() {
        return GoThreadResult::err("Lua pointer or function pointer is null".to_string());
    }

    let lua = unsafe { &mut *ptr };
    let lua_func = unsafe { &*func };

    match lua.create_thread(lua_func.clone()) {
        Ok(thread) => GoThreadResult::ok(Box::into_raw(Box::new(thread))),
        Err(e) => GoThreadResult::err(format!("{e}")),
    }
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_thread_status(t: *mut mluau::Thread) -> u8 {
    // Safety: Assume t is a valid, non-null pointer to a Lua thread
    if t.is_null() {
        return 0; // Return 0 for null pointer
    }

    let lua_t = unsafe { &*t };

    // Get the status of the Lua thread
    match lua_t.status() {
        mluau::ThreadStatus::Resumable => 0,
        mluau::ThreadStatus::Running => 1,
        mluau::ThreadStatus::Finished => 2,
        mluau::ThreadStatus::Error => 3,
    }
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_thread_resume(ptr: *mut mluau::Thread, args: *mut GoMultiValue) -> GoMultiValueResult  {
    if ptr.is_null() {
        return GoMultiValueResult::err("Function pointer is null".to_string());
    }

    let th = unsafe { &*ptr };
    
    // Safety: Go side must ensure values cannot be used after it is set
    // here as a return value
    let values = unsafe { Box::from_raw(args) };
    let values_mv = values.values.into_inner().unwrap();
    let res = th.resume::<mluau::MultiValue>(values_mv);
    match res {
        Ok(mv) => GoMultiValueResult::ok(GoMultiValue::inst(mv)),
        Err(e) => GoMultiValueResult::err(format!("{e}"))
    }
}


#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_thread_to_pointer(t: *mut mluau::Thread) -> usize {
    // Safety: Assume thread is a valid, non-null pointer to a Lua thread
    if t.is_null() {
        return 0;
    }

    let lua_t = unsafe { &*t };

    let ptr = lua_t.to_pointer();

    ptr as usize
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_thread_equals(t: *mut mluau::Thread, t2: *mut mluau::Thread) -> bool {
    if t.is_null() || t2.is_null() {
        return t.is_null() && t2.is_null(); // If both are null, they are equal
    }

    let t1 = unsafe { &*t };
    let t2 = unsafe { &*t2 };

    t1 == t2
}

#[unsafe(no_mangle)]
pub extern "C-unwind" fn luago_free_thread(t: *mut mluau::Thread) {
    // Safety: Assume t is a valid, non-null pointer to a Lua thread
    if t.is_null() {
        return;
    }

    // Re-box the Lua thread to manage its memory automatically.
    unsafe { drop(Box::from_raw(t)) };
}