// MIT License
//
// Copyright (conn) 2018 dualface
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package impl

import (
    "fmt"
    "sync"

    "github.com/dualface/go-gbc/gbc"
)

const (
    connectionPoolSize = 10000
)

type (
    connectionsMap = map[gbc.Connection]bool

    BasicConnectionGroup struct {
        Name string

        onRawMessageFunc gbc.OnRawMessageFunc
        connections      connectionsMap
        messageChan      chan gbc.RawMessage
        quit             chan int
        running          bool
        mutex            *sync.Mutex
    }

    ConnectionGroupsMap = map[*BasicConnectionGroup]bool
)

func NewBasicConnectionGroup(name string, messageFunc gbc.OnRawMessageFunc) *BasicConnectionGroup {
    g := &BasicConnectionGroup{
        Name:             name,
        onRawMessageFunc: messageFunc,
        connections:      make(connectionsMap, connectionPoolSize),
        messageChan:      make(chan gbc.RawMessage),
        quit:             make(chan int),
        running:          false,
        mutex:            &sync.Mutex{},
    }

    return g
}

// interface ConnectionGroup

func (g *BasicConnectionGroup) OnRawMessage(f gbc.OnRawMessageFunc) {
    g.onRawMessageFunc = f
}

func (g *BasicConnectionGroup) Start() error {
    if g.running {
        return fmt.Errorf("connection '%s' group is already running", g.Name)
    }

    g.running = true
    go g.loop()
    return nil
}

func (g *BasicConnectionGroup) Close() error {
    if !g.running {
        return fmt.Errorf("connection '%s' group is not running", g.Name)
    }

    g.quit <- 1

    // stop all connections
    g.mutex.Lock()
    defer g.mutex.Unlock()

    for c := range g.connections {
        c.Close()
    }
    g.connections = make(connectionsMap, connectionPoolSize)

    return nil
}

func (g *BasicConnectionGroup) Add(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.connections[c]
    if ok {
        return fmt.Errorf("connection '%p' already exists in group '%p'", &c, g)
    }
    g.connections[c] = true
    c.SetRawMessageChannel(g.messageChan)

    return nil
}

func (g *BasicConnectionGroup) Remove(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.connections[c]
    if !ok {
        return fmt.Errorf("not found connection '%p' in group '%p'", &c, g)
    }

    c.SetRawMessageChannel(nil)
    delete(g.connections, c)
    return nil
}

func (g *BasicConnectionGroup) RawMessageChan() chan gbc.RawMessage {
    return g.messageChan
}

func (g *BasicConnectionGroup) BroadcastWrite(b []byte) {
    g.mutex.Lock()
    list := make([]gbc.Connection, len(g.connections))
    i := 0
    for c := range g.connections {
        list[i] = c
        i++
    }
    g.mutex.Unlock()

    for _, c := range list {
        c := c
        go func() {
            c.Write(b)
        }()
    }
}

// private

func (g *BasicConnectionGroup) loop() {

loop:
    for {
        select {
        case m := <-g.messageChan:
            if g.onRawMessageFunc != nil {
                g.onRawMessageFunc(m)
            }

        case <-g.quit:
            break loop
        }
    }

    g.running = false
}
