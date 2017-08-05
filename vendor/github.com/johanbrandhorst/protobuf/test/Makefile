regenerate:
	# Generate protofiles
	protoc proto/test/test.proto \
		--js_out=import_style=commonjs,binary:./client/ \
		--gopherjs_out=plugins=grpc,Mgoogle/protobuf/empty.proto=github.com/johanbrandhorst/protobuf/ptypes/empty:./client/ \
		--go_out=plugins=grpc:./server/

	# Replace top level import with global reference
	sed -i "s;require('google-protobuf');\$$global;g" ./client/proto/test/test_pb.js
	# Replace any well known type imports with correct namespace
	sed -i -E "s;require\('google-protobuf/.*'\);\$$global.proto.google.protobuf;g" ./client/proto/test/test_pb.js
	# Remove export statement
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);\$$//g" ./client/proto/test/test_pb.js

	# Move original file
	mv ./client/proto/test/test_pb.js ./client/proto/test/test_pb.inc.js

	# Regenerate dependencies
	(cd ../protoc-gen-gopherjs/test && make regenerate)

	# Generate GopherJS files
	go generate ./client/...

test:
	ginkgo -v .
