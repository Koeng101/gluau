use std::ffi::c_void;

use crate::result::wrap_failable;

#[repr(C)]
pub enum LuaValueType {
    Nil = 0,
    Boolean = 1,
    LightUserData = 2,
    Integer = 3,
    Number = 4,
    Vector = 5,
    String = 6,
    Table = 7,
    Function = 8,
    Thread = 9,
    UserData = 10,
    Buffer = 11,
    Other = 12,
}

impl LuaValueType {
    pub fn from_value(value: &mluau::Value) -> Self {
        match value {
            mluau::Value::Nil => LuaValueType::Nil,
            mluau::Value::Boolean(_) => LuaValueType::Boolean,
            mluau::Value::LightUserData(_) => LuaValueType::LightUserData,
            mluau::Value::Integer(_) => LuaValueType::Integer,
            mluau::Value::Number(_) => LuaValueType::Number,
            mluau::Value::Vector(_) => LuaValueType::Vector,
            mluau::Value::String(_) => LuaValueType::String,
            mluau::Value::Table(_) => LuaValueType::Table,
            mluau::Value::Function(_) => LuaValueType::Function,
            mluau::Value::Thread(_) => LuaValueType::Thread,
            mluau::Value::UserData(_) => LuaValueType::UserData,
            mluau::Value::Buffer(_) => LuaValueType::Buffer,
            mluau::Value::Error(_) => LuaValueType::Other,
            mluau::Value::Other(_) => LuaValueType::Other, // TODO: Handle other types
        }
    }
}

#[repr(C)]
#[derive(Clone, Copy)]
pub union LuaValueData {
    boolean: bool,
    light_userdata: *mut c_void,
    integer: i64,
    number: f64,
    vector: [f32; 3], 
    string: *mut mluau::String,
    table: *mut mluau::Table,
    function: *mut mluau::Function,
    thread: *mut mluau::Thread,
    userdata: *mut mluau::AnyUserData,
    buffer: *mut mluau::Buffer,
    other: *mut c_void, // Placeholder for other types
}

#[repr(C)]
pub struct GoLuaValue {
    tag: LuaValueType,
    data: LuaValueData,
}

