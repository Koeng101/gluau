use std::{ffi::{c_char, CString}, panic::AssertUnwindSafe};

use crate::{multivalue::GoMultiValue, value::GoLuaValue};

pub trait Errorable {
    fn error_variant(s: String) -> Self;
}

#[repr(C)]
pub struct GoNoneResult {
    error: *mut c_char
}

impl GoNoneResult {
    pub fn ok() -> Self {
        Self {
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoNoneResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_none_result_free(ptr: *mut GoNoneResult) {
    if ptr.is_null() {
        return;
    }

    unsafe { drop(Box::from_raw(ptr)); }
}

#[repr(C)]
pub struct GoBoolResult {
    value: bool,
    error: *mut c_char
}

impl GoBoolResult {
    pub fn ok(b: bool) -> Self {
        Self {
            value: b,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: false,
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoBoolResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoI64Result {
    value: i64,
    error: *mut c_char
}

impl GoI64Result {
    pub fn ok(v: i64) -> Self {
        Self {
            value: v,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: 0,
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoI64Result {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoUsizePtrResult {
    value: usize,
    error: *mut c_char
}

impl GoUsizePtrResult {
    pub fn ok(v: usize) -> Self {
        Self {
            value: v,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: 0,
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoUsizePtrResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoStringResult {
    value: *mut mluau::String,
    error: *mut c_char
}

impl GoStringResult {
    pub fn ok(s: *mut mluau::String) -> Self {
        Self {
            value: s,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoStringResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoTableResult {
    value: *mut mluau::Table,
    error: *mut c_char
}

impl GoTableResult {
    pub fn ok(t: *mut mluau::Table) -> Self {
        Self {
            value: t,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoTableResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoFunctionResult {
    value: *mut mluau::Function,
    error: *mut c_char
}

impl GoFunctionResult {
    pub fn ok(f: *mut mluau::Function) -> Self {
        Self {
            value: f,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoFunctionResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoAnyUserDataResult {
    value: *mut mluau::AnyUserData,
    error: *mut c_char
}

impl GoAnyUserDataResult {
    pub fn ok(f: *mut mluau::AnyUserData) -> Self {
        Self {
            value: f,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoAnyUserDataResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoMultiValueResult {
    value: *mut GoMultiValue,
    error: *mut c_char
}

impl GoMultiValueResult {
    pub fn ok(f: *mut GoMultiValue) -> Self {
        Self {
            value: f,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoMultiValueResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoThreadResult {
    value: *mut mluau::Thread,
    error: *mut c_char
}

impl GoThreadResult {
    pub fn ok(t: *mut mluau::Thread) -> Self {
        Self {
            value: t,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: std::ptr::null_mut(),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoThreadResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

#[repr(C)]
pub struct GoValueResult {
    value: GoLuaValue,
    error: *mut c_char
}

impl GoValueResult {
    pub fn ok(v: GoLuaValue) -> Self {
        Self {
            value: v,
            error: std::ptr::null_mut(),
        }
    }

    pub fn err(error: String) -> Self {
        Self {
            value: GoLuaValue::from_owned(mluau::Value::Nil),
            error: to_c_string(error),
        }
    }
}

impl Errorable for GoValueResult {
    fn error_variant(s: String) -> Self {
        Self::err(s)
    }
}

/// Given a error string, return a heap allocated error
/// 
/// Useful for API's which have no return
pub fn to_c_string(error: String) -> *mut c_char {
    let error_str = error.replace('\0', ""); // Ensure no null characters in the string
    let error_cstr = CString::new(error_str).unwrap_or_else(|_| CString::new("Invalid error string").unwrap());
    CString::into_raw(error_cstr)
}

// Creates a new CString given string and length
#[unsafe(no_mangle)]
pub extern "C" fn luago_string_new(s: *const c_char, len: usize) -> *mut c_char {
    if s.is_null() || len == 0 {
        let c_string = CString::new("").unwrap_or_else(|_| CString::new("Invalid string").unwrap());
        return CString::into_raw(c_string);
    }
    // Safety: Assume s points to a valid C string of length len.
    let slice = unsafe { std::slice::from_raw_parts(s as *const u8, len) };
    let c_string = CString::new(slice).unwrap_or_else(|_| CString::new("Invalid string").unwrap());
    // Convert CString to raw pointer
    CString::into_raw(c_string)
}

/// Frees the memory for an error string created by Rust.
#[unsafe(no_mangle)]
pub extern "C" fn luago_string_free(error_ptr: *mut c_char) {
    if !error_ptr.is_null() {
        // Reconstruct the CString from the raw pointer and let it drop,
        // which deallocates the memory.
        unsafe { drop(CString::from_raw(error_ptr)); }
    }
}

/// Helper to wrap a Errorable in a catch_unwind
pub fn wrap_failable<T: Errorable>(f: impl FnOnce() -> T) -> T {
    match std::panic::catch_unwind(AssertUnwindSafe(|| f())) {
        Ok(t) => t,
        Err(e) => {
            if let Some(s) = e.downcast_ref::<&str>() {
                T::error_variant(s.to_string())
            } else if let Some(s) = e.downcast_ref::<String>() {
                T::error_variant(s.to_string())
            } else {
                T::error_variant("unknown panic reason".to_string())
            }
        }
    }
}

impl Errorable for () {
    fn error_variant(_s: String) -> Self {
        ()
    }
}

impl Errorable for u8 {
    fn error_variant(_s: String) -> Self {
        0
    }
}

impl Errorable for usize {
    fn error_variant(_s: String) -> Self {
        0
    }
}

impl Errorable for bool {
    fn error_variant(_s: String) -> Self {
        false
    }
}


impl<T> Errorable for *mut T {
    fn error_variant(_s: String) -> Self {
        std::ptr::null_mut()
    }
}

impl Errorable for GoLuaValue {
    fn error_variant(_s: String) -> Self {
        GoLuaValue::from_owned(mluau::Value::Nil)
    }
}