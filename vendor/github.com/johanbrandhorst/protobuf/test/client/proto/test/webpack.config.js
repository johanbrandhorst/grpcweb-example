module.exports = {
    entry: "./test_pb.js",
    output: {
        filename: "test_pb.inc.js",
    },
    externals: {
        "google-protobuf": "window",
        "google-protobuf/google/protobuf/empty_pb.js": "window.proto.google.protobuf"
    }
};
