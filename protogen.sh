#!/usr/bin/env bash
# Copyright 2017 Johan Brandhorst. All Rights Reserved.
# See LICENSE for licensing terms.

# Generate protofiles
protoc proto/library/book_service.proto \
    --js_out=import_style=commonjs,binary:./client/ \
    --go_out=plugins=grpc:./server/

# Replace imports and exports
python3 -c "with open('./client/proto/library/book_service_pb.js') as f:
    c = f.read()
    c = c.replace('require(\'google-protobuf\')', '\$global')
    c = c.replace('goog.object.extend(exports, proto.library);', '')
    with open('./client/proto/library/book_service_pb.inc.js', 'w') as out:
        out.write(c)"

# Remove original file
rm ./client/proto/library/book_service_pb.js
