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
	"github.com/dualface/go-cli-colorlog"
	"github.com/dualface/go-gbc/gbc"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
)

type (
	BasicConnectionManager struct {
		name  string
		cgp   gbc.ConnectionGroupPolicy
		cgs   gbc.ConnectionGroupsMap
		cc    chan string
		mutex *sync.Mutex
	}
)

func NewBasicConnectionManager(name string) gbc.ConnectionManager {
	cm := &BasicConnectionManager{
		name:  name,
		cgp:   NewAllInOneConnectionGroupPolicy(),
		cgs:   make(gbc.ConnectionGroupsMap),
		cc:    make(chan string),
		mutex: &sync.Mutex{},
	}

	return cm
}

func (cm *BasicConnectionManager) SetConnectionGroupPolicy(p gbc.ConnectionGroupPolicy) {
	cm.cgp = p
}

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

	// waiting for all messages has processed

	// close all connections
	for cg := range cm.cgs {
		cg.CloseAllConnections()
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

		conn := NewBasicConnection(rawConn)
		cg := cm.cgp.GetGroup(conn)
		cm.cgs[cg] = true
		cg.AddConnection(conn)

		// start connection message loop
		conn.Start()
	}
}
