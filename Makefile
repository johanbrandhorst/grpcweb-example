regenerate:

	protoc -I. -Ivendor/ proto/library/book_service.proto \
    	--gopherjs_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/johanbrandhorst/protobuf/ptypes/timestamp:$$GOPATH/src \
    	--go_out=plugins=grpc:$$GOPATH/src
	go generate ./client/...
