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

function getSubQueryList(queryName) {
  if (queryName.includes('Instance'))
    return 'listCrownlabsPolitoItV1alpha2NamespacedInstance';
  else return `${queryName}List`;
}

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
 * The function starts from the respective query type and generates an `Update` version
 * for the subscription, this type is made up of two fields: `updateType` and `payload`.
 * So, the function first extends the schema with the new Subscription type then starts to resolve it
 * whether the published event passes the check.
 * First things first, the function resolve `subscriptionField` that correspond to the main type,
 * to do so it waits to receive events related to labels and then filtering the right published event.
 *
 * Into the filter:
 * 1. you retrieve information about the subscription used for performing a check
 * on permission of the user about watching this resource.
 * 2. If the `subscription type` has some wrapped type (sublabels.length > 0):
 *   2.1 fieldsCheck = false because you must perform the checks on fields of the object(s)
 *   2.2 You retrieve from the `baseSchema` the respective query used to resolve the object
 *       starting from namespace and optionally the name of the subscription variables.
 *   2.3 After the object was resolved you start to check the respective name and namespace
 *       of the resolved object with the published object for each field and
 *       each item ( in the case of a list of items ) of the resolved object.
 * Finally, you return the value of the published events regarding this subscription or not.
 * The value returned is based on the result of:
 * 1. `fieldsCheck` value: for subscription with wrapped fields.
 * 2. Checks of name and namespace of the published event in the case of
 *    the subscription has not wrapped fields.
 * 3. `checkPermission()` value: that checks if the user can watch this resource.
 * If the returned value is true the payload object is passed to the resolver
 * that use it to return the updated value at the client.
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
          /*
           * Listening of events on label and sub-labels
           */
          () => pubsubAsyncIterator(label, ...sublabels),
          async (payload, variables, context, info) => {
            graphqlLogger(`[i] Validate ${info.fieldName} subscription`);

            /*
             * Some variables used for subscription with wrapped types:
             * @variable {*} fiekdsCheck: used to notify if name and namespace of the published event
             * are equal to the respective name and namespace of the resolved object. Starts === true in the case
             * of the subscription has not wrapped fields.
             * @variable {*} isList: used to notify if the subscription is on a list of items or on a single object.
             * @variable {*} found: used in the case of isList === true to notify that the respective
             * object on the list was found and payload.apiObj = item; so, no other check and overwrite on
             * payload.apiObj must be performed.
             */
            let fieldsCheck = true;
            const isList = variables.name === undefined;
            let found = false;

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
              fieldsCheck = false;
              graphqlLogger(
                `[i] Search for ${targetType} main query object (is list: ${isList})`
              );
              const mainQueryObj = baseSchema.getQueryType().getFields()[
                isList ? getSubQueryList(subQueryType) : subQueryType
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

              let targetObjField;

              fieldWrapper.forEach(fw => {
                if (!fieldsCheck) {
                  /*
                   * If the subscription is 'List' version, you find a match
                   * for each wrapped type and item
                   */
                  if (isList) {
                    for (let item of newApiObj.items) {
                      /*
                       * If found is true means that the item on the list
                       * was found and the payload.apiObj was updated
                       */
                      if (!found) {
                        /*
                         * Checks if the published event is about the main type
                         */
                        graphqlLogger(
                          `[i] Checks if the published event is about the main type of the item: ${item}`
                        );
                        fieldsCheck =
                          fieldsCheck ||
                          (item.metadata.namespace ===
                            payload.apiObj.metadata.namespace &&
                            (item.metadata.name === undefined ||
                              item.metadata.name ===
                                payload.apiObj.metadata.name));

                        if (fieldsCheck) {
                          graphqlLogger(
                            `[i] Item found: ${item} on the list, event publisched is about main type`
                          );
                          payload.apiObj = item;
                          found = true;
                          break;
                        }

                        /*
                         * Checks if the published event is about the wrapped type.
                         * The function getQueryField() retrieve the wrapper type
                         * in the query object for a given wrapped type
                         */
                        graphqlLogger(
                          `[i] Checks if the published event is about the wrapped type: ${fw} of the item: ${item}`
                        );
                        targetObjField = getQueryField(item, fw);
                        if (typeof targetObjField === 'object') {
                          fieldsCheck =
                            fieldsCheck ||
                            (targetObjField.namespace ===
                              payload.apiObj.metadata.namespace &&
                              (targetObjField.name === undefined ||
                                targetObjField.name ===
                                  payload.apiObj.metadata.name));
                        }
                        if (fieldsCheck) {
                          graphqlLogger(
                            `[i] Item found: ${item} on the list, event publisched is about wrapped type: ${fw}`
                          );
                          payload.apiObj = item;
                          found = true;
                          break;
                        }
                      }
                    }
                  } else {
                    graphqlLogger('[i] Check for single item');
                    /*
                     * Checking whether the published event is about the main type
                     */
                    graphqlLogger(
                      `[i] Checks if the published event is about the main type of the item: ${payload.apiObj}`
                    );
                    fieldsCheck =
                      fieldsCheck ||
                      (payload.apiObj.metadata.namespace ===
                        newApiObj.metadata.namespace &&
                        (newApiObj.metadata.name === undefined ||
                          payload.apiObj.metadata.name ===
                            newApiObj.metadata.name));

                    /*
                     * Checks if the published event is about the wrapped type.
                     */
                    graphqlLogger(
                      `[i] Checks if the published event is about the wrapped type: ${fw} of the item: ${payload.apiObj}`
                    );
                    targetObjField = getQueryField(newApiObj, fw);
                    if (typeof targetObjField === 'object') {
                      fieldsCheck =
                        fieldsCheck ||
                        (targetObjField.namespace ===
                          payload.apiObj.metadata.namespace &&
                          (targetObjField.name === undefined ||
                            targetObjField.name ===
                              payload.apiObj.metadata.name));
                    }

                    if (fieldsCheck) {
                      graphqlLogger(`[i] Item found: ${payload.apiObj}`);
                      payload.apiObj = newApiObj;
                    }
                  }
                }
              });
            }

            /*
             * if all checks are passed the event published is about this subscription.
             * So, the new values are sent on the WebSocket to the client
             */
            return (
              fieldsCheck &&
              variables.namespace === payload.apiObj.metadata.namespace &&
              (variables.name === undefined ||
                variables.name === payload.apiObj.metadata.name) &&
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
