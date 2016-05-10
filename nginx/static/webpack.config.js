var path = require("path");
var webpack = require('webpack');
var ExtractTextPlugin = require("extract-text-webpack-plugin");

var PROD = JSON.parse(process.env.PROD_ENV || '0');

var config = {
  entry: path.resolve(__dirname, 'js/app.jsx'),
  resolve: {
    alias: {}
  },
  output: {
    path: path.resolve(__dirname, 'build'),
    filename: 'bundle.js'
  },
  module: {
    noParse: [],
    loaders: [
      {
        test: /\.jsx?$/,
        exclude: /node_modules/,
        loader: 'babel',
        query: {
          presets: ['react', 'es2015']
        },
      },
      {
        test: /\.css$/, // Only .css files
        loader: ExtractTextPlugin.extract("style-loader", "css-loader")
        // loader: 'style-loader!css-loader' // Run both loaders
      },
      {
        test: /jquery\.js$/,
        loader: 'expose?$'
      },
      {
        test: /jquery\.js$/,
        loader: 'expose?jQuery'
      },
      {
        test: require.resolve("react"),
        loader: "expose?React"
      },
    ]
  },
    plugins: PROD ? [
        new ExtractTextPlugin("[name].css")
    ,
      new webpack.optimize.UglifyJsPlugin({
          compress: { warnings: false }
      })
    ] : [
        new ExtractTextPlugin("[name].css")
    ]
};

module.exports = config;
