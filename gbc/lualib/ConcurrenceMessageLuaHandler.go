package lualib

import (
    "fmt"
    "path/filepath"
    "strconv"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
    "github.com/dualface/go-gbc/gbc/impl"
    "github.com/dualface/go-gbc/gbc/protoconv"
    "github.com/yuin/gopher-lua"
)

type (
    ConcurrenceLuaMessageHandler struct {
        availLuaStates chan lua.LValue
        luaStates      map[string]*lua.LState
        messageChan    map[string]chan lua.LValue
        luaDir         string
        luaFile        string
    }
)

func NewConcurrenceLuaMessageHandler(concurrence int, luaDir string, luaFile string) *ConcurrenceLuaMessageHandler {
    if concurrence < 1 {
        concurrence = 1
    }

    r := &ConcurrenceLuaMessageHandler{
        availLuaStates: make(chan lua.LValue, concurrence),
        luaStates:      make(map[string]*lua.LState, concurrence),
        messageChan:    make(map[string]chan lua.LValue, concurrence),
        luaDir:         luaDir,
        luaFile:        luaFile,
    }
    if !filepath.IsAbs(r.luaFile) {
        r.luaFile = filepath.Clean(filepath.Join(r.luaDir, r.luaFile))
    }

    for i := 0; i < concurrence; i++ {
        id := strconv.Itoa(i + 1)
        r.messageChan[id] = make(chan lua.LValue)
        r.luaStates[id] = r.createLuaState(id, r.messageChan[id])
    }

    return r
}

// interface MessagePipeline

func (r *ConcurrenceLuaMessageHandler) ReceiveRawMessage(m gbc.RawMessage) error {
    // avoid blocking caller
    go func() {
        avail := <-r.availLuaStates
        id := avail.String()

        L, ok := r.luaStates[id]
        if !ok {
            clog.PrintError("get invalid Lua worker id: %s", id)
            return
        }

        v, err := r.convertMessageToLuaValue(L, m)
        if err != nil {
            r.availLuaStates <- avail
            clog.PrintWarn(err.Error())
        } else {
            r.messageChan[id] <- v
        }
    }()
    return nil
}

// private

func (r *ConcurrenceLuaMessageHandler) convertMessageToLuaValue(L *lua.LState, m gbc.RawMessage) (lua.LValue, error) {
    msg, ok := m.(*impl.CommandMessage)
    if !ok {
        return nil, fmt.Errorf("%T only support CommandMessage", r)
    }

    switch msg.DataType() {
    case impl.CommandMessageProtobufType:
        pb, err := protoconv.UnmarshalCommandMessageToProto(msg)
        if err != nil {
            return nil, err
        }
        lv := ToLuaValue(L, pb)
        return lv, nil

    case impl.CommandMessageClangType:
        return nil, fmt.Errorf("%T not support DataType %d", r, msg.DataType())
    default:
        return nil, fmt.Errorf("%T not support DataType %d", r, msg.DataType())
    }

    return lua.LNil, nil
}

func (r *ConcurrenceLuaMessageHandler) createLuaState(id string, mc chan lua.LValue) *lua.LState {
    s := lua.NewState()
    table := s.NewTable()
    table.RawSetString("ID", lua.LString(id))
    table.RawSetString("BOOT_DIR", lua.LString(r.luaDir))
    table.RawSetString("MESSAGE_CHAN", lua.LChannel(mc))
    table.RawSetString("WORKER_CHAN", lua.LChannel(r.availLuaStates))
    s.SetGlobal("WORKER", table)

    go func() {
        err := s.DoFile(r.luaFile)
        if err != nil {
            clog.PrintWarn(err.Error())
        }
    }()

    return s
}
