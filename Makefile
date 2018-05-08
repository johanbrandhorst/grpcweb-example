regenerate:

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

generate_cert:
	cd insecure && go run "$$(go env GOROOT)/src/crypto/tls/generate_cert.go" \
		--host=localhost,127.0.0.1 \
		--ecdsa-curve=P256 \
		--ca=true
