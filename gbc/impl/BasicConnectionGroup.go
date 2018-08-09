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

type (
    connectionsMap = map[gbc.Connection]bool

    BasicConnectionGroup struct {
        conf        *gbc.ConnectionGroupConfig
        connections connectionsMap
        handler     gbc.RawMessageHandler
        mc          chan gbc.RawMessage
        cc          chan string
        running     bool
        mutex       *sync.Mutex
    }

    ConnectionGroupsMap = map[*BasicConnectionGroup]bool
)

func NewBasicConnectionGroup(conf *gbc.ConnectionGroupConfig) gbc.ConnectionGroup {
    if conf == nil {
        conf = gbc.DefaultConnectionGroupConfig
    }
    g := &BasicConnectionGroup{
        conf:        conf,
        connections: make(connectionsMap, conf.PoolInitSize),
        mc:          make(chan gbc.RawMessage, conf.MessageQueueInitSize),
        cc:          make(chan string),
        running:     true,
        mutex:       &sync.Mutex{},
    }

    go g.loop()

    return g
}

// interface ConnectionGroup

func (g *BasicConnectionGroup) Add(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.connections[c]
    if ok {
        return fmt.Errorf("connection '%p' already exists in group '%p'", &c, g)
    }
    g.connections[c] = true

    // forward message from connection
    c.SetRawMessageReceiver(g.mc)

    return nil
}

func (g *BasicConnectionGroup) Remove(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.connections[c]
    if !ok {
        return fmt.Errorf("not found connection '%p' in group '%p'", &c, g)
    }

    c.SetRawMessageReceiver(nil)

    delete(g.connections, c)
    return nil
}

func (g *BasicConnectionGroup) RemoveAll() {
    g.mutex.Lock()
    m := g.connections
    g.connections = make(connectionsMap, g.conf.PoolInitSize)
    g.mutex.Unlock()

    for c := range m {
        c.SetRawMessageReceiver(nil)
    }
}

func (g *BasicConnectionGroup) CloseAll() {
    if g.running {
        g.cc <- "stop"
    }

    list := g.makeTempConnList()
    g.connections = make(connectionsMap, g.conf.PoolInitSize)

    for _, c := range list {
        c.SetRawMessageReceiver(nil)
    }

    go func() {
        for _, c := range list {
            c.Close()
        }
    }()
}

// broadcast to all list
func (g *BasicConnectionGroup) WriteBytes(b []byte) ([]byte, error) {
    list := g.makeTempConnList()

    for _, c := range list {
        go func() {
            c.WriteBytes(b)
        }()
    }
    return nil, nil
}

func (g *BasicConnectionGroup) SetRawMessageHandler(h gbc.RawMessageHandler) {
    g.handler = h
}

// private

func (g *BasicConnectionGroup) makeTempConnList() []gbc.Connection {
    g.mutex.Lock()
    // copy connections to tmp array
    clist := make([]gbc.Connection, len(g.connections))
    i := 0
    for c := range g.connections {
        clist[i] = c
        i++
    }
    g.mutex.Unlock()

    return clist
}

func (g *BasicConnectionGroup) loop() {

loop:
    for {
        select {
        case m := <-g.mc:
            if g.handler != nil {
                g.handler.WriteRawMessage(m)
            }

        case c := <-g.cc:
            if c == "stop" {
                break loop
            }
        }
    }

    g.running = false
}
