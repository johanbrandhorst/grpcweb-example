package main

import (
	"context"
	"fmt"
	"io"
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
	"github.com/johanbrandhorst/protobuf/ptypes/empty"
	"github.com/johanbrandhorst/protobuf/test/client/proto/test"
	"github.com/johanbrandhorst/protobuf/test/recoverer"
	"github.com/johanbrandhorst/protobuf/test/shared"
)

//go:generate gopherjs build main.go -m -o html/index.js

var uri = strings.TrimSuffix(dom.GetWindow().Document().BaseURI(), shared.GopherJSServer+"/")

func typeTests() {
	qunit.Module("Integration Types tests")

	qunit.Test("Simple type factory", func(assert qunit.QUnitAssert) {
		qunit.Expect(8)

		req := new(test.PingRequest).New("1234", 10, 1, test.PingRequest_CODE, true, true, true, 100)
		assert.Equal(req.GetValue(), "1234", "Value is set as expected")
		assert.Equal(req.GetResponseCount(), 10, "ResponseCount is set as expected")
		assert.Equal(req.GetErrorCodeReturned(), 1, "ErrorCodeReturned is set as expected")
		assert.Equal(req.GetFailureType(), test.PingRequest_CODE, "ErrorCodeReturned is set as expected")
		assert.Equal(req.GetCheckMetadata(), true, "CheckMetadata is set as expected")
		assert.Equal(req.GetSendHeaders(), true, "SendHeaders is set as expected")
		assert.Equal(req.GetSendTrailers(), true, "SendTrailers is set as expected")
		assert.Equal(req.GetMessageLatencyMs(), 100, "MessageLatencyMs is set as expected")
	})

	qunit.Test("Complex type factory", func(assert qunit.QUnitAssert) {
		qunit.Expect(7)

		es := new(test.ExtraStuff).New(
			map[int32]string{1234: "The White House", 5678: "The Empire State Building"},
			&test.ExtraStuff_FirstName{FirstName: "Allison"},
			[]uint32{1234, 5678})
		addrs := es.GetAddresses()
		assert.Equal(addrs[1234], "The White House", "Address 1234 is set as expected")
		assert.Equal(addrs[5678], "The Empire State Building", "Address 5678 is set as expected")
		crdnrs := es.GetCardNumbers()
		assert.Equal(crdnrs[0], 1234, "CardNumber #1 is set as expected")
		assert.Equal(crdnrs[1], 5678, "CardNumber #2 is set as expected")
		assert.Equal(es.GetFirstName(), "Allison", "FirstName is set as expected")
		assert.Equal(es.GetIdNumber(), 0, "IdNumber is not set, as expected")
		if _, ok := es.GetTitle().(*test.ExtraStuff_FirstName); !ok {
			assert.Ok(false, "GetTitle did not return a struct of type *test.ExtraStuff_FirstName as expected")
		} else {
			assert.Ok(true, "GetTitle did return a struct of type *test.ExtraStuff_FirstName as expected")
		}
	})

	qunit.Test("Simple setters and getters", func(assert qunit.QUnitAssert) {
		qunit.Expect(16)

		req := &test.PingRequest{
			Object: js.Global.Get("proto").Get("test").Get("PingRequest").New(),
		}
		assert.Equal(req.GetCheckMetadata(), false, "CheckMetadata was unset")
		req.SetCheckMetadata(true)
		assert.Equal(req.GetCheckMetadata(), true, "CheckMetadata was set correctly")

		assert.Equal(req.GetErrorCodeReturned(), 0, "ErrorCodeReturned was unset")
		req.SetErrorCodeReturned(1)
		assert.Equal(req.GetErrorCodeReturned(), 1, "ErrorCodeReturned was set correctly")

		assert.Equal(req.GetFailureType(), test.PingRequest_NONE, "FailureType was unset")
		req.SetFailureType(test.PingRequest_DROP)
		assert.Equal(req.GetFailureType(), test.PingRequest_DROP, "FailureType was set correctly")

		assert.Equal(req.GetMessageLatencyMs(), 0, "MessageLatencyMs was unset")
		req.SetMessageLatencyMs(1)
		assert.Equal(req.GetMessageLatencyMs(), 1, "MessageLatencyMs was set correctly")

		assert.Equal(req.GetResponseCount(), 0, "ResponseCount was unset")
		req.SetResponseCount(1)
		assert.Equal(req.GetResponseCount(), 1, "ResponseCount was set correctly")

		assert.Equal(req.GetSendHeaders(), false, "SendHeaders was unset")
		req.SetSendHeaders(true)
		assert.Equal(req.GetSendHeaders(), true, "SendHeaders was set correctly")

		assert.Equal(req.GetSendTrailers(), false, "SendTrailers was unset")
		req.SetSendTrailers(true)
		assert.Equal(req.GetSendTrailers(), true, "SendTrailers was set correctly")

		assert.Equal(req.GetValue(), "", "Value was unset")
		req.SetValue("something")
		assert.Equal(req.GetValue(), "something", "Value was set correctly")
	})

	qunit.Test("Map getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		req := &test.ExtraStuff{
			Object: js.Global.Get("proto").Get("test").Get("ExtraStuff").New(),
		}
		assert.Equal(len(req.GetAddresses()), 0, "Addresses was unset")
		req.SetAddresses(map[int32]string{
			1234: "The White House",
			5678: "The Empire State Building",
		})
		addrs := req.GetAddresses()
		assert.Equal(len(addrs), 2, "Addresses was the correct size")
		assert.Equal(addrs[1234], "The White House", "Address 1234 was set correctly")
		assert.Equal(addrs[5678], "The Empire State Building", "Address 5678 was set correctly")

		req.ClearAddresses()
		assert.Equal(len(req.GetAddresses()), 0, "Addresses aws unset")
	})

	qunit.Test("Array getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		req := &test.ExtraStuff{
			Object: js.Global.Get("proto").Get("test").Get("ExtraStuff").New(),
		}
		assert.Equal(len(req.GetCardNumbers()), 0, "CardNumbers was unset")
		req.SetCardNumbers([]uint32{
			1234,
			5678,
		})
		crdnrs := req.GetCardNumbers()
		assert.Equal(len(crdnrs), 2, "CardNumbers was the correct size")
		assert.Equal(crdnrs[0], 1234, "CardNumber #1 was set correctly")
		assert.Equal(crdnrs[1], 5678, "CardNumber #2 was set correctly")

		req.ClearCardNumbers()
		assert.Equal(len(req.GetCardNumbers()), 0, "CardNumbers was unset")
	})

	qunit.Test("Oneof getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(20)

		req := &test.ExtraStuff{
			Object: js.Global.Get("proto").Get("test").Get("ExtraStuff").New(),
		}
		assert.Equal(req.GetTitle(), nil, "Title was unset")
		assert.Equal(req.GetFirstName(), "", "FirstName was unset")
		assert.Equal(req.GetIdNumber(), 0, "IdNumber was unset")
		assert.Equal(req.HasFirstName(), false, "HasFirstName was false")
		assert.Equal(req.HasIdNumber(), false, "HasIdNumber was false")

		req.SetTitle(&test.ExtraStuff_FirstName{FirstName: "Allison"})
		fn, ok := req.GetTitle().(*test.ExtraStuff_FirstName)
		if !ok {
			assert.Ok(false, "Title was not of type *test.ExtraStuff_FirstName")
		} else {
			assert.Ok(true, "Title was of type *test.ExtraStuff_FirstName")
			assert.Equal(fn.FirstName, "Allison", "Title FirstName was set correctly")
			assert.Equal(req.GetFirstName(), "Allison", "FirstName was set correctly")
			assert.Equal(req.GetIdNumber(), 0, "IdNumber was still unset")
			assert.Equal(req.HasFirstName(), true, "HasFirstName was true")
			assert.Equal(req.HasIdNumber(), false, "HasIdNumber was false")
		}

		req.SetIdNumber(100)
		id, ok := req.GetTitle().(*test.ExtraStuff_IdNumber)
		if !ok {
			assert.Ok(false, "Title was not of type *test.ExtraStuff_IdNumber")
		} else {
			assert.Ok(true, "Title was of type *test.ExtraStuff_IdNumber")
			assert.Equal(id.IdNumber, 100, "Title IdNumber was set correctly")
			assert.Equal(req.GetFirstName(), "", "FirstName was unset")
			assert.Equal(req.GetIdNumber(), 100, "IdNumber was set correctly")
			assert.Equal(req.HasIdNumber(), true, "HasIdNumber was true")
			assert.Equal(req.HasFirstName(), false, "HasFirstName was false")
		}

		req.ClearIdNumber()
		assert.Equal(req.GetTitle(), nil, "Title was unset")
		assert.Equal(req.HasFirstName(), false, "HasFirstName was false")
		assert.Equal(req.HasIdNumber(), false, "HasIdNumber was false")
	})
}

