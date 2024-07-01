const { fetchJson, logger } = require('./utils');

async function getOpenApiSpec(baseUrl, paths, token) {
  for (const path of paths) {
    try {
      const url = [baseUrl, path].join('/');
      const res = await fetchJson(url, token ? { Authorization: `Bearer ${token}` } : {});
      if (res.statusCode >= 200 && res.statusCode < 300 && res.body) {
        logger.warn({ baseUrl, path }, 'OpenApi retrieved');
        return res.body;
      }
      logger.warn({ res, baseUrl, path }, 'OpenApi invalid response');
    } catch (e) {
      logger.warn({ baseUrl, path, err: e.message }, 'Could not retrieve OpenApi Spec');
    }
  }
  logger.error({ baseUrl, paths }, 'OpenApi retrieval failed');
  throw new Error('ENOAPI');
}

module.exports.getOpenApiSpec = getOpenApiSpec;
