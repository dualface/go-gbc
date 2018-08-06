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
    RawMessageParser struct {
        headerBuf    []byte
        headerOffset int
        newMsg       bool
        msg          *RawMessageImpl
    }
)

func NewRawMessageParser() *RawMessageParser {
    p := &RawMessageParser{
        headerBuf: make([]byte, RawMessageHeaderLen),
        newMsg:    true,
    }
    return p
}

func (p *RawMessageParser) WriteBytes(b []byte) (writeLen int, err error) {
    avail := len(b)

    if p.newMsg {
        p.msg = nil

        if p.headerOffset < RawMessageHeaderLen {
            // fill header
            writeLen = avail
            if writeLen > RawMessageHeaderLen-p.headerOffset {
                writeLen = RawMessageHeaderLen - p.headerOffset
            }

            copy(p.headerBuf[p.headerOffset:], b[0:writeLen])
            p.headerOffset += writeLen
            if p.headerOffset < RawMessageHeaderLen {
                // if header not filled, return
                return
            }

            b = b[writeLen:]
            avail -= writeLen
        }

        // header is ready, create msg
        p.headerOffset = 0
        p.newMsg = true
        p.msg, err = NewRawMessageFromHeaderBuf(p.headerBuf)
        if err != nil {
            return
        }
    }

    // determine write len
    appendLen := p.msg.RemainsBytes()
    if appendLen > avail {
        appendLen = avail
    }
    writeLen += appendLen

    // write to message
    _, err = p.msg.WriteBytes(b[0:appendLen])
    return
}

func (p *RawMessageParser) FetchMessage() gbc.RawMessage {
    if p.msg != nil && p.msg.RemainsBytes() == 0 {
        m := p.msg
        p.msg = nil
        return m
    }
    return nil
}
