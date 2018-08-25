package lualib

import (
    "fmt"
    "path/filepath"
    "reflect"
    "strconv"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
    "github.com/dualface/go-gbc/gbc/impl"
    "github.com/dualface/go-gbc/gbc/protoconv"
    "github.com/yuin/gopher-lua"
    "layeh.com/gopher-luar"
)

type (
    ConcurrenceLuaHandler struct {
        availLuaStates     chan lua.LValue
        luaStates          map[string]*lua.LState
        messageToLuaChan   map[string]chan lua.LValue
        messageFromLuaChan map[string]chan lua.LValue
        luaDir             string
        luaFile            string
    }
)

func NewConcurrenceLuaHandler(concurrence int, luaDir string, luaFile string) *ConcurrenceLuaHandler {
    if concurrence < 1 {
        concurrence = 1
    }

    h := &ConcurrenceLuaHandler{
        availLuaStates:     make(chan lua.LValue, concurrence),
        luaStates:          make(map[string]*lua.LState, concurrence),
        messageToLuaChan:   make(map[string]chan lua.LValue, concurrence),
        messageFromLuaChan: make(map[string]chan lua.LValue, concurrence),
        luaDir:             luaDir,
        luaFile:            luaFile,
    }
    if !filepath.IsAbs(h.luaFile) {
        h.luaFile = filepath.Clean(filepath.Join(h.luaDir, h.luaFile))
    }

    for i := 0; i < concurrence; i++ {
        id := strconv.Itoa(i + 1)
        h.messageToLuaChan[id] = make(chan lua.LValue)
        h.messageFromLuaChan[id] = make(chan lua.LValue)
        h.luaStates[id] = h.createLuaState(id)
    }

    return h
}

func (h *ConcurrenceLuaHandler) RegisterModuleLoader(loader func(*lua.LState)) {
    for _, L := range h.luaStates {
        loader(L)
    }
}

func (h *ConcurrenceLuaHandler) RegisterType(name string, vt interface{}) {
    for _, L := range h.luaStates {
        L.SetGlobal(name, luar.NewType(L, vt))
    }
}

func (h *ConcurrenceLuaHandler) RegisterGlobalVar(name string, v interface{}) {
    for _, L := range h.luaStates {
        L.SetGlobal(name, luar.New(L, v))
    }
}

func (h *ConcurrenceLuaHandler) RegisterGlobalFunc(name string, f lua.LGFunction) {
    for _, L := range h.luaStates {
        L.SetGlobal(name, L.NewFunction(f))
    }
}

func (h *ConcurrenceLuaHandler) Start() {
    for id, L := range h.luaStates {
        L := L
        go func() {
            err := L.DoFile(h.luaFile)
            if err != nil {
                clog.PrintWarn("Lua state %L has failed, %L", id, err.Error())
            }
        }()
    }
}

// interface RawMessageReceiver

func (h *ConcurrenceLuaHandler) ReceiveRawMessage(m gbc.RawMessage) error {
    // avoid blocking caller
    go func() {
        avail := <-h.availLuaStates
        id := avail.String()

        L, ok := h.luaStates[id]
        if !ok {
            clog.PrintError("get invalid Lua worker id: %s", id)
            return
        }

        v, err := h.convertMessageToLuaValue(L, m)
        if err != nil {
            h.availLuaStates <- avail
            clog.PrintWarn(err.Error())
        } else {
            h.messageToLuaChan[id] <- v
        }
    }()
    return nil
}

// private

func (h *ConcurrenceLuaHandler) convertMessageToLuaValue(L *lua.LState, m gbc.RawMessage) (lua.LValue, error) {
    msg, ok := m.(*impl.CommandMessage)
    if !ok {
        return nil, fmt.Errorf("%T only support CommandMessage", h)
    }

    switch msg.DataType() {
    case impl.CommandMessageProtobufType:
        pb, err := protoconv.UnmarshalCommandMessageToProto(msg)
        if err != nil {
            return nil, err
        }
        lv := luar.New(L, pb)
        typeName := lua.LString(reflect.TypeOf(pb).String())
        tb := L.NewTable()
        tb.RawSetString("type", lua.LString(typeName))
        tb.RawSetString("msg", lv)
        return tb, nil

    default:
        return nil, fmt.Errorf("%T not support DataType %d", h, msg.DataType())
    }

    return lua.LNil, nil
}

func (h *ConcurrenceLuaHandler) createLuaState(id string) *lua.LState {
    L := lua.NewState()
    worker := L.NewTable()
    worker.RawSetString("ID", lua.LString(id))
    worker.RawSetString("BOOT_DIR", lua.LString(h.luaDir))
    worker.RawSetString("INPUT_CHAN", lua.LChannel(h.messageToLuaChan[id]))
    worker.RawSetString("OUTPUT_CHAN", lua.LChannel(h.messageFromLuaChan[id]))
    worker.RawSetString("WORKER_CHAN", lua.LChannel(h.availLuaStates))
    L.SetGlobal("WORKER", worker)
    return L
}
