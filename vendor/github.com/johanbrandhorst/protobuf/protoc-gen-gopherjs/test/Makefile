golden:
	gofmt -w grpc_test/test.pb.gopherjs.go
	diff -w grpc_test/test.pb.gopherjs.go grpc_test/test.pb.gopherjs.go.golden
	gofmt -w types/types.pb.gopherjs.go
	diff -w types/types.pb.gopherjs.go types/types.pb.gopherjs.go.golden

regenerate:
	# Invoke protoc once to generate three independent .pb.go files in the same package.
	protoc -I ${GOPATH}/src -I .\
		--gopherjs_out=${GOPATH}/src \
		multi/multi1.proto multi/multi2.proto multi/multi3.proto
	protoc -I ${GOPATH}/src -I .\
		--gopherjs_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:${GOPATH}/src \
		grpc_test/test.proto
	protoc -I ${GOPATH}/src -I .\
		--gopherjs_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:${GOPATH}/src \
		types/types.proto

test: regenerate golden
