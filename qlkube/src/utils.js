const { GraphQLError } = require('graphql');
const https = require('https');
const http = require('http');
const pino = require('pino');

const logger = pino({
  base: false,
  translateTime: true,
  timestamp: () => `,"time":"${new Date(Date.now()).toISOString()}"`,
  formatters: {
    level: (level) => ({ level }),
  },
});

/**
 * Converts query name in the query type
 * e.g. query name: itPolitoCrownlabsV1alpha2Instance
 *      query type: ItPolitoCrownlabsV1alpha2Instance
 */
function capitalizeType(name) {
  return name && name[0].toUpperCase() + name.slice(1);
}

function uncapitalizeType(name) {
  return name && name[0].toLowerCase() + name.slice(1);
}

function getBearerToken(connectionParams) {
  if (!connectionParams) throw new Error('Parameter connectionParams cannot be empty!');
  const auth = connectionParams.authorization || connectionParams.Authorization;
  if (!auth) {
    throw new GraphQLError('Token Error! Token absent!', {
      extensions: {
        code: 'FORBIDDEN',
        connectionParams,
        http: { status: 403 },
      },
    });
  }
  const token = auth.split(/\s+/)[1];
  if (!token) {
    throw new GraphQLError('Token Error! Token not valid!', {
      extensions: {
        code: 'FORBIDDEN',
        http: { status: 403 },
      },
    });
  }

  return token;
}

const graphqlQueryRegistry = {
  // Fires whenever a GraphQL request is received from a client.
  async requestDidStart(requestContext) {
    const { variables, operationName } = requestContext.request;
    if (operationName !== 'IntrospectionQuery') {
      logger.info({ variables, operationName }, 'Request started');
    }
  },
};

/**
 * Starting from a query object and a target field,
 * the algorithm search in deep into the object the specific field.
 * So, after a check to ensure that queryObj is an object, you retrieve all its keys.
 * For each key a check with target field is performed,
 * thus the object target is stored and returned.
 * Otherwise, you use a recursive strategy to analize the subfields.
 */
function getQueryField(queryObj, targetField) {
  if (typeof queryObj !== 'object' || queryObj === null) return null;

  let objTarget = null;
  const keys = Object.keys(queryObj);
  keys.forEach((key) => {
    if (key === targetField) {
      objTarget = queryObj[key];
    } else {
      const result = getQueryField(queryObj[key], targetField);
      if (result) {
        objTarget = result;
      }
    }
  });

  return objTarget;
}

function normalizeField(field) {
  if (!/[a-z.]+\/[A-Za-z]+/.test(field)) return field;
  const [domain, path] = field.split('/');
  const bits = domain.split('.');
  return uncapitalizeType(
    bits
      .filter((e) => e)
      .map(capitalizeType)
      .join('') + (path || ''),
  );
}

function normalizeAllFieldsRecursive(obj) {
  if (typeof obj !== 'object' || obj === null) return obj;

  const normalizedObj = {};

  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      normalizedObj[normalizeField(key)] = normalizeAllFieldsRecursive(obj[key]);
    }
  }

  return normalizedObj;
}

function getUid() {
  const head = Date.now().toString(36);
  const tail = Math.random().toString(36).substr(2);

  return head + tail;
}

async function fetchJson(url, headers) {
  return new Promise((resolve, reject) => {
    const {
      hostname, port, pathname, protocol,
    } = new URL(url);
    (protocol === 'https:' ? https : http).get({
      headers, hostname, port, path: pathname, timeout: 10000,
    }, (res) => {
      let body = '';

      res.on('data', (chunk) => { body += chunk; });

      res.on('end', () => {
        try {
          resolve({
            ...res,
            body: JSON.parse(body),
          });
        } catch (error) {
          logger.error({
            error, pathname, hostname, port,
          }, 'fetchJson failure');
          reject(error);
        }
      });
    }).on('error', reject);
  });
}

module.exports = {
  capitalizeType,
  uncapitalizeType,
  getBearerToken,
  graphqlQueryRegistry,
  logger,
  getQueryField,
  getUid,
  fetchJson,
  normalizeField,
  normalizeAllFieldsRecursive,
};
