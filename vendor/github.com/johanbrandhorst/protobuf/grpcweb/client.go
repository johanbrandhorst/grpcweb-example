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
	"context"
	"io"

	"github.com/gopherjs/gopherjs/js"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/browserheaders"
	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// Client encapsulates all gRPC calls to a
// host-service combination.
type Client struct {
	host    string
	service string
}

// NewClient creates a new Client.
func NewClient(host, service string, opts ...DialOption) *Client {
	c := &Client{
		host:    host,
		service: service,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// RPCCall performs a unary call to an endpoint, blocking until a
// reply has been received or the context was canceled.
func (c Client) RPCCall(ctx context.Context, method string, req []byte, opts ...CallOption) ([]byte, error) {
	respChan := make(chan []byte, 1)
	errChan := make(chan error, 1)

	onMsg := func(in []byte) {
		respChan <- in
	}
	onEnd := func(s *status.Status) {
		if s.Code != codes.OK {
			errChan <- s
		} else {
			errChan <- io.EOF // Success!
		}
	}
	err := invoke(ctx, c.host, c.service, method, req, onMsg, onEnd, opts...)
	if err != nil {
		return nil, err
	}

	select {
	case err := <-errChan:
		// Wait until we've gotten the result from onEnd
		if err == io.EOF {
			select {
			// Now check for the response - should already be
			// here, but can't be too careful
			case resp := <-respChan:
				return resp, nil
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Stream performs a server-to-client streaming RPC call, returning
// a struct which exposes a Go gRPC like streaming interface.
// It is non-blocking.
func (c Client) Stream(ctx context.Context, method string, req []byte, opts ...CallOption) (*StreamClient, error) {
	srv := &StreamClient{
		ctx:      ctx,
		messages: make(chan []byte, 10), // Buffer up to 10 messages
		errors:   make(chan error, 1),
	}

	onMsg := func(in []byte) { srv.messages <- in }
	onEnd := func(s *status.Status) {
		if s.Code != codes.OK {
			srv.errors <- s
		} else {
			srv.errors <- io.EOF
		}
	}
	err := invoke(ctx, c.host, c.service, method, req, onMsg, onEnd, opts...)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

// Invoke populates the necessary JS structures and performs the gRPC-web call.
// It attempts to catch any JS errors thrown.
func invoke(ctx context.Context, host, service, method string, req []byte, onMsg onMessageFunc, onEnd onEndFunc, opts ...CallOption) (err error) {
	methodDesc := newMethodDescriptor(newService(service), method, newResponseType())

	c := &callInfo{}
	rawOnEnd := func(code int, msg string, trailers *browserheaders.BrowserHeaders) {
		s := status.New(codes.Code(code), msg, trailers)
		c.trailers = trailers

		// Perform CallOptions required after call
		for _, o := range opts {
			o.after(c)
		}

		onEnd(s)
	}
	onHeaders := func(headers *browserheaders.BrowserHeaders) {
		c.headers = headers
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	props := newProperties(host, false, newRequest(req), browserheaders.New(md), onHeaders, onMsg, rawOnEnd)

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

	// Perform CallOptions required before call
	for _, o := range opts {
		if err := o.before(c); err != nil {
			return status.FromError(err)
		}
	}

	js.Global.Get("grpc").Call("invoke", methodDesc, props)

	return nil
}
