const { withFilter } = require('apollo-server');
const { gql } = require('apollo-server-core');
const { addResolversToSchema } = require('@graphql-tools/schema');
const { extendSchema } = require('graphql/utilities');
const { pubsubAsyncIterator } = require('./pubsub.js');
const { subscriptions } = require('./subscriptions.js');
const { capitalizeType } = require('./utils.js');
const { canWatchResource } = require('./watch.js');

let cacheSubscriptions = {};
const TEN_MINUTES = 10 * 60 * 1000;

function decorateEnum(baseSchema, enumName, values) {
  if (!baseSchema) throw 'Parameter baseSchema cannot be empty!';
  if (!enumName) throw 'Parameter enumName cannot be empty!';
  if (!values) throw 'Parameter values cannot be empty!';

  if (baseSchema._typeMap[enumName] !== undefined)
    throw 'Enum type is already present in the schema!';

  let enumType = `enum ${enumName} {`;
  values.forEach(val => {
    enumType += `
      ${val}`;
  });
  enumType += `
    }`;
  const extension = gql`
    ${enumType}
  `;

  const newSchema = extendSchema(baseSchema, extension);
  return newSchema;
}

function decorateSubscription(baseSchema, targetType, enumType, kubeApiUrl) {
  if (!baseSchema) throw 'Parameter baseSchema cannot be empty!';
  if (!targetType) throw 'Parameter targetType cannot be empty!';
  if (!enumType) throw 'Parameter enumType cannot be empty!';

  if (baseSchema.getQueryType().getFields()[targetType] === undefined)
    throw 'Target type not found into the schema';

  const subscriptionField = `${targetType}Update`;
  const label = targetType;

  const subType =
    baseSchema._typeMap.Subscription === undefined
      ? 'type Subscription'
      : 'extend type Subscription';

  const subscriptionType = capitalizeType(subscriptionField);
  targetType = capitalizeType(targetType);

  const extension = gql`
  type  ${subscriptionType} {
    updateType: ${enumType}
    payload: ${targetType}
  }
    ${subType} {
      ${subscriptionField}(name: String, namespace: String!): ${subscriptionType}
    }
  `;

  const resolvers = {
    Subscription: {
      [subscriptionField]: {
        subscribe: withFilter(
          () => pubsubAsyncIterator(label),
          (payload, variables, info, context) => {
            const resourceApi = subscriptions.filter(sub => {
              return `${sub.type}Update` === context.fieldName;
            })[0];
            return (
              payload.apiObj.metadata.namespace === variables.namespace &&
              (variables.name === undefined ||
                payload.apiObj.metadata.name === variables.name) &&
              checkPermission(
                info.token,
                resourceApi.group,
                resourceApi.resource,
                variables.namespace,
                variables.name,
                kubeApiUrl
              )
            );
          }
        ),
        resolve: async (payload, args, context, info) => {
          return payload;
        },
      },
    },
    [subscriptionType]: {
      updateType: (payload, args, context, info) => {
        return payload.type;
      },
      payload: (payload, args, context, info) => {
        return payload.apiObj;
      },
    },
  };

  const extendedSchema = extendSchema(baseSchema, extension);
  const newSchema = addResolversToSchema(extendedSchema, resolvers);
  newSchema._subscriptionType = newSchema._typeMap.Subscription;
  return newSchema;
}

async function checkPermission(
  token,
  group = '',
  resource,
  namespace,
  name = '',
  kubeApiUrl
) {
  if (!token) throw 'Parameter token cannot be empty!';
  if (!resource) throw 'Parameter resource cannot be empty!';
  if (!namespace) throw 'Parameter namespace cannot be empty!';
  if (!kubeApiUrl) throw 'Parameter kubeApiUrl cannot be empty!';

  const keyCache = `${token}_${group}_${resource}_${namespace}_${name}`;
  const lastSub = cacheSubscriptions[keyCache];
  const canUserWatchResourceCached =
    lastSub &&
    !(
      Date.now() - lastSub > TEN_MINUTES && delete cacheSubscriptions[keyCache]
    );

  if (canUserWatchResourceCached) {
    return true;
  } else {
    const canUserWatchResource = await canWatchResource(
      kubeApiUrl,
      token,
      resource,
      group,
      namespace,
      name
    );

    if (canUserWatchResource) {
      cacheSubscriptions[keyCache] = Date.now();
      return true;
    }
  }
  return false;
}

function clearCache() {
  const currentTimestamp = Date.now();
  Object.keys(cacheSubscriptions).forEach(e => {
    currentTimestamp - cacheSubscriptions[e] > TEN_MINUTES &&
      delete cacheSubscriptions[e];
  });
}

function setupSubscriptions(subscriptions, schema, kubeApiUrl) {
  if (!subscriptions) throw 'Parameter subscriptions cannot be empty!';
  if (!schema) throw 'Parameter schema cannot be empty!';
  if (!kubeApiUrl) throw 'Parameter kubeApiUrl cannot be empty!';

  let newSchema = decorateEnum(schema, 'UpdateType', [
    'ADDED',
    'MODIFIED',
    'DELETED',
  ]);

  subscriptions.forEach(sub => {
    newSchema = decorateSubscription(
      newSchema,
      sub.type,
      'UpdateType',
      kubeApiUrl
    );
  });

  setInterval(() => {
    clearCache();
  }, TEN_MINUTES * 2.2);

  return newSchema;
}

module.exports = {
  decorateSubscription,
  setupSubscriptions,
};
