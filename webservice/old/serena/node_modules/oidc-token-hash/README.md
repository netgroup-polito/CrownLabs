# oidc-token-hash

[![build][travis-image]][travis-url] [![codecov][codecov-image]][codecov-url]

oidc-token-hash validates (and generates) ID Token `_hash` claims such as `at_hash` or `c_hash`.

## Usage

Validating
```js
const oidcTokenHash = require('oidc-token-hash');

const at_hash = 'x7vk7f6BvQj0jQHYFIk4ag';
const access_token = 'YmJiZTAwYmYtMzgyOC00NzhkLTkyOTItNjJjNDM3MGYzOWIy9sFhvH8K_x8UIHj1osisS57f5DduL-ar_qw5jl3lthwpMjm283aVMQXDmoqqqydDSqJfbhptzw8rUVwkuQbolw';

oidcTokenHash(at_hash, access_token, 'RS256'); // => true
oidcTokenHash(at_hash, 'foobar', 'RS256'); // => false
oidcTokenHash.valid('foobar', access_token, 'RS256'); // => false
```

Generating
```js
// access_token from first example
oidcTokenHash.generate(access_token, 'RS256'); // => 'x7vk7f6BvQj0jQHYFIk4ag'
oidcTokenHash.generate(access_token, 'HS384'); // => 'ups_76_7CCye_J1WIyGHKVG7AAs2olYm'
oidcTokenHash.generate(access_token, 'ES512'); // => 'EGEAhGYyfuwDaVTifvrWSoD5MSy_5hZPy6I7Vm-7pTQ'
```

## Changelog
- 3.0.2 - removed `base64url` dependency
- 3.0.1 - `base64url` comeback
- 3.0.0 - drop lts/4 support, replace base64url dependency
- 2.0.0 - rather then assuming the alg based on the hash length `#valid()` now requires a third
  argument with the JOSE header `alg` value, resulting in strict validation
- 1.0.0 - initial release

[travis-image]: https://api.travis-ci.com/panva/oidc-token-hash.svg?branch=master
[travis-url]: https://travis-ci.com/panva/oidc-token-hash
[codecov-image]: https://img.shields.io/codecov/c/github/panva/oidc-token-hash/master.svg
[codecov-url]: https://codecov.io/gh/panva/oidc-token-hash
