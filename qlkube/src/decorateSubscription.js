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
  graphqlLogger,
  uncapitalizeType,
  getUid,
} = require('./utils');

let cacheSubscriptions = {};
const TEN_MINUTES = 10 * 60 * 1000;

function checkVariables(subVar, payloadVar) {
  for (const key in subVar) if (subVar[key] !== payloadVar[key]) return false;
  return true;
}

function checkMetadata(metadataItemA, metadataItemB) {
  return (
    metadataItemA.namespace === metadataItemB.namespace &&
    metadataItemA.name === metadataItemB.name
  );
}

function splitLabelSelector(labelSelector) {
  let key;
  let values = [];

  /*
   * Split labelSelector into key and values
   * First case: label selector into the form "crownlabs.polito.it/workspace=sid"
   * Second case: label selector into the form "crownlabs.polito.it/workspace in (sid)"
   */
  if (labelSelector.includes('=')) {
    let [sel0, sel1] = labelSelector.split('=').map(s => s.trim());
    key = sel0;
    values.push(sel1);
  }

  if (labelSelector.includes(' in ')) {
    let [sel0, sel1] = labelSelector
      .replaceAll('(', ' ')
      .replaceAll(')', ' ')
      .split(/\s+in\s+/)
      .map(s => s.trim());
    key = sel0;
    values = sel1.split(',').map(v => v.trim());
  }

  return { key, values };
}

function checkLabelSelector(subVar, payloadVar, ruid) {
  let { labels } = payloadVar;
  let { labelSelector } = subVar;
  let result = false;

  graphqlLogger(
    `[i] (${ruid}) Perform check on ${JSON.stringify(
      labels
    )} labels with ${labelSelector} label selector`
  );

  /*
   * No labelSelector, so all resources are watched
   */
  if (!labelSelector) {
    graphqlLogger(
      `[i] (${ruid}) The result of the check is true due to empty labelSelector`
    );
    return true;
  }
  /*
   * The resource has no labels, so no match
   */
  if (!labels) {
    graphqlLogger(
      `[i] (${ruid}) The result of the check is false due to empty labels`
    );
    return false;
  }

  const { key, values } = splitLabelSelector(labelSelector);

  for (const k in labels) {
    if (k.includes(key) && values.includes(labels[k])) result = true;
  }

  graphqlLogger(`[i] (${ruid}) The result of the check is ${result}`);

  return result;
}

function overrideSubQueryList(queryName) {
  if (queryName.includes('Instance'))
    return 'listCrownlabsPolitoItV1alpha2NamespacedInstance';
  else return queryName;
}

function getSublabels(targetType) {
  return wrappers
    .map(({ parents, type }) => parents.includes(targetType) && type)
    .filter(s => s);
}

function getFieldWrapper(targetType) {
  return wrappers
    .map(
      ({ parents, fieldWrapper }) =>
        parents.includes(targetType) && fieldWrapper
    )
    .filter(s => s)
    .map(s => uncapitalizeType(s));
}

function getSubType(baseSchema) {
  return baseSchema._typeMap.Subscription === undefined
    ? 'type Subscription'
    : 'extend type Subscription';
}

