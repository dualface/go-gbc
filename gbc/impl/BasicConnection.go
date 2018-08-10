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
    "io"
    "net"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
)

const (
    readBufferSize   = 1024 * 4 // 4KB
    readFailureLimit = 3
)

type (
    BasicConnection struct {
        RawConn      net.Conn
        InputFilter  gbc.InputFilter
        OutputFilter gbc.OutputFilter
        MessageChan  chan gbc.RawMessage
    }
)

func NewBasicConnection(rawConn net.Conn, i gbc.InputFilter) *BasicConnection {
    conn := &BasicConnection{
        RawConn:     rawConn,
        InputFilter: i,
    }
    return conn
}

// interface Connection

func (c *BasicConnection) Start() error {
    if c.InputFilter == nil {
        clog.PrintWarn("connection '%s' not set input filter", c.RawConn.RemoteAddr().String())
    } else {
        c.InputFilter.SetRawMessageChannel(c.MessageChan)
    }

    if c.MessageChan == nil {
        clog.PrintWarn("connection '%s' not set raw message chan", c.RawConn.RemoteAddr().String())
    }

    go c.loop()
    return nil
}

func (c *BasicConnection) Close() error {
    c.InputFilter = nil
    c.OutputFilter = nil
    c.MessageChan = nil
    return c.RawConn.Close()
}

func (c *BasicConnection) Write(b []byte) (writeLen int, err error) {
    var output []byte
    if c.OutputFilter != nil {
        output, err = c.OutputFilter.WriteBytes(b)
        if err != nil {
            return
        }
        writeLen, err = c.RawConn.Write(output)
    }
    return
}

func (c *BasicConnection) SetRawMessageChannel(mc chan gbc.RawMessage) {
    c.MessageChan = mc
    if c.InputFilter != nil {
        c.InputFilter.SetRawMessageChannel(mc)
    }
}

// private

func (c *BasicConnection) loop() {
    failure := 0
    // use double buffer
    halfBufSize := readBufferSize
    buf := make([]byte, halfBufSize*2, halfBufSize*2)
    offset := 0

    for {
        if failure >= readFailureLimit {
            // stop read
            break
        }

        avail, err := c.RawConn.Read(buf[offset : offset+halfBufSize])
        if err != nil {
            if err != io.EOF {
                clog.PrintWarn("reading failed on %s, %s", c.RawConn.RemoteAddr(), err)
                failure++
                continue // try again
            } else if avail == 0 {
                break // conn closed
            }
        } else {
            // reset read failure counter
            failure = 0
        }

        if avail > 0 && c.InputFilter != nil {
            _, err := c.InputFilter.WriteBytes(buf[offset : offset+avail])
            if err != nil {
                clog.PrintWarn("parsing bytes failed, %s", err)
            }

            offset += avail
        }

        if offset >= halfBufSize {
            offset = 0
        }
    }
}
