package test

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/rusco/qunit"

	test "github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/grpc_test"
	multitest "github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi"
	"github.com/johanbrandhorst/protobuf/test/recoverer"
)

// This test is imported and run by the root level tests

func GenTypesTest() {
	defer recoverer.Recover() // recovers any panics and fails tests

	qunit.Module("GopherJS Protobuf Generator tests")

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
		qunit.Expect(5)

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
