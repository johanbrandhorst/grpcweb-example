package test

import (
	"github.com/johanbrandhorst/protobuf/grpcweb"
	"github.com/rusco/qunit"

	"github.com/johanbrandhorst/protobuf/test/recoverer"
)

// This test is imported and run by the root level tests

func GRPCWebTest() {
	defer recoverer.Recover() // recovers any panics and fails tests

	qunit.Module("gRPC-Web tests")

	qunit.Test("Creating a new client", func(assert qunit.QUnitAssert) {
		qunit.Expect(1)

		c := grpcweb.NewClient("bla", "bla")
		assert.NotEqual(c, nil, "NewClient creates a non-nil client")
	})
}
