# p-pipe [![Build Status](https://travis-ci.org/sindresorhus/p-pipe.svg?branch=master)](https://travis-ci.org/sindresorhus/p-pipe)

> Compose promise-returning & async functions into a reusable pipeline


## Install

```
$ npm install p-pipe
```


## Usage

```js
const pPipe = require('p-pipe');

const addUnicorn = async string => `${string} Unicorn`;
const addRainbow = async string => `${string} Rainbow`;

const pipeline = pPipe(addUnicorn, addRainbow);

(async () => {
	console.log(await pipeline('❤️'));
	//=> '❤️ Unicorn Rainbow'
})();
```


## API

### pPipe(input…)

The `input` functions are applied from left to right.

#### input

Type: `Function`

Expected to return a `Promise` or any value.


## Related

- [p-each-series](https://github.com/sindresorhus/p-each-series) - Iterate over promises serially
- [p-series](https://github.com/sindresorhus/p-series) - Run promise-returning & async functions in series
- [p-waterfall](https://github.com/sindresorhus/p-waterfall) - Run promise-returning & async functions in series, each passing its result to the next
- [More…](https://github.com/sindresorhus/promise-fun)


## License

MIT © [Sindre Sorhus](https://sindresorhus.com)
