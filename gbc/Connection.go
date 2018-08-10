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

package gbc

import (
    "net"
)

type (
    Connection interface {
        // all received message from connection will forward to this channel
        RawMessageChannelSetter

        // start reading data from connection
        Start() error

        // close the connection
        Close() error

        // write bytes to connection
        Write([]byte) (int, error)
    }

    ConnectionGroup interface {
        // set handler function for incoming rawMessage
        RawMessageReceiverSetter

        // start message loop
        Start() error

        // close all connections in group
        Close() error

        // add connection to group
        Add(c Connection) error

        // remove connection from group
        Remove(c Connection) error

        // get message channel for group
        RawMessageChan() chan RawMessage

        // write bytes to all connections in group
        BroadcastWrite([]byte)
    }

    // when new connection accepted, call this function
    OnConnectFunc func(net.Conn) Connection

    ConnectionManager interface {
        // set handler function for incoming connect
        OnConnect(OnConnectFunc)

        // start accepting new connections
        Start(l net.Listener) error

        // stop network server
        Stop()
    }
)
