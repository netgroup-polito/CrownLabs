const { gql } = require('apollo-server-core');
const { extendSchema } = require('graphql/utilities');
const { addResolversToSchema } = require('@graphql-tools/schema');
const { PubSub, withFilter } = require('apollo-server');
const { capitalizeType } = require('./utils.js');

const pubsub = new PubSub();

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

function decorateSubscription(baseSchema, targetType, enumType) {
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
          () => pubsub.asyncIterator([label]),
          (payload, variables, info, context) => {
            return (
              payload.apiObj.metadata.namespace === variables.namespace &&
              (variables.name === undefined ||
                payload.apiObj.metadata.name === variables.name)
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

function setupSubscriptions(subscriptions, schema) {
  if (!subscriptions) throw 'Parameter subscriptions cannot be empty!';
  if (!schema) throw 'Parameter schema cannot be empty!';

  let newSchema = decorateEnum(schema, 'UpdateType', [
    'ADDED',
    'MODIFIED',
    'DELETED',
  ]);

  subscriptions.forEach(e => {
    newSchema = decorateSubscription(newSchema, e.type, 'UpdateType');
  });

  return newSchema;
}

function publishEvent(label, value) {
  pubsub.publish(label, value);
}

module.exports = {
  decorateSubscription,
  setupSubscriptions,
  publishEvent,
};
