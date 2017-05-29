const webpack = require('webpack');
const path = require('path');

module.exports = {
  entry:  "./node_modules/google-protobuf/google-protobuf.js",
  output: {
    path: path.resolve(__dirname),
    filename: 'jspb.inc.js',
    libraryTarget: 'this',
  },
  resolve: {
    extensions: [".js"]
  },
  plugins: [
    new webpack.optimize.UglifyJsPlugin()
  ]
};
