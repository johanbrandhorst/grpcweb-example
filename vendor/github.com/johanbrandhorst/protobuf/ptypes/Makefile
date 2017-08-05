URL="https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/"

update:
	wget ${URL}/empty.proto -O empty/empty.proto
	wget ${URL}/timestamp.proto -O timestamp/timestamp.proto
	wget ${URL}/struct.proto -O struct/struct.proto
	wget ${URL}/any.proto -O any/any.proto
	wget ${URL}/duration.proto -O duration/duration.proto
	wget ${URL}/wrappers.proto -O wrappers/wrappers.proto

	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" empty/empty.proto
	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" timestamp/timestamp.proto
	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" struct/struct.proto
	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" any/any.proto
	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" duration/duration.proto
	sed -i "s;golang/protobuf;johanbrandhorst/protobuf;g" wrappers/wrappers.proto

regenerate:

	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./empty/empty.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./empty/empty_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./empty/empty_pb.js
	# Rename file
	mv ./empty/empty_pb.js ./empty/empty_pb.inc.js

	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./timestamp/timestamp.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./timestamp/timestamp_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./timestamp/timestamp_pb.js
	# Rename file
	mv ./timestamp/timestamp_pb.js ./timestamp/timestamp_pb.inc.js


	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./struct/struct.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./struct/struct_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./struct/struct_pb.js
	# Rename file
	mv ./struct/struct_pb.js ./struct/struct_pb.inc.js


	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./any/any.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./any/any_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./any/any_pb.js
	# Rename file
	mv ./any/any_pb.js ./any/any_pb.inc.js


	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./duration/duration.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./duration/duration_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./duration/duration_pb.js
	# Rename file
	mv ./duration/duration_pb.js ./duration/duration_pb.inc.js


	protoc --gopherjs_out=:${GOPATH}/src --js_out=import_style=commonjs,binary:. ./wrappers/wrappers.proto
	# Replace imports and exports in ProtobufJS file
	sed -i "s/require('google-protobuf')/\$$global/g" ./wrappers/wrappers_pb.js
	sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$$//g" ./wrappers/wrappers_pb.js
	# Rename file
	mv ./wrappers/wrappers_pb.js ./wrappers/wrappers_pb.inc.js

clean:
	rm -f empty/*.js empty/*.go
	rm -f timestamp/*.js timestamp/*.go
	rm -f struct/*.js struct/*.go
	rm -f any/*.js any/*.go
	rm -f duration/*.js duration/*.go
	rm -f wrappers/*.js wrappers/*.go
