const { ForbiddenError } = require('@apollo/server');
const { withFilter } = require('graphql-subscriptions');
const { gql } = require('graphql-tag');
const { addResolversToSchema } = require('@graphql-tools/schema');
const { extendSchema } = require('graphql/utilities');
const { pubsubAsyncIterator } = require('./pubsub');
const { subscriptions } = require('./subscriptions');
const { canWatchResource } = require('./informer');
const { wrappers } = require('./wrappers');
const {
  capitalizeType,
  uncapitalizeType,
  getUid,
  logger,
  normalizeAllFieldsRecursive,
} = require('./utils');

const cacheSubscriptions = {};
const TEN_MINUTES = 10 * 60 * 1000;

async function runTypeResolver(ruid, schema, targetType, vars, ctx, info) {
  logger.info({
    ruid, targetType, field: info.fieldName, vars,
  }, 'Searching main query object');
  const mainQueryObj = schema.getQueryType().getFields()[targetType];
  if (!mainQueryObj) throw new Error('Query object not found');

  return mainQueryObj.resolve(null, vars, ctx, info);
}

async function checkPermission(
  ruid,
  token,
  group,
  resource,
  namespace,
  name,
) {
  if (!token) throw new Error('Parameter token cannot be empty!');
  if (!resource) throw new Error('Parameter resource cannot be empty!');

  const keyCache = `t:${token.split('.')[2]}_g:${group}_r:${resource}_ns:${namespace}_n:${name}`;
  const lastSub = cacheSubscriptions[keyCache];
  const canUserWatchResourceCached = lastSub && !(
    (Date.now() - lastSub > TEN_MINUTES) && delete cacheSubscriptions[keyCache]
  );

  if (canUserWatchResourceCached) {
    logger.info({ ruid, keyCache }, 'User can watch resources. Token cached');
    return true;
  }

  logger.info({ ruid, keyCache }, 'Token not in cache - checking validity on k8s');
  const canUserWatchResource = await canWatchResource(token, resource, group, namespace);

  if (!canUserWatchResource) {
    throw new ForbiddenError('Token Error! You cannot watch this resource');
  }

  cacheSubscriptions[keyCache] = Date.now();
  logger.info({ ruid, keyCache }, 'Token check passed, caching');
  return true;
}

function checkVariables(subVar, payloadVar) {
  for (const key in subVar) if (subVar[key] !== payloadVar[key]) return false;
  return true;
}

function checkMetadata(metadataItemA, metadataItemB) {
  return (
    metadataItemA.namespace === metadataItemB.namespace
    && metadataItemA.name === metadataItemB.name
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
    const [sel0, sel1] = labelSelector.split('=').map((s) => s.trim());
    key = sel0;
    values.push(sel1);
  }

  if (labelSelector.includes(' in ')) {
    const [sel0, sel1] = labelSelector
      .replaceAll('(', ' ')
      .replaceAll(')', ' ')
      .split(/\s+in\s+/)
      .map((s) => s.trim());
    key = sel0;
    values = sel1.split(',').map((v) => v.trim());
  }

  return { key, values };
}

function checkLabelSelector(subVar, payloadVar, ruid) {
  const { labels } = payloadVar;
  const { labelSelector } = subVar;

  logger.info({ ruid, labels, labelSelector }, 'Checking label selector');

  // No labelSelector, so all resources are watched
  if (!labelSelector) {
    logger.info({ ruid, labels }, 'No label selector, check passed');
    return true;
  }
  // The resource has no labels, so no match
  if (!labels) {
    logger.info({ ruid, labelSelector }, 'No labels, check failed');
    return false;
  }

  let result = false;
  const { key, values } = splitLabelSelector(labelSelector);

  for (const k in labels) {
    if (k === key && values.includes(labels[k])) {
      result = true;
    }
  }

  logger.info({ ruid, result }, 'Label selector check completed');

  return result;
}

function overrideSubQueryList(queryName) {
  if (queryName.includes('Instance')) return 'listCrownlabsPolitoItV1alpha2NamespacedInstance';
  return queryName;
}

function getSublabels(targetType) {
  return wrappers
    .map(({ parents, type }) => parents.includes(targetType) && type)
    .filter((s) => s);
}

function getFieldWrapper(targetType) {
  return wrappers
    .map(
      ({ parents, fieldWrapper }) => parents.includes(targetType) && fieldWrapper,
    )
    .filter((s) => s)
    .map((s) => uncapitalizeType(s));
}

/**
  * In case of an existing Subscription type in the schema,
  * you must extend it for adding a new subscription
  */
