const webpack = require('webpack');
const path = require('path');

module.exports = {
  entry: {
    jspb: "./node_modules/google-protobuf/google-protobuf.js",
    grpc: "./node_modules/grpc-web-client/dist/index.js",
  },
  output: {
    path: path.resolve(__dirname),
    filename: '[name].inc.js',
    libraryTarget: 'this',
  },
  resolve: {
    extensions: [".js"]
  },
  plugins: [
    new webpack.optimize.UglifyJsPlugin()
  ]
};
