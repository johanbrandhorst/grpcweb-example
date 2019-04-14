generate:

	protoc -I. -Ivendor/ proto/library/book_service.proto \
    	--gopherjs_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/johanbrandhorst/protobuf/ptypes/timestamp:$$GOPATH/src \
    	--go_out=plugins=grpc:$$GOPATH/src
	go generate ./client/...

install:
	go install ./vendor/github.com/golang/protobuf/protoc-gen-go \
		./vendor/github.com/johanbrandhorst/protobuf/protoc-gen-gopherjs \
		./vendor/myitcv.io/react/cmd/reactGen \
		./vendor/myitcv.io/immutable/cmd/immutableGen \
		./vendor/github.com/foobaz/go-zopfli \
		./vendor/github.com/gopherjs/gopherjs

deploy:
	gcloud builds submit --tag gcr.io/grpc-web-serverless/grpcweb-example
