# Websocket to bi-directional gRPC stream proxy

[![GoDoc](https://godoc.org/github.com/johanbrandhorst/protobuf/wsproxy?status.svg)](https://godoc.org/github.com/johanbrandhorst/protobuf/wsproxy)

This package implements a Websocket to gRPC streaming proxy.
It supports both client side streaming and bi-directional streams.

## Using the proxy
The proxy exposes the `WrapServer` method, which creates
a `http.Handler`. It could be used like so:

```go
gs := grpc.NewServer()
... // Perform registration to gs
handler := wsproxy.WrapServer(gs)
http.ListenAndServe("localhost:8080", handler)
```

Most of the time you'll want to set up TLS as well.
[See the test setup](../test/server/main.go) for an example of this.

## Connecting to the proxy
A Websocket client wishing to perform gRPC streaming
should send a websocket upgrade request to the same URL
as would have been used for the gRPC streaming request.
This will usually look something like this:

```
https://host/PackageName.ServiceName/MethodName
```

The websocket upgrade request will be intercepted and a gRPC
client stream will be opened to the underlying gRPC server.

## The "protocol"
After the connection has been established, the following message format is expected:

1. Each incoming message is a proto encoded binary blob with a gRPC style 5 byte
header prefix indicating the compression status and the size of the message.
This message is passed directly to the gRPC stream.
A message consisting of the ASCII string "clos" indicates a wish
to close the stream. This tells the stream the last message has been seen.
1. Replies are sent as encoded proto binary blobs with previously described gRPC
style header prefixed.
1. Errors are sent with a websocket Close control message.
gRPC error codes are mapped by adding 4000 to the error code.
1. Trailers and Headers are currently unsupported.
