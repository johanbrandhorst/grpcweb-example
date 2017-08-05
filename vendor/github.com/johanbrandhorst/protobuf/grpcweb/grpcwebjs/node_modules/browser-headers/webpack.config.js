const path = require('path');

module.exports = {
  entry: "./test/BrowserHeaders.spec.ts",
  output: {
    path: path.resolve(__dirname, 'test', 'build'),
    filename: 'integration-tests.js',
  },
  devtool: 'source-map',
  module: {
    rules: [
      {
        test: /\.js$/,
        include: /src|test|node_modules/,
        loader: 'babel-loader?cacheDirectory'
      },
      {
        test: /\.ts$/,
        include: /src|test|node_modules/,
        loader: "babel-loader?cacheDirectory!ts-loader"
      }
    ]
  },
  plugins: [],
  resolve: {
    extensions: [".ts", ".js"]
  }
};