function getSubType(baseSchema) {
  return !baseSchema.getType('Subscription')
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

  if (baseSchema.getType(enumName)) {
    throw new Error('Enum type is already present in the schema!');
  }

  // Create the enum type adding its values
  let enumType = `enum ${enumName} {`;
  values.forEach((val) => {
    enumType += `
      ${val}`;
  });
  enumType += `
    }`;
  const extension = gql`
    ${enumType}
  `;

  return extendSchema(baseSchema, extension);
}

/**
 * Function used to add a new subscription at the schema
 * @param {*} baseSchema: The schema that should be extended
 * @param {*} targetType: The respective query name of the subscription that should be created
 * @param {*} enumType: Enum type of the watched object containing the state of that
 * Permission checking is done through the integrated k8s client.
 * The function starts from the respective query type and generates an `Update` version
 * for the subscription, this type is made up of two fields: `updateType` and `payload`.
 * So, the function first extends the schema with the new Subscription type
 * then starts to resolve it whether the published event passes the check.
 * First things first, the function resolve `subscriptionField` that correspond to the main type,
 * to do so it waits to receive events related to labels and then
 * filtering the right published event.
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
function decorateSubscription(baseSchema, targetType, enumType) {
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');
  if (!targetType) throw new Error('Parameter targetType cannot be empty!');
  if (!enumType) throw new Error('Parameter enumType cannot be empty!');

  if (baseSchema.getQueryType().getFields()[targetType] === undefined) throw new Error('Target type not found into the schema');

  const subscriptionField = `${targetType}Update`;
  const label = targetType;

  /*
   * The name of the query is used to retrieve
   * the new data if it has wrapped types
   */
  const subQueryType = targetType;

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
            try {
              const ruid = getUid();
              const { fieldName } = info;
              payload.ruid = ruid;

              logger.info({ ruid, fieldName, variables }, 'Validating subscription');

              if (!(payload.apiObj?.metadata)) {
                logger.error({ ruid, apiObj: payload.apiObj }, 'Bad payload');
                return false;
              }

              if (!checkVariables(variables, payload.apiObj.metadata)) {
                const { name, labels, namespace } = payload.apiObj.metadata;
                logger.info({
                  ruid, fieldName, variables, meta: { name, labels, namespace },
                }, 'Variables check NOT PASSED');
                return false;
              }

              /*
              * Retrieve information about the subscription
              */
              const resApiMainType = getResourceApiMainType(info.fieldName);

              const checkPermissionResult = await checkPermission(
                ruid,
                context.token,
                resApiMainType.group,
                resApiMainType.resource,
                variables.namespace,
                variables.name,
              );
              if (!checkPermissionResult) {
                logger.info({ ruid, fieldName }, 'Permissions check NOT PASSED');
                return false;
              }

              if (payload.type === 'DELETED') {
                logger.info({ ruid, fieldName, type: payload.type }, 'Subscription valid');
                return true;
              }

              const newApiObj = await runTypeResolver(ruid, baseSchema, subQueryType, {
                name: payload.apiObj.metadata.name,
                namespace: payload.apiObj.metadata.namespace,
              }, context, info);

              if (newApiObj) {
                logger.info({ ruid, fieldName }, 'Main query object vars resolved');
                payload.apiObj = newApiObj;
                return true;
              }
              logger.error({ ruid, fieldName }, 'Error during main query obj resolution');
              return false;
            } catch (e) {
              logger.error(null, 'Exception catched during subscription filtering');
              // eslint-disable-next-line no-console
              console.error(e);
              return false;
            }
          },
        ),
        /**
         * The values obtained from the watcher or resolved
         * in the case of wrapped types are now passed
         * at the son fields of the main type
         */
        resolve: async (payload, args, context, info) => {
          logger.info({ ruid: payload.ruid, fieldName: info.fieldName }, 'Resolving subscription');
          if (payload.type === 'DELETED') {
            return normalizeAllFieldsRecursive(payload);
          }
          return payload;
        },
      },
    },
    [subscriptionType]: {
      /** Retrieve from the father the enum type */
      updateType: (payload, _args, _context, _info) => payload.type,
      /** Retrieve from the father the new values */
      payload: (payload, _args, _context, _info) => payload.apiObj
      ,
    },
  };

  const extendedSchema = extendSchema(baseSchema, extension);
  const newSchema = addResolversToSchema({
    schema: extendedSchema,
    resolvers,
  });
  // eslint-disable-next-line no-underscore-dangle
  newSchema._subscriptionType = newSchema._typeMap.Subscription;
  return newSchema;
}

