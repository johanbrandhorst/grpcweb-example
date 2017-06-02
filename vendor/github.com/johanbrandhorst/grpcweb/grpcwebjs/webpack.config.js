const webpack = require('webpack');
const path = require('path');

module.exports = {
  entry: "./node_modules/grpc-web-client/dist/index.js",
  output: {
    path: path.resolve(__dirname),
    filename: 'grpc.inc.js',
    libraryTarget: 'this',
  },
  resolve: {
    extensions: [".js"]
  },
  plugins: [
    new webpack.optimize.UglifyJsPlugin()
  ]
};
