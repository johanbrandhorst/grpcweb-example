golden:
	gofmt -w grpc_test/test.pb.gopherjs.go
	diff -w grpc_test/test.pb.gopherjs.go grpc_test/test.pb.gopherjs.go.golden

regenerate:
	# Invoke protoc once to generate three independent .pb.go files in the same package.
	protoc --gopherjs_out=. --js_out=import_style=commonjs,binary:. \
		multi/multi1.proto multi/multi2.proto multi/multi3.proto
	protoc \
		--gopherjs_out=plugins=grpc,Mmulti/multi1.proto=github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs/test/multi:. \
		--js_out=import_style=commonjs,binary:. grpc_test/test.proto

	# Replace any well known type imports with correct namespace
	sed -i -E "s;require\('google-protobuf/.*'\);\$$global.proto.google.protobuf;g" \
		./grpc_test/test_pb.js \
		./multi/multi1_pb.js \
		./multi/multi2_pb.js \
		./multi/multi3_pb.js
	# Replace any well known type imports with correct namespace
	sed -i -E "s;require\('../multi.*'\);\$$global.proto.multitest;g" \
		./grpc_test/test_pb.js \
		./multi/multi1_pb.js \
		./multi/multi2_pb.js \
		./multi/multi3_pb.js
	# Replace any other imports with global namespace
	sed -i -E "s;require\('.*'\);\$$global;g" \
		./grpc_test/test_pb.js \
		./multi/multi1_pb.js \
		./multi/multi2_pb.js \
		./multi/multi3_pb.js
	# Remove export statements
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);\$$//g" \
		./grpc_test/test_pb.js \
		./multi/multi1_pb.js \
		./multi/multi2_pb.js \
		./multi/multi3_pb.js

	# Move exportSymbol in multi1_pb to avoid undefined proto type
	sed -i "s/goog.exportSymbol('proto.multitest.Multi1', null, global);//g" \
		./multi/multi1_pb.js
	sed -i "12igoog.exportSymbol('proto.multitest.Multi1', null, global);" \
		./multi/multi1_pb.js


	# Move original files
	mv ./grpc_test/test_pb.js ./grpc_test/test_pb.inc.js
	mv ./multi/multi1_pb.js ./multi/multi1_pb.inc.js
	mv ./multi/multi2_pb.js ./multi/multi2_pb.inc.js
	mv ./multi/multi3_pb.js ./multi/multi3_pb.inc.js

test: regenerate golden
