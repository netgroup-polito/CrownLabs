const { makeExecutableSchema } = require('@graphql-tools/schema');
const k8s = require('@kubernetes/client-node');
const { regexes } = require('./nodesLabelsRegexes');
const { kubeClient } = require('./informer');
const { logger } = require('./utils');

const apiClient = kubeClient.makeApiClient(k8s.CoreV1Api);
const nodesLabels = new Set();

async function updateNodesLabels() {
  try {
    nodesLabels.clear();
    const nodes = await apiClient.listNode();
    nodes.body.items.forEach((node) => {
      const { labels } = node.metadata;
      if (labels) {
        Object.keys(labels).forEach((label) => {
          regexes.forEach((regex) => {
            if (regex.test(label)) {
              nodesLabels.add(`${label}=${labels[label]}`);
            }
          });
        });
      }
    });
  } catch (e) {
    logger.error(e.message, 'Node labels error');
  }
}

updateNodesLabels();
setInterval(updateNodesLabels, 6000);

const typeDefs = `
type Query {
    getLabels: [Label!]
}
type Label {
    key: String!
    value: String
}`;
const resolvers = {
  Query: {
    getLabels: async () => { // parent, args, context, info
      const labels = [];
      nodesLabels.forEach((label) => {
        const [key, value] = label.split('=');
        labels.push({ key, value });
      });
      return labels;
    },
  },
};

module.exports = {
  schema: makeExecutableSchema({ typeDefs, resolvers }),
};
