const k8s = require('@kubernetes/client-node');
const fetch = require('node-fetch');
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

function kwatch(api, label) {
  if (!api) throw new Error('Parameter api cannot be empty!');
  if (!label) throw new Error('Parameter label cannot be empty!');
  watch
    .watch(
      api,
      { allowWatchBookmarks: false },
      (type, apiObj, watchObj) => {
        graphqlLogger(`[i] Publish event on ${label} label`);
        publishEvent(label, {
          apiObj,
          type,
        });
      },
      err => {
        if (err?.message === 'Not Found') {
          console.error(`No items for ${api} found to watch, restarting in 2s`);
        } else {
          console.error('Error when watching', err);
        }
        setTimeout(() => kwatch(api, label), 2000);
      }
    )
    .then(req => {});
}

module.exports = { kwatch, canWatchResource };
