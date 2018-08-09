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
    "sync"

    "github.com/dualface/go-gbc/gbc"
)

type (
    BasicRawMessagePipeline struct {
        handlers []gbc.RawMessageHandler
    }
)

func NewBasicRawMessagePipeline() *BasicRawMessagePipeline {
    p := &BasicRawMessagePipeline{
        handlers: []gbc.RawMessageHandler{},
    }
    return p
}

// interface RawMessagePipeline

func (p *BasicRawMessagePipeline) WriteRawMessage(m gbc.RawMessage) (err error) {
    for _, h := range p.handlers {
        err = h.WriteRawMessage(m)
        if err != nil {
            break
        }
    }
    return
}

func (p *BasicRawMessagePipeline) Stop() {
    for _, h := range p.handlers {
        h.Stop()
    }
}

func (p *BasicRawMessagePipeline) WaitForComplete() {
    wait := sync.WaitGroup{}
    for _, h := range p.handlers {
        wait.Add(1)
        go func() {
            h.WaitForComplete()
            wait.Done()
        }()
    }
    wait.Wait()
}

func (p *BasicRawMessagePipeline) Append(h gbc.RawMessageHandler) {
    p.handlers = append(p.handlers, h)
}
