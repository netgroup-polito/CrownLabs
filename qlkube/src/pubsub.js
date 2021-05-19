const { PubSub } = require('apollo-server');

const pubsub = new PubSub();

function publishEvent(label, value) {
  pubsub.publish(label, value);
}

function pubsubAsyncIterator(label) {
  return pubsub.asyncIterator([label]);
}

module.exports = { publishEvent, pubsubAsyncIterator };
