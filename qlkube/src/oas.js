const got = require('got');
const promiseAny = require('promise-any');
const logger = require('pino')({ useLevelLabels: true });

const openApiPaths = ['/openapi/v2', '/swagger.json'];

// execute parallel requests to possible open api endpoints and return first success
module.exports = async function getOpenApiSpec(url, token) {
  let gotProms = [];
  for (let p of openApiPaths) {
    const gotProm = got(p, {
      baseUrl: url,
      json: true,
      timeout: 5 * 1000,
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => {
        logger.info(
          { url, path: p },
          'successfully retrieved open api spec from this path'
        );
        return r.body;
      })
      .catch(err => {
        logger.info(
          { cause: err.message, url, path: p },
          'failed to retrieve open api spec from this path - will try another'
        );
        throw err;
      });
    gotProms.push(gotProm);
  }
  return promiseAny(gotProms);
};
