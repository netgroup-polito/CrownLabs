const { ForbiddenError } = require('apollo-server-core');

function capitalizeType(name) {
  return name[0].toUpperCase() + name.slice(1);
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
        `[i]Request started:\n[+] Variables: ${JSON.stringify(
          requestContext.request.variables
        )}\n[+] Query: ${requestContext.request.operationName}`
      );
  },
};

function graphqlLogger(msg) {
  if (process.env.DEBUG) {
    console.log(msg);
  }
}

module.exports = {
  capitalizeType,
  getBearerToken,
  graphqlQueryRegistry,
  graphqlLogger,
};
