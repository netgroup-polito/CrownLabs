const { PubSub } = require('apollo-server');

const maxListeners = parseInt(process.env.MAX_LISTENERS) || 100;
const pubsub = new PubSub();
pubsub.ee.setMaxListeners(maxListeners);

function publishEvent(label, value) {
  pubsub.publish(label, value);
}

function pubsubAsyncIterator(...labels) {
  return pubsub.asyncIterator(labels);
}

module.exports = { publishEvent, pubsubAsyncIterator };
