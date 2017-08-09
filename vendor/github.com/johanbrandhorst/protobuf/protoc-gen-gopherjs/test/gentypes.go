package test

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/rusco/qunit"

	test "github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/grpc_test"
	"github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi"
	"github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/types"
	"github.com/johanbrandhorst/protobuf/test/recoverer"
)

// GenTypesTest is imported and run by the root level tests
func GenTypesTest() {
	defer recoverer.Recover() // recovers any panics and fails tests

	qunit.Module("GopherJS Protobuf Generator tests")

	gRPCTypeTests()

	singleTypeTests()

	repeatedTypeTests()

	oneofTypeTests()

	mapTypeTests()
}

func gRPCTypeTests() {
	qunit.Test("Simple type factory", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := new(test.Simple).New(1234, 1.5, test.Days_MONDAY, "Alfred")
		assert.Equal(t.GetKey(), 1234, "New sets Key to 1234")
		assert.Equal(t.GetDeadline(), 1.5, "New sets Deadline to 1.5")
		assert.Equal(t.GetDay(), test.Days_MONDAY, "New sets Day to MONDAY")
		assert.Equal(t.GetName(), "Alfred", "New sets Name to Alfred")
	})

	qunit.Test("Simple setters and getters", func(assert qunit.QUnitAssert) {
		qunit.Expect(8)

		t := &test.Simple{
			Object: js.Global.Get("proto").Get("my").Get("test").Get("Simple").New(),
		}
		assert.Equal(t.GetKey(), 0, "Key is unset")
		t.SetKey(1234)
		assert.Equal(t.GetKey(), 1234, "SetKey sets Key to 1234")

		assert.Equal(t.GetDeadline(), 0, "Deadline is unset")
		t.SetDeadline(1.5)
		assert.Equal(t.GetDeadline(), 1.5, "SetDeadline sets Deadline to 1.5")

		assert.Equal(t.GetDay(), test.Days(0), "Day is unset")
		t.SetDay(test.Days_TUESDAY)
		assert.Equal(t.GetDay(), test.Days_TUESDAY, "SetDay sets Day to TUESDAY")

		assert.Equal(t.GetName(), "", "Name is unset")
		t.SetName("Alfred")
		assert.Equal(t.GetName(), "Alfred", "SetName sets Name to Alfred")
	})

	qunit.Test("Complex type factory", func(assert qunit.QUnitAssert) {
		qunit.Expect(15)

		t := new(test.Complex).New(
			[]*test.Complex_Communique{
				new(test.Complex_Communique).New(false, &test.Complex_Communique_Delta_{Delta: 1234}),
				new(test.Complex_Communique).New(true, &test.Complex_Communique_Today{Today: test.Days_TUESDAY}),
			},
			map[int32]string{1234: "The White House", 5678: "The Empire State Building"},
			new(multitest.Multi1).New(
				new(multitest.Multi2).New(2345, multitest.Multi2_BLUE),
				multitest.Multi2_RED,
				multitest.Multi3_FEZ),
		)
		assert.Equal(t.HasMulti(), true, "New populates the Multi")
		assert.NotEqual(t.GetMulti(), nil, "New sets the Multi")
		assert.Equal(t.GetMulti().GetColor(), multitest.Multi2_RED, "New sets the color of the Multi to RED")
		assert.Equal(t.GetMulti().GetHatType(), multitest.Multi3_FEZ, "New sets the hat type of the Multi to FEZ")
		assert.Equal(t.GetMulti().GetMulti2().GetColor(), multitest.Multi2_BLUE, "New sets the color of the Multi.Multi2 to BLUE")
		assert.Equal(t.GetMulti().GetMulti2().GetRequiredValue(), 2345, "New sets the color of the Multi.Multi2 to BLUE")
		ck := t.GetCompactKeys()
		assert.Equal(ck[1234], "The White House", `New sets key 1234 to "The White House"`)
		assert.Equal(ck[5678], "The Empire State Building", `New sets key 5678 to "The Empire State Building"`)
		l := t.GetCommunique()
		assert.Equal(len(l), 2, "New adds two elements to the communique")
		assert.Equal(l[0].GetMakeMeCry(), false, "New sets MakeMeCry on the first element")
		assert.Equal(l[0].GetDelta(), 1234, "First element of the communique is 1234")
		assert.Equal(l[1].GetMakeMeCry(), true, "New sets MakeMeCry on the second element")
		assert.Equal(l[1].GetToday(), test.Days_TUESDAY, "Second element of the communique is TUESDAY")
		_, ok := l[0].GetUnionThing().(*test.Complex_Communique_Delta_)
		assert.Ok(ok, "First element of the communique is a Delta")
		_, ok = l[1].GetUnionThing().(*test.Complex_Communique_Today)
		assert.Ok(ok, "Second element of the communique is a Today")
	})

	qunit.Test("Map getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &test.Complex{
			Object: js.Global.Get("proto").Get("my").Get("test").Get("Complex").New(),
		}
		assert.Equal(len(t.GetCompactKeys()), 0, "CompactKeys is unset")
		t.SetCompactKeys(map[int32]string{
			1234: "The White House",
			5678: "The Empire State Building",
		})
		ck := t.GetCompactKeys()
		assert.Equal(len(ck), 2, "SetCompactKeys sets CompactKeys to the correct size")
		assert.Equal(ck[1234], "The White House", `SetCompactKeys sets key 1234 to "The White House"`)
		assert.Equal(ck[5678], "The Empire State Building", `SetCompactKeys sets key 5678 to "The Empire State Building"`)

		t.ClearCompactKeys()
		assert.Equal(len(t.GetCompactKeys()), 0, "ClearCompactKeys removes all entries in the map")
	})

	qunit.Test("Array getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(7)

		t := &test.Complex{
			Object: js.Global.Get("proto").Get("my").Get("test").Get("Complex").New(),
		}
		assert.Equal(len(t.GetCommunique()), 0, "Communique is unset")
		t.SetCommunique([]*test.Complex_Communique{
			new(test.Complex_Communique).New(false, &test.Complex_Communique_Height{Height: 1.5}),
			new(test.Complex_Communique).New(true, &test.Complex_Communique_Today{Today: test.Days_TUESDAY}),
		})
		comm := t.GetCommunique()
		assert.Equal(len(comm), 2, "Communique was the correct size")
		assert.Equal(comm[0].GetHeight(), 1.5, "Communique #1 was set correctly")
		assert.Equal(comm[1].GetToday(), test.Days_TUESDAY, "Communique #2 was set correctly")

		t.AddCommunique(
			new(test.Complex_Communique).New(false, &test.Complex_Communique_Name{Name: "Daisy"}),
			2,
		)
		comm = t.GetCommunique()
		assert.Equal(len(comm), 3, "Communique was the correct size")
		assert.Equal(comm[2].GetName(), "Daisy", "Communique #2 was set correctly")

		t.ClearCommunique()
		assert.Equal(len(t.GetCommunique()), 0, "Communique was unset")
	})

	qunit.Test("Oneof getters and setters", func(assert qunit.QUnitAssert) {
		qunit.Expect(35)

		t := new(test.Complex_Communique).New(false, nil)
		assert.Equal(t.GetMakeMeCry(), false, "MakeMeCry is unset")
		assert.Equal(len(t.GetData()), 0, "Data is unset")
		assert.Equal(t.GetDelta(), 0, "Delta is unset")
		assert.Equal(t.GetHeight(), 0, "Height is unset")
		assert.Equal(t.GetMaybe(), false, "Maybe is unset")
		assert.Equal(t.GetName(), "", "Name is unset")
		assert.Equal(t.GetNumber(), 0, "Number is unset")
		assert.Equal(t.GetTempC(), 0, "TempC is unset")
		assert.Equal(t.GetToday(), test.Days(0), "Today is unset")
		assert.Equal(t.GetUnionThing(), nil, "UnionThing is unset")

		t.SetUnionThing(&test.Complex_Communique_TempC{TempC: 1.5})
		tc, ok := t.GetUnionThing().(*test.Complex_Communique_TempC)
		assert.Ok(ok, "SetUnionThing sets UnionThing to TempC")
		assert.Equal(tc.TempC, 1.5, "SetUnionThing sets UnionThing.TempC correctly")
		assert.Equal(t.GetTempC(), 1.5, "SetUnionThing sets UnionThing.TempC correctly")
		assert.Equal(t.HasTempC(), true, "SetUnionThing sets UnionThing.TempC correctly")
		assert.Equal(len(t.GetData()), 0, "Data is unset")
		assert.Equal(t.GetDelta(), 0, "Delta is unset")
		assert.Equal(t.GetHeight(), 0, "Height is unset")
		assert.Equal(t.GetMaybe(), false, "Maybe is unset")
		assert.Equal(t.GetName(), "", "Name is unset")
		assert.Equal(t.GetNumber(), 0, "Number is unset")
		assert.Equal(t.GetToday(), test.Days(0), "Today is unset")

		t.SetMaybe(true)
		mb, ok := t.GetUnionThing().(*test.Complex_Communique_Maybe)
		assert.Ok(ok, "SetUnionThing sets UnionThing to Maybe")
		assert.Equal(mb.Maybe, true, "SetUnionThing sets UnionThing.Maybe correctly")
		assert.Equal(t.GetMaybe(), true, "SetUnionThing sets UnionThing.Maybe correctly")
		assert.Equal(t.HasMaybe(), true, "SetUnionThing sets UnionThing.Maybe correctly")
		assert.Equal(len(t.GetData()), 0, "Data is unset")
		assert.Equal(t.GetDelta(), 0, "Delta is unset")
		assert.Equal(t.GetHeight(), 0, "Height is unset")
		assert.Equal(t.GetName(), "", "Name is unset")
		assert.Equal(t.GetNumber(), 0, "Number is unset")
		assert.Equal(t.GetTempC(), 0, "SetMaybe clears TempC")
		assert.Equal(t.HasTempC(), false, "SetMaybe clears TempC")
		assert.Equal(t.GetToday(), test.Days(0), "Today is unset")

		t.ClearMaybe()
		assert.Equal(t.GetUnionThing(), nil, "ClearMaybe clears UnionThing")
		assert.Equal(t.HasMaybe(), false, "ClearMaybe clears Maybe")
	})
}

