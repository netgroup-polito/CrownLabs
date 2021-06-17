// This file serves development/testing purposes only.
// When qlkube is deployed with Helm, it is overwritten by an equivalent
// one automatically generated from the configuration therein specified.

const subscriptions = [
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'instances',
    type: 'itPolitoCrownlabsV1alpha2Instance',
  },
  {
    api: 'apis',
    group: 'crownlabs.polito.it',
    version: 'v1alpha2',
    resource: 'templates',
    type: 'itPolitoCrownlabsV1alpha2Template',
  },
];

module.exports = { subscriptions };
