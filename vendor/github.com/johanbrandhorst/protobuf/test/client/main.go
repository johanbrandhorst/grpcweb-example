package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/rusco/qunit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"honnef.co/go/js/dom"

	"github.com/johanbrandhorst/protobuf/grpcweb"
	metatest "github.com/johanbrandhorst/protobuf/grpcweb/metadata/test"
	"github.com/johanbrandhorst/protobuf/grpcweb/status"
	grpctest "github.com/johanbrandhorst/protobuf/grpcweb/test"
	gentest "github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test"
	"github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi"
	"github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/types"
	"github.com/johanbrandhorst/protobuf/ptypes/empty"
	"github.com/johanbrandhorst/protobuf/test/client/proto/test"
	"github.com/johanbrandhorst/protobuf/test/recoverer"
	"github.com/johanbrandhorst/protobuf/test/shared"
)

//go:generate gopherjs build main.go -o html/index.js

var uri string

func init() {
	u, err := url.Parse(dom.GetWindow().Document().BaseURI())
	if err != nil {
		panic(err)
	}
	uri = u.Scheme + "://" + u.Hostname()
}

func typeTests() {
	qunit.Module("Integration Types tests")

	qunit.Test("PingRequest Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
		req := &test.PingRequest{
			Value:             "1234",
			ResponseCount:     10,
			ErrorCodeReturned: 1,
			FailureType:       test.PingRequest_CODE,
			CheckMetadata:     true,
			SendHeaders:       true,
			SendTrailers:      true,
			MessageLatencyMs:  100,
		}

		marshalled := req.Marshal()
		newReq, err := new(test.PingRequest).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})

	qunit.Test("ExtraStuff Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
		req := &test.ExtraStuff{
			Addresses: map[int32]string{
				1234: "The White House",
				5678: "The Empire State Building",
			},
			Title: &test.ExtraStuff_FirstName{
				FirstName: "Allison",
			},
			CardNumbers: []uint32{
				1234, 5678,
			},
		}

		marshalled := req.Marshal()
		newReq, err := new(test.ExtraStuff).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})
}

