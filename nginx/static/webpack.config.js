var path = require("path");
var ExtractTextPlugin = require("extract-text-webpack-plugin");

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
    plugins: [
        new ExtractTextPlugin("[name].css")
    ]
};

module.exports = config;
