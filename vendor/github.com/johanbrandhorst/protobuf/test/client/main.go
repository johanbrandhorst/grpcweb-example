package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/rusco/qunit"

	"github.com/johanbrandhorst/protobuf/test/client/proto/test"
)

//go:generate gopherjs build main.go -m -o html/index.js

func main() {
	qunit.Module("grpcweb")

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
