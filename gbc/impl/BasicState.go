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

type State string

const (
    // Idle => Starting => Running => Stopping => Idle
    Idle     State = "idle"
    Starting State = "starting"
    Running  State = "running"
    Stopping State = "stopping"
)

type (
    BasicState struct {
        state State
        seq   []State
    }
)

func NewBasicState() *BasicState {
    s := &BasicState{
        state: Idle,
        seq:   []State{Idle, Starting, Running, Stopping, Idle},
    }
    return s
}

func (s *BasicState) Current() State {
    return s.state
}

func (s *BasicState) To(next State) bool {
    for i, state := range s.seq {
        if state != s.state {
            continue
        }

        if s.seq[i+1] != next {
            return false
        }

        break
    }
    s.state = next
    return true
}
