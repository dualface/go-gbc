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
    "github.com/dualface/go-gbc/gbc"
)

type (
    ConcurrenceMessageHandler struct {
        semaphore        chan int
        onRawMessageFunc gbc.OnRawMessageFunc
    }
)

func NewConcurrenceMessageHandler(concurrence int, f gbc.OnRawMessageFunc) *ConcurrenceMessageHandler {
    if concurrence <= 1 {
        concurrence = 1
    }

    r := &ConcurrenceMessageHandler{
        semaphore:        make(chan int, concurrence),
        onRawMessageFunc: f,
    }

    return r
}

// interface MessagePipeline

func (r *ConcurrenceMessageHandler) ReceiveRawMessage(m gbc.RawMessage) error {
    // avoid blocking caller
    go func() {
        r.semaphore <- 1
        r.onRawMessageFunc(m)
        <-r.semaphore
    }()
    return nil
}
