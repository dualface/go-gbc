// MIT License
//
// Copyright (c) 2018 dualface
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
    HandlerFunc func(gbc.RawMessage)

    BasicMessagePipeline struct {
        Name string

        state       *BasicState
        quit        chan int
        inputCh     chan gbc.RawMessage
        wait        sync.WaitGroup
        handlerFunc HandlerFunc
    }
)

func NewBasicMessagePipeline(name string, inputChSize int, handler HandlerFunc) gbc.MessagePipeline {
    if inputChSize <= 0 {
        // unbuffered channel
        inputChSize = 0
    }

    p := &BasicMessagePipeline{
        Name:        name,
        state:       NewBasicState(),
        quit:        make(chan int),
        inputCh:     make(chan gbc.RawMessage, inputChSize),
        wait:        sync.WaitGroup{},
        handlerFunc: handler,
    }

    return p
}

// interface MessagePipeline

func (p *BasicMessagePipeline) Start() (err error) {
    err = p.state.To(Starting)
    if err == nil {
        go p.loop()
    }
    return
}

func (p *BasicMessagePipeline) Stop() {
    if p.state.To(Stopping) == nil {
        p.quit <- 1
    }
}

func (p *BasicMessagePipeline) WaitForComplete() {
    if p.state.To(Idle) == nil {
        p.wait.Wait()
    }
}

func (p *BasicMessagePipeline) OnMessage(m gbc.RawMessage) error {
    if p.state.Current() == Running {
        return p.makeErr("can not handling message")
    }

    // avoid blocking Pipeline
    go func() {
        // goroutine will be blocking, if channel buffer is full
        p.inputCh <- m
    }()

    return nil
}

// private

func (p *BasicMessagePipeline) loop() {
    if p.state.To(Running) != nil {
        return
    }

loop:
    for {
        select {
        case msg := <-p.inputCh:
            p.wait.Add(1)
            go func() {
                p.handlerFunc(msg)
                p.wait.Done()
            }()
        case <-p.quit:
            break loop
        }
    }

    p.wait.Wait()
    p.state.To(Idle)
}

func (p *BasicMessagePipeline) makeErr(msg string) error {
    return fmt.Errorf("[%s][%s], %s", p.Name, p.state, msg)
}
