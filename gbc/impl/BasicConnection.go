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
	"io"
	"net"
)

type (
	BasicConnection struct {
		conn   net.Conn
		mc     chan *gbc.RawMessage
		parser *RawMessageParser
	}
)

func NewBasicConnection(c net.Conn) gbc.Connection {
	cc := &BasicConnection{
		conn:   c,
		parser: NewRawMessageParser(),
	}
	return cc
}

func (c *BasicConnection) SetMessageChan(mc chan *gbc.RawMessage) {
	c.mc = mc
}

func (c *BasicConnection) Start() {
	go c.loop()
}

func (c *BasicConnection) Close() error {
	return c.conn.Close()
}

// private

func (c *BasicConnection) loop() {
	failure := 0
	// use double buffer
	halfBufSize := gbc.ConnectionReadBufferSize
	buf := make([]byte, halfBufSize*2, halfBufSize*2)
	offset := 0

	for {
		if failure >= gbc.ConnectionReadFailureLimit {
			// stop read
			break
		}

		avail, err := c.conn.Read(buf[offset : offset+halfBufSize])
		if err != nil {
			if err != io.EOF {
				clog.PrintWarn("reading failed on %s, %s", c.conn.RemoteAddr(), err)
				failure++
				continue // try again
			} else if avail == 0 {
				break // conn closed
			}
		} else {
			// reset read failure counter
			failure = 0
		}

		for avail > 0 {
			writeLen, err := c.parser.WriteBytes(buf[offset : offset+avail])
			avail -= writeLen
			offset += writeLen
			if err != nil {
				clog.PrintWarn("parsing bytes failed, %s", err)
			}

			msg := c.parser.FetchMessage()
			if msg != nil {
				clog.PrintDebug("%s", msg)
			}
		}

		if offset >= halfBufSize {
			offset = 0
		}
	}
}
