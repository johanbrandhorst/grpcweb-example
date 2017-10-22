package grpcweb

import (
	"context"

	"github.com/gopherjs/gopherjs/js"
	"google.golang.org/grpc/codes"
	gmd "google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/metadata"
	"github.com/johanbrandhorst/protobuf/grpcweb/status"
)

// Invoke populates the necessary JS structures and performs the gRPC-web call.
// It attempts to catch any JS errors thrown.
func invoke(ctx context.Context, host, service, method string, req []byte, onMsg onMessageFunc, onEnd onEndFunc, opts ...CallOption) (err error) {
	methodDesc := newMethodDescriptor(newService(service), method, newResponseType())

	c := &callInfo{}
	rawOnEnd := func(code int, msg string, trailers metadata.Metadata) {
		s := &status.Status{
			Code:     codes.Code(code),
			Message:  msg,
			Trailers: trailers.MD,
		}
		c.trailers = trailers.MD

		// Perform CallOptions required after call
		for _, o := range opts {
			o.after(c)
		}

		onEnd(s)
	}
	onHeaders := func(headers metadata.Metadata) {
		c.headers = headers.MD
	}

	md, _ := gmd.FromOutgoingContext(ctx)
	props := newProperties(host, false, newRequest(req), metadata.New(md), onHeaders, onMsg, rawOnEnd)

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
