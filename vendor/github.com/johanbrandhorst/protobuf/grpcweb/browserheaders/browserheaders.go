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

package browserheaders

import (
	"github.com/gopherjs/gopherjs/js"
	"google.golang.org/grpc/metadata"

	// Include gRPC-web JS objects
	_ "github.com/johanbrandhorst/protobuf/grpcweb/grpcwebjs"
)

// BrowserHeaders encasulates the Improbable BrowserHeaders.
type BrowserHeaders struct {
	*js.Object
	md metadata.MD `js:"keyValueMap"`
}

// New initializes and populates a new BrowserHeaders.
func New(headers metadata.MD) *BrowserHeaders {
	b := &BrowserHeaders{
		Object: js.Global.Get("BrowserHeaders").New(),
	}
	for k, v := range headers {
		for _, h := range v {
			b.Set(k, h)
		}
	}

	return b
}

// Set sets the value of the key to value. It
// overwrites any values for that key.
func (b *BrowserHeaders) Set(key, value string) {
	b.Call("set", key, value)
}

// Append adds the value to the key without overwriting
// existing values.
func (b *BrowserHeaders) Append(key, value string) {
	b.Call("append", key, value)
}

// Get gets all values associated with key. Mutating
// the returned slice will not modify the contents of the key.
func (b *BrowserHeaders) Get(key string) (value []string) {
	// JavaScript Array types are converted to `[]interface{}`
	// so this copies the values into a separate slice
	b.Call("get", key).Call("forEach", func(v string) {
		value = append(value, v)
	})
	return value
}

// Delete deletes the key.
func (b *BrowserHeaders) Delete(key string) {
	b.Call("delete", key)
}

// DeleteValueFromKey deletes the value from the key.
// If the value is the last value in the key, it also deletes the key.
func (b *BrowserHeaders) DeleteValueFromKey(key, value string) {
	b.Call("delete", key, value)
}

// HasKey returns whether the BrowserHeaders has the key.
func (b *BrowserHeaders) HasKey(key string) bool {
	return b.Call("has", key).Bool()
}

// HasKeyWithValue returns whether the BrowserHeaders has the key
// and whether the value is among the values associated with the key.
func (b *BrowserHeaders) HasKeyWithValue(key, value string) bool {
	return b.Call("has", key, value).Bool()
}

// ForEach runs the callback function against all keys and values.
func (b *BrowserHeaders) ForEach(callback func(string, []string)) {
	b.Call("forEach", callback)
}

// Len returns the number of keys.
func (b *BrowserHeaders) Len() int {
	return b.md.Len()
}
