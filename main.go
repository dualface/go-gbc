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
    "time"

    "github.com/dualface/go-gbc/gbc"
    "github.com/dualface/go-gbc/gbc/impl"
)

const (
    Host = "localhost"
    Port = "27010"
)

func main() {
    rand.Seed(time.Now().Unix())

    addr := fmt.Sprintf("%s:%s", Host, Port)
    l, err := net.Listen("tcp", addr)
    if err == nil {
        handler := impl.NewBasicRawMessageHandler(3, rawMessageHandler)
        policy := impl.NewAllInOneConnectionGroupPolicy(nil, handler)
        cm := impl.NewBasicConnectionManager(connectHandler, policy)
        err = cm.Start(l)
    }

    if err != nil {
        fmt.Printf("[ERR] %s\n\n", err)
        os.Exit(1)
    }
}

func connectHandler(rawConn net.Conn) gbc.Connection {
    c := impl.NewBasicConnection(rawConn, nil)
    c.Pipeline.Append(impl.NewBase64DecodeFilter())
    c.Pipeline.Append(impl.NewXORFilter([]byte{0xff}))
    c.Pipeline.Append(impl.NewRawMessageInputFilter())
    return c
}

func rawMessageHandler(m gbc.RawMessage) {
    fmt.Printf("%+v\n", m)
    time.Sleep(time.Second * time.Duration(rand.Intn(3)+1))
}