// Safety note:
//
// Previous versions of this included methods that converted between references
// of GoLuaValue and mluau::Value. This is annoying as it could lead to memory leaks
//
// Instead, the current code always takes ownership of the GoLuaValue and converts it to 
// a mluau::Value.
impl GoLuaValue {
    // Clones the GoLuaValue
    pub fn clone(&self) -> Self {
        match self.tag {
            // Primitives, no cloning is actually needed.
            LuaValueType::Nil => GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } },
            LuaValueType::Boolean => GoLuaValue { tag: LuaValueType::Boolean, data: LuaValueData { boolean: unsafe { self.data.boolean } } },
            LuaValueType::LightUserData => GoLuaValue { tag: LuaValueType::LightUserData, data: LuaValueData { light_userdata: unsafe { self.data.light_userdata } } },
            LuaValueType::Integer => GoLuaValue { tag: LuaValueType::Integer, data: LuaValueData { integer: unsafe { self.data.integer } } },
            LuaValueType::Number => GoLuaValue { tag: LuaValueType::Number, data: LuaValueData { number: unsafe { self.data.number } } },
            LuaValueType::Vector => GoLuaValue { tag: LuaValueType::Vector, data: LuaValueData { vector: unsafe { self.data.vector } } },
            // Complex types, we need to clone to increment the refcount internally
            LuaValueType::String => {
                let string_ptr = unsafe { self.data.string };
                if string_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    // Safety: Avoid free'ing the string pointer here, as it is managed by Go
                    let string_ptr = unsafe { &*string_ptr };
                    GoLuaValue { tag: LuaValueType::String, data: LuaValueData { string: Box::into_raw(Box::new(string_ptr.clone())) } }
                }
            },
            LuaValueType::Table => {
                let table_ptr = unsafe { self.data.table };
                if table_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    // Safety: Avoid free'ing the table pointer here, as it is managed by Go
                    let table_ptr = unsafe { &*table_ptr };
                    GoLuaValue { tag: LuaValueType::Table, data: LuaValueData { table: Box::into_raw(Box::new(table_ptr.clone())) } }
                }
            }
            LuaValueType::Function => {
                let function_ptr = unsafe { self.data.function };
                if function_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    // Safety: Avoid free'ing the function pointer here, as it is managed by Go
                    let function_ptr = unsafe { &*function_ptr };
                    GoLuaValue { tag: LuaValueType::Function, data: LuaValueData { function: Box::into_raw(Box::new(function_ptr.clone())) } }
                }
            },
            LuaValueType::Thread => {
                let thread_ptr = unsafe { self.data.thread };
                if thread_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    // Safety: Avoid free'ing the thread pointer here, as it is managed by Go
                    let thread_ptr = unsafe { &*thread_ptr };
                    GoLuaValue { tag: LuaValueType::Thread, data: LuaValueData { thread: Box::into_raw(Box::new(thread_ptr.clone())) } }
                }
            },
            LuaValueType::UserData => {
                let userdata_ptr = unsafe { self.data.userdata };
                if userdata_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    // Safety: Avoid free'ing the userdata pointer here, as it is managed by Go
                    let userdata_ptr = unsafe { &*userdata_ptr };
                    GoLuaValue { tag: LuaValueType::UserData, data: LuaValueData { userdata: Box::into_raw(Box::new(userdata_ptr.clone())) } }
                }
            },
            LuaValueType::Buffer => {
                let buffer_ptr = unsafe { self.data.buffer };
                if buffer_ptr.is_null() {
                    GoLuaValue { tag: LuaValueType::Nil, data: LuaValueData { boolean: false } }
                } else {
                    let buffer_ptr = unsafe { &*buffer_ptr };
                    GoLuaValue { tag: LuaValueType::Buffer, data: LuaValueData { buffer: Box::into_raw(Box::new(buffer_ptr.clone())) } }
                }
            },
            // TODO: Support this better later on
            LuaValueType::Other => GoLuaValue { tag: LuaValueType::Other, data: LuaValueData { boolean: false } },
        }
    }

    /// Converts a GoLuaValue to a mluau::Value.
    /// # Safety
    /// This function destroys the GoLuaValue and transfers ownership of the data to mluau::Value.
    /// Use of clone may be needed
    pub fn to_value_from_owned(self) -> mluau::Value {
        match self.tag {
            LuaValueType::Nil => mluau::Value::Nil,
            LuaValueType::Boolean => mluau::Value::Boolean(unsafe { self.data.boolean }),
            LuaValueType::LightUserData => mluau::Value::LightUserData(mluau::LightUserData(unsafe { self.data.light_userdata })),
            LuaValueType::Integer => mluau::Value::Integer(unsafe { self.data.integer }),
            LuaValueType::Number => mluau::Value::Number(unsafe { self.data.number }),
            LuaValueType::Vector => mluau::Value::Vector(unsafe { mluau::Vector::new(self.data.vector[0], self.data.vector[1], self.data.vector[2]) }),
            LuaValueType::String => {
                let string_ptr = unsafe { self.data.string };
                if string_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let string_ptr = unsafe { Box::from_raw(string_ptr) };
                    mluau::Value::String(*string_ptr)
                }
            },
            LuaValueType::Table => {
                let table_ptr = unsafe { self.data.table };
                if table_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let table_ptr = unsafe { Box::from_raw(table_ptr) };
                    mluau::Value::Table(*table_ptr)
                }
            },
            LuaValueType::Function => {
                let function_ptr = unsafe { self.data.function };
                if function_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let function_ptr = unsafe { Box::from_raw(function_ptr) };
                    mluau::Value::Function(*function_ptr)
                }
            },
            LuaValueType::Thread => {
                let thread_ptr = unsafe { self.data.thread };
                if thread_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let thread_ptr = unsafe { Box::from_raw(thread_ptr) };
                    mluau::Value::Thread(*thread_ptr)
                }
            },
            LuaValueType::UserData => {
                let userdata_ptr = unsafe { self.data.userdata };
                if userdata_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let userdata_ptr = unsafe { Box::from_raw(userdata_ptr) };
                    mluau::Value::UserData(*userdata_ptr)
                }
            },
            LuaValueType::Buffer => {
                let buffer_ptr = unsafe { self.data.buffer };
                if buffer_ptr.is_null() {
                    mluau::Value::Nil
                } else {
                    let buffer_ptr = unsafe { Box::from_raw(buffer_ptr) };
                    mluau::Value::Buffer(*buffer_ptr)
                }
            },
            LuaValueType::Other => {
                // Handle other types, currently returning Nil
                mluau::Value::Nil
            },
        }
    }

    /// Drops a GoLuaValue.
    /// # Safety
    /// This function destroys the GoLuaValue
    pub fn drop_owned(self) {
        match self.tag {
            LuaValueType::Nil | LuaValueType::Boolean | LuaValueType::LightUserData | LuaValueType::Integer | LuaValueType::Number | LuaValueType::Vector => {
                // No-op for now
            },
            LuaValueType::String => {
                let string_ptr = unsafe { self.data.string };
                if !string_ptr.is_null() {
                    // Safety: Avoid free'ing the string pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(string_ptr)) };
                }
            },
            LuaValueType::Table => {
                let table_ptr = unsafe { self.data.table };
                if !table_ptr.is_null() {
                    // Safety: Avoid free'ing the table pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(table_ptr)) };
                }
            },
            LuaValueType::Function => {
                let function_ptr = unsafe { self.data.function };
                if !function_ptr.is_null() {
                    // Safety: Avoid free'ing the function pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(function_ptr)) };
                }
            },
            LuaValueType::Thread => {
                let thread_ptr = unsafe { self.data.thread };
                if !thread_ptr.is_null() {
                    // Safety: Avoid free'ing the thread pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(thread_ptr)) };
                }
            },
            LuaValueType::UserData => {
                let userdata_ptr = unsafe { self.data.userdata };
                if !userdata_ptr.is_null() {
                    // Safety: Avoid free'ing the userdata pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(userdata_ptr)) };
                }
            },
            LuaValueType::Buffer => {
                let buffer_ptr = unsafe { self.data.buffer };
                if !buffer_ptr.is_null() {
                    // Safety: Avoid free'ing the buffer pointer here, as it is managed by Go
                    unsafe { drop(Box::from_raw(buffer_ptr)) };
                }
            },
            LuaValueType::Other => {
                // Handle other types, currently no-op
            },
        }
    }

    pub fn from_owned(value: mluau::Value) -> Self {
        let tag = LuaValueType::from_value(&value);
        let data = match value {
            mluau::Value::Nil => LuaValueData { boolean: false },
            mluau::Value::Boolean(b) => LuaValueData { boolean: b },
            mluau::Value::LightUserData(ptr) => LuaValueData { light_userdata: ptr.0 },
            mluau::Value::Integer(i) => LuaValueData { integer: i },
            mluau::Value::Number(n) => LuaValueData { number: n },
            mluau::Value::Vector(v) => LuaValueData { vector: [v.x(), v.y(), v.z()] },
            mluau::Value::String(s) => LuaValueData { string: Box::into_raw(Box::new(s)) },
            mluau::Value::Table(t) => LuaValueData { table: Box::into_raw(Box::new(t)) },
            mluau::Value::Function(f) => LuaValueData { function: Box::into_raw(Box::new(f)) },
            mluau::Value::Thread(t) => LuaValueData { thread: Box::into_raw(Box::new(t)) },
            mluau::Value::UserData(ud) => LuaValueData { userdata: Box::into_raw(Box::new(ud)) },
            mluau::Value::Buffer(buf) => LuaValueData { buffer: Box::into_raw(Box::new(buf)) },
            mluau::Value::Error(_) => LuaValueData { other: std::ptr::null_mut() }, // This variant will never actually happen as disable_error_userdata is set
            mluau::Value::Other(_) => LuaValueData { other: std::ptr::null_mut() }, // TODO: Handle other types
        };
        GoLuaValue { tag, data }
    }
}

// Clones a GoLuaValue
#[unsafe(no_mangle)]
pub extern "C" fn luago_value_clone(value: GoLuaValue) -> GoLuaValue {
    wrap_failable(|| {
        let cloned_value = value.clone();
        cloned_value
    })
}

// Destroys a GoLuaValue
#[unsafe(no_mangle)]
pub extern "C" fn luago_value_destroy(value: GoLuaValue) {
    wrap_failable(|| {
        value.drop_owned();
    })
}
