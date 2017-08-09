const webpack = require("webpack");

module.exports = {
    entry: {
        any: './any/any_pb.js',
        duration: './duration/duration_pb.js',
        empty: './empty/empty_pb.js',
        struct: './struct/struct_pb.js',
        timestamp: './timestamp/timestamp_pb.js',
        wrappers: './wrappers/wrappers_pb.js',
    },
    output: {
        filename: '[name]/[name]_pb.inc.js',
    },
    externals: {
        "google-protobuf": "window"
    },
    plugins: [
        new webpack.optimize.UglifyJsPlugin()
    ]
};
