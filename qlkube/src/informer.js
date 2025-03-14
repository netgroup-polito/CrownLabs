const k8s = require('@kubernetes/client-node');
const { publishEvent } = require('./pubsub');
const { logger } = require('./utils');

const kc = new k8s.KubeConfig();
kc.loadFromDefault();

/**
 * @type k8s.AuthorizationV1Api
 */
const authApi = kc.makeApiClient(k8s.AuthorizationV1Api);

async function canWatchResource(
  token,
  resource,
  group,
  namespace,
) {
  try {
    authApi.setApiKey(k8s.AuthorizationV1ApiApiKeys.BearerToken, token);
    const res = await authApi.createSelfSubjectAccessReview({
      spec: {
        resourceAttributes: {
          namespace,
          verb: 'watch',
          group,
          resource,
        },
      },
    });
    if (res.response.errored) {
      logger.error(res.response, 'Permission assertion error received');
      return false;
    }
    return res.body.status.allowed;
  } catch (e) {
    logger.error(e.message, 'Permission assertion request error');
    // eslint-disable-next-line no-console
    console.error(e);
    return false;
  }
}

/**
 * @type k8s.CustomObjectsApi
 */

const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);

function kinformer(sub) {
  const {
    api, group, version, resource,
  } = sub;

  const resourceApi = `/${api}${group ? `/${group}` : ''
  }/${version}/${resource}`;

  const listFn = () => k8sApi.listClusterCustomObject(group, version, resource);

  logger.info({ resourceApi }, 'Instantiating informer');
  const informer = k8s.makeInformer(kc, resourceApi, listFn);

  Object.entries({
    add: 'ADDED',
    update: 'MODIFIED',
    delete: 'DELETED',
  }).forEach(([evt, type]) => {
    // create an informer for each of the events above
    informer.on(evt, (apiObj) => {
      logger.info({ sub: sub.type, type }, 'Forwarding event');
      publishEvent(sub.type, { apiObj, type });
    });
  });

  informer.on('error', (err) => {
    logger.info({ resourceApi, err }, 'Watching error, restarting in 5 secs');
    setTimeout(() => informer.start(), 5000);
  });

  informer.start();
}

module.exports = { kinformer, canWatchResource };
