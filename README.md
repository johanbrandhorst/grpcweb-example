# GopherJS over Improbable gRPC-Web to Go gRPC backend
[![CircleCI](https://circleci.com/gh/johanbrandhorst/grpcweb-example.svg?style=svg)](https://circleci.com/gh/johanbrandhorst/grpcweb-example)

An example implementation of a
[GopherJS React](https://myitcv.io/react)
client talking to a Go gRPC server using the Improbable gRPC-Web implementation and
the [wsproxy](https://github.com/johanbrandhorst/protobuf/tree/master/wsproxy)
through the
[protoc-gen-gopherjs](https://github.com/johanbrandhorst/protobuf/tree/master/protoc-gen-gopherjs)
bindings generator.

A demo of this application can be found on
[https://grpcweb.jbrandhorst.com](https://grpcweb.jbrandhorst.com).

## Developing
To run the server on `https://localhost:10000`:

```
$ go run main.go
```

If you want to make any changes to the client, you'll need to install GopherJS:

```
$ go get -u github.com/gopherjs/gopherjs
```

Then you'll need to also install some vendored generators:

```
$ go install ./vendor/myitcv.io/react/cmd/reactGen ./vendor/myitcv.io/immutable/cmd/immutableGen
```

After that, any changes you make to the proto file in `./proto/` or the JS client code
in `./client/` should be followed by

```
$ make regenerate
```

You may need to generate the client code twice as the first time will run `reactGen` and
`immutableGen` which might be necessary for the subsequent `gopherjs build` to work.
