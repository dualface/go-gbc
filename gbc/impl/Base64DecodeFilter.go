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
    "encoding/base64"
    "fmt"
)

const (
    base64TupleLen = 4
    numberOfTuples = 3
    tupleBuffSize  = base64TupleLen * numberOfTuples
)

type (
    Base64DecodeFilter struct {
        tupleBuff  []byte
        tupleAvail int
        decodeBuff []byte
    }
)

func NewBase64DecodeFilter() *Base64DecodeFilter {
    h := &Base64DecodeFilter{
        tupleBuff:  make([]byte, tupleBuffSize),
        decodeBuff: make([]byte, tupleBuffSize),
    }
    return h
}

// interface Filter

func (f *Base64DecodeFilter) fillTupleBuffer(input []byte) int {
    if f.tupleAvail == 0 {
        return 0
    }

    // fill tuple buffer
    avail := len(input)
    used := tupleBuffSize - f.tupleAvail
    if used > avail {
        used = avail
    }

    copy(f.tupleBuff[f.tupleAvail:], input[:used])
    f.tupleAvail += used
    return used
}

func (f *Base64DecodeFilter) decodeTupleBuffer() (int, error) {
    if f.tupleAvail < base64TupleLen {
        return 0, nil
    }

    usedTuple := f.tupleAvail - (f.tupleAvail % base64TupleLen)
    decodeLen, err := base64.StdEncoding.Decode(f.decodeBuff, f.tupleBuff[0:usedTuple])

    if f.tupleAvail > usedTuple {
        // move remains tuple to head of buffer
        copy(f.tupleBuff, f.tupleBuff[usedTuple:f.tupleAvail])
        f.tupleAvail -= usedTuple
    } else {
        f.tupleAvail = 0
    }

    return decodeLen, err
}

func (f *Base64DecodeFilter) WriteBytes(input []byte) ([]byte, error) {
    avail := len(input)
    output := input
    write := output

    used := f.fillTupleBuffer(input)
    decodeLen1, _ := f.decodeTupleBuffer()

    if used == avail {
        // no more tuples
        return f.decodeBuff[:decodeLen1], nil
    }

    if used < decodeLen1 {
        return nil, fmt.Errorf("unknown error")
    } else if used >= decodeLen1 && decodeLen1 != 0 {
        // copy decoded bytes to input buffer
        copy(write, f.decodeBuff[:decodeLen1])
        write = write[decodeLen1:]
    }

    // decode remains tuples
    decodeLen2, _ := base64.StdEncoding.Decode(write, input[used:])
    // copy not decoded tuples to tuple buffer
    used += decodeLen2 * 8 / 6
    copy(f.tupleBuff[f.tupleAvail:], input[used:])
    f.tupleAvail += avail - used

    return output[:decodeLen1+decodeLen2], nil
}
