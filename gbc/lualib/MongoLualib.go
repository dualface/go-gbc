package lualib

import (
    "github.com/globalsign/mgo"
    "github.com/yuin/gopher-lua"
    "layeh.com/gopher-luar"
)

func LuaMongoLoader(L *lua.LState) {
    L.PreloadModule("mongodblib", func(L *lua.LState) int {
        mongo := L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
            "Dial": dial,
        })

        mongo.RawSetString("Session", luar.NewType(L, &mgo.Session{}))
        mongo.RawSetString("Database", luar.NewType(L, &mgo.Database{}))
        mongo.RawSetString("Collection", luar.NewType(L, &mgo.Collation{}))
        mongo.RawSetString("Query", luar.NewType(L, &mgo.Query{}))

        L.Push(mongo)
        return 1
    })
}

func dial(L *lua.LState) int {
    if L.GetTop() < 1 {
        L.Push(lua.LNil)
        L.Push(lua.LString("invalid arguments"))
        return 2
    }

    url := L.CheckString(-1)
    session, err := mgo.Dial(url)
    if err != nil {
        L.Push(lua.LNil)
        L.Push(lua.LString(err.Error()))
        return 2
    }

    L.Push(luar.New(L, session))
    return 1
}
