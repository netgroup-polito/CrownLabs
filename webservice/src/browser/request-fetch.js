const extend = require('extend');
const fetch = require('portable-fetch');
const qs = require('qs');
const querystring = require('querystring');

/* A stub for the `http.ClientResponse` type */
const FakeResponse = (function() {
    function FakeResponse(fetchResponse) {
        this._fetchResponse = fetchResponse;
    }

    FakeResponse.prototype = {
        get statusCode() {
            return this._fetchResponse.status;
        },
    };

    return FakeResponse;
})();

/* Copied from the 'requests' library */
function initParams(uri, options, callback) {
    if (typeof options === 'function') {
        callback = options;
    }

    const params = {};
    if (typeof options === 'object') {
        extend(params, options, { uri: uri });
    } else if (typeof uri === 'string') {
        extend(params, { uri: uri });
    } else {
        extend(params, uri);
    }

    params.callback = callback || params.callback;

    return params;
}

function requestOptionsToFetchOptions(options) {
    const supportedKeys = [
        'method',
        'headers',
        'uri',
        'json',
        'body',
        /* Unused here but 'allowed' */
        'callback',
        'useQuerystring',
        'qs',
    ];

    Object.keys(options).forEach(function(key) {
        if (supportedKeys.indexOf(key) === -1) {
            throw new Error('Unsupported option: ' + key);
        }
    });

    const result = {
        method: 'GET',
        headers: {},
    };

    if (typeof options.method !== 'undefined') {
        result.method = options.method;
    }
    if (typeof options.headers !== 'undefined') {
        result.headers = Object.assign(result.headers, options.headers);
    }
    if (typeof options.body !== 'undefined') {
        if (options.json) {
            result.headers['Content-Type'] = options.headers['Content-Type'] || 'application/json';
            result.body = JSON.stringify(options.body);
        } else {
            result.body = options.body;
        }
    }

    return result;
}

function hasQs(params) {
    if (typeof params.qs === 'undefined') {
        return false;
    }

    if (params.qs.constructor !== Object) {
        throw new TypeError('Invalid type for qs');
    }

    return Object.keys(params.qs).length !== 0;
}

module.exports = function request(uri, options, callback) {
    if (typeof uri === 'undefined') {
        throw new Error('undefined is not a valid uri or options object.');
    }

    const params = initParams(uri, options, callback);

    if (params.method === 'HEAD' && typeof params.body !== 'undefined') {
        throw new Error('HTTP HEAD requests MUST NOT include a request body.');
    }

    const fetchUri = !hasQs(params)
        ? params.uri
        : params.uri + '?' + (params.useQuerystring ? querystring.stringify : qs.stringify)(params.qs);

    fetch(fetchUri, requestOptionsToFetchOptions(params))
        .then(function(response) {
            (params.json ? response.json() : response.text())
                .then(function(body) {
                    params.callback(null, new FakeResponse(response), body);
                })
                .catch(function(error) {
                    params.callback(error, new FakeResponse(response), null);
                });
        })
        .catch(function(error) {
            params.callback(error, null, null);
        });
};
