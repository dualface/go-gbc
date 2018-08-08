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
    RawMessageInputFilter struct {
        headerBuf    []byte
        headerOffset int
        mc           chan gbc.RawMessage
        msg          *RawMessageImpl
    }
)

func NewRawMessageInputFilter() *RawMessageInputFilter {
    p := &RawMessageInputFilter{
        headerBuf: make([]byte, RawMessageHeaderLen),
    }
    return p
}

// interface InputFilter

func (f *RawMessageInputFilter) WriteBytes(input []byte) (output []byte, err error) {
    avail := len(input)

    for ; avail > 0; {

        if f.msg == nil {
            if f.headerOffset < RawMessageHeaderLen {
                // fill header
                writeLen := avail
                if writeLen > RawMessageHeaderLen-f.headerOffset {
                    writeLen = RawMessageHeaderLen - f.headerOffset
                }

                copy(f.headerBuf[f.headerOffset:], input[0:writeLen])
                f.headerOffset += writeLen
                if f.headerOffset < RawMessageHeaderLen {
                    // if header not filled, return
                    return
                }

                input = input[writeLen:]
                avail -= writeLen
            }

            // header is ready, create msg
            f.headerOffset = 0
            f.msg, err = NewRawMessageFromHeaderBuf(f.headerBuf)
            if err != nil {
                return
            }
        }

        // determine write len
        appendLen := f.msg.RemainsBytes()
        if appendLen > avail {
            appendLen = avail
        }
        avail -= appendLen

        // write to message
        _, err = f.msg.WriteBytes(input[0:appendLen])
        if err != nil {
            return
        }
        input = input[appendLen:]

        // if get a message, send it
        if f.msg.RemainsBytes() == 0 {
            f.mc <- f.msg
            f.msg = nil
        }
    }

    return
}

func (f *RawMessageInputFilter) SetRawMessageReceiver(c chan gbc.RawMessage) {
    f.mc = c
}
