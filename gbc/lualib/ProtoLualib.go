package lualib

import (
    "github.com/golang/protobuf/proto"
    "github.com/yuin/gopher-lua"
    "layeh.com/gopher-luar"
)

func LuaProtoLoader(L *lua.LState) {
    L.PreloadModule("proto", func(L *lua.LState) int {
        p := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
            "Bool":    protoBool,
            "Int":     protoInt,
            "Int32":   protoInt32,
            "Int64":   protoInt64,
            "Uint32":  protoUint32,
            "Uint64":  protoUint64,
            "Float32": protoFloat32,
            "Float64": protoFloat64,
            "String":  protoString,
        })

        L.Push(p)
        return 1
    })
}

// private

func protoBool(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Bool", 1)
    }

    L.Push(luar.New(L, proto.Bool(L.CheckBool(-1))))
    return 1
}

func protoInt(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Int", 1)
    }

    L.Push(luar.New(L, proto.Int(L.CheckInt(-1))))
    return 1
}

func protoInt32(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Int32", 1)
    }

    L.Push(luar.New(L, proto.Int32(int32(L.CheckInt(-1)))))
    return 1
}

func protoInt64(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Int64", 1)
    }

    L.Push(luar.New(L, proto.Int64(L.CheckInt64(-1))))
    return 1
}

func protoUint32(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Uint32", 1)
    }

    L.Push(luar.New(L, proto.Uint32(uint32(L.CheckInt64(-1)))))
    return 1
}

func protoUint64(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Uint64", 1)
    }

    L.Push(luar.New(L, proto.Uint64(uint64(L.CheckInt64(-1)))))
    return 1
}

func protoFloat32(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Float32", 1)
    }

    L.Push(luar.New(L, proto.Float32(float32(L.CheckNumber(-1)))))
    return 1
}

func protoFloat64(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "Float64", 1)
    }

    L.Push(luar.New(L, proto.Float64(float64(L.CheckNumber(-1)))))
    return 1
}

func protoString(L *lua.LState) int {
    if L.GetTop() < 1 {
        raiseInvalidArgumentsError(L, "String", 1)
    }

    L.Push(luar.New(L, proto.String(L.CheckString(-1))))
    return 1
}

func raiseInvalidArgumentsError(L *lua.LState, name string, expected int) {
    L.RaiseError("proto.%s() invalid number of function arguments (%d expected, got %d)", name, expected, L.GetTop())
}
