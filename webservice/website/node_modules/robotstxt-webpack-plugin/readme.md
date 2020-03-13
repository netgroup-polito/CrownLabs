[![NPM version](https://img.shields.io/npm/v/robotstxt-webpack-plugin.svg)](https://www.npmjs.org/package/robotstxt-webpack-plugin)
[![Travis Build Status](https://img.shields.io/travis/itgalaxy/robotstxt-webpack-plugin/master.svg?label=build)](https://travis-ci.org/itgalaxy/robotstxt-webpack-plugin)
[![dependencies Status](https://david-dm.org/itgalaxy/robotstxt-webpack-plugin/status.svg)](https://david-dm.org/itgalaxy/robotstxt-webpack-plugin)
[![devDependencies Status](https://david-dm.org/itgalaxy/robotstxt-webpack-plugin/dev-status.svg)](https://david-dm.org/itgalaxy/robotstxt-webpack-plugin?type=dev)

# robotstxt-webpack-plugin

Generating `robots.txt` using webpack.

Why your need [robots.txt](https://support.google.com/webmasters/answer/6062608?hl=en)?

Webpack plugin for [generate-robotstxt](https://github.com/itgalaxy/generate-robotstxt/) package.

## Getting Started

To begin, you'll need to install `robotstxt-webpack-plugin`:

```console
npm install --save-dev robotstxt-webpack-plugin
```

**webpack.config.js**

```js
const RobotstxtPlugin = require("robotstxt-webpack-plugin");

const options = {}; // see options below

module.exports = {
  plugins: [new RobotstxtPlugin(options)]
};
```

## Options

- `filePath` - (optional) path for robots.txt (should be contain full path include `robots.txt` file name, example - `path/to/robots.txt`).
- `General options` - see [generate-robotstxt](https://github.com/itgalaxy/generate-robotstxt) options.

## Related

- [generate-robotstxt](https://github.com/itgalaxy/generate-robotstxt) - api for this package.

## Contribution

Feel free to push your code if you agree with publishing under the MIT license.

## Changelog

[MIT](./CHANGELOG.md)

## License

[MIT](./LICENSE)
