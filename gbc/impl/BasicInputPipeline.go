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
    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
)

type (
    BasicInputPipeline struct {
        filters []gbc.Filter
        mc      chan gbc.RawMessage
    }
)

func NewBasicInputPipeline() *BasicInputPipeline {
    p := &BasicInputPipeline{
        filters: []gbc.Filter{},
    }
    return p
}

// interface InputPipeline

func (p *BasicInputPipeline) WriteBytes(input []byte) (output []byte, err error) {
    for _, f := range p.filters {
        output, err = f.WriteBytes(input)

        if err != nil {
            clog.PrintWarn("filter '%T' failed, %s", f, err.Error())
            break
        }
        if len(output) == 0 {
            break
        }

        input = output
    }

    return
}

func (p *BasicInputPipeline) Append(f gbc.Filter) {
    p.setupFilter(f)
    p.filters = append(p.filters, f)
}

func (p *BasicInputPipeline) SetRawMessageReceiver(c chan gbc.RawMessage) {
    p.mc = c
    for _, f := range p.filters {
        p.setupFilter(f)
    }
}

// private

func (p *BasicInputPipeline) setupFilter(f gbc.Filter) {
    mr, ok := f.(gbc.RawMessageChannelSender)
    if ok {
        mr.SetRawMessageReceiver(p.mc)
    }
}
