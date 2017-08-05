# protobuf
GopherJS Bindings for ProtobufJS and gRPC-Web

## Packages
This repo consists of 3 major parts:

### GopherJS ProtobufJS Bindings (jspb)
This is a simple GopherJS binding around the node `google-protobuf` package.
Importing it into any GopherJS source allows usage of ProtobufJS functionality.

### GopherJS gRPC-Web Client Bindings (grpcweb)
This is a GopherJS binding around the Improbable gRPC-Web client.

### GopherJS Protobuf Generator (protoc-gen-gopherjs)
This is a GopherJS client code generator for the Google Protobuf format.
It generates code for interfacing with any gRPC services exposing a
gRPC-Web spec compatible interface. It uses `jspb` and `grpcweb`.
It is the main entrypoint for using the protobuf/gRPC GopherJS bindings.

## Contributions
Contributions are very welcome, please submit issues and PRs for review.

## An example
See [the example repo](https://github.com/johanbrandhorst/grpcweb-example) for an
example use of `protoc-gen-gopherjs`.
