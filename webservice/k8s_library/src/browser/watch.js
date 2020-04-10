const fetch = require('portable-fetch');
const byline = require('byline');
const stream = require('stream');
const util = require('util');
const querystring = require('querystring');

module.exports.watch = function watch(config, path, queryParams, callback, done) {
    const url = config.getCurrentCluster().server + path;
    queryParams.watch = true;

    const requestOptions = {
        headers: {},
    };

    config.applyToRequest(requestOptions);

    const keys = Object.keys(requestOptions);
    if (keys.length !== 1 && keys[0] !== 'headers') {
        throw new Error('Unexpected request options passed');
    }

    const stream = new byline.LineStream();
    stream.on('data', function (data) {
        if (data instanceof Buffer) {
            obj = JSON.parse(data.toString());
        } else {
            obj = JSON.parse(data);
        }

        if (typeof obj === 'object' && obj.object) {
            callback(obj.type, obj.object);
        } else {
            throw new Error('unexpected ' + typeof obj + ': ' + JSON.stringify(obj));
        }
    });

    stream.on('end', function () {
        done(null);
    });

    stream.on('error', function (error) {
        done(error);
    });

    const fetchOptions = {
        method: 'GET',
        headers: requestOptions.headers,
    };

    fetch(url + '?' + querystring.stringify(queryParams), fetchOptions)
        .then(function (response) {
            new FetchReaderStream(response.body.getReader(), {}, done).pipe(stream);
        })
        .catch(function (error) {
            done(error);
        })
        .catch(() => {
            done(null);
        })
};

const FetchReaderStream = (function () {
    function FetchReaderStream(reader, options, doneFunc) {
        if (!(this instanceof FetchReaderStream)) {
            return new FetchReaderStream(reader);
        }

        this._doneFunc = doneFunc;
        this._isDestroyed = false;
        this._reader = reader;
        stream.Readable.call(this, options);
    }

    util.inherits(FetchReaderStream, stream.Readable);

    FetchReaderStream.prototype._read = function (_size) {
        const self = this;

        function loop() {
            self._reader.read()
                .then(function (data) {
                    if (!self._isDestroyed) {
                        if (data.done) {
                            self.push(null);
                            return;
                        }

                        const keepPushing = self.push(data.value);
                        if (!keepPushing) {
                            return;
                        }

                        loop();
                    }
                })
                .catch(function () {
                    if (!self._isDestroyed) {
                        self._isDestroyed = true;
                        self._doneFunc(null);
                    }
                });
        }

        loop();
    };

    return FetchReaderStream;
})();
