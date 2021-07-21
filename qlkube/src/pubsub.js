const { PubSub } = require('apollo-server');

const pubsub = new PubSub();

function publishEvent(label, value) {
  pubsub.publish(label, value);
}

function pubsubAsyncIterator(...labels) {
  return pubsub.asyncIterator(labels);
}

module.exports = { publishEvent, pubsubAsyncIterator };
