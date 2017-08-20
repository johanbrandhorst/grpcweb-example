regenerate:
	# Generate protofiles
	protoc -I. -I$$GOPATH/src proto/test/test.proto \
		--gopherjs_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/johanbrandhorst/protobuf/ptypes/empty:$$GOPATH/src \
		--go_out=plugins=grpc:$$GOPATH/src
	protoc -I. -I$$GOPATH/src -I../protoc-gen-gopherjs/test \
		../protoc-gen-gopherjs/test/multi/multi1.proto \
		../protoc-gen-gopherjs/test/multi/multi2.proto \
		../protoc-gen-gopherjs/test/multi/multi3.proto \
		--go_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:./server/proto
	protoc -I. -I$$GOPATH/src -I../protoc-gen-gopherjs/test \
		../protoc-gen-gopherjs/test/types/types.proto \
		--go_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/test/server/proto/multi:./server/proto

	# Regenerate dependencies
	(cd ../protoc-gen-gopherjs/test && make regenerate)

	# Generate GopherJS files
	go generate ./client/...

test:
	ginkgo -v .
