const { withFilter } = require('apollo-server');
const { gql, ForbiddenError } = require('apollo-server-core');
const { addResolversToSchema } = require('@graphql-tools/schema');
const { extendSchema } = require('graphql/utilities');
const { pubsubAsyncIterator } = require('./pubsub');
const { subscriptions } = require('./subscriptions');
const { canWatchResource } = require('./watch');
const { wrappers } = require('./wrappers');
const {
  capitalizeType,
  getQueryField,
  graphqlLogger,
  uncapitalizeType,
} = require('./utils');

let cacheSubscriptions = {};
const TEN_MINUTES = 10 * 60 * 1000;

/**
 * Function used to add an enum type in the schema
 * @param {*} baseSchema: The schema that should be extended
 * @param {*} enumName: Name of the enum type
 * @param {*} values: Possible values of the enum type
 */
function decorateEnum(baseSchema, enumName, values) {
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');
  if (!enumName) throw new Error('Parameter enumName cannot be empty!');
  if (!values) throw new Error('Parameter values cannot be empty!');

  if (baseSchema._typeMap[enumName] !== undefined)
    throw new Error('Enum type is already present in the schema!');

  // Create the enum type adding its values
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

/**
 * Function used to add a new subscription at the schema
 * @param {*} baseSchema: The schema that should be extended
 * @param {*} targetType: The respective query name of the subscription that should be created
 * @param {*} enumType: Enum type of the watched object containing the state of that
 * @param {*} kubeApiUrl: Url of Kubernetes for checking the permission about obtaining a subscription on a resource
 */
function decorateSubscription(baseSchema, targetType, enumType, kubeApiUrl) {
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');
  if (!targetType) throw new Error('Parameter targetType cannot be empty!');
  if (!enumType) throw new Error('Parameter enumType cannot be empty!');

  if (baseSchema.getQueryType().getFields()[targetType] === undefined)
    throw new Error('Target type not found into the schema');

  const subscriptionField = `${targetType}Update`;
  const label = targetType;
  /*
   * retrieve sub-labels about wrapped fields
   * in other to listen to changes on them
   */
  const sublabels = wrappers
    .map(wtype => {
      if (wtype['parents'].includes(targetType)) return wtype['type'];
    })
    .filter(s => {
      return s;
    });
  /*
   * retrieve the name of the wrapped fields
   */
  const fieldWrapper = wrappers
    .map(wtype => {
      if (wtype['parents'].includes(targetType)) return wtype['fieldWrapper'];
    })
    .filter(s => {
      return s;
    })
    .map(s => {
      return uncapitalizeType(s);
    });
  /*
   * The name of the query is used to retrieve
   * the new data if it has wrapped types
   */
  const subQueryType = targetType;

  /*
   * In case of an existing Subscription type in the schema,
   * you must extend it for adding a new subscription
   */
  const subType =
    baseSchema._typeMap.Subscription === undefined
      ? 'type Subscription'
      : 'extend type Subscription';

  /*
   * Converts query name in the query type
   * e.g. query name: itPolitoCrownlabsV1alpha2Instance
   *      query type: ItPolitoCrownlabsV1alpha2Instance
   */
  const subscriptionType = capitalizeType(subscriptionField);
  targetType = capitalizeType(targetType);

  const extension = gql`
  type ${subscriptionType} {
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
          () => pubsubAsyncIterator(label, ...sublabels),
          async (payload, variables, context, info) => {
            graphqlLogger(`[i] Validate ${info.fieldName} subscription`);

            let subfieldsCheck = false;

            /*
             * Retrieve information about the subscription
             */
            const resourceApiMainType = subscriptions.filter(sub => {
              return `${sub.type}Update` === info.fieldName;
            })[0];

            /*
             * Check if the subscription has some wrapped types.
             * If so, more operations must be performed
             * in other to check whether the published event is related
             * to the subscription and the main type must be resolved again
             * due to the composition of the wrapped query
             */
            if (sublabels.length > 0) {
              graphqlLogger(`[i] Search for ${targetType} main query object`);
              const mainQueryObj = baseSchema.getQueryType().getFields()[
                subQueryType
              ];
              if (!mainQueryObj) throw new Error('Query object not found');

              graphqlLogger(
                `[i] Resolve main query object with variables: ${JSON.stringify(
                  variables
                )}`
              );
              const newApiObj = await mainQueryObj.resolve(
                variables,
                variables,
                context,
                info
              );

              graphqlLogger(`[i] Main query object resolved`);
              graphqlLogger(
                `[i] Checking whether watched object is the main query object`
              );
              subfieldsCheck =
                subfieldsCheck ||
                (payload.apiObj.metadata.namespace === variables.namespace &&
                  (variables.name === undefined ||
                    payload.apiObj.metadata.name === variables.name));

              if (!subfieldsCheck) {
                let targetObjField;

                fieldWrapper.forEach(fw => {
                  targetObjField = getQueryField(newApiObj, fw);
                  if (typeof targetObjField === 'object') {
                    subfieldsCheck =
                      subfieldsCheck ||
                      (targetObjField.namespace ===
                        payload.apiObj.metadata.namespace &&
                        (targetObjField.name === undefined ||
                          targetObjField.name ===
                            payload.apiObj.metadata.name));
                  }
                });
              }
              if (subfieldsCheck) payload.apiObj = newApiObj;
            }

            /*
             * if all checks are passed the event published is about this subscription.
             * So, the new values are sent on the WebSocket to the client
             */
            return (
              subfieldsCheck &&
              payload.apiObj.metadata.namespace === variables.namespace &&
              (variables.name === undefined ||
                payload.apiObj.metadata.name === variables.name) &&
              checkPermission(
                context.token,
                resourceApiMainType.group,
                resourceApiMainType.resource,
                variables.namespace,
                variables.name,
                kubeApiUrl
              )
            );
          }
        ),
        resolve: async (payload, args, context, info) => {
          /*
           * The values obtained from the watcher or resolved
           * in the case of wrapped types are now passed
           * at the son fields of the main type
           */
          graphqlLogger(`[i] Resolve ${info.fieldName} subscription`);
          return payload;
        },
      },
    },
    [subscriptionType]: {
      updateType: (payload, args, context, info) => {
        /*
         * Retrieve from the father the enum tyme
         */
        return payload.type;
      },
      payload: (payload, args, context, info) => {
        /*
         * Retrieve from the father the new values
         */
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
  if (!token) throw new Error('Parameter token cannot be empty!');
  if (!resource) throw new Error('Parameter resource cannot be empty!');
  if (!namespace) throw new Error('Parameter namespace cannot be empty!');
  if (!kubeApiUrl) throw new Error('Parameter kubeApiUrl cannot be empty!');

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

    if (!canUserWatchResource)
      throw new ForbiddenError('Token Error! You cannot watch this resource');
    cacheSubscriptions[keyCache] = Date.now();
    return true;
  }
}

function clearCache() {
  const currentTimestamp = Date.now();
  Object.keys(cacheSubscriptions).forEach(e => {
    currentTimestamp - cacheSubscriptions[e] > TEN_MINUTES &&
      delete cacheSubscriptions[e];
  });
}

function setupSubscriptions(subscriptions, schema, kubeApiUrl) {
  if (!subscriptions)
    throw new Error('Parameter subscriptions cannot be empty!');
  if (!schema) throw new Error('Parameter schema cannot be empty!');
  if (!kubeApiUrl) throw new Error('Parameter kubeApiUrl cannot be empty!');

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