func singleTypeTests() {
	qunit.Test("SingleInt32 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleInt32 := t.GetSingleInt32()
		assert.Equal(singleInt32, 0, "SingleInt32 is 0")
		singleInt32 = 10
		t.SetSingleInt32(singleInt32)
		singleInt32 = t.GetSingleInt32()
		assert.Equal(singleInt32, 10, "SetSingleInt32 sets the correct value")
	})

	qunit.Test("SingleInt64 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleInt64 := t.GetSingleInt64()
		assert.Equal(singleInt64, 0, "SingleInt64 is 0")
		singleInt64 = 10
		t.SetSingleInt64(singleInt64)
		singleInt64 = t.GetSingleInt64()
		assert.Equal(singleInt64, 10, "SetSingleInt64 sets the correct value")
	})

	qunit.Test("SingleUint32 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleUint32 := t.GetSingleUint32()
		assert.Equal(singleUint32, 0, "SingleUint32 is 0")
		singleUint32 = 10
		t.SetSingleUint32(singleUint32)
		singleUint32 = t.GetSingleUint32()
		assert.Equal(singleUint32, 10, "SetSingleUint32 sets the correct value")
	})

	qunit.Test("SingleUint64 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleUint64 := t.GetSingleUint64()
		assert.Equal(singleUint64, 0, "SingleUint64 is 0")
		singleUint64 = 10
		t.SetSingleUint64(singleUint64)
		singleUint64 = t.GetSingleUint64()
		assert.Equal(singleUint64, 10, "SetSingleUint64 sets the correct value")
	})

	qunit.Test("SingleSint32 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleSint32 := t.GetSingleSint32()
		assert.Equal(singleSint32, 0, "SingleSint32 is 0")
		singleSint32 = 10
		t.SetSingleSint32(singleSint32)
		singleSint32 = t.GetSingleSint32()
		assert.Equal(singleSint32, 10, "SetSingleSint32 sets the correct value")
	})

	qunit.Test("SingleSint64 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleSint64 := t.GetSingleSint64()
		assert.Equal(singleSint64, 0, "SingleSint64 is 0")
		singleSint64 = 10
		t.SetSingleSint64(singleSint64)
		singleSint64 = t.GetSingleSint64()
		assert.Equal(singleSint64, 10, "SetSingleSint64 sets the correct value")
	})

	qunit.Test("SingleFixed32 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleFixed32 := t.GetSingleFixed32()
		assert.Equal(singleFixed32, 0, "SingleFixed32 is 0")
		singleFixed32 = 10
		t.SetSingleFixed32(singleFixed32)
		singleFixed32 = t.GetSingleFixed32()
		assert.Equal(singleFixed32, 10, "SetSingleFixed32 sets the correct value")
	})

	qunit.Test("SingleFixed64 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleFixed64 := t.GetSingleFixed64()
		assert.Equal(singleFixed64, 0, "SingleFixed64 is 0")
		singleFixed64 = 10
		t.SetSingleFixed64(singleFixed64)
		singleFixed64 = t.GetSingleFixed64()
		assert.Equal(singleFixed64, 10, "SetSingleFixed64 sets the correct value")
	})

	qunit.Test("SingleSfixed32 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleSfixed32 := t.GetSingleSfixed32()
		assert.Equal(singleSfixed32, 0, "SingleSfixed32 is 0")
		singleSfixed32 = 10
		t.SetSingleSfixed32(singleSfixed32)
		singleSfixed32 = t.GetSingleSfixed32()
		assert.Equal(singleSfixed32, 10, "SetSingleSfixed32 sets the correct value")
	})

	qunit.Test("SingleSfixed64 Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleSfixed64 := t.GetSingleSfixed64()
		assert.Equal(singleSfixed64, 0, "SingleSfixed64 is 0")
		singleSfixed64 = 10
		t.SetSingleSfixed64(singleSfixed64)
		singleSfixed64 = t.GetSingleSfixed64()
		assert.Equal(singleSfixed64, 10, "SetSingleSfixed64 sets the correct value")
	})

	qunit.Test("SingleFloat Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleFloat := t.GetSingleFloat()
		assert.Equal(singleFloat, 0, "SingleFloat is 0")
		singleFloat = 1.5
		t.SetSingleFloat(singleFloat)
		singleFloat = t.GetSingleFloat()
		assert.Equal(singleFloat, 1.5, "SetSingleFloat sets the correct value")
	})

	qunit.Test("SingleDouble Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleDouble := t.GetSingleDouble()
		assert.Equal(singleDouble, 0, "SingleDouble is 0")
		singleDouble = 1.5
		t.SetSingleDouble(singleDouble)
		singleDouble = t.GetSingleDouble()
		assert.Equal(singleDouble, 1.5, "SetSingleDouble sets the correct value")
	})

	qunit.Test("SingleBool Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleBool := t.GetSingleBool()
		assert.Equal(singleBool, false, "SingleBool is false")
		singleBool = true
		t.SetSingleBool(singleBool)
		singleBool = t.GetSingleBool()
		assert.Equal(singleBool, true, "SetSingleBool sets the correct value")
	})

	qunit.Test("SingleString Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleString := t.GetSingleString()
		assert.Equal(singleString, "", "SingleString is empty")
		singleString = "a string"
		t.SetSingleString(singleString)
		singleString = t.GetSingleString()
		assert.Equal(singleString, "a string", "SetSingleString sets the correct value")
	})

	qunit.Test("SingleBytes Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleBytes := t.GetSingleBytes()
		assert.Equal(len(singleBytes), 0, "Size of singleBytes is 0")
		singleBytes = []byte{0x01, 0x02}
		t.SetSingleBytes(singleBytes)
		singleBytes = t.GetSingleBytes()
		assert.Equal(len(singleBytes), 2, "SetSingleBytes adds the correct number of elements")
		// Can't compare against []byte literal
		assert.Equal(singleBytes[0], 0x01, "SetSingleBytes sets the value")
		assert.Equal(singleBytes[1], 0x02, "SetSingleBytes sets the value")
	})

	qunit.Test("SingleImportedMessage Get, Set, Has, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(8)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleImportedMessage := t.GetSingleImportedMessage()
		assert.Equal(t.HasSingleImportedMessage(), false, "SingleImportedMessage is unset")
		assert.Equal(singleImportedMessage, nil, "SingleImportedMessage is unset")
		singleImportedMessage = new(multitest.Multi1).New(
			nil,
			multitest.Multi2_BLUE,
			multitest.Multi3_FEDORA,
		)
		t.SetSingleImportedMessage(singleImportedMessage)
		assert.Equal(t.HasSingleImportedMessage(), true, "SetSingleImportedMessage sets the message")
		singleImportedMessage = t.GetSingleImportedMessage()
		assert.Equal(singleImportedMessage.GetColor(), multitest.Multi2_BLUE, "SetSingleImportedMessage sets the correct color")
		assert.Equal(singleImportedMessage.GetHatType(), multitest.Multi3_FEDORA, "SetSingleImportedMessage sets the correct hat")
		assert.Equal(singleImportedMessage.GetMulti2(), nil, "SetSingleImportedMessage sets the correct multi2")
		t.ClearSingleImportedMessage()
		assert.Equal(t.HasSingleImportedMessage(), false, "ClearSingleImportedMessage clears the message")
		assert.Equal(t.GetSingleImportedMessage(), nil, "ClearSingleImportedMessage clears the message")
	})

	qunit.Test("SingleNestedMessage Get, Set, Has, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleNestedMessage := t.GetSingleNestedMessage()
		assert.Equal(t.HasSingleNestedMessage(), false, "SingleNestedMessage is unset")
		assert.Equal(singleNestedMessage, nil, "SingleNestedMessage is unset")
		singleNestedMessage = new(types.TestAllTypes_NestedMessage).New(
			10,
		)
		t.SetSingleNestedMessage(singleNestedMessage)
		assert.Equal(t.HasSingleNestedMessage(), true, "SetSingleNestedMessage sets the message")
		singleNestedMessage = t.GetSingleNestedMessage()
		assert.Equal(singleNestedMessage.GetB(), 10, "SetSingleNestedMessage sets the correct b")
		t.ClearSingleNestedMessage()
		assert.Equal(t.HasSingleNestedMessage(), false, "ClearSingleNestedMessage clears the message")
		assert.Equal(t.GetSingleNestedMessage(), nil, "ClearSingleNestedMessage clears the message")
	})

	qunit.Test("SingleForeignMessage Get, Set, Has, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleForeignMessage := t.GetSingleForeignMessage()
		assert.Equal(t.HasSingleForeignMessage(), false, "SingleForeignMessage is unset")
		assert.Equal(singleForeignMessage, nil, "SingleForeignMessage is unset")
		singleForeignMessage = new(types.ForeignMessage).New(
			10,
		)
		t.SetSingleForeignMessage(singleForeignMessage)
		assert.Equal(t.HasSingleForeignMessage(), true, "SetSingleForeignMessage sets the message")
		singleForeignMessage = t.GetSingleForeignMessage()
		assert.Equal(singleForeignMessage.GetC(), 10, "SetSingleForeignMessage sets the correct b")
		t.ClearSingleForeignMessage()
		assert.Equal(t.HasSingleForeignMessage(), false, "ClearSingleForeignMessage clears the message")
		assert.Equal(t.GetSingleForeignMessage(), nil, "ClearSingleForeignMessage clears the message")
	})

	qunit.Test("SingleNestedEnum Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleNestedEnum := t.GetSingleNestedEnum()
		assert.Equal(singleNestedEnum, types.TestAllTypes_NESTED_ENUM_UNSPECIFIED, "SingleNestedEnum is unset")
		singleNestedEnum = types.TestAllTypes_BAR
		t.SetSingleNestedEnum(singleNestedEnum)
		singleNestedEnum = t.GetSingleNestedEnum()
		assert.Equal(singleNestedEnum, types.TestAllTypes_BAR, "SetSingleNestedEnum sets the value")
	})

	qunit.Test("SingleForeignEnum Get, Set", func(assert qunit.QUnitAssert) {
		qunit.Expect(2)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}
		singleForeignEnum := t.GetSingleForeignEnum()
		assert.Equal(singleForeignEnum, types.ForeignEnum_FOREIGN_UNSPECIFIED, "SingleForeignEnum is unset")
		singleForeignEnum = types.ForeignEnum_FOREIGN_FOO
		t.SetSingleForeignEnum(singleForeignEnum)
		singleForeignEnum = t.GetSingleForeignEnum()
		assert.Equal(singleForeignEnum, types.ForeignEnum_FOREIGN_FOO, "SetSingleForeignEnum sets the value")
	})
}

