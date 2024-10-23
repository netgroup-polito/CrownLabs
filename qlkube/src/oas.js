const logger = require('pino')({ useLevelLabels: true });

const openApiPaths = ['openapi/v2', 'swagger2.json', 'swagger.json'];

// execute parallel requests to possible open api endpoints and return first success
module.exports.getOpenApiSpec = async (url, token) => {
  const { got } = await import('got');
  for (let p of openApiPaths) {
    const res = await got(p, {
      prefixUrl: url,
      responseType: 'json',
      timeout: { request: 5000 },
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
      .then(r => {
        if (
          r.headers['content-type'] &&
          r.headers['content-type'].includes('application/json')
        ) {
          logger.info(
            { url, path: p },
            'successfully retrieved open api spec from this path'
          );
          return r.body;
        }
      })
      .catch(err => {
        if (err.response && err.response.statusCode === 404) {
          logger.info(
            { cause: err.message, url, path: p },
            'failed to retrieve open api spec from this path - will try another'
          );
          return null;
        } else {
          if (
            err.response?.headers['content-type'] &&
            !err.response?.headers['content-type'].includes('application/json')
          ) {
            logger.info(
              { cause: err.message, url, path: p },
              'failed to retrieve open api spec from this path - will try another'
            );
            return null;
          } else {
            if (err.response && err.response.statusCode === 403) {
              logger.info(
                { cause: err.message, url, path: p },
                'failed to retrieve open api spec from this path (403) - will try another'
              );
              return null;
            } else {
              throw err;
            }
          }
        }
      });
    if (res) {
      return res;
    }
  }

  throw new Error('Failed to retrieve OpenAPI spec from any path');
};
