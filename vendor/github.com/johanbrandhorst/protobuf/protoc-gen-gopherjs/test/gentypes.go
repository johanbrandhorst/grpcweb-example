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

	gRPCMarshal()

	typeMarshal()

	mapMarshal()
}

func gRPCMarshal() {
	qunit.Test("Simple Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
		req := &test.Simple{
			Key:      1234,
			Deadline: 1.5,
			Day:      test.Days_MONDAY,
			Name:     "Alfred",
		}
		marshalled := req.Marshal()
		newReq, err := new(test.Simple).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})

	qunit.Test("Complex Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
		req := &test.Complex{
			Communique: []*test.Complex_Communique{
				{
					MakeMeCry: false,
					UnionThing: &test.Complex_Communique_Delta_{
						Delta: 1234,
					},
				},
				{
					MakeMeCry: true,
					UnionThing: &test.Complex_Communique_Today{
						Today: test.Days_TUESDAY,
					},
				},
			},
			CompactKeys: map[int32]string{
				1234: "The White House",
				5678: "The Empire State Building",
			},
			Multi: &multi.Multi1{
				Multi2: &multi.Multi2{
					RequiredValue: 2345,
					Color:         multi.Multi2_BLUE,
				},
				Color:   multi.Multi2_RED,
				HatType: multi.Multi3_FEZ,
			},
		}

		marshalled := req.Marshal()
		newReq, err := new(test.Complex).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})
}

func typeMarshal() {
	qunit.Test("TestAllTypes Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
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

		marshalled := req.Marshal()
		newReq, err := new(types.TestAllTypes).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})
}

func mapMarshal() {
	qunit.Test("TestMap Marshal and Unmarshal", func(assert qunit.QUnitAssert) {
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

		marshalled := req.Marshal()
		newReq, err := new(types.TestMap).Unmarshal(marshalled)
		if err != nil {
			assert.Ok(false, "Unexpected error returned: "+err.Error()+"\n"+err.(*js.Error).Stack())
		}
		assert.DeepEqual(req, newReq, "Marshalling and unmarshalling results in the same struct")
	})
}
