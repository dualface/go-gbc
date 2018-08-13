package lualib

import (
    "reflect"
    "strings"

    "github.com/yuin/gopher-lua"
)

// convert golang value to lua value
func ToLuaValue(L *lua.LState, v interface{}) lua.LValue {
    vt := reflect.TypeOf(v)
    vr := reflect.ValueOf(v)
    return toLuaValueByReflect(L, vt, vr)
}

// private

func toLuaValueByReflect(L *lua.LState, vt reflect.Type, vr reflect.Value) lua.LValue {
    switch vr.Kind() {
    case reflect.Ptr:
        return getPtrV(vr.Type(), vr.Elem(), vr.Elem().Kind(), L)

    case reflect.Array, reflect.Slice:
        table := L.NewTable()
        l := vr.Len()
        for i := 0; i < l; i++ {
            table.Append(toLuaValueByReflect(L, vt, vr.Index(i)))
        }
        return table

    case reflect.Struct:
        t := vr.Type()
        table := L.NewTable()
        l := vr.NumField()
        for i := 0; i < l; i++ {
            vr := vr.Field(i)
            if !vr.CanSet() || strings.Contains(t.Field(i).Name, "XXX_") {
                continue
            }
            table.RawSetString(t.Field(i).Name, toLuaValueByReflect(L, vt, vr))
        }
        return table

    case reflect.Bool:
        return lua.LBool(vr.Bool())

    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return lua.LNumber(vr.Int())

    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
        return lua.LNumber(vr.Uint())

    case reflect.Float32, reflect.Float64:
        return lua.LNumber(vr.Float())

    case reflect.Complex64, reflect.Complex128:
        table := L.NewTable()
        x := vr.Complex()
        table.RawSetString("real", lua.LNumber(real(x)))
        table.RawSetString("imag", lua.LNumber(imag(x)))
        return table
    }

    return lua.LNil
}

// private

func getPtrV(vt reflect.Type, vr reflect.Value, t reflect.Kind, L *lua.LState) lua.LValue {
    switch t {
    case reflect.Bool:
        return lua.LBool(vr.Interface().(bool))
    case reflect.Int:
        return lua.LNumber(vr.Interface().(int))
    case reflect.Int8:
        return lua.LNumber(vr.Interface().(int8))
    case reflect.Int16:
        return lua.LNumber(vr.Interface().(int16))
    case reflect.Int32:
        return lua.LNumber(vr.Interface().(int32))
    case reflect.Int64:
        return lua.LNumber(vr.Interface().(int64))
    case reflect.Uint:
        return lua.LNumber(vr.Interface().(uint))
    case reflect.Uint8:
        return lua.LNumber(vr.Interface().(uint8))
    case reflect.Uint16:
        return lua.LNumber(vr.Interface().(uint16))
    case reflect.Uint32:
        return lua.LNumber(vr.Interface().(uint32))
    case reflect.Uint64:
        return lua.LNumber(vr.Interface().(uint64))
    case reflect.Float32:
        return lua.LNumber(vr.Interface().(float32))
    case reflect.Float64:
        return lua.LNumber(vr.Interface().(float64))
    case reflect.String:
        return lua.LString(vr.Interface().(string))
    case reflect.Uintptr:
    case reflect.Complex64, reflect.Complex128:
        table := L.NewTable()
        x := vr.Complex()
        table.RawSetString("real", lua.LNumber(real(x)))
        table.RawSetString("imag", lua.LNumber(imag(x)))
    case reflect.Array, reflect.Slice:
        table := L.NewTable()
        l := vr.Len()
        for i := 0; i < l; i++ {
            table.Append(toLuaValueByReflect(L, vt, vr.Index(i)))
        }
        return table
    case reflect.Chan:
    case reflect.Func:
    case reflect.Interface:
    case reflect.Map:
    case reflect.Ptr:
        return getPtrV(vr.Type(), vr.Elem(), vr.Elem().Kind(), L)
    case reflect.Struct:
        table := L.NewTable()
        t := vr.Type()
        l := vr.NumField()
        for i := 0; i < l; i++ {
            v := vr.Field(i)
            if !v.CanSet() || strings.Contains(t.Field(i).Name, "XXX_") {
                continue
            }
            table.RawSetString(t.Field(i).Name, toLuaValueByReflect(L, vt, v))
        }
        return table
    }
    return lua.LNil
}
