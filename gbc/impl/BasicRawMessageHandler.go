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

    BasicRawMessageHandler struct {
        state       *BasicState
        quit        chan int
        input       chan gbc.RawMessage
        semaphore   chan int
        wait        sync.WaitGroup
        handlerFunc HandlerFunc
    }
)

func NewBasicRawMessageHandler(concurrence int, handler HandlerFunc) gbc.RawMessageHandler {
    if concurrence <= 0 {
        // unbuffered channel
        concurrence = 0
    }

    h := &BasicRawMessageHandler{
        state:       NewBasicState(),
        quit:        make(chan int),
        input:       make(chan gbc.RawMessage),
        semaphore:   make(chan int, concurrence),
        wait:        sync.WaitGroup{},
        handlerFunc: handler,
    }

    h.state.To(Starting)
    go h.loop()

    return h
}

// interface MessagePipeline

func (h *BasicRawMessageHandler) Stop() {
    if h.state.To(Stopping) {
        h.quit <- 1
    }
}

func (h *BasicRawMessageHandler) WaitForComplete() {
    if h.state.To(Idle) {
        h.wait.Wait()
    }
}

func (h *BasicRawMessageHandler) WriteRawMessage(m gbc.RawMessage) error {
    if h.state.Current() != Running {
        return fmt.Errorf("'%T' current is not running", h)
    }

    // avoid blocking caller
    go func() {
        h.input <- m
    }()

    return nil
}

// private

func (h *BasicRawMessageHandler) loop() {
    if !h.state.To(Running) {
        return
    }

loop:
    for {
        select {
        case msg := <-h.input:
            h.wait.Add(1)
            go func() {
                h.semaphore <- 1
                h.handlerFunc(msg)
                h.wait.Done()
                <-h.semaphore
            }()
        case <-h.quit:
            break loop
        }
    }

    h.wait.Wait()
    h.state.To(Idle)
}