func serverTests(label, serverAddr, emptyServerAddr string) {
	qunit.Module(fmt.Sprintf("%s Integration tests", label))

	qunit.AsyncTest("Unary server call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:         "test",
				ResponseCount: 1,
			}
			resp, err := c.Ping(context.Background(), req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Ping error seen: "+st.Message)
				return
			}

			qunit.Ok(true, "Request succeeded")
			if resp.GetValue() != "test" {
				qunit.Ok(false, fmt.Sprintf("Value was not as expected, was %q", resp.GetValue()))
			}
			if resp.GetCounter() != 1 {
				qunit.Ok(false, fmt.Sprintf("Counter was not as expected, was %q", resp.GetCounter()))
			}
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call with metadata", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:         "test",
				ResponseCount: 1,
				CheckMetadata: true,
			}
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(shared.ClientMDTestKey, shared.ClientMDTestValue))
			resp, err := c.Ping(ctx, req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Ping error seen: "+st.Message)
				return
			}

			qunit.Ok(true, "Request succeeded")
			if resp.GetValue() != "test" {
				qunit.Ok(false, fmt.Sprintf("Value was not as expected, was %q", resp.GetValue()))
			}
			if resp.GetCounter() != 1 {
				qunit.Ok(false, fmt.Sprintf("Counter was not as expected, was %q", resp.GetCounter()))
			}
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call expecting headers and trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:         "test",
				ResponseCount: 1,
				SendHeaders:   true,
				SendTrailers:  true,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			resp, err := c.Ping(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Ping error seen: "+st.Message)
				return
			}

			qunit.Ok(true, "Request succeeded")

			if resp.GetValue() != "test" {
				qunit.Ok(false, fmt.Sprintf("Value was not as expected, was %q", resp.GetValue()))
			}
			if resp.GetCounter() != 1 {
				qunit.Ok(false, fmt.Sprintf("Counter was not as expected, was %q", resp.GetCounter()))
			}

			if len(headers[strings.ToLower(shared.ServerMDTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 1 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey1)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Header 1 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey1)]))
			}
			if len(headers[strings.ToLower(shared.ServerMDTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 2 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey2)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Header 2 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey2)]))
			}

			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 1 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Trailer 1 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey1)]))
			}
			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 2 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Trailer 2 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey2)]))
			}
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call expecting only headers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:         "test",
				ResponseCount: 1,
				SendHeaders:   true,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			resp, err := c.Ping(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Ping error seen: "+st.Message)
				return
			}

			qunit.Ok(true, "Request succeeded")

			if resp.GetValue() != "test" {
				qunit.Ok(false, fmt.Sprintf("Value was not as expected, was %q", resp.GetValue()))
			}
			if resp.GetCounter() != 1 {
				qunit.Ok(false, fmt.Sprintf("Counter was not as expected, was %q", resp.GetCounter()))
			}

			if len(headers[strings.ToLower(shared.ServerMDTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 1 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey1)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Header 1 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey1)]))
			}
			if len(headers[strings.ToLower(shared.ServerMDTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 2 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey2)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Header 2 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey2)]))
			}

			// Trailers always include the grpc-status, anything else is unexpected
			if trailers.Len() > 1 {
				qunit.Ok(false, fmt.Sprintf("Unexpected trailer provided, size of trailers was %d", trailers.Len()))
			}
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call expecting only trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:         "test",
				ResponseCount: 1,
				SendTrailers:  true,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			resp, err := c.Ping(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Ping error seen: "+st.Message)
				return
			}

			qunit.Ok(true, "Request succeeded")

			if resp.GetValue() != "test" {
				qunit.Ok(false, fmt.Sprintf("Value was not as expected, was %q", resp.GetValue()))
			}
			if resp.GetCounter() != 1 {
				qunit.Ok(false, fmt.Sprintf("Counter was not as expected, was %q", resp.GetCounter()))
			}

			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 1 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Trailer 1 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey1)]))
			}
			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 2 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Trailer 2 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey2)]))
			}

			// Headers always include the content-type, anything else is unexpected
			if headers.Len() > 1 {
				qunit.Ok(false, fmt.Sprintf("Unexpected header provided, size of headers was %d", headers.Len()))
			}
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call returning gRPC error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				Value:             "",
				ResponseCount:     0,
				ErrorCodeReturned: uint32(codes.InvalidArgument),
				FailureType:       test.PingRequest_CODE,
			}
			_, err := c.PingError(context.Background(), req)
			if err == nil {
				qunit.Ok(false, "Expected error, returned nil")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.InvalidArgument {
				qunit.Ok(false, fmt.Sprintf("Unexpected code returned, was %s", st.Code))
			}

			qunit.Ok(true, "Error was as expected")
		}()

		return nil
	})

	qunit.AsyncTest("Unary server call returning network error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := &test.PingRequest{
				FailureType: test.PingRequest_DROP,
			}
			_, err := c.PingError(context.Background(), req)
			if err == nil {
				qunit.Ok(false, "Expected error, returned nil")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.Internal {
				qunit.Ok(false, fmt.Sprintf("Unexpected code returned, was %s", st.Code))
			}
			if st.Message != "Response closed without grpc-status (Headers only)" {
				qunit.Ok(false, fmt.Sprintf("Unexpected message returned, was %q", st.Message))
			}

			qunit.Ok(true, "Error was as expected")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				MessageLatencyMs: 1,
			}
			srv, err := c.PingList(context.Background(), req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}

			var pings []*test.PingResponse
			for {
				ping, err := srv.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}

					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
					return
				}

				pings = append(pings, ping)
			}

			if len(pings) != int(req.GetResponseCount()) {
				qunit.Ok(false, fmt.Sprintf("Unexpected number of replies: expected 20, saw %d", len(pings)))
			}

			for i, ping := range pings {
				if int(ping.GetCounter()) != i {
					qunit.Ok(false, fmt.Sprintf("Unexpected count in ping #%d, was %d", i, ping.GetCounter()))
				}
				if ping.GetValue() != fmt.Sprintf(`test %d`, i) {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in ping #%d, was %q", i, ping.GetValue()))
				}
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call with metadata", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				MessageLatencyMs: 1,
			}
			ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(shared.ClientMDTestKey, shared.ClientMDTestValue))
			srv, err := c.PingList(ctx, req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}

			var pings []*test.PingResponse
			for {
				ping, err := srv.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}

					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
					return
				}

				pings = append(pings, ping)
			}

			if len(pings) != int(req.GetResponseCount()) {
				qunit.Ok(false, fmt.Sprintf("Unexpected number of replies: expected 20, saw %d", len(pings)))
			}

			for i, ping := range pings {
				if int(ping.GetCounter()) != i {
					qunit.Ok(false, fmt.Sprintf("Unexpected count in ping #%d, was %d", i, ping.GetCounter()))
				}
				if ping.GetValue() != fmt.Sprintf(`test %d`, i) {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in ping #%d, was %q", i, ping.GetValue()))
				}
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call expecting headers and trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				SendHeaders:      true,
				SendTrailers:     true,
				MessageLatencyMs: 1,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			srv, err := c.PingList(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}

			var pings []*test.PingResponse
			for {
				ping, err := srv.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}

					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
					return
				}

				pings = append(pings, ping)
			}

			if len(pings) != int(req.GetResponseCount()) {
				qunit.Ok(false, fmt.Sprintf("Unexpected number of replies: expected 20, saw %d", len(pings)))
			}

			for i, ping := range pings {
				if int(ping.GetCounter()) != i {
					qunit.Ok(false, fmt.Sprintf("Unexpected count in ping #%d, was %d", i, ping.GetCounter()))
				}
				if ping.GetValue() != fmt.Sprintf(`test %d`, i) {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in ping #%d, was %q", i, ping.GetValue()))
				}
			}

			if len(headers[strings.ToLower(shared.ServerMDTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 1 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey1)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Header 1 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey1)]))
			}
			if len(headers[strings.ToLower(shared.ServerMDTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 2 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey2)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Header 2 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey2)]))
			}

			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 1 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Trailer 1 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey1)]))
			}
			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 2 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Trailer 2 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey2)]))
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call expecting only headers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				SendHeaders:      true,
				MessageLatencyMs: 1,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			srv, err := c.PingList(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}

			var pings []*test.PingResponse
			for {
				ping, err := srv.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}

					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
					return
				}

				pings = append(pings, ping)
			}

			if len(pings) != int(req.GetResponseCount()) {
				qunit.Ok(false, fmt.Sprintf("Unexpected number of replies: expected 20, saw %d", len(pings)))
			}

			for i, ping := range pings {
				if int(ping.GetCounter()) != i {
					qunit.Ok(false, fmt.Sprintf("Unexpected count in ping #%d, was %d", i, ping.GetCounter()))
				}
				if ping.GetValue() != fmt.Sprintf(`test %d`, i) {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in ping #%d, was %q", i, ping.GetValue()))
				}
			}

			if len(headers[strings.ToLower(shared.ServerMDTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 1 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey1)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Header 1 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey1)]))
			}
			if len(headers[strings.ToLower(shared.ServerMDTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Header 2 was not as expected, was %v", len(headers[strings.ToLower(shared.ServerMDTestKey2)])))
			}
			if headers[strings.ToLower(shared.ServerMDTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Header 2 was not as expected, was %q", headers[strings.ToLower(shared.ServerMDTestKey2)]))
			}

			// Trailers always include the grpc-status, anything else is unexpected
			if trailers.Len() > 1 {
				qunit.Ok(false, fmt.Sprintf("Unexpected trailer provided, size of trailers was %d", trailers.Len()))
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call expecting only trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				SendTrailers:     true,
				MessageLatencyMs: 1,
			}
			headers, trailers := metadata.New(nil), metadata.New(nil)
			srv, err := c.PingList(context.Background(), req, grpcweb.Header(&headers), grpcweb.Trailer(&trailers))
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}

			var pings []*test.PingResponse
			for {
				ping, err := srv.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}

					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
					return
				}

				pings = append(pings, ping)
			}

			if len(pings) != int(req.GetResponseCount()) {
				qunit.Ok(false, fmt.Sprintf("Unexpected number of replies: expected 20, saw %d", len(pings)))
			}

			for i, ping := range pings {
				if int(ping.GetCounter()) != i {
					qunit.Ok(false, fmt.Sprintf("Unexpected count in ping #%d, was %d", i, ping.GetCounter()))
				}
				if ping.GetValue() != fmt.Sprintf(`test %d`, i) {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in ping #%d, was %q", i, ping.GetValue()))
				}
			}

			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 1 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey1)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey1)][0] != shared.ServerMDTestValue1 {
				qunit.Ok(false, fmt.Sprintf("Trailer 1 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey1)]))
			}
			if len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)]) != 1 {
				qunit.Ok(false, fmt.Sprintf("Size of Trailer 2 was not as expected, was %v", len(trailers[strings.ToLower(shared.ServerTrailerTestKey2)])))
			}
			if trailers[strings.ToLower(shared.ServerTrailerTestKey2)][0] != shared.ServerMDTestValue2 {
				qunit.Ok(false, fmt.Sprintf("Trailer 2 was not as expected, was %q", trailers[strings.ToLower(shared.ServerTrailerTestKey2)]))
			}

			// Headers always include the content-type, anything else is unexpected
			if headers.Len() > 1 {
				qunit.Ok(false, fmt.Sprintf("Unexpected header provided, size of headers was %d", headers.Len()))
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Server Streaming call returning network error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := &test.PingRequest{
				Value:            "test",
				ResponseCount:    20,
				FailureType:      test.PingRequest_DROP,
				MessageLatencyMs: 1,
			}
			srv, err := c.PingList(context.Background(), req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingList error seen: "+st.Message)
				return
			}
			_, err = srv.Recv()
			if err == nil {
				qunit.Ok(false, "Expected error, returned nil")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.Internal {
				qunit.Ok(false, fmt.Sprintf("Unexpected code returned, was %s", st.Code))
			}
			if st.Message != "Response closed without grpc-status (Headers only)" {
				qunit.Ok(false, fmt.Sprintf("Unexpected message returned, was %q", st.Message))
			}

			qunit.Ok(true, "Error was as expected")
		}()

		return nil
	})

	qunit.AsyncTest("Unary call to empty server", func() interface{} {
		c := test.NewTestServiceClient(uri + emptyServerAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			_, err := c.PingEmpty(context.Background(), &empty.Empty{})
			if err == nil {
				qunit.Ok(false, "Expected error, returned nil")
				return
			}

			st := status.FromError(err)
			if st.Message != "unknown service test.TestService" {
				qunit.Ok(false, "Unexpected error, saw "+st.Message)
			}

			qunit.Ok(true, "Error was as expected")
		}()

		return nil
	})

	qunit.AsyncTest("Unary call to echo server with many types", func() interface{} {
		c := types.NewEchoServiceClient(uri + serverAddr)
		req := &types.TestAllTypes{
			SingleInt32:       1,
			SingleInt64:       2,
			SingleUint32:      3,
			SingleUint64:      4,
			SingleSint32:      5,
			SingleSint64:      6,
			SingleFixed32:     7,
			SingleFixed64:     8,
			SingleSfixed32:    9,
			SingleSfixed64:    10,
			SingleFloat:       10.5,
			SingleDouble:      11.5,
			SingleBool:        true,
			SingleString:      "Alfred",
			SingleBytes:       []byte("Megan"),
			SingleNestedEnum:  types.TestAllTypes_BAR,
			SingleForeignEnum: types.ForeignEnum_FOREIGN_BAR,
			SingleImportedMessage: &multi.Multi1{
				Color:   multi.Multi2_GREEN,
				HatType: multi.Multi3_FEDORA,
			},
			SingleNestedMessage: &types.TestAllTypes_NestedMessage{
				B: 12,
			},
			SingleForeignMessage: &types.ForeignMessage{
				C: 13,
			},
			RepeatedInt32:       []int32{14, 15},
			RepeatedInt64:       []int64{16, 17},
			RepeatedUint32:      []uint32{18, 19},
			RepeatedUint64:      []uint64{20, 21},
			RepeatedSint32:      []int32{22, 23},
			RepeatedSint64:      []int64{24, 25},
			RepeatedFixed32:     []uint32{26, 27},
			RepeatedFixed64:     []uint64{28, 29},
			RepeatedSfixed32:    []int32{30, 31},
			RepeatedSfixed64:    []int64{32, 33},
			RepeatedFloat:       []float32{34.33, 35.34},
			RepeatedDouble:      []float64{36.35, 37.36},
			RepeatedBool:        []bool{true, false, true},
			RepeatedString:      []string{"Alfred", "Robin", "Simon"},
			RepeatedBytes:       [][]byte{[]byte("David"), []byte("Henrik")},
			RepeatedNestedEnum:  []types.TestAllTypes_NestedEnum{types.TestAllTypes_BAR, types.TestAllTypes_BAZ},
			RepeatedForeignEnum: []types.ForeignEnum{types.ForeignEnum_FOREIGN_BAR, types.ForeignEnum_FOREIGN_BAZ},
			RepeatedImportedMessage: []*multi.Multi1{
				{
					Color:   multi.Multi2_RED,
					HatType: multi.Multi3_FEZ,
				},
				{
					Color:   multi.Multi2_GREEN,
					HatType: multi.Multi3_FEDORA,
				},
			},
			RepeatedNestedMessage: []*types.TestAllTypes_NestedMessage{
				{
					B: 38,
				},
				{
					B: 39,
				},
			},
			RepeatedForeignMessage: []*types.ForeignMessage{
				{
					C: 40,
				},
				{
					C: 41,
				},
			},
			OneofField: &types.TestAllTypes_OneofImportedMessage{
				OneofImportedMessage: &multi.Multi1{
					Multi2: &multi.Multi2{
						RequiredValue: 42,
						Color:         multi.Multi2_BLUE,
					},
					Color:   multi.Multi2_RED,
					HatType: multi.Multi3_FEDORA,
				},
			},
		}

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			resp, err := c.EchoAllTypes(context.Background(), req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected error:"+st.Error())
				return
			}
			if !reflect.DeepEqual(req, resp) {
				qunit.Ok(false, fmt.Sprintf("response and request differed: Req:\n%v\nResp:\n%v", req, resp))
				return
			}

			qunit.Ok(true, "Request and Response matched")
		}()

		return nil
	})

	qunit.AsyncTest("Unary call to echo server with many maps", func() interface{} {
		c := types.NewEchoServiceClient(uri + serverAddr)
		req := &types.TestMap{
			MapInt32Int32: map[int32]int32{
				1: 2,
				3: 4,
			},
			MapInt64Int64: map[int64]int64{
				5: 6,
				7: 8,
			},
			MapUint32Uint32: map[uint32]uint32{
				9:  10,
				11: 12,
			},
			MapUint64Uint64: map[uint64]uint64{
				13: 14,
				15: 16,
			},
			MapSint32Sint32: map[int32]int32{
				17: 18,
				19: 20,
			},
			MapSint64Sint64: map[int64]int64{
				21: 22,
				23: 24,
			},
			MapFixed32Fixed32: map[uint32]uint32{
				25: 26,
				27: 28,
			},
			MapFixed64Fixed64: map[uint64]uint64{
				29: 30,
				31: 32,
			},
			MapSfixed32Sfixed32: map[int32]int32{
				33: 34,
				35: 36,
			},
			MapSfixed64Sfixed64: map[int64]int64{
				37: 38,
				39: 40,
			},
			MapInt32Float: map[int32]float32{
				41:  42.41,
				432: 44.43,
			},
			MapInt32Double: map[int32]float64{
				45: 46.45,
				47: 48.47,
			},
			MapBoolBool: map[bool]bool{
				true:  false,
				false: false,
			},
			MapStringString: map[string]string{
				"Henrik": "David",
				"Simon":  "Robin",
			},
			MapInt32Bytes: map[int32][]byte{
				49: []byte("Astrid"),
				50: []byte("Ebba"),
			},
			MapInt32Enum: map[int32]types.MapEnum{
				51: types.MapEnum_MAP_ENUM_BAR,
				52: types.MapEnum_MAP_ENUM_BAZ,
			},
			MapInt32ForeignMessage: map[int32]*types.ForeignMessage{
				53: {C: 54},
				55: {C: 56},
			},
			MapInt32ImportedMessage: map[int32]*multi.Multi1{
				57: {
					Multi2: &multi.Multi2{
						RequiredValue: 58,
						Color:         multi.Multi2_RED,
					},
					Color:   multi.Multi2_GREEN,
					HatType: multi.Multi3_FEZ,
				},
				59: {
					Color:   multi.Multi2_BLUE,
					HatType: multi.Multi3_FEDORA,
				},
			},
		}

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			resp, err := c.EchoMaps(context.Background(), req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected error:"+st.Error())
				return
			}
			if !reflect.DeepEqual(req, resp) {
				qunit.Ok(false, fmt.Sprintf("response and request differed: Req:\n%v\nResp:\n%v", req, resp))
				return
			}

			qunit.Ok(true, "Request and Response matched")
		}()

		return nil
	})

}