func repeatedTypeTests() {
	qunit.Test("RepeatedInt32 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedInt32 := t.GetRepeatedInt32()
		assert.Equal(len(repeatedInt32), 0, "Size of repeatedInt32 is 0")
		repeatedInt32 = append(repeatedInt32, 1)
		t.SetRepeatedInt32(repeatedInt32)
		repeatedInt32 = t.GetRepeatedInt32()
		assert.Equal(len(repeatedInt32), 1, "SetRepeatedInt32 adds the correct number of elements")
		assert.Equal(repeatedInt32[0], 1, "SetRepeatedInt32 sets the slices value")
		t.ClearRepeatedInt32()
		repeatedInt32 = t.GetRepeatedInt32()
		assert.Equal(len(repeatedInt32), 0, "ClearRepeatedInt32 clears all entries in the slice")
		t.AddRepeatedInt32(2, 0)
		repeatedInt32 = t.GetRepeatedInt32()
		assert.Equal(repeatedInt32[0], 2, "AddRepeatedInt32 appends a value to the slice")
	})

	qunit.Test("RepeatedInt64 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedInt64 := t.GetRepeatedInt64()
		assert.Equal(len(repeatedInt64), 0, "Size of repeatedInt64 is 0")
		repeatedInt64 = append(repeatedInt64, 1)
		t.SetRepeatedInt64(repeatedInt64)
		repeatedInt64 = t.GetRepeatedInt64()
		assert.Equal(len(repeatedInt64), 1, "SetRepeatedInt64 adds the correct number of elements")
		assert.Equal(repeatedInt64[0], 1, "SetRepeatedInt64 sets the slices value")
		t.ClearRepeatedInt64()
		repeatedInt64 = t.GetRepeatedInt64()
		assert.Equal(len(repeatedInt64), 0, "ClearRepeatedInt64 clears all entries in the slice")
		t.AddRepeatedInt64(2, 0)
		repeatedInt64 = t.GetRepeatedInt64()
		assert.Equal(repeatedInt64[0], 2, "AddRepeatedInt64 appends a value to the slice")
	})

	qunit.Test("RepeatedUint32 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedUint32 := t.GetRepeatedUint32()
		assert.Equal(len(repeatedUint32), 0, "Size of repeatedUint32 is 0")
		repeatedUint32 = append(repeatedUint32, 1)
		t.SetRepeatedUint32(repeatedUint32)
		repeatedUint32 = t.GetRepeatedUint32()
		assert.Equal(len(repeatedUint32), 1, "SetRepeatedUint32 adds the correct number of elements")
		assert.Equal(repeatedUint32[0], 1, "SetRepeatedUint32 sets the slices value")
		t.ClearRepeatedUint32()
		repeatedUint32 = t.GetRepeatedUint32()
		assert.Equal(len(repeatedUint32), 0, "ClearRepeatedUint32 clears all entries in the slice")
		t.AddRepeatedUint32(2, 0)
		repeatedUint32 = t.GetRepeatedUint32()
		assert.Equal(repeatedUint32[0], 2, "AddRepeatedUint32 appends a value to the slice")
	})

	qunit.Test("RepeatedUint64 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedUint64 := t.GetRepeatedUint64()
		assert.Equal(len(repeatedUint64), 0, "Size of repeatedUint64 is 0")
		repeatedUint64 = append(repeatedUint64, 1)
		t.SetRepeatedUint64(repeatedUint64)
		repeatedUint64 = t.GetRepeatedUint64()
		assert.Equal(len(repeatedUint64), 1, "SetRepeatedUint64 adds the correct number of elements")
		assert.Equal(repeatedUint64[0], 1, "SetRepeatedUint64 sets the slices value")
		t.ClearRepeatedUint64()
		repeatedUint64 = t.GetRepeatedUint64()
		assert.Equal(len(repeatedUint64), 0, "ClearRepeatedUint64 clears all entries in the slice")
		t.AddRepeatedUint64(2, 0)
		repeatedUint64 = t.GetRepeatedUint64()
		assert.Equal(repeatedUint64[0], 2, "AddRepeatedUint64 appends a value to the slice")
	})

	qunit.Test("RepeatedSint32 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedSint32 := t.GetRepeatedSint32()
		assert.Equal(len(repeatedSint32), 0, "Size of repeatedSint32 is 0")
		repeatedSint32 = append(repeatedSint32, 1)
		t.SetRepeatedSint32(repeatedSint32)
		repeatedSint32 = t.GetRepeatedSint32()
		assert.Equal(len(repeatedSint32), 1, "SetRepeatedSint32 adds the correct number of elements")
		assert.Equal(repeatedSint32[0], 1, "SetRepeatedSint32 sets the slices value")
		t.ClearRepeatedSint32()
		repeatedSint32 = t.GetRepeatedSint32()
		assert.Equal(len(repeatedSint32), 0, "ClearRepeatedSint32 clears all entries in the slice")
		t.AddRepeatedSint32(2, 0)
		repeatedSint32 = t.GetRepeatedSint32()
		assert.Equal(repeatedSint32[0], 2, "AddRepeatedSint32 appends a value to the slice")
	})

	qunit.Test("RepeatedSint64 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedSint64 := t.GetRepeatedSint64()
		assert.Equal(len(repeatedSint64), 0, "Size of repeatedSint64 is 0")
		repeatedSint64 = append(repeatedSint64, 1)
		t.SetRepeatedSint64(repeatedSint64)
		repeatedSint64 = t.GetRepeatedSint64()
		assert.Equal(len(repeatedSint64), 1, "SetRepeatedSint64 adds the correct number of elements")
		assert.Equal(repeatedSint64[0], 1, "SetRepeatedSint64 sets the slices value")
		t.ClearRepeatedSint64()
		repeatedSint64 = t.GetRepeatedSint64()
		assert.Equal(len(repeatedSint64), 0, "ClearRepeatedSint64 clears all entries in the slice")
		t.AddRepeatedSint64(2, 0)
		repeatedSint64 = t.GetRepeatedSint64()
		assert.Equal(repeatedSint64[0], 2, "AddRepeatedSint64 appends a value to the slice")
	})
	qunit.Test("RepeatedFixed32 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedFixed32 := t.GetRepeatedFixed32()
		assert.Equal(len(repeatedFixed32), 0, "Size of repeatedFixed32 is 0")
		repeatedFixed32 = append(repeatedFixed32, 1)
		t.SetRepeatedFixed32(repeatedFixed32)
		repeatedFixed32 = t.GetRepeatedFixed32()
		assert.Equal(len(repeatedFixed32), 1, "SetRepeatedFixed32 adds the correct number of elements")
		assert.Equal(repeatedFixed32[0], 1, "SetRepeatedFixed32 sets the slices value")
		t.ClearRepeatedFixed32()
		repeatedFixed32 = t.GetRepeatedFixed32()
		assert.Equal(len(repeatedFixed32), 0, "ClearRepeatedFixed32 clears all entries in the slice")
		t.AddRepeatedFixed32(2, 0)
		repeatedFixed32 = t.GetRepeatedFixed32()
		assert.Equal(repeatedFixed32[0], 2, "AddRepeatedFixed32 appends a value to the slice")
	})

	qunit.Test("RepeatedFixed64 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedFixed64 := t.GetRepeatedFixed64()
		assert.Equal(len(repeatedFixed64), 0, "Size of repeatedFixed64 is 0")
		repeatedFixed64 = append(repeatedFixed64, 1)
		t.SetRepeatedFixed64(repeatedFixed64)
		repeatedFixed64 = t.GetRepeatedFixed64()
		assert.Equal(len(repeatedFixed64), 1, "SetRepeatedFixed64 adds the correct number of elements")
		assert.Equal(repeatedFixed64[0], 1, "SetRepeatedFixed64 sets the slices value")
		t.ClearRepeatedFixed64()
		repeatedFixed64 = t.GetRepeatedFixed64()
		assert.Equal(len(repeatedFixed64), 0, "ClearRepeatedFixed64 clears all entries in the slice")
		t.AddRepeatedFixed64(2, 0)
		repeatedFixed64 = t.GetRepeatedFixed64()
		assert.Equal(repeatedFixed64[0], 2, "AddRepeatedFixed64 appends a value to the slice")
	})

	qunit.Test("RepeatedSfixed32 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedSfixed32 := t.GetRepeatedSfixed32()
		assert.Equal(len(repeatedSfixed32), 0, "Size of repeatedSfixed32 is 0")
		repeatedSfixed32 = append(repeatedSfixed32, 1)
		t.SetRepeatedSfixed32(repeatedSfixed32)
		repeatedSfixed32 = t.GetRepeatedSfixed32()
		assert.Equal(len(repeatedSfixed32), 1, "SetRepeatedSfixed32 adds the correct number of elements")
		assert.Equal(repeatedSfixed32[0], 1, "SetRepeatedSfixed32 sets the slices value")
		t.ClearRepeatedSfixed32()
		repeatedSfixed32 = t.GetRepeatedSfixed32()
		assert.Equal(len(repeatedSfixed32), 0, "ClearRepeatedSfixed32 clears all entries in the slice")
		t.AddRepeatedSfixed32(2, 0)
		repeatedSfixed32 = t.GetRepeatedSfixed32()
		assert.Equal(repeatedSfixed32[0], 2, "AddRepeatedSfixed32 appends a value to the slice")
	})

	qunit.Test("RepeatedSfixed64 Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedSfixed64 := t.GetRepeatedSfixed64()
		assert.Equal(len(repeatedSfixed64), 0, "Size of repeatedSfixed64 is 0")
		repeatedSfixed64 = append(repeatedSfixed64, 1)
		t.SetRepeatedSfixed64(repeatedSfixed64)
		repeatedSfixed64 = t.GetRepeatedSfixed64()
		assert.Equal(len(repeatedSfixed64), 1, "SetRepeatedSfixed64 adds the correct number of elements")
		assert.Equal(repeatedSfixed64[0], 1, "SetRepeatedSfixed64 sets the slices value")
		t.ClearRepeatedSfixed64()
		repeatedSfixed64 = t.GetRepeatedSfixed64()
		assert.Equal(len(repeatedSfixed64), 0, "ClearRepeatedSfixed64 clears all entries in the slice")
		t.AddRepeatedSfixed64(2, 0)
		repeatedSfixed64 = t.GetRepeatedSfixed64()
		assert.Equal(repeatedSfixed64[0], 2, "AddRepeatedSfixed64 appends a value to the slice")
	})

	qunit.Test("RepeatedFloat Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedFloat := t.GetRepeatedFloat()
		assert.Equal(len(repeatedFloat), 0, "Size of repeatedFloat is 0")
		repeatedFloat = append(repeatedFloat, 1.5)
		t.SetRepeatedFloat(repeatedFloat)
		repeatedFloat = t.GetRepeatedFloat()
		assert.Equal(len(repeatedFloat), 1, "SetRepeatedFloat adds the correct number of elements")
		assert.Equal(repeatedFloat[0], 1.5, "SetRepeatedFloat sets the slices value")
		t.ClearRepeatedFloat()
		repeatedFloat = t.GetRepeatedFloat()
		assert.Equal(len(repeatedFloat), 0, "ClearRepeatedFloat clears all entries in the slice")
		t.AddRepeatedFloat(1.5, 0)
		repeatedFloat = t.GetRepeatedFloat()
		assert.Equal(repeatedFloat[0], 1.5, "AddRepeatedFloat appends a value to the slice")
	})

	qunit.Test("RepeatedDouble Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedDouble := t.GetRepeatedDouble()
		assert.Equal(len(repeatedDouble), 0, "Size of repeatedDouble is 0")
		repeatedDouble = append(repeatedDouble, 1.5)
		t.SetRepeatedDouble(repeatedDouble)
		repeatedDouble = t.GetRepeatedDouble()
		assert.Equal(len(repeatedDouble), 1, "SetRepeatedDouble adds the correct number of elements")
		assert.Equal(repeatedDouble[0], 1.5, "SetRepeatedDouble sets the slices value")
		t.ClearRepeatedDouble()
		repeatedDouble = t.GetRepeatedDouble()
		assert.Equal(len(repeatedDouble), 0, "ClearRepeatedDouble clears all entries in the slice")
		t.AddRepeatedDouble(1.5, 0)
		repeatedDouble = t.GetRepeatedDouble()
		assert.Equal(repeatedDouble[0], 1.5, "AddRepeatedDouble appends a value to the slice")
	})

	qunit.Test("RepeatedBool Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedBool := t.GetRepeatedBool()
		assert.Equal(len(repeatedBool), 0, "Size of repeatedBool is 0")
		repeatedBool = append(repeatedBool, true)
		t.SetRepeatedBool(repeatedBool)
		repeatedBool = t.GetRepeatedBool()
		assert.Equal(len(repeatedBool), 1, "SetRepeatedBool adds the correct number of elements")
		assert.Equal(repeatedBool[0], true, "SetRepeatedBool sets the slices value")
		t.ClearRepeatedBool()
		repeatedBool = t.GetRepeatedBool()
		assert.Equal(len(repeatedBool), 0, "ClearRepeatedBool clears all entries in the slice")
		t.AddRepeatedBool(false, 0)
		repeatedBool = t.GetRepeatedBool()
		assert.Equal(repeatedBool[0], false, "AddRepeatedBool appends a value to the slice")
	})

	qunit.Test("RepeatedString Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedString := t.GetRepeatedString()
		assert.Equal(len(repeatedString), 0, "Size of repeatedString is 0")
		repeatedString = append(repeatedString, "Daisy")
		t.SetRepeatedString(repeatedString)
		repeatedString = t.GetRepeatedString()
		assert.Equal(len(repeatedString), 1, "SetRepeatedString adds the correct number of elements")
		assert.Equal(repeatedString[0], "Daisy", "SetRepeatedString sets the slices value")
		t.ClearRepeatedString()
		repeatedString = t.GetRepeatedString()
		assert.Equal(len(repeatedString), 0, "ClearRepeatedString clears all entries in the slice")
		t.AddRepeatedString("Robert", 0)
		repeatedString = t.GetRepeatedString()
		assert.Equal(repeatedString[0], "Robert", "AddRepeatedString appends a value to the slice")
	})

	qunit.Test("RepeatedBytes Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(7)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedBytes := t.GetRepeatedBytes()
		assert.Equal(len(repeatedBytes), 0, "Size of repeatedBytes is 0")
		repeatedBytes = append(repeatedBytes, []byte{0x01, 0x02})
		t.SetRepeatedBytes(repeatedBytes)
		repeatedBytes = t.GetRepeatedBytes()
		assert.Equal(len(repeatedBytes), 1, "SetRepeatedBytes adds the correct number of elements")
		// Can't compare against []byte literal
		assert.Equal(repeatedBytes[0][0], 0x01, "SetRepeatedBytes sets the slices value")
		assert.Equal(repeatedBytes[0][1], 0x02, "SetRepeatedBytes sets the slices value")
		t.ClearRepeatedBytes()
		repeatedBytes = t.GetRepeatedBytes()
		assert.Equal(len(repeatedBytes), 0, "ClearRepeatedBytes clears all entries in the slice")
		t.AddRepeatedBytes([]byte{0x01, 0x02}, 0)
		// Can't compare against []byte literal
		repeatedBytes = t.GetRepeatedBytes()
		assert.Equal(repeatedBytes[0][0], 0x01, "AddRepeatedBytes appends the byte slice")
		assert.Equal(repeatedBytes[0][1], 0x02, "AddRepeatedBytes appends the byte slice")
	})

	qunit.Test("RepeatedImportedMessage Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(10)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedImportedMessage := t.GetRepeatedImportedMessage()
		assert.Equal(len(repeatedImportedMessage), 0, "Size of repeatedImportedMessage is 0")
		repeatedImportedMessage = append(repeatedImportedMessage, new(multitest.Multi1).New(
			nil,
			multitest.Multi2_GREEN,
			multitest.Multi3_FEZ,
		))
		t.SetRepeatedImportedMessage(repeatedImportedMessage)
		repeatedImportedMessage = t.GetRepeatedImportedMessage()
		assert.Equal(len(repeatedImportedMessage), 1, "SetRepeatedImportedMessage adds the correct number of elements")
		assert.Equal(repeatedImportedMessage[0].GetColor(), multitest.Multi2_GREEN, "SetRepeatedImportedMessage sets the slices value")
		assert.Equal(repeatedImportedMessage[0].GetHatType(), multitest.Multi3_FEZ, "SetRepeatedImportedMessage sets the slices value")
		assert.Equal(repeatedImportedMessage[0].GetMulti2(), nil, "SetRepeatedImportedMessage sets the slices value")
		t.ClearRepeatedImportedMessage()
		repeatedImportedMessage = t.GetRepeatedImportedMessage()
		assert.Equal(len(repeatedImportedMessage), 0, "ClearRepeatedImportedMessage clears all entries in the slice")
		t.AddRepeatedImportedMessage(
			new(multitest.Multi1).New(
				nil,
				multitest.Multi2_BLUE,
				multitest.Multi3_FEDORA,
			),
			0,
		)
		repeatedImportedMessage = t.GetRepeatedImportedMessage()
		assert.Equal(len(repeatedImportedMessage), 1, "AddRepeatedImportedMessage adds a value to the slice")
		assert.Equal(repeatedImportedMessage[0].GetColor(), multitest.Multi2_BLUE, "AddRepeatedImportedMessage appends a value to the slice")
		assert.Equal(repeatedImportedMessage[0].GetHatType(), multitest.Multi3_FEDORA, "AddRepeatedImportedMessage appends a value to the slice")
		assert.Equal(repeatedImportedMessage[0].GetMulti2(), nil, "AddRepeatedImportedMessage appends a value to the slice")
	})

	qunit.Test("RepeatedNestedMessage Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedNestedMessage := t.GetRepeatedNestedMessage()
		assert.Equal(len(repeatedNestedMessage), 0, "Size of repeatedNestedMessage is 0")
		repeatedNestedMessage = append(repeatedNestedMessage, new(types.TestAllTypes_NestedMessage).New(
			10,
		))
		t.SetRepeatedNestedMessage(repeatedNestedMessage)
		repeatedNestedMessage = t.GetRepeatedNestedMessage()
		assert.Equal(len(repeatedNestedMessage), 1, "SetRepeatedNestedMessage adds the correct number of elements")
		assert.Equal(repeatedNestedMessage[0].GetB(), 10, "SetRepeatedNestedMessage sets the slices value")
		t.ClearRepeatedNestedMessage()
		repeatedNestedMessage = t.GetRepeatedNestedMessage()
		assert.Equal(len(repeatedNestedMessage), 0, "ClearRepeatedNestedMessage clears all entries in the slice")
		t.AddRepeatedNestedMessage(
			new(types.TestAllTypes_NestedMessage).New(
				10,
			),
			0,
		)
		repeatedNestedMessage = t.GetRepeatedNestedMessage()
		assert.Equal(len(repeatedNestedMessage), 1, "AddRepeatedNestedMessage adds a value to the slice")
		assert.Equal(repeatedNestedMessage[0].GetB(), 10, "AddRepeatedNestedMessage appends a value to the slice")
	})

	qunit.Test("RepeatedForeignMessage Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedForeignMessage := t.GetRepeatedForeignMessage()
		assert.Equal(len(repeatedForeignMessage), 0, "Size of repeatedForeignMessage is 0")
		repeatedForeignMessage = append(repeatedForeignMessage, new(types.ForeignMessage).New(
			10,
		))
		t.SetRepeatedForeignMessage(repeatedForeignMessage)
		repeatedForeignMessage = t.GetRepeatedForeignMessage()
		assert.Equal(len(repeatedForeignMessage), 1, "SetRepeatedForeignMessage adds the correct number of elements")
		assert.Equal(repeatedForeignMessage[0].GetC(), 10, "SetRepeatedForeignMessage sets the slices value")
		t.ClearRepeatedForeignMessage()
		repeatedForeignMessage = t.GetRepeatedForeignMessage()
		assert.Equal(len(repeatedForeignMessage), 0, "ClearRepeatedForeignMessage clears all entries in the slice")
		t.AddRepeatedForeignMessage(
			new(types.ForeignMessage).New(
				10,
			),
			0,
		)
		repeatedForeignMessage = t.GetRepeatedForeignMessage()
		assert.Equal(len(repeatedForeignMessage), 1, "AddRepeatedForeignMessage adds a value to the slice")
		assert.Equal(repeatedForeignMessage[0].GetC(), 10, "AddRepeatedForeignMessage appends a value to the slice")
	})

	qunit.Test("RepeatedNestedEnum Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedNestedEnum := t.GetRepeatedNestedEnum()
		assert.Equal(len(repeatedNestedEnum), 0, "Size of repeatedNestedEnum is 0")
		repeatedNestedEnum = append(repeatedNestedEnum, types.TestAllTypes_FOO)
		t.SetRepeatedNestedEnum(repeatedNestedEnum)
		repeatedNestedEnum = t.GetRepeatedNestedEnum()
		assert.Equal(len(repeatedNestedEnum), 1, "SetRepeatedNestedEnum adds the correct number of elements")
		assert.Equal(repeatedNestedEnum[0], types.TestAllTypes_FOO, "SetRepeatedNestedEnum sets the slices value")
		t.ClearRepeatedNestedEnum()
		repeatedNestedEnum = t.GetRepeatedNestedEnum()
		assert.Equal(len(repeatedNestedEnum), 0, "ClearRepeatedNestedEnum clears all entries in the slice")
		t.AddRepeatedNestedEnum(types.TestAllTypes_FOO, 0)
		repeatedNestedEnum = t.GetRepeatedNestedEnum()
		assert.Equal(len(repeatedNestedEnum), 1, "AddRepeatedNestedEnum adds a value to the slice")
		assert.Equal(repeatedNestedEnum[0], types.TestAllTypes_FOO, "AddRepeatedNestedEnum appends a value to the slice")
	})

	qunit.Test("RepeatedForeignEnum Get, Set, Add, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		repeatedForeignEnum := t.GetRepeatedForeignEnum()
		assert.Equal(len(repeatedForeignEnum), 0, "Size of repeatedForeignEnum is 0")
		repeatedForeignEnum = append(repeatedForeignEnum, types.ForeignEnum_FOREIGN_BAR)
		t.SetRepeatedForeignEnum(repeatedForeignEnum)
		repeatedForeignEnum = t.GetRepeatedForeignEnum()
		assert.Equal(len(repeatedForeignEnum), 1, "SetRepeatedForeignEnum adds the correct number of elements")
		assert.Equal(repeatedForeignEnum[0], types.ForeignEnum_FOREIGN_BAR, "SetRepeatedForeignEnum sets the slices value")
		t.ClearRepeatedForeignEnum()
		repeatedForeignEnum = t.GetRepeatedForeignEnum()
		assert.Equal(len(repeatedForeignEnum), 0, "ClearRepeatedForeignEnum clears all entries in the slice")
		t.AddRepeatedForeignEnum(types.ForeignEnum_FOREIGN_BAR, 0)
		repeatedForeignEnum = t.GetRepeatedForeignEnum()
		assert.Equal(len(repeatedForeignEnum), 1, "AddRepeatedForeignEnum adds a value to the slice")
		assert.Equal(repeatedForeignEnum[0], types.ForeignEnum_FOREIGN_BAR, "AddRepeatedForeignEnum appends a value to the slice")
	})
}

