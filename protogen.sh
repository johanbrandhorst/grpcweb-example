#!/usr/bin/env bash
# Copyright 2017 Johan Brandhorst. All Rights Reserved.
# See LICENSE for licensing terms.

# Generate protofiles
protoc proto/library/book_service.proto \
    --gopherjs_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/johanbrandhorst/protobuf/ptypes/timestamp:./client/ \
    --go_out=plugins=grpc:./server/

