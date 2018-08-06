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

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
)

type (
    connectionsMap = map[gbc.Connection]bool

    BasicConnectionGroup struct {
        conf    *gbc.ConnectionGroupConfig
        cmap    connectionsMap
        mc      chan gbc.RawMessage
        cc      chan string
        running bool
        mutex   *sync.Mutex
    }

    ConnectionGroupsMap = map[*BasicConnectionGroup]bool
)

func NewBasicConnectionGroup(conf *gbc.ConnectionGroupConfig) gbc.ConnectionGroup {
    if conf == nil {
        conf = gbc.DefaultConnectionGroupConfig
    }
    g := &BasicConnectionGroup{
        conf:    conf,
        cmap:    make(connectionsMap, conf.PoolInitSize),
        mc:      make(chan gbc.RawMessage, conf.MessageQueueInitSize),
        cc:      make(chan string),
        running: true,
        mutex:   &sync.Mutex{},
    }

    go g.loop()

    return g
}

// interface ConnectionGroup

func (g *BasicConnectionGroup) Add(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.cmap[c]
    if ok {
        return fmt.Errorf("connection '%p' already exists in group '%p'", &c, g)
    }
    g.cmap[c] = true

    // forward message from connection
    c.SetMessageChan(g.mc)

    return nil
}

func (g *BasicConnectionGroup) Remove(c gbc.Connection) error {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    _, ok := g.cmap[c]
    if !ok {
        return fmt.Errorf("not found connection '%p' in group '%p'", &c, g)
    }

    c.SetMessageChan(nil)

    delete(g.cmap, c)
    return nil
}

func (g *BasicConnectionGroup) RemoveAll() {
    g.mutex.Lock()
    cmap := g.cmap
    g.cmap = make(connectionsMap, g.conf.PoolInitSize)
    g.mutex.Unlock()

    for c := range cmap {
        c.SetMessageChan(nil)
    }
}

func (g *BasicConnectionGroup) CloseAll() {
    if g.running {
        g.cc <- "stop"
    }

    clist := g.makeConnList()
    g.cmap = make(connectionsMap, g.conf.PoolInitSize)

    for _, c := range clist {
        c.SetMessageChan(nil)
    }

    go func() {
        for _, c := range clist {
            c.Close()
        }
    }()
}

func (g *BasicConnectionGroup) Broadcast(b []byte) {
    clist := g.makeConnList()

    go func() {
        // write message to all connections
        for _, c := range clist {
            c.Write(b)
        }
    }()
}

func (g *BasicConnectionGroup) makeConnList() []gbc.Connection {
    g.mutex.Lock()
    // copy connections to tmp array
    clist := make([]gbc.Connection, len(g.cmap))
    i := 0
    for c := range g.cmap {
        clist[i] = c
        i++
    }
    g.mutex.Unlock()

    return clist
}

// private

func (g *BasicConnectionGroup) loop() {

loop:
    for {
        select {
        case msg := <-g.mc:
            clog.PrintDebug("%s", msg)

        case c := <-g.cc:
            if c == "stop" {
                break loop
            }
        }
    }

    g.running = false
}