func serverTests(label, serverAddr, emptyServerAddr string) {
	qunit.Module(fmt.Sprintf("%s Integration tests", label))

	qunit.AsyncTest("Unary server call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			req := new(test.PingRequest).New(
				"test", 1, 0, test.PingRequest_NONE, false, false, false, 0)
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

			req := new(test.PingRequest).New(
				"test", 1, 0, test.PingRequest_NONE, true, false, false, 0)
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

			req := new(test.PingRequest).New(
				"test", 1, 0, test.PingRequest_NONE, false, true, true, 0)
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

			req := new(test.PingRequest).New(
				"test", 1, 0, test.PingRequest_NONE, false, true, false, 0)
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

			req := new(test.PingRequest).New(
				"test", 1, 0, test.PingRequest_NONE, false, false, true, 0)
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

			req := new(test.PingRequest).New(
				"", 0, uint32(codes.InvalidArgument), test.PingRequest_CODE, false, false, false, 0)
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

			req := new(test.PingRequest).New(
				"", 0, 0, test.PingRequest_DROP, false, false, false, 0)
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

	qunit.AsyncTest("Streaming server call", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_NONE, false, false, false, 1)
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

	qunit.AsyncTest("Streaming server call with metadata", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_NONE, true, false, false, 1)
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

	qunit.AsyncTest("Streaming server call expecting headers and trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_NONE, false, true, true, 1)
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

	qunit.AsyncTest("Streaming server call expecting only headers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_NONE, false, true, false, 1)
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

	qunit.AsyncTest("Streaming server call expecting only trailers", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_NONE, false, false, true, 1)
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

	qunit.AsyncTest("Streaming server call returning network error", func() interface{} {
		c := test.NewTestServiceClient(uri + serverAddr)

		go func() {
			defer recoverer.Recover() // recovers any panics and fails tests
			defer qunit.Start()

			// Send 20 messages with 1ms wait before each
			req := new(test.PingRequest).New(
				"test", 20, 0, test.PingRequest_DROP, false, false, false, 1)
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

			_, err := c.PingEmpty(context.Background(), new(empty.Empty).New())
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
}

func main() {
	defer recoverer.Recover() // recovers any panics and fails tests

	typeTests()
	serverTests("HTTP2", shared.HTTP2Server, shared.EmptyHTTP2Server)
	serverTests("HTTP1", shared.HTTP1Server, shared.EmptyHTTP1Server)

	// protoc-gen-gopherjs tests
	gentest.GenTypesTest()

	// grpcweb metadata tests
	metatest.MetadataTest()

	// grpcweb tests
	grpctest.GRPCWebTest()
}
