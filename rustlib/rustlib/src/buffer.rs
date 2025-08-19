use std::ffi::c_char;

use crate::{result::{wrap_failable, GoBufferResult}, string::LuaStringBytes};

#[unsafe(no_mangle)]
pub extern "C" fn luago_create_buffer(ptr: *mut mluau::Lua, s: *const c_char, len: usize) -> GoBufferResult  {
    wrap_failable(|| {
        // Safety: Assume ptr is a valid, non-null pointer to a Lua
        // and that s points to a valid C string of length len.
        let lua = unsafe { &*ptr };

        let res = if s.is_null() {
            // Create an empty string if the pointer is null.
            lua.create_buffer(&[])
        } else {
            let slice = unsafe { std::slice::from_raw_parts(s as *const u8, len) };
            lua.create_buffer(slice)
        };

        match res {
            Ok(buf) => GoBufferResult::ok(Box::into_raw(Box::new(buf))),
            Err(err) => crate::result::GoBufferResult::err(format!("{err}"))
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_to_pointer(buf: *mut mluau::Buffer) -> usize {
    wrap_failable(|| {
        // Safety: Assume string is a valid, non-null pointer to a Lua String
        if buf.is_null() {
            return 0;
        }

        let buf = unsafe { &*buf };

        let ptr = buf.to_pointer();

        ptr as usize
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_equals(t: *mut mluau::Buffer, t2: *mut mluau::Buffer) -> bool {
    wrap_failable(|| {
        if t.is_null() || t2.is_null() {
            return t.is_null() && t2.is_null(); // If both are null, they are equal
        }

        let t1 = unsafe { &*t };
        let t2 = unsafe { &*t2 };

        t1 == t2
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_len(t: *mut mluau::Buffer) -> usize {
    wrap_failable(|| {
        if t.is_null() {
            return 0
        }

        let t1 = unsafe { &*t };

        t1.len()
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_to_bytes(buf: *mut mluau::Buffer) -> LuaStringBytes {
    wrap_failable(|| {
        // Safety: Assume string is a valid, non-null pointer to a Lua String
        if buf.is_null() {
            return LuaStringBytes {
                data: std::ptr::null(),
                size: 0,
            };
        }

        let lua_string = unsafe { &*buf };
        
        // Return a pointer to the bytes of the Lua String.
        let bytes = lua_string.to_vec().into_boxed_slice();
        let bytes_ptr = bytes.as_ptr();
        let bytes_len = bytes.len();
        std::mem::forget(bytes); // Prevent deallocation of the bytes
        LuaStringBytes {
            data: bytes_ptr,
            size: bytes_len,
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_read_bytes(buf: *mut mluau::Buffer, offset: usize, len: usize) -> LuaStringBytes {
    wrap_failable(|| {
        // Safety: Assume string is a valid, non-null pointer to a Lua String
        if buf.is_null() {
            return LuaStringBytes {
                data: std::ptr::null(),
                size: 0,
            };
        }

        let lua_string = unsafe { &*buf };
        
        // Return a pointer to the bytes of the Lua String.
        let bytes = lua_string.read_bytes_to_vec(offset, len).into_boxed_slice();
        let bytes_ptr = bytes.as_ptr();
        let bytes_len = bytes.len();
        std::mem::forget(bytes); // Prevent deallocation of the bytes
        LuaStringBytes {
            data: bytes_ptr,
            size: bytes_len,
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_write_bytes(buf: *mut mluau::Buffer, offset: usize, data: *const c_char, len: usize) {
    wrap_failable(|| {
        let lua_string = unsafe { &*buf };
        let slice = unsafe { std::slice::from_raw_parts(data as *const u8, len) };
        lua_string.write_bytes(offset, slice);
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_buffer_free_bytes(buf: LuaStringBytes) {
    wrap_failable(|| {
        if buf.data.is_null() {
            return; // Nothing to free
        }

        let s = std::ptr::slice_from_raw_parts_mut(buf.data as *mut u8, buf.size);
        unsafe {
            drop(Box::from_raw(s));
        }
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn luago_free_buffer(buf: *mut mluau::Buffer) {
    wrap_failable(|| {
        // Safety: Assume buf is a valid, non-null pointer to a Lua buffer
        if buf.is_null() {
            return;
        }

        // Re-box the Lua buffer pointer to manage its memory automatically.
        unsafe { drop(Box::from_raw(buf)) };
    })
}