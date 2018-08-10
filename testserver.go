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

package main

import (
    "fmt"
    "math/rand"
    "net"
    "os"
    "os/signal"
    "time"

    "github.com/dualface/go-cli-colorlog"
    "github.com/dualface/go-gbc/gbc"
    "github.com/dualface/go-gbc/gbc/impl"
)

const (
    Bind = "localhost"
    Port = "27010"
)

func main() {
    rand.Seed(time.Now().Unix())

    // a worker pool, max 3 concurrence jobs
    handler := impl.NewConcurrenceRawMessageHandler(3, func(m gbc.RawMessage) error {
        fmt.Printf("%+v\n", m)
        time.Sleep(time.Second / 2)
        return nil
    })

    // start listening on specified addr
    addr := fmt.Sprintf("%s:%s", Bind, Port)
    l, err := net.Listen("tcp", addr)
    if err == nil {
        cm := impl.NewBasicConnectionManager()

        // use filter chain process incoming bytes
        cm.OnConnect(func(rawConn net.Conn) gbc.Connection {
            p := impl.NewBasicInputPipeline()
            p.Append(impl.NewBase64DecodeFilter())
            p.Append(impl.NewXORFilter([]byte{0xff}))

            // RawMessageInputFilter fetch rawMessage from bytes stream,
            // and send rawMessage to connection group
            p.Append(impl.NewRawMessageInputFilter())
            return impl.NewBasicConnection(rawConn, p)
        })

        // forward message to worker pool
        cm.DefaultGroup.OnRawMessage(handler.ReceiveRawMessage)

        // handle CTRL+C
        signCh := make(chan os.Signal)
        signal.Notify(signCh, os.Interrupt)
        go func() {
            <-signCh
            // sig is a ^C, handle it
            clog.PrintInfo("signal os.Interrupt captured")
            cm.Stop()
        }()

        err = cm.Start(l)
    }

    if err != nil {
        fmt.Printf("[ERR] %s\n\n", err)
        os.Exit(1)
    }
}
