package test

import (
	"github.com/rusco/qunit"
	gmd "google.golang.org/grpc/metadata"

	"github.com/johanbrandhorst/protobuf/grpcweb/metadata"
	"github.com/johanbrandhorst/protobuf/test/recoverer"
)

// This test is imported and run by the root level tests

func MetadataTest() {
	defer recoverer.Recover() // recovers any panics and fails tests

	qunit.Module("Metadata tests")

	qunit.Test("Creating a new metadata type", func(assert qunit.QUnitAssert) {
		qunit.Expect(1)

		h := metadata.New(nil)
		assert.Equal(h.MD.Len(), 0, "Len of an empty browserheader is 0")
	})

	qunit.Test("Creating a new metadata type with metadata", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		h := metadata.New(gmd.Pairs("one", "1", "two", "2", "one", "11"))
		assert.Equal(h.MD.Len(), 2, "Len is 2")
		assert.Equal(len(h.MD["one"]), 2, `Size of value of key "one" is 2`)
		assert.Equal(h.MD["one"][0], "1", `First value of "one" is "1"`)
		assert.Equal(h.MD["one"][1], "11", `Second value of "one" is "11"`)
		assert.Equal(len(h.MD["two"]), 1, `Size of value of key "two" is 1`)
		assert.Equal(h.MD["two"][0], "2", `Value of "two" is "2"`)
	})
}