func bidiServerTests(serverAddr string) {
	qunit.Module("Client streaming tests")

	qunit.AsyncTest("Client Streaming call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingClientStream(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingClientStream error seen: "+st.Message)
				return
			}

			for i := 0; i < 20; i++ {
				req := &test.PingRequest{
					Value:            "test",
					MessageLatencyMs: 1,
				}
				err := srv.Send(req)
				if err != nil {
					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
					return
				}
			}

			ping, err := srv.CloseAndRecv()
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected CloseAndRecv error seen:"+st.Message)
				return
			}
			if ping.GetValue() != "Closed" {
				qunit.Ok(false, fmt.Sprintf("Unexpected value in response ping, was %q", ping.GetValue()))
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Client Streaming call mid send grpc error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingClientStreamError(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingClientStreamError error seen: "+st.Message)
				return
			}

			req := &test.PingRequest{
				Value:            "test",
				MessageLatencyMs: 1,
			}
			err = srv.Send(req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Trigger error
			req = &test.PingRequest{
				FailureType:       test.PingRequest_CODE,
				ErrorCodeReturned: uint32(codes.DataLoss),
			}
			err = srv.Send(req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Shouldn't error
			err = srv.Send(req)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Should return the error
			_, err = srv.CloseAndRecv()
			if err == nil {
				qunit.Ok(false, "Unexpected nil error")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.DataLoss {
				qunit.Ok(false, fmt.Sprintf("Unexpected code in error, was %q, expected %q, error: %v", st.Code, codes.DataLoss, st.Error()))
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Client Streaming call after send grpc error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingClientStreamError(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingClientStreamError error seen: "+st.Message)
				return
			}

			_, err = srv.CloseAndRecv()
			if err == nil {
				qunit.Ok(false, "Unexpected nil error")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.Internal {
				qunit.Ok(false, fmt.Sprintf("Unexpected code in error, was %q, expected %q", st.Code, codes.Internal))
				return
			}
			if st.Message != "error" {
				qunit.Ok(false, fmt.Sprintf("Unexpected message in error, was %q, expected %q", st.Message, "error"))
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Bi-directional streaming call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingBidiStream(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingBidiStream error seen: "+st.Message)
				return
			}

			for i := 0; i < 10; i++ {
				req := &test.PingRequest{
					Value:            "test",
					MessageLatencyMs: 1,
				}
				err := srv.Send(req)
				if err != nil {
					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
					return
				}

				ping, err := srv.Recv()
				if err != nil {
					st := status.FromError(err)
					qunit.Ok(false, "Unexpected Recv error seen: "+st.Message)
					return
				}
				if ping.GetValue() != req.Value {
					qunit.Ok(false, fmt.Sprintf("Unexpected value in response ping, was %q, expected %q", ping.GetValue(), req.Value))
					return
				}
			}

			err = srv.CloseSend()
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected CloseSend error seen: "+st.Message)
				return
			}

			_, err = srv.Recv()
			if err != io.EOF {
				qunit.Ok(false, "Recv after CloseSend did not return io.EOF, got: "+err.Error())
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Bi-directional streaming call mid send grpc error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingBidiStreamError(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingBidiStreamError error seen: "+st.Message)
				return
			}

			req1 := &test.PingRequest{
				Value:            "test",
				MessageLatencyMs: 1,
			}
			err = srv.Send(req1)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Trigger error
			req2 := &test.PingRequest{
				FailureType:       test.PingRequest_CODE,
				ErrorCodeReturned: uint32(codes.DataLoss),
			}
			err = srv.Send(req2)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Shouldn't error
			err = srv.Send(req2)
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			// Shouldn't error
			ping, err := srv.Recv()
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Recv error seen:"+st.Message)
				return
			}
			if ping.GetValue() != req1.Value {
				qunit.Ok(false, fmt.Sprintf("Unexpected value in response ping, was %q, expected %q", ping.GetValue(), req1.Value))
				return
			}

			// Should error
			_, err = srv.Recv()
			if err == nil {
				qunit.Ok(false, "Unexpected nil error")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.DataLoss {
				qunit.Ok(false, fmt.Sprintf("Unexpected code in error, was %q, expected %q", st.Code, codes.DataLoss))
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})

	qunit.AsyncTest("Bi-directional streaming call after send grpc error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			srv, err := c.PingBidiStreamError(context.Background())
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected PingBidiStreamError error seen: "+st.Message)
				return
			}

			err = srv.CloseSend()
			if err != nil {
				st := status.FromError(err)
				qunit.Ok(false, "Unexpected Send error seen:"+st.Message)
				return
			}

			_, err = srv.Recv()
			if err == nil {
				qunit.Ok(false, "Unexpected nil error")
				return
			}

			st := status.FromError(err)
			if st.Code != codes.Internal {
				qunit.Ok(false, fmt.Sprintf("Unexpected code in error, was %q, expected %q", st.Code, codes.Internal))
				return
			}
			if st.Message != "error" {
				qunit.Ok(false, fmt.Sprintf("Unexpected message in error, was %q, expected %q", st.Message, "error"))
				return
			}

			qunit.Ok(true, "Request succeeded")
		}()

		return nil
	})
}

func main() {
	defer recoverer.Recover() // recovers any panics and fails tests

	typeTests()
	serverTests("HTTP2", shared.HTTP2Server, shared.EmptyHTTP2Server)
	serverTests("HTTP1", shared.HTTP1Server, shared.EmptyHTTP1Server)
	bidiServerTests(shared.HTTP2Server)

	// protoc-gen-gopherjs tests
	gentest.GenTypesTest()

	// grpcweb metadata tests
	metatest.MetadataTest()

	// grpcweb tests
	grpctest.GRPCWebTest()
}
