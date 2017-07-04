#!/usr/bin/env bash
# Copyright 2017 Johan Brandhorst. All Rights Reserved.
# See LICENSE for licensing terms.

# Generate protofiles
protoc proto/library/book_service.proto \
    --js_out=import_style=commonjs,binary:./client/ \
    --gopherjs_out=plugins=grpc,Mgoogle/protobuf/timestamp.proto=github.com/johanbrandhorst/protobuf/ptypes/timestamp:./client/ \
    --go_out=plugins=grpc:./server/

# Replace top level import with global reference
sed -i "s;require('google-protobuf');\$global;g" ./client/proto/library/book_service_pb.js
# Replace any well known type imports with correct namespace
sed -i -E "s;require\('google-protobuf/.*'\);\$global.proto.google.protobuf;g" ./client/proto/library/book_service_pb.js
# Remove export statement
sed -i -E "s/goog\.object\.extend\(exports, proto\..*\);$//g" ./client/proto/library/book_service_pb.js

# Move original file
mv ./client/proto/library/book_service_pb.js ./client/proto/library/book_service_pb.inc.js