function decorateLabelsSubscription(
  baseSchema,
  targetType,
  queryName,
) {
  if (!baseSchema) throw new Error('Parameter baseSchema cannot be empty!');
  if (!targetType) throw new Error('Parameter targetType cannot be empty!');

  if (baseSchema.getQueryType().getFields()[targetType] === undefined) throw new Error('Target type not found into the schema');

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

  if (sublabels.length > 0 && !fieldWrapper) {
    logger.error({ sublabels, fieldWrapper }, 'mismatch of sublabels and fieldWrapper variables');
    throw new Error('ESUBSETUP');
  }

  const subType = getSubType(baseSchema);

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
            try {
              const ruid = getUid();
              const { fieldName } = info;
              payload.ruid = ruid;

              logger.info({ ruid, fieldName, variables }, 'Validating labeled subscription');

              if (
                payload.apiObj === undefined
                || payload.apiObj.metadata === undefined
              ) {
                logger.error({ ruid, apiObj: payload.apiObj }, 'Bad payload');
                return false;
              }

              if (!checkLabelSelector(variables, payload.apiObj.metadata, ruid)) {
                logger.error({
                  ruid, fieldName, variables, labels: payload.apiObj.metadata.labels,
                }, 'Labels check not passed');
                return false;
              }

              /*
               * Retrieve information about the subscription
               */
              const resourceApiMainType = getResourceApiMainType(
                info.fieldName.replace('Labels', ''),
              );

              const checkPermissionResult = await checkPermission(
                ruid,
                context.token,
                resourceApiMainType.group,
                resourceApiMainType.resource,
                payload.apiObj.metadata.namespace,
                payload.apiObj.metadata.name,
              );
              if (!checkPermissionResult) {
                logger.info({ ruid, fieldName }, 'Permissions check NOT PASSED');
                return false;
              }

              if (payload.type === 'DELETED') {
                logger.info({ ruid, fieldName, type: payload.type }, 'Subscription valid');
                return true;
              }

              const newApiObj = await runTypeResolver(
                ruid,
                baseSchema,
                overrideSubQueryList(queryName),
                {
                  name: payload.apiObj.metadata.name,
                  namespace: payload.apiObj.metadata.namespace,
                },
                context,
                info,
              );

              if (newApiObj) {
                logger.info({ ruid, fieldName }, 'Main query object vars resolved');
                newApiObj.items.forEach((item) => {
                  if (checkMetadata(item.metadata, payload.apiObj.metadata)) {
                    payload.apiObj = item;
                  }
                });
                return true;
              }
              logger.error({ ruid, fieldName }, 'Error during main query obj resolution');
              return false;
            } catch (e) {
              logger.error(null, 'Exception catched during subscription filtering');
              // eslint-disable-next-line no-console
              console.error(e);
              return false;
            }
          },
        ),
        /**
         * The values obtained from the watcher or resolved
         * in the case of wrapped types are now passed
         * at the son fields of the main type
         */
        resolve: async (payload, args, context, info) => {
          logger.info({ ruid: payload.ruid, fieldName: info.fieldName }, 'Resolving subscription');
          if (payload.type === 'DELETED') {
            return normalizeAllFieldsRecursive(payload);
          }
          return payload;
        },
      },
    },
    [subscriptionType]: {
      /** Retrieve from the father the enum type */
      updateType: (payload, _args, _context, _info) => payload.type,
      /** Retrieve from the father the new values */
      payload: (payload, _args, _context, _info) => payload.apiObj
      ,
    },
  };

  return addResolversToSchema({
    schema: extendSchema(baseSchema, extension),
    resolvers,
  });
}

function clearCache() {
  logger.info('Starting cache cleanup');
  const currentTimestamp = Date.now();
  Object.keys(cacheSubscriptions).forEach(
    (e) => (currentTimestamp - cacheSubscriptions[e] > TEN_MINUTES) && delete cacheSubscriptions[e],
  );
  logger.info({ durationMs: Date.now() - currentTimestamp }, 'Cache cleanup completed');
}

function setupSubscriptions(subs, schema) {
  if (!subs) throw new Error('Parameter subscriptions cannot be empty!');
  if (!schema) throw new Error('Parameter schema cannot be empty!');

  let newSchema = decorateEnum(schema, 'UpdateType', [
    'ADDED',
    'MODIFIED',
    'DELETED',
  ]);

  subs.forEach((sub) => {
    newSchema = decorateSubscription(
      newSchema,
      sub.type,
      'UpdateType',
    );
    if (sub.listMapping) {
      newSchema = decorateLabelsSubscription(
        newSchema,
        sub.type,
        sub.listMapping,
      );
    }
  });

  setInterval(clearCache, TEN_MINUTES * 2.2);

  return newSchema;
}

module.exports = {
  decorateSubscription,
  setupSubscriptions,
};
