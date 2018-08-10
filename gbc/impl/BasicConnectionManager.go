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
    "net"
    "strings"
    "sync"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
)

type (
    BasicConnectionManager struct {
        DefaultGroup *BasicConnectionGroup

        onConnectFunc gbc.OnConnectFunc
        groups        map[gbc.ConnectionGroup]bool
        quit          chan int
        mutex         *sync.Mutex
    }
)

func NewBasicConnectionManager() *BasicConnectionManager {
    cm := &BasicConnectionManager{
        DefaultGroup: NewBasicConnectionGroup("incoming", nil),
        mutex:        &sync.Mutex{},
    }
    return cm
}

// interface ConnectionManager

func (cm *BasicConnectionManager) OnConnect(f gbc.OnConnectFunc) {
    cm.onConnectFunc = f
}

func (cm *BasicConnectionManager) Start(l net.Listener) (err error) {
    clog.PrintInfo("listening at: %s", l.Addr().String())

    cm.groups = make(map[gbc.ConnectionGroup]bool)
    cm.quit = make(chan int)

    // handle connect
    cm.DefaultGroup.Start()
    go cm.startAcceptConnect(l)

    // waiting for quit
loop:
    for {
        select {
        case <-cm.quit:
            break loop
        }
    }

    // stop accept new connect
    l.Close()

    // close all groups and clear
    for group := range cm.groups {
        group.Close()
    }
    cm.groups = make(map[gbc.ConnectionGroup]bool)

    clog.PrintInfo("closed")
    return
}

func (cm *BasicConnectionManager) Stop() {
    if cm.quit != nil {
        cm.quit <- 1
        close(cm.quit)
        cm.quit = nil
    }
}

// private

func (cm *BasicConnectionManager) startAcceptConnect(l net.Listener) {
    for {
        rawConn, err := l.Accept()
        if err != nil {
            if !strings.Contains(err.Error(), "use of closed network connection") {
                clog.PrintWarn(err.Error())
            }
            continue
        }

        var conn gbc.Connection
        if cm.onConnectFunc == nil {
            conn = NewBasicConnection(rawConn, NewRawMessageInputFilter())
        } else {
            conn = cm.onConnectFunc(rawConn)
        }
        cm.DefaultGroup.Add(conn)
        conn.Start()
    }
}
