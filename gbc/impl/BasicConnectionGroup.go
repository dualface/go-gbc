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
	"github.com/dualface/go-cli-colorlog"
	"github.com/dualface/go-gbc/gbc"
	"sync"
)

type (
	connectionsMap = map[gbc.Connection]bool

	BasicConnectionGroup struct {
		cmap  connectionsMap
		mc    chan *gbc.RawMessage
		mutex *sync.Mutex
	}

	ConnectionGroupsMap = map[*BasicConnectionGroup]bool
)

func NewBasicConnectionGroup() gbc.ConnectionGroup {
	cg := &BasicConnectionGroup{
		cmap:  make(connectionsMap, gbc.ConnectionsPoolInitSize),
		mc:    make(chan *gbc.RawMessage, gbc.MessageQueueInitSize),
		mutex: &sync.Mutex{},
	}
	return cg
}

// interface ConnectionGroup

func (cg *BasicConnectionGroup) AddConnection(c gbc.Connection) error {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	_, ok := cg.cmap[c]
	if ok {
		return fmt.Errorf("connection '%p' already exists in group '%p'", &c, cg)
	}
	c.SetMessageChan(cg.mc)
	cg.cmap[c] = true
	return nil
}

func (cg *BasicConnectionGroup) RemoveConnection(c gbc.Connection) error {
	cg.mutex.Lock()
	defer cg.mutex.Unlock()

	_, ok := cg.cmap[c]
	if !ok {
		return fmt.Errorf("not found connection '%p' in group '%p'", &c, cg)
	}
	c.SetMessageChan(nil)
	delete(cg.cmap, c)
	return nil
}

func (cg *BasicConnectionGroup) RemoveAllConnections() {
	cg.mutex.Lock()
	cmap := cg.cmap
	cg.cmap = make(connectionsMap, gbc.ConnectionsPoolInitSize)
	cg.mutex.Unlock()

	for c := range cmap {
		c.SetMessageChan(nil)
	}
}

func (cg *BasicConnectionGroup) CloseAllConnections() {
	cg.mutex.Lock()
	cmap := cg.cmap
	cg.cmap = make(connectionsMap, gbc.ConnectionsPoolInitSize)
	cg.mutex.Unlock()

	for c := range cmap {
		c.SetMessageChan(nil)
		err := c.Close()
		if err != nil {
			clog.PrintWarn("close connection failed, %s", err)
		}
	}
}
