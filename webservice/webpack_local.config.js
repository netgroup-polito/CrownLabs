const webpack = require('webpack');
const webpackDockerConfig = require('./webpack_docker.config');

webpackDockerConfig.plugins.push(
  new webpack.DefinePlugin({
    OIDC_PROVIDER_URL: JSON.stringify(process.env.OIDC_PROVIDER_URL),
    OIDC_CLIENT_ID: JSON.stringify(process.env.OIDC_CLIENT_ID),
    APISERVER_URL: JSON.stringify(process.env.APISERVER_URL),
    OIDC_REDIRECT_URI: JSON.stringify(process.env.OIDC_REDIRECT_URI),
    OIDC_CLIENT_SECRET: JSON.stringify(process.env.OIDC_CLIENT_SECRET)
  })
);
module.exports = webpackDockerConfig;
