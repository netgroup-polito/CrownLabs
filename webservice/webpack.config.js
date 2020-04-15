const HtmlWebpackPlugin = require('html-webpack-plugin');
const RobotstxtPlugin = require('robotstxt-webpack-plugin');
const path = require('path');
const webpack = require('webpack');

module.exports = {
  context: __dirname,
  entry: ['@babel/polyfill', './src/index.js'],
  output: {
    path: path.resolve(__dirname, 'dist'),
    filename: 'bundle.js',
    publicPath: '/'
  },
  devServer: {
    host: '0.0.0.0',
    port: 8000,
    historyApiFallback: true
  },
  module: {
    rules: [
      {
        test: /\.js$/,
        use: 'babel-loader'
      },
      {
        test: /\.css$/,
        use: ['style-loader', 'css-loader']
      },
      {
        test: /\.(gif|png|jpe?g|svg)$/i,
        use: [
          {
            loader: 'file-loader',
            options: {
              esModule: false
            }
          }
        ]
      },
      {
        test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'url-loader?limit=10000&mimetype=application/font-woff'
      },
      {
        test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/,
        loader: 'file-loader'
      }
    ]
  },
  plugins: [
    new HtmlWebpackPlugin({
      filename: 'index.html',
      title: 'CrownLabs',
      meta: {
        viewport: 'width=device-width, initial-scale=1',
        'theme-color': '#000000',
        description: 'CrownLabs website'
      },
      favicon: 'src/assets/crown.png'
    }),
    new webpack.DefinePlugin({
      OIDC_PROVIDER_URL: JSON.stringify(process.env.OIDC_PROVIDER_URL),
      OIDC_CLIENT_ID: JSON.stringify(process.env.OIDC_CLIENT_ID),
      APISERVER_URL: JSON.stringify(process.env.APISERVER_URL),
      OIDC_REDIRECT_URI: JSON.stringify(process.env.OIDC_REDIRECT_URI),
      OIDC_CLIENT_SECRET: JSON.stringify(process.env.OIDC_CLIENT_SECRET)
    }),
    new RobotstxtPlugin({
      'User-agent': '*',
      Disallow: ''
    })
  ]
};
