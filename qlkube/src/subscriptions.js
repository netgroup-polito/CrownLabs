const subscriptions = [
  { resource: '/api/v1/pods', type: 'ioK8sApiCoreV1Pod' },
  { resource: '/api/v1/nodes', type: 'ioK8sApiCoreV1Node' },
];

module.exports = { subscriptions };
