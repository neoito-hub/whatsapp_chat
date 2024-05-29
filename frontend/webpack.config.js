/* eslint-disable import/extensions */
/* eslint-disable no-underscore-dangle */
import HtmlWebpackPlugin from 'html-webpack-plugin'
import path from 'path'
import webpack from 'webpack'
import dotenv from 'dotenv'
import getFederationConfig from './federation.config.js'

const { ModuleFederationPlugin } = webpack.container

// For testing take pull from Appblocks/@appblocks/node-sdk and npm install from path
// import { env } from '@appblocks/node-sdk';
// env.init();

const env = process.env.NODE_ENV || 'dev'
const Dotenv = dotenv.config({
  path: `./.env.${env}`,
})

const __dirname = path.resolve()

const port = 3011

export default {
  entry: './src/index',
  mode: 'development',
  devServer: {
    port,
    static: path.join(__dirname, 'dist'),
    hot: true,
    // watchContentBase: true,
    historyApiFallback: true,
  },
  externals: {
    env: JSON.stringify(process.env),
  },
  output: {
    publicPath: 'auto',
  },
  module: {
    rules: [
      {
        test: /.js$/,
        loader: 'babel-loader',
        options: {
          presets: ['@babel/preset-react'],
        },
      },
      {
        test: /\.(jpg|png|svg)$/,
        use: {
          loader: 'url-loader',
        },
      },
      {
        test: /\.s[ac]ss$/i,
        use: [
          'style-loader',
          'css-loader',
          {
            loader: 'sass-loader',
          },
        ],
      },
      {
        test: /\.css$/i,
        use: ['style-loader', 'css-loader'],
      },
      {
        test: /.m?js/,
        type: 'javascript/auto',
      },
      {
        test: /.m?js/,
        resolve: {
          fullySpecified: false,
        },
      },
    ],
  },
  plugins: [
    new webpack.DefinePlugin({
      process: { env: JSON.stringify(Dotenv.parsed) },
    }),
    // new webpack.DefinePlugin({
    //   process: { env: JSON.stringify(process.env) },
    // }),
    new ModuleFederationPlugin(getFederationConfig()),
    new HtmlWebpackPlugin({
      template: './public/index.html',
    }),
  ],
}
