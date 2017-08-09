golden:
	gofmt -w grpc_test/test.pb.gopherjs.go
	diff -w grpc_test/test.pb.gopherjs.go grpc_test/test.pb.gopherjs.go.golden
	gofmt -w types/types.pb.gopherjs.go
	diff -w types/types.pb.gopherjs.go types/types.pb.gopherjs.go.golden

regenerate:
	# Invoke protoc once to generate three independent .pb.go files in the same package.
	protoc --gopherjs_out=. --js_out=import_style=commonjs,binary:. \
		multi/multi1.proto multi/multi2.proto multi/multi3.proto
	protoc \
		--gopherjs_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:. \
		--js_out=import_style=commonjs,binary:. grpc_test/test.proto
	protoc \
		--gopherjs_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:. \
		--js_out=import_style=commonjs,binary:. types/types.proto

	(cd grpc_test && webpack && rm test_pb.js)
	(cd types && webpack && rm types_pb.js)
	(cd multi && webpack && rm multi1_pb.js && rm multi2_pb.js && rm multi3_pb.js)

test: regenerate golden