func oneofTypeTests() {
	qunit.Test("OneofField Get, Set, Has, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(186) // 💯

		t := &types.TestAllTypes{
			Object: js.Global.Get("proto").Get("types").Get("TestAllTypes").New(),
		}

		checkIsClear := func() {
			assert.Equal(t.HasOneofBytes(), false, "OneofField is unset")
			assert.Equal(t.HasOneofImportedMessage(), false, "OneofField is unset")
			assert.Equal(t.HasOneofNestedMessage(), false, "OneofField is unset")
			assert.Equal(t.HasOneofString(), false, "OneofField is unset")
			assert.Equal(t.HasOneofUint32(), false, "OneofField is unset")
			oneofField := t.GetOneofField()
			assert.Equal(oneofField, nil, "OneofField is unset")
		}

		checkIsClear()

		checkOneofBytesSet := func(setter string, val []byte) {
			assert.Equal(t.HasOneofBytes(), true, setter+" sets OneofBytes")
			assert.Equal(t.HasOneofImportedMessage(), false, setter+" does not set OneofImportedMessage")
			assert.Equal(t.HasOneofNestedMessage(), false, setter+" does not set OneofNestedMessage")
			assert.Equal(t.HasOneofString(), false, setter+" does not set OneofString")
			assert.Equal(t.HasOneofUint32(), false, setter+" does not set OneofUint32")

			oneofField := t.GetOneofField()
			oneofBytesMsg, ok := oneofField.(*types.TestAllTypes_OneofBytes)
			assert.Ok(ok, "GetOneofField returns type OneofBytes")
			assert.Equal(oneofBytesMsg.OneofBytes[0], val[0], setter+" sets OneofBytes")

			assert.Equal(t.GetOneofBytes()[0], val[0], setter+" sets OneofBytes")
			assert.Equal(t.GetOneofImportedMessage(), nil, setter+" does not set OneofImportedMessage")
			assert.Equal(t.GetOneofNestedMessage(), nil, setter+" does not set OneofNestedMessage")
			assert.Equal(t.GetOneofString(), "", setter+" does not set OneofString")
			assert.Equal(t.GetOneofUint32(), 0, setter+" does not set OneofUint32")
		}

		t.SetOneofField(&types.TestAllTypes_OneofBytes{OneofBytes: []byte{0x01}})
		checkOneofBytesSet("SetOneofField", []byte{0x01})
		t.ClearOneofBytes()
		checkIsClear()
		t.SetOneofBytes([]byte{0x01})
		checkOneofBytesSet("SetOneofBytes", []byte{0x01})
		t.ClearOneofBytes()
		checkIsClear()

		checkOneofImportedMessageSet := func(setter string, val *multitest.Multi1) {
			assert.Equal(t.HasOneofBytes(), false, setter+" does not set OneofBytes")
			assert.Equal(t.HasOneofImportedMessage(), true, setter+" sets OneofImportedMessage")
			assert.Equal(t.HasOneofNestedMessage(), false, setter+" does not set OneofNestedMessage")
			assert.Equal(t.HasOneofString(), false, setter+" does not set OneofString")
			assert.Equal(t.HasOneofUint32(), false, setter+" does not set OneofUint32")

			oneofField := t.GetOneofField()
			oneofImportedMessage, ok := oneofField.(*types.TestAllTypes_OneofImportedMessage)
			assert.Ok(ok, "GetOneofField returns type OneofImportedMessage")
			assert.Equal(oneofImportedMessage.OneofImportedMessage, val, setter+" sets OneofImportedMessage")

			assert.Equal(len(t.GetOneofBytes()), 0, setter+" does not set OneofBytes")
			assert.Equal(t.GetOneofImportedMessage(), val, setter+" sets OneofImportedMessage")
			assert.Equal(t.GetOneofNestedMessage(), nil, setter+" does not set OneofNestedMessage")
			assert.Equal(t.GetOneofString(), "", setter+" does not set OneofString")
			assert.Equal(t.GetOneofUint32(), 0, setter+" does not set OneofUint32")
		}

		importedMessage := new(multitest.Multi1).New(
			nil,
			multitest.Multi2_BLUE,
			multitest.Multi3_FEDORA,
		)
		t.SetOneofField(&types.TestAllTypes_OneofImportedMessage{OneofImportedMessage: importedMessage})
		checkOneofImportedMessageSet("SetOneofField", importedMessage)
		t.ClearOneofImportedMessage()
		checkIsClear()
		t.SetOneofImportedMessage(importedMessage)
		checkOneofImportedMessageSet("SetOneofImportedMessage", importedMessage)
		t.ClearOneofImportedMessage()
		checkIsClear()

		checkOneofNestedMessageSet := func(setter string, val *types.TestAllTypes_NestedMessage) {
			assert.Equal(t.HasOneofBytes(), false, setter+" does not set OneofBytes")
			assert.Equal(t.HasOneofImportedMessage(), false, setter+" does not set OneofImportedMessage")
			assert.Equal(t.HasOneofNestedMessage(), true, setter+" sets OneofNestedMessage")
			assert.Equal(t.HasOneofString(), false, setter+" does not set OneofString")
			assert.Equal(t.HasOneofUint32(), false, setter+" does not set OneofUint32")

			oneofField := t.GetOneofField()
			oneofNestedMessage, ok := oneofField.(*types.TestAllTypes_OneofNestedMessage)
			assert.Ok(ok, "GetOneofField returns type OneofNestedMessage")
			assert.Equal(oneofNestedMessage.OneofNestedMessage, val, setter+" sets OneofNestedMessage")

			assert.Equal(len(t.GetOneofBytes()), 0, setter+" does not set OneofBytes")
			assert.Equal(t.GetOneofImportedMessage(), nil, setter+" does not set OneofImportedMessage")
			assert.Equal(t.GetOneofNestedMessage(), val, setter+" sets OneofNestedMessage")
			assert.Equal(t.GetOneofString(), "", setter+" does not set OneofString")
			assert.Equal(t.GetOneofUint32(), 0, setter+" does not set OneofUint32")
		}

		nestedMessage := new(types.TestAllTypes_NestedMessage).New(
			10,
		)
		t.SetOneofField(&types.TestAllTypes_OneofNestedMessage{OneofNestedMessage: nestedMessage})
		checkOneofNestedMessageSet("SetOneofField", nestedMessage)
		t.ClearOneofNestedMessage()
		checkIsClear()
		t.SetOneofNestedMessage(nestedMessage)
		checkOneofNestedMessageSet("SetOneofNestedMessage", nestedMessage)
		t.ClearOneofNestedMessage()
		checkIsClear()

		checkOneofStringSet := func(setter string, val string) {
			assert.Equal(t.HasOneofBytes(), false, setter+" does not set OneofBytes")
			assert.Equal(t.HasOneofImportedMessage(), false, setter+" does not set OneofImportedMessage")
			assert.Equal(t.HasOneofNestedMessage(), false, setter+" does not set OneofNestedMessage")
			assert.Equal(t.HasOneofString(), true, setter+" sets OneofString")
			assert.Equal(t.HasOneofUint32(), false, setter+" does not set OneofUint32")

			oneofField := t.GetOneofField()
			oneofStringMsg, ok := oneofField.(*types.TestAllTypes_OneofString)
			assert.Ok(ok, "GetOneofField returns type OneofString")
			assert.Equal(oneofStringMsg.OneofString, val, setter+" sets OneofString")

			assert.Equal(len(t.GetOneofBytes()), 0, setter+" does not set OneofBytes")
			assert.Equal(t.GetOneofImportedMessage(), nil, setter+" does not set OneofImportedMessage")
			assert.Equal(t.GetOneofNestedMessage(), nil, setter+" does not set OneofNestedMessage")
			assert.Equal(t.GetOneofString(), val, setter+" sets OneofString")
			assert.Equal(t.GetOneofUint32(), 0, setter+" does not set OneofUint32")
		}

		t.SetOneofField(&types.TestAllTypes_OneofString{OneofString: "Daisy"})
		checkOneofStringSet("SetOneofField", "Daisy")
		t.ClearOneofString()
		checkIsClear()
		t.SetOneofString("Daisy")
		checkOneofStringSet("SetOneofString", "Daisy")
		t.ClearOneofString()
		checkIsClear()

		checkOneofUint32Set := func(setter string, val uint32) {
			assert.Equal(t.HasOneofBytes(), false, setter+" does not set OneofBytes")
			assert.Equal(t.HasOneofImportedMessage(), false, setter+" does not set OneofImportedMessage")
			assert.Equal(t.HasOneofNestedMessage(), false, setter+" does not set OneofNestedMessage")
			assert.Equal(t.HasOneofString(), false, setter+" does not set OneofString")
			assert.Equal(t.HasOneofUint32(), true, setter+" sets OneofUint32")

			oneofField := t.GetOneofField()
			oneofUint32Msg, ok := oneofField.(*types.TestAllTypes_OneofUint32)
			assert.Ok(ok, "GetOneofField returns type OneofUint32")
			assert.Equal(oneofUint32Msg.OneofUint32, val, setter+" sets OneofUint32")

			assert.Equal(len(t.GetOneofBytes()), 0, setter+" does not set OneofBytes")
			assert.Equal(t.GetOneofImportedMessage(), nil, setter+" does not set OneofImportedMessage")
			assert.Equal(t.GetOneofNestedMessage(), nil, setter+" does not set OneofNestedMessage")
			assert.Equal(t.GetOneofString(), "", setter+" does not set OneofString")
			assert.Equal(t.GetOneofUint32(), val, setter+" sets OneofUint32")
		}

		t.SetOneofField(&types.TestAllTypes_OneofUint32{OneofUint32: 10})
		checkOneofUint32Set("SetOneofField", 10)
		t.ClearOneofUint32()
		checkIsClear()
		t.SetOneofUint32(10)
		checkOneofUint32Set("SetOneofUint32", 10)
		t.ClearOneofUint32()
		checkIsClear()
	})
}

