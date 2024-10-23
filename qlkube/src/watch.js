const k8s = require('@kubernetes/client-node');
const { publishEvent } = require('./pubsub');
const { graphqlLogger } = require('./utils');
const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const watch = new k8s.Watch(kc);

// resource needs to be PLURAL
async function canWatchResource(
  apiServerUrl,
  token,
  resource,
  group,
  namespace,
  name
) {
  const { fetch } = await import('node-fetch');
  graphqlLogger(`[i] Check resource ${resource} in namespace ${namespace}`);
  return fetch(
    `${apiServerUrl}/apis/authorization.k8s.io/v1/selfsubjectaccessreviews`,
    {
      method: 'POST',
      body: JSON.stringify({
        kind: 'SelfSubjectAccessReview',
        apiVersion: 'authorization.k8s.io/v1',
        spec: {
          resourceAttributes: {
            namespace,
            verb: 'watch',
            group,
            resource,
            name,
          },
        },
      }),
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    }
  )
    .then(res => res.json())
    .then(body => {
      if (body.status) {
        return body.status.allowed === true;
      }
      return false;
    })
    .catch(err => {
      console.error('ERROR WHEN CHECKING IF USER CAN CHECK', err);
      throw new Error(
        'ERROR WHEN CHECKING IF USER CAN CHECK',
        err.message,
        err
      );
    });
}

module.exports = { canWatchResource };
