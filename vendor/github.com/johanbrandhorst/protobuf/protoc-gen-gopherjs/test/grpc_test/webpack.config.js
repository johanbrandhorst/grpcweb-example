module.exports = {
    entry: "./test_pb.js",
    output: {
        filename: "test_pb.inc.js",
    },
    externals: {
        "google-protobuf": "window",
        "../multi/multi1_pb.js": "window.proto.multitest"
    }
};
