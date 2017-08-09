module.exports = {
    entry: [
        "./multi3_pb.js",
        "./multi2_pb.js",
        "./multi1_pb.js"
    ],
    output: {
        filename: "multitest_pb.inc.js",
    },
    externals: {
        "google-protobuf": "window"
    }
};
