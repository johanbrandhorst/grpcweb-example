# GopherJS Bindings for ProtobufJS and gRPC-Web
[![CircleCI](https://circleci.com/gh/johanbrandhorst/protobuf/tree/master.svg?style=svg)](https://circleci.com/gh/johanbrandhorst/protobuf/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/johanbrandhorst/protobuf)](https://goreportcard.com/report/github.com/johanbrandhorst/protobuf)

### [GopherJS Protobuf Generator](./protoc-gen-gopherjs/README.md)
This is a GopherJS client code generator for the Google Protobuf format.
It generates code for interfacing with any gRPC services exposing a
gRPC-Web spec compatible interface. It uses `jspb` and `grpcweb`.
It is the main entrypoint for using the protobuf/gRPC GopherJS bindings.

### [GopherJS ProtobufJS Bindings](./jspb/README.md)
This is a simple GopherJS binding around the node `google-protobuf` package.
Importing it into any GopherJS source allows usage of ProtobufJS functionality.

### [GopherJS gRPC-Web Client Bindings](./grpcweb/README.md)
This is a GopherJS binding around the Improbable gRPC-Web client and
the websocket proxy.


## Contributions
Contributions are very welcome, please submit issues and PRs for review.

## Demo
See [the example repo](https://github.com/johanbrandhorst/grpcweb-example)
and [the demo website](https://grpcweb.jbrandhorst.com)
for an example use of the Protobuf and gRPC-Web bindings.
