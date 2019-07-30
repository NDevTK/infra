
const path = require('path');

module.exports = {
  entry: {
    'main': './demo.ts',
  },
  mode: 'development',
  devtool: 'inline-source-map',
  module: {
    rules: [
      {
        test: /\.tsx?$/,
        loader: 'ts-loader',
      },
    ],
  },
  devServer: {
    publicPath: '/dist/',
  },
  resolve: {
    modules: [__dirname, 'node_modules'],
    extensions: ['.ts', '.js'],
  },
  output: {
    filename: '[name].min.js',
    path: path.resolve(__dirname, 'dist'),
  },
};
