// Copyright (c) 2017 Johan Brandhorst

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package grpcweb

import (
	"github.com/gopherjs/gopherjs/js"
)

// ProtoMessage is implemented by... all *js.Objects.
// But it'll do for an interface to Serialize and Deserialize
type ProtoMessage interface {
	Call(string, ...interface{}) *js.Object
}

// Serialize marshals the provided ProtoMessage into
// a slice of bytes using the serializeBinary ProtobufJS function,
// returning an error if one was thrown.
func Serialize(m ProtoMessage) (resp []byte, err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	return m.Call("serializeBinary").Interface().([]byte), err
}

// Deserialize unmarshals the provided ProtoMessage bytes into
// a generic *js.Object of the provided type,
// returning an error if one was thrown.
func Deserialize(m ProtoMessage, rawBytes []byte) (o *js.Object, err error) {
	// Recover any thrown JS errors
	defer func() {
		e := recover()
		if e == nil {
			return
		}

		if e, ok := e.(*js.Error); ok {
			err = e
		} else {
			panic(e)
		}
	}()

	return m.Call("deserializeBinary", rawBytes), err
}
