const webpack = require("webpack");

module.exports = {
    entry: "./book_service_pb.js",
    output: {
        filename: "book_service_pb.inc.js",
    },
    externals: {
        "google-protobuf": "window",
        "google-protobuf/google/protobuf/timestamp_pb.js": "window.proto.google.protobuf"
    },
    plugins: [
        new webpack.optimize.UglifyJsPlugin()
    ]
};
