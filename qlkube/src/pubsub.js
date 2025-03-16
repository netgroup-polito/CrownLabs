const { PubSub } = require('graphql-subscriptions');

const maxListeners = parseInt(process.env.MAX_LISTENERS) || 100;
const pubsub = new PubSub();
pubsub.ee.setMaxListeners(maxListeners);

function publishEvent(label, value) {
  pubsub.publish(label, value);
}

function pubsubAsyncIterator(...labels) {
  return pubsub.asyncIterableIterator(labels);
}

module.exports = { publishEvent, pubsubAsyncIterator };
