#!/usr/bin/env bash
# Copyright 2017 Johan Brandhorst. All Rights Reserved.
# See LICENSE for licensing terms.

# Generate protofiles
protoc proto/library/book_service.proto \
    --js_out=import_style=commonjs,binary:./client/ \
    --gopherjs_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/johanbrandhorst/protobuf/ptypes/timestamp:./client/ \
    --go_out=plugins=grpc:./server/

(cd client/proto/library && webpack && rm book_service_pb.js)

