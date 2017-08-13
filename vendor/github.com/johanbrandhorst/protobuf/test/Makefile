regenerate:
	# Generate protofiles
	protoc proto/test/test.proto \
		--gopherjs_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/johanbrandhorst/protobuf/ptypes/empty:./client/ \
		--go_out=plugins=grpc:./server/
	protoc ../protoc-gen-gopherjs/test/multi/multi1.proto ../protoc-gen-gopherjs/test/multi/multi2.proto ../protoc-gen-gopherjs/test/multi/multi3.proto \
		-I../protoc-gen-gopherjs/test \
		--go_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:./server/proto
	protoc ../protoc-gen-gopherjs/test/types/types.proto \
		-I../protoc-gen-gopherjs/test \
		--go_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/test/server/proto/multi:./server/proto

	# Regenerate dependencies
	(cd ../protoc-gen-gopherjs/test && make regenerate)

	# Generate GopherJS files
	go generate ./client/...

test:
	ginkgo -v .
