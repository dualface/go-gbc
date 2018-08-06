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
    "os"
    "os/signal"
    "strings"
    "sync"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
)

type (
    BasicConnectionManager struct {
        name   string
        ch     gbc.OnConnectHandler
        policy gbc.ConnectionGroupPolicy
        groups gbc.ConnectionGroupsMap
        cc     chan string
        mutex  *sync.Mutex
    }
)

func NewBasicConnectionManager(name string, handler gbc.OnConnectHandler, policy gbc.ConnectionGroupPolicy) gbc.ConnectionManager {
    if handler == nil {
        handler = defaultOnConnectHandler
    }
    if policy == nil {
        policy = NewAllInOneConnectionGroupPolicy(nil)
    }
    cm := &BasicConnectionManager{
        name:   name,
        ch:     handler,
        policy: policy,
        groups: make(gbc.ConnectionGroupsMap),
        cc:     make(chan string),
        mutex:  &sync.Mutex{},
    }

    return cm
}

// interface ConnectionManager

func (cm *BasicConnectionManager) Start(l net.Listener) (err error) {
    clog.PrintInfo("[%s] listening at: %s", cm.name, l.Addr().String())

    // handle CTRL+C
    signCh := make(chan os.Signal)
    signal.Notify(signCh, os.Interrupt)
    go func() {
        <-signCh
        // sig is a ^C, handle it
        clog.PrintInfo("[%s] signal os.Interrupt captured", cm.name)
        cm.cc <- "stop"
    }()

    // handle connect
    go cm.startAcceptConnect(l)

    // waiting for connect, command
loop:
    for {
        select {
        case cmd := <-cm.cc:
            cmd = strings.ToLower(cmd)
            clog.PrintInfo("[%s] get command '%s'", cm.name, cmd)
            if cmd == "stop" {
                break loop
            }
        }
    }

    // stop accept new connect
    l.Close()

    // close all connections
    for group := range cm.groups {
        group.CloseAll()
    }

    clog.PrintInfo("[%s] shutdown", cm.name)
    return
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

        // add connection to group
        conn := cm.ch(rawConn)
        cg := cm.policy.GetGroup(conn)
        cg.Add(conn)
        cm.groups[cg] = true

        // start connection message loop
        conn.Start()
    }
}

func defaultOnConnectHandler(rawConn net.Conn) gbc.Connection {
    return NewBasicConnection(rawConn, nil)
}