function getResourceApiMainType(fieldName) {
  return subscriptions.find(({ type }) => `${type}Update` === fieldName);
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
   * The name of the query is used to retrieve
   * the new data if it has wrapped types
   */
  const subQueryType = targetType;

  /*
   * In case of an existing Subscription type in the schema,
   * you must extend it for adding a new subscription
   */
  const subType = getSubType(baseSchema);
  /*
   * Retrieve information about the resource
   */
  const resourceApiMainType = getResourceApiMainType(subscriptionField);

  /*
   * Converts query name in the query type
   * e.g. query name: itPolitoCrownlabsV1alpha2Instance
   *      query type: ItPolitoCrownlabsV1alpha2Instance
   */
  const subscriptionType = capitalizeType(subscriptionField);
  targetType = capitalizeType(targetType);

  const isTenant = resourceApiMainType.resource === 'tenants';

  const extension = gql`
  type ${subscriptionType} {
    updateType: ${enumType}
    payload: ${targetType}
  }
    ${subType} {
      ${subscriptionField}(
        name: ${isTenant ? 'String!' : 'String'}
        namespace: ${isTenant ? 'String' : 'String!'}
      ): ${subscriptionType}
    }
  `;

  const resolvers = {
    Subscription: {
      [subscriptionField]: {
        subscribe: withFilter(
          /*
           * Listening of events on label
           */
          () => pubsubAsyncIterator(label),
          async (payload, variables, context, info) => {
            const ruid = getUid();
            payload.ruid = ruid;

            graphqlLogger(
              `[i] (${ruid}) Validate ${
                info.fieldName
              } subscription with ${JSON.stringify(variables)} variables`
            );

            if (
              payload.apiObj === undefined ||
              payload.apiObj.metadata === undefined
            ) {
              graphqlLogger(`[e] (${ruid}) Error: Bad payload received`);
              return false;
            }

            if (!checkVariables(variables, payload.apiObj.metadata)) {
              graphqlLogger(
                `[i] (${ruid}) Check on variables about subscription: ${
                  info.fieldName
                } with ${JSON.stringify(
                  variables
                )} variables and payload with ${
                  payload.apiObj.metadata.namespace
                } namespace and ${payload.apiObj.metadata.name} name not passed`
              );
              return false;
            }

            /*
             * Retrieve information about the subscription
             */
            const resourceApiMainType = getResourceApiMainType(info.fieldName);

            const checkPermissionResult = await checkPermission(
              context.token,
              resourceApiMainType.group,
              resourceApiMainType.resource,
              variables.namespace,
              variables.name,
              kubeApiUrl,
              ruid
            );
            if (!checkPermissionResult) {
              graphqlLogger(
                `[i] (${ruid}) Check on permission about subscription: ${
                  info.fieldName
                } with ${JSON.stringify(variables)} variables not passed`
              );
              return false;
            }

            if (payload.type === 'DELETED') {
              graphqlLogger(
                `[i] (${ruid}) ${
                  info.fieldName
                } subscription with ${JSON.stringify(
                  variables
                )} variables have DELETE type`
              );
              return true;
            }

            graphqlLogger(
              `[i] (${ruid}) Search for ${targetType} main query object of ${
                info.fieldName
              } with ${JSON.stringify(variables)} variables`
            );
            const mainQueryObj = baseSchema.getQueryType().getFields()[
              subQueryType
            ];
            if (!mainQueryObj) throw new Error('Query object not found');

            graphqlLogger(
              `[i] (${ruid}) Resolve main query object of ${
                info.fieldName
              } with ${JSON.stringify(variables)} variables`
            );

            const mainQueryObjVar = {
              name: payload.apiObj.metadata.name,
              namespace: payload.apiObj.metadata.namespace,
            };
            const newApiObj = await mainQueryObj.resolve(
              mainQueryObjVar,
              mainQueryObjVar,
              context,
              info
            );

            if (newApiObj) {
              graphqlLogger(
                `[i] (${ruid}) Main query object of ${
                  info.fieldName
                } with ${JSON.stringify(mainQueryObjVar)} variables resolved`
              );
              payload.apiObj = newApiObj;
              return true;
            } else {
              graphqlLogger(
                `[e] (${ruid}) Error during the resolution of the main query object of ${
                  info.fieldName
                } with ${JSON.stringify(mainQueryObjVar)} variables`
              );
              return false;
            }
          }
        ),
        resolve: async (payload, args, context, info) => {
          /*
           * The values obtained from the watcher or resolved
           * in the case of wrapped types are now passed
           * at the son fields of the main type
           */
          graphqlLogger(
            `[i] (${payload.ruid}) Resolve ${info.fieldName} subscription`
          );
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

function decorateLabelsSubscription(
  baseSchema,
  targetType,
  kubeApiUrl,
  queryName
) {
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');
  if (!targetType) throw new Error('Parameter targetType cannot be empty!');

  if (baseSchema.getQueryType().getFields()[targetType] === undefined)
    throw new Error('Target type not found into the schema');

  const subscriptionField = `${targetType}LabelsUpdate`;
  const label = targetType;
  /*
   * retrieve sub-labels about wrapped fields
   * in other to listen to changes on them
   */
  const sublabels = getSublabels(targetType);
  /*
   * retrieve the name of the wrapped fields
   */
  const fieldWrapper = getFieldWrapper(targetType);

  if (sublabels.length > 0 && !fieldWrapper)
    graphqlLogger(
      `[e] Error: mismatch of sublabels and fieldWrapper variables`
    );

  /*
   * In case of an existing Subscription type in the schema,
   * you must extend it for adding a new subscription
   */
  const subType = getSubType(baseSchema);

  /*
   * Converts query name in the query type
   * e.g. query name: itPolitoCrownlabsV1alpha2Instance
   *      query type: ItPolitoCrownlabsV1alpha2Instance
   */
  const subscriptionType = capitalizeType(`${targetType}Update`);

  const extension = gql`
    ${subType} {
        ${subscriptionField}(labelSelector: String): ${subscriptionType}
    }
    `;

  const resolvers = {
    Subscription: {
      [subscriptionField]: {
        subscribe: withFilter(
          /*
           * Listening of events on label
           */
          () => pubsubAsyncIterator(label),
          async (payload, variables, context, info) => {
            const ruid = getUid();
            payload.ruid = ruid;

            graphqlLogger(
              `[i] (${ruid}) Validate ${
                info.fieldName
              } subscription with ${JSON.stringify(variables)} variables`
            );

            if (
              payload.apiObj === undefined ||
              payload.apiObj.metadata === undefined
            ) {
              graphqlLogger(`[e] (${ruid}) Error: Bad payload received`);
              return false;
            }

            if (!checkLabelSelector(variables, payload.apiObj.metadata, ruid)) {
              graphqlLogger(
                `[i] (${ruid}) Check on labels about subscription: ${
                  info.fieldName
                } with ${JSON.stringify(
                  variables
                )} variables and payload with ${JSON.stringify(
                  payload.apiObj.metadata.labels
                )} labels not passed`
              );
              return false;
            }

            /*
             * Retrieve information about the subscription
             */
            const resourceApiMainType = getResourceApiMainType(
              info.fieldName.replace('Labels', '')
            );

            const checkPermissionResult = await checkPermission(
              context.token,
              resourceApiMainType.group,
              resourceApiMainType.resource,
              payload.apiObj.metadata.namespace,
              payload.apiObj.metadata.name,
              kubeApiUrl,
              ruid
            );
            if (!checkPermissionResult) {
              graphqlLogger(
                `[i] (${ruid}) Check on permission about subscription: ${
                  info.fieldName
                } with ${JSON.stringify(variables)} variables not passed`
              );
              return false;
            }

            if (payload.type === 'DELETED') {
              graphqlLogger(
                `[i] (${ruid}) ${
                  info.fieldName
                } subscription with ${JSON.stringify(
                  variables
                )} variables have DELETE type`
              );
              return true;
            }

            graphqlLogger(
              `[i] (${ruid}) Search for ${targetType} main query object of ${
                info.fieldName
              } with ${JSON.stringify(variables)} variables`
            );
            const mainQueryObj = baseSchema.getQueryType().getFields()[
              overrideSubQueryList(queryName)
            ];
            if (!mainQueryObj) throw new Error('Query object not found');

            graphqlLogger(
              `[i] (${ruid}) Resolve main query object of ${
                info.fieldName
              } with ${JSON.stringify(variables)} variables`
            );
            const mainQueryObjVar = {
              namespace: payload.apiObj.metadata.namespace,
              labelSelector: variables.labelSelector,
            };
            const newApiObj = await mainQueryObj.resolve(
              mainQueryObjVar,
              mainQueryObjVar,
              context,
              info
            );

            if (newApiObj) {
              graphqlLogger(
                `[i] (${ruid}) Main query object of ${
                  info.fieldName
                } with ${JSON.stringify(mainQueryObjVar)} variables resolved`
              );
              graphqlLogger(
                `[i] (${ruid}) newApiObj: ${JSON.stringify(newApiObj)} `
              );
              newApiObj.items.forEach(item => {
                if (checkMetadata(item.metadata, payload.apiObj.metadata)) {
                  payload.apiObj = item;
                }
              });
              return true;
            } else {
              graphqlLogger(
                `[e] (${ruid}) Error during the resolution of the main query object of ${
                  info.fieldName
                } with variables: ${JSON.stringify(mainQueryObjVar)}`
              );
              return false;
            }
          }
        ),
        resolve: async (payload, args, context, info) => {
          /*
           * The values obtained from the watcher or resolved
           * in the case of wrapped types are now passed
           * at the son fields of the main type
           */
          graphqlLogger(
            `[i] (${payload.ruid}) Resolve ${info.fieldName} subscription`
          );
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
  namespace = '',
  name = '',
  kubeApiUrl,
  ruid
) {
  if (!token) throw new Error('Parameter token cannot be empty!');
  if (!resource) throw new Error('Parameter resource cannot be empty!');
  if (!kubeApiUrl) throw new Error('Parameter kubeApiUrl cannot be empty!');

  graphqlLogger(
    `[i] (${ruid}) CheckPermission function is starting to generate the key for cache with values (group: ${group}, resource: ${resource}, namespace: ${namespace}, name: ${name})`
  );
  const keyCache = `${token}_${group}_${resource}_${namespace}_${name}`;
  const lastSub = cacheSubscriptions[keyCache];
  const canUserWatchResourceCached =
    lastSub &&
    !(
      Date.now() - lastSub > TEN_MINUTES && delete cacheSubscriptions[keyCache]
    );

  if (canUserWatchResourceCached) {
    graphqlLogger(
      `[i] (${ruid}) User can watch resources. Token already in cache with values (group: ${group}, resource: ${resource}, namespace: ${namespace}, name: ${name})`
    );
    return true;
  } else {
    graphqlLogger(
      `[i] (${ruid}) Token not in cache for values (group: ${group}, resource: ${resource}, namespace: ${namespace}, name: ${name})`
    );
    const canUserWatchResource = await canWatchResource(
      kubeApiUrl,
      token,
      resource,
      group,
      namespace || undefined,
      name
    );

    if (!canUserWatchResource)
      throw new ForbiddenError('Token Error! You cannot watch this resource');
    cacheSubscriptions[keyCache] = Date.now();
    graphqlLogger(
      `[i] (${ruid}) User with values (group: ${group}, resource: ${resource}, namespace: ${namespace}, name: ${name}) pass check. Token added in cache.`
    );
    return true;
  }
}

function clearCache() {
  graphqlLogger('[i] Starts clearCache function');
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
    if (sub.listMapping) {
      newSchema = decorateLabelsSubscription(
        newSchema,
        sub.type,
        kubeApiUrl,
        sub.listMapping
      );
    }
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
