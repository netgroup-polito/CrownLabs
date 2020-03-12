const HtmlWebpackPlugin = require('html-webpack-plugin');
const path = require('path');
const webpack = require('webpack');
const ManifestPlugin = require('webpack-manifest-plugin');
const RobotstxtPlugin = require("robotstxt-webpack-plugin");

module.exports = {
    context: __dirname,
    entry: './src/index.js',
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: 'bundle.js',
        publicPath: '/',
    },
    devServer: {
        host: "0.0.0.0",
        port: 8000,
        historyApiFallback: true,
        contentBase: './public'
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                use: 'babel-loader'
            },
            {
                test: /\.css$/,
                use: ['style-loader', 'css-loader'],
            },
            {
                test: /\.(gif|png|jpe?g|svg)$/i,
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            esModule: false,
                        },
                    },
                ],
            },
            {
                test: /\.woff(2)?(\?v=[0-9]\.[0-9]\.[0-9])?$/,
                loader: "url-loader?limit=10000&mimetype=application/font-woff"
            },
            {test: /\.(ttf|eot|svg)(\?v=[0-9]\.[0-9]\.[0-9])?$/, loader: "file-loader"}
        ]
    },
    plugins: [
        new HtmlWebpackPlugin({
            template: path.resolve(__dirname, 'public/index.html'),
            filename: 'index.html'
        }),
        new webpack.DefinePlugin({
            OIDC_PROVIDER_URL: JSON.stringify(process.env.OIDC_PROVIDER_URL),
            OIDC_CLIENT_ID: JSON.stringify(process.env.OIDC_CLIENT_ID),
            APISERVER_URL: JSON.stringify(process.env.APISERVER_URL),
        }),
        new ManifestPlugin({
            short_name: "React App",
            name: "Create React App Sample",
            icons: [
                {
                    src: "favicon.ico",
                    sizes: "64x64 32x32 24x24 16x16",
                    type: "image/x-icon"
                },
                {
                    src: "logo192.png",
                    type: "image/png",
                    sizes: "192x192"
                },
                {
                    src: "logo512.png",
                    type: "image/png",
                    sizes: "512x512"
                }
            ],
            start_url: ".",
            display: "standalone",
            theme_color: "#000000",
            background_color: "#ffffff"
        }),
        new RobotstxtPlugin({
            "User-agent": "*",
            "Disallow": ""
        })
    ]
};
