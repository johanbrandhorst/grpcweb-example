regenerate:
	# Generate protofiles
	protoc proto/test/test.proto \
		--js_out=import_style=commonjs,binary:./client/ \
		--gopherjs_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/johanbrandhorst/protobuf/ptypes/empty:./client/ \
		--go_out=plugins=grpc:./server/


	(cd ./client/proto/test && webpack && rm test_pb.js)

	# Regenerate dependencies
	(cd ../protoc-gen-gopherjs/test && make regenerate)

	# Generate GopherJS files
	go generate ./client/...

test:
	ginkgo -v .
