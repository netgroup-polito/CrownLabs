const { makeExecutableSchema } = require('@graphql-tools/schema');
const k8s = require('@kubernetes/client-node');
const { regexes } = require('./nodesLabelsRegexes');
const { kubeClient } = require('./informer');
const { logger } = require('./utils');

const apiClient = kubeClient.makeApiClient(k8s.CoreV1Api);
const nodesLabels = new Set();
const nodesLabelsPairs = [];

async function updateNodesLabels() {
  logger.info('Updating nodes labels');
  try {
    const nodes = await apiClient.listNode();
    nodesLabels.clear();
    nodesLabelsPairs.length = 0;
    nodes.body.items.forEach((node) => {
      const { labels } = node.metadata;
      if (labels) {
        Object.keys(labels).forEach((label) => {
          regexes.forEach((regex) => {
            const labelPair = `${label}=${labels[label]}`;
            if (regex.test(label)) {
              if (!nodesLabels.has(labelPair)) {
                nodesLabels.add(labelPair);
                nodesLabelsPairs.push({ key: label, value: labels[label] });
              }
            }
          });
        });
      }
    });
    logger.info('Node labels updated');
  } catch (e) {
    logger.error(e.message, 'Node labels error');
  }
}

updateNodesLabels();
setInterval(updateNodesLabels, 60000);

const typeDefs = `
type Query {
    getLabels: [Label!]
}
type Label {
    key: String!
    value: String!
}`;
const resolvers = {
  Query: {
    getLabels: () => nodesLabelsPairs,
  },
};

module.exports = {
  schema: makeExecutableSchema({ typeDefs, resolvers }),
};
