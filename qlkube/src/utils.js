const { ForbiddenError } = require('apollo-server-core');

function capitalizeType(name) {
  return name[0].toUpperCase() + name.slice(1);
}

function uncapitalizeType(name) {
  return name[0].toLowerCase() + name.slice(1);
}

function getBearerToken(connectionParams) {
  if (!connectionParams)
    throw new Error('Parameter connectionParams cannot be empty!');
  const auth = connectionParams.authorization || connectionParams.Authorization;
  if (!auth) throw new ForbiddenError('Token Error! Token absent!');
  const token = auth.split(/\s+/)[1];
  if (!token) throw new ForbiddenError('Token Error! Token not valid!');

  return token;
}

const graphqlQueryRegistry = {
  // Fires whenever a GraphQL request is received from a client.
  async requestDidStart(requestContext) {
    if (requestContext.request.operationName !== 'IntrospectionQuery')
      graphqlLogger(
        `[i] Request started (Variables: ${JSON.stringify(
          requestContext.request.variables
        )}, Query: ${requestContext.request.operationName})`
      );
  },
};

function getTimestamp() {
  let today = new Date();
  let time = `${today.getHours()}:${today.getMinutes()}:${today.getSeconds()}:${today.getMilliseconds()}`;
  return time;
}

function graphqlLogger(msg) {
  if (process.env.DEBUG) {
    console.log(`(${getTimestamp()}) ${msg}`);
  }
}

/**
 * Starting from a query object and a target field,
 * the algorithm search in deep into the object the specific field.
 * So, after a check to ensure that queryObj is an object, you retrieve all its keys.
 * For each key a check with target field is performed, thus the object target is stored and returned.
 * Otherwise, you use a recursive strategy to analize the subfields.
 */

function getQueryField(queryObj, targetField) {
  if (typeof queryObj !== 'object' || queryObj === null) return null;

  let objTarget = null;
  const keys = Object.keys(queryObj);
  keys.forEach(key => {
    if (key === targetField) {
      objTarget = queryObj[key];
      return;
    } else {
      const result = getQueryField(queryObj[key], targetField);
      if (result) {
        objTarget = result;
        return;
      }
    }
  });

  return objTarget;
}

function getUid() {
  const head = Date.now().toString(36);
  const tail = Math.random().toString(36).substr(2);

  return head + tail;
}

module.exports = {
  capitalizeType,
  uncapitalizeType,
  getBearerToken,
  graphqlQueryRegistry,
  graphqlLogger,
  getQueryField,
  getUid,
};