func mapTypeTests() {
	qunit.Test("Int32Int32Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32int32Map := t.GetMapInt32Int32()
		assert.Equal(len(int32int32Map), 0, "Size of int32int32Map is 0")
		int32int32Map[1] = 10
		t.SetMapInt32Int32(int32int32Map)
		int32int32Map = t.GetMapInt32Int32()
		assert.Equal(len(int32int32Map), 1, "SetMapInt32Int32 adds the correct number of keys")
		assert.Equal(int32int32Map[1], 10, "SetMapInt32Int32 sets the keys value")
		t.ClearMapInt32Int32()
		int32int32Map = t.GetMapInt32Int32()
		assert.Equal(len(int32int32Map), 0, "ClearMapInt32Int32 clears all entries in the map")
	})

	qunit.Test("Int64Int64Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int64int64Map := t.GetMapInt64Int64()
		assert.Equal(len(int64int64Map), 0, "Size of int64int64Map is 0")
		int64int64Map[1] = 10
		t.SetMapInt64Int64(int64int64Map)
		int64int64Map = t.GetMapInt64Int64()
		assert.Equal(len(int64int64Map), 1, "SetMapInt64Int64 adds the correct number of keys")
		assert.Equal(int64int64Map[1], 10, "SetMapInt64Int64 sets the keys value")
		t.ClearMapInt64Int64()
		int64int64Map = t.GetMapInt64Int64()
		assert.Equal(len(int64int64Map), 0, "ClearMapInt64Int64 clears all entries in the map")
	})

	qunit.Test("Uint32Uint32Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		uint32uint32Map := t.GetMapUint32Uint32()
		assert.Equal(len(uint32uint32Map), 0, "Size of uint32uint32Map is 0")
		uint32uint32Map[1] = 10
		t.SetMapUint32Uint32(uint32uint32Map)
		uint32uint32Map = t.GetMapUint32Uint32()
		assert.Equal(len(uint32uint32Map), 1, "SetMapUint32Uint32 adds the correct number of keys")
		assert.Equal(uint32uint32Map[1], 10, "SetMapUint32Uint32 sets the keys value")
		t.ClearMapUint32Uint32()
		uint32uint32Map = t.GetMapUint32Uint32()
		assert.Equal(len(uint32uint32Map), 0, "ClearMapUint32Uint32 clears all entries in the map")
	})

	qunit.Test("Uint64Uint64Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		uint64uint64Map := t.GetMapUint64Uint64()
		assert.Equal(len(uint64uint64Map), 0, "Size of uint64uint64Map is 0")
		uint64uint64Map[1] = 10
		t.SetMapUint64Uint64(uint64uint64Map)
		uint64uint64Map = t.GetMapUint64Uint64()
		assert.Equal(len(uint64uint64Map), 1, "SetMapUint64Uint64 adds the correct number of keys")
		assert.Equal(uint64uint64Map[1], 10, "SetMapUint64Uint64 sets the keys value")
		t.ClearMapUint64Uint64()
		uint64uint64Map = t.GetMapUint64Uint64()
		assert.Equal(len(uint64uint64Map), 0, "ClearMapUint64Uint64 clears all entries in the map")
	})

	qunit.Test("Sint32Sint32Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		sint32sint32Map := t.GetMapSint32Sint32()
		assert.Equal(len(sint32sint32Map), 0, "Size of sint32sint32Map is 0")
		sint32sint32Map[1] = -10
		t.SetMapSint32Sint32(sint32sint32Map)
		sint32sint32Map = t.GetMapSint32Sint32()
		assert.Equal(len(sint32sint32Map), 1, "SetMapSint32Sint32 adds the correct number of keys")
		assert.Equal(sint32sint32Map[1], -10, "SetMapSint32Sint32 sets the keys value")
		t.ClearMapSint32Sint32()
		sint32sint32Map = t.GetMapSint32Sint32()
		assert.Equal(len(sint32sint32Map), 0, "ClearMapSint32Sint32 clears all entries in the map")
	})

	qunit.Test("Sint64Sint64Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		sint64sint64Map := t.GetMapSint64Sint64()
		assert.Equal(len(sint64sint64Map), 0, "Size of sint64sint64Map is 0")
		sint64sint64Map[1] = -10
		t.SetMapSint64Sint64(sint64sint64Map)
		sint64sint64Map = t.GetMapSint64Sint64()
		assert.Equal(len(sint64sint64Map), 1, "SetMapSint64Sint64 adds the correct number of keys")
		assert.Equal(sint64sint64Map[1], -10, "SetMapSint64Sint64 sets the keys value")
		t.ClearMapSint64Sint64()
		sint64sint64Map = t.GetMapSint64Sint64()
		assert.Equal(len(sint64sint64Map), 0, "ClearMapSint64Sint64 clears all entries in the map")
	})

	qunit.Test("Fixed32Fixed32Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		fixed32fixed32Map := t.GetMapFixed32Fixed32()
		assert.Equal(len(fixed32fixed32Map), 0, "Size of fixed32fixed32Map is 0")
		fixed32fixed32Map[1] = 10
		t.SetMapFixed32Fixed32(fixed32fixed32Map)
		fixed32fixed32Map = t.GetMapFixed32Fixed32()
		assert.Equal(len(fixed32fixed32Map), 1, "SetMapFixed32Fixed32 adds the correct number of keys")
		assert.Equal(fixed32fixed32Map[1], 10, "SetMapFixed32Fixed32 sets the keys value")
		t.ClearMapFixed32Fixed32()
		fixed32fixed32Map = t.GetMapFixed32Fixed32()
		assert.Equal(len(fixed32fixed32Map), 0, "ClearMapFixed32Fixed32 clears all entries in the map")
	})

	qunit.Test("Fixed64Fixed64Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		fixed64fixed64Map := t.GetMapFixed64Fixed64()
		assert.Equal(len(fixed64fixed64Map), 0, "Size of fixed64fixed64Map is 0")
		fixed64fixed64Map[1] = 10
		t.SetMapFixed64Fixed64(fixed64fixed64Map)
		fixed64fixed64Map = t.GetMapFixed64Fixed64()
		assert.Equal(len(fixed64fixed64Map), 1, "SetMapFixed64Fixed64 adds the correct number of keys")
		assert.Equal(fixed64fixed64Map[1], 10, "SetMapFixed64Fixed64 sets the keys value")
		t.ClearMapFixed64Fixed64()
		fixed64fixed64Map = t.GetMapFixed64Fixed64()
		assert.Equal(len(fixed64fixed64Map), 0, "ClearMapFixed64Fixed64 clears all entries in the map")
	})

	qunit.Test("Sfixed32Sfixed32Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		sfixed32sfixed32Map := t.GetMapSfixed32Sfixed32()
		assert.Equal(len(sfixed32sfixed32Map), 0, "Size of sfixed32sfixed32Map is 0")
		sfixed32sfixed32Map[1] = 10
		t.SetMapSfixed32Sfixed32(sfixed32sfixed32Map)
		sfixed32sfixed32Map = t.GetMapSfixed32Sfixed32()
		assert.Equal(len(sfixed32sfixed32Map), 1, "SetMapSfixed32Sfixed32 adds the correct number of keys")
		assert.Equal(sfixed32sfixed32Map[1], 10, "SetMapSfixed32Sfixed32 sets the keys value")
		t.ClearMapSfixed32Sfixed32()
		sfixed32sfixed32Map = t.GetMapSfixed32Sfixed32()
		assert.Equal(len(sfixed32sfixed32Map), 0, "ClearMapSfixed32Sfixed32 clears all entries in the map")
	})

	qunit.Test("Sfixed64Sfixed64Map Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		sfixed64sfixed64Map := t.GetMapSfixed64Sfixed64()
		assert.Equal(len(sfixed64sfixed64Map), 0, "Size of sfixed64sfixed64Map is 0")
		sfixed64sfixed64Map[1] = 10
		t.SetMapSfixed64Sfixed64(sfixed64sfixed64Map)
		sfixed64sfixed64Map = t.GetMapSfixed64Sfixed64()
		assert.Equal(len(sfixed64sfixed64Map), 1, "SetMapSfixed64Sfixed64 adds the correct number of keys")
		assert.Equal(sfixed64sfixed64Map[1], 10, "SetMapSfixed64Sfixed64 sets the keys value")
		t.ClearMapSfixed64Sfixed64()
		sfixed64sfixed64Map = t.GetMapSfixed64Sfixed64()
		assert.Equal(len(sfixed64sfixed64Map), 0, "ClearMapSfixed64Sfixed64 clears all entries in the map")
	})

	qunit.Test("Int32FloatMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32floatMap := t.GetMapInt32Float()
		assert.Equal(len(int32floatMap), 0, "Size of int32floatMap is 0")
		int32floatMap[1] = 1.5
		t.SetMapInt32Float(int32floatMap)
		int32floatMap = t.GetMapInt32Float()
		assert.Equal(len(int32floatMap), 1, "SetMapInt32Float adds the correct number of keys")
		assert.Equal(int32floatMap[1], 1.5, "SetMapInt32Float sets the keys value")
		t.ClearMapInt32Float()
		int32floatMap = t.GetMapInt32Float()
		assert.Equal(len(int32floatMap), 0, "ClearMapInt32Float clears all entries in the map")
	})

	qunit.Test("Int32DoubleMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32doubleMap := t.GetMapInt32Double()
		assert.Equal(len(int32doubleMap), 0, "Size of int32doubleMap is 0")
		int32doubleMap[1] = 1.5
		t.SetMapInt32Double(int32doubleMap)
		int32doubleMap = t.GetMapInt32Double()
		assert.Equal(len(int32doubleMap), 1, "SetMapInt32Double adds the correct number of keys")
		assert.Equal(int32doubleMap[1], 1.5, "SetMapInt32Double sets the keys value")
		t.ClearMapInt32Double()
		int32doubleMap = t.GetMapInt32Double()
		assert.Equal(len(int32doubleMap), 0, "ClearMapInt32Double clears all entries in the map")
	})

	qunit.Test("BoolBoolMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		boolboolMap := t.GetMapBoolBool()
		assert.Equal(len(boolboolMap), 0, "Size of boolboolMap is 0")
		boolboolMap[true] = false
		t.SetMapBoolBool(boolboolMap)
		boolboolMap = t.GetMapBoolBool()
		assert.Equal(len(boolboolMap), 1, "SetMapBoolBool adds the correct number of keys")
		assert.Equal(boolboolMap[true], false, "SetMapBoolBool sets the keys value")
		t.ClearMapBoolBool()
		boolboolMap = t.GetMapBoolBool()
		assert.Equal(len(boolboolMap), 0, "ClearMapBoolBool clears all entries in the map")
	})

	qunit.Test("StringStringMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		stringstringMap := t.GetMapStringString()
		assert.Equal(len(stringstringMap), 0, "Size of stringstringMap is 0")
		stringstringMap["daisy"] = "jonas"
		t.SetMapStringString(stringstringMap)
		stringstringMap = t.GetMapStringString()
		assert.Equal(len(stringstringMap), 1, "SetMapStringString adds the correct number of keys")
		assert.Equal(stringstringMap["daisy"], "jonas", "SetMapStringString sets the keys value")
		t.ClearMapStringString()
		stringstringMap = t.GetMapStringString()
		assert.Equal(len(stringstringMap), 0, "ClearMapStringString clears all entries in the map")
	})

	qunit.Test("Int32BytesMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(5)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32BytesMap := t.GetMapInt32Bytes()
		assert.Equal(len(int32BytesMap), 0, "Size of int32BytesMap is 0")
		int32BytesMap[1] = []byte{0x01, 0x02}
		t.SetMapInt32Bytes(int32BytesMap)
		int32BytesMap = t.GetMapInt32Bytes()
		assert.Equal(len(int32BytesMap), 1, "SetMapInt32Bytes adds the correct number of keys")
		// Can't compare against []byte literal
		assert.Equal(int32BytesMap[1][0], 0x01, "SetMapInt32Bytes sets the keys value")
		assert.Equal(int32BytesMap[1][1], 0x02, "SetMapInt32Bytes sets the keys value")
		t.ClearMapInt32Bytes()
		int32BytesMap = t.GetMapInt32Bytes()
		assert.Equal(len(int32BytesMap), 0, "ClearMapInt32Bytes clears all entries in the map")
	})

	qunit.Test("Int32EnumMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32enumMap := t.GetMapInt32Enum()
		assert.Equal(len(int32enumMap), 0, "Size of int32enumMap is 0")
		int32enumMap[1] = types.MapEnum_MAP_ENUM_BAR
		t.SetMapInt32Enum(int32enumMap)
		int32enumMap = t.GetMapInt32Enum()
		assert.Equal(len(int32enumMap), 1, "SetMapInt32Enum adds the correct number of keys")
		assert.Equal(int32enumMap[1], types.MapEnum_MAP_ENUM_BAR, "SetMapInt32Enum sets the keys value")
		t.ClearMapInt32Enum()
		int32enumMap = t.GetMapInt32Enum()
		assert.Equal(len(int32enumMap), 0, "ClearMapInt32Enum clears all entries in the map")
	})

	qunit.Test("Int32ForeignMessageMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(4)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32foreignmessageMap := t.GetMapInt32ForeignMessage()
		assert.Equal(len(int32foreignmessageMap), 0, "Size of int32foreignmessageMap is 0")
		int32foreignmessageMap[1] = new(types.ForeignMessage).New(10)
		t.SetMapInt32ForeignMessage(int32foreignmessageMap)
		int32foreignmessageMap = t.GetMapInt32ForeignMessage()
		assert.Equal(len(int32foreignmessageMap), 1, "SetMapInt32ForeignMessage adds the correct number of keys")
		assert.Equal(int32foreignmessageMap[1].GetC(), 10, "SetMapInt32ForeignMessage sets the keys value")
		t.ClearMapInt32ForeignMessage()
		int32foreignmessageMap = t.GetMapInt32ForeignMessage()
		assert.Equal(len(int32foreignmessageMap), 0, "ClearMapInt32ForeignMessage clears all entries in the map")
	})

	qunit.Test("Int32ImportedMessageMap Get, Set, Clear", func(assert qunit.QUnitAssert) {
		qunit.Expect(6)

		t := &types.TestMap{
			Object: js.Global.Get("proto").Get("types").Get("TestMap").New(),
		}

		int32importedmessageMap := t.GetMapInt32ImportedMessage()
		assert.Equal(len(int32importedmessageMap), 0, "Size of int32importedmessageMap is 0")
		int32importedmessageMap[1] = new(multitest.Multi1).New(
			nil,
			multitest.Multi2_GREEN,
			multitest.Multi3_FEDORA,
		)
		t.SetMapInt32ImportedMessage(int32importedmessageMap)
		int32importedmessageMap = t.GetMapInt32ImportedMessage()
		assert.Equal(len(int32importedmessageMap), 1, "SetMapInt32ImportedMessage adds the correct number of keys")
		assert.Equal(int32importedmessageMap[1].GetMulti2(), nil, "SetMapInt32ImportedMessage sets the keys value")
		assert.Equal(int32importedmessageMap[1].GetColor(), multitest.Multi2_GREEN, "SetMapInt32ImportedMessage sets the keys value")
		assert.Equal(int32importedmessageMap[1].GetHatType(), multitest.Multi3_FEDORA, "SetMapInt32ImportedMessage sets the keys value")
		t.ClearMapInt32ImportedMessage()
		int32importedmessageMap = t.GetMapInt32ImportedMessage()
		assert.Equal(len(int32importedmessageMap), 0, "ClearMapInt32ImportedMessage clears all entries in the map")
	})
}
