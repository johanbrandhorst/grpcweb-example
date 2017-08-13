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
	protoc --gopherjs_out=:${GOPATH}/src ./empty/empty.proto
	protoc --gopherjs_out=:${GOPATH}/src ./timestamp/timestamp.proto
	protoc --gopherjs_out=:${GOPATH}/src ./struct/struct.proto
	protoc --gopherjs_out=:${GOPATH}/src ./any/any.proto
	protoc --gopherjs_out=:${GOPATH}/src ./duration/duration.proto
	protoc --gopherjs_out=:${GOPATH}/src ./wrappers/wrappers.proto

clean:
	rm -f empty/*.go
	rm -f timestamp/*.go
	rm -f struct/*.go
	rm -f any/*.go
	rm -f duration/*.go
	rm -f wrappers/*.go
