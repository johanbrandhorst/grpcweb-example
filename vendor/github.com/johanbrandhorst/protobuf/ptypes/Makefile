URL="https://raw.githubusercontent.com/google/protobuf/master/src/google/protobuf/"

update:
	wget ${URL}/empty.proto -O empty/empty.proto
	wget ${URL}/timestamp.proto -O timestamp/timestamp.proto
	wget ${URL}/struct.proto -O struct/struct.proto
	wget ${URL}/any.proto -O any/any.proto
	wget ${URL}/duration.proto -O duration/duration.proto
	wget ${URL}/wrappers.proto -O wrappers/wrappers.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' empty/empty.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/empty";' empty/empty.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' timestamp/timestamp.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/timestamp";' timestamp/timestamp.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' struct/struct.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/struct;structpb";' struct/struct.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' any/any.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/any";' any/any.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' duration/duration.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/duration";' duration/duration.proto

	sed -i '/package google.protobuf/aimport "github.com/johanbrandhorst/protobuf/proto/gopherjs.proto";' wrappers/wrappers.proto
	sed -i '/option go_package/aoption (gopherjs.gopherjs_package) = "github.com/johanbrandhorst/protobuf/ptypes/wrappers";' wrappers/wrappers.proto

regenerate:
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./empty/empty.proto
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./timestamp/timestamp.proto
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./struct/struct.proto
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./any/any.proto
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./duration/duration.proto
	protoc -I. -I$$GOPATH/src --gopherjs_out=:${GOPATH}/src ./wrappers/wrappers.proto

clean:
	rm -f empty/*.go
	rm -f timestamp/*.go
	rm -f struct/*.go
	rm -f any/*.go
	rm -f duration/*.go
	rm -f wrappers/*.go
