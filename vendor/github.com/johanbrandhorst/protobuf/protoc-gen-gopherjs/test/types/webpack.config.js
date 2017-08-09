module.exports = {
    entry: "./types_pb.js",
    output: {
        filename: "types_pb.inc.js",
    },
    externals: {
        "google-protobuf": "window",
        "../multi/multi1_pb.js": "window.proto.multitest",
    }
};
