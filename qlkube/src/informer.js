const k8s = require('@kubernetes/client-node');
const { publishEvent } = require('./pubsub');
const { graphqlLogger } = require('./utils');

const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);

function kinformer(sub) {
  const { api, group, version, resource } = sub;

  const resourceApi = `/${api}${
    group ? `/${group}` : ''
  }/${version}/${resource}`;

  const listFn = () => k8sApi.listClusterCustomObject(group, version, resource);

  graphqlLogger(`[i] Make informer for api ${resourceApi}`);
  const informer = k8s.makeInformer(kc, resourceApi, listFn);

  informer.on('add', apiObj => {
    graphqlLogger(`[i] Publish event on ${sub.type} label`);
    publishEvent(sub.type, {
      apiObj,
      type: 'ADDED',
    });
  });
  informer.on('update', apiObj => {
    graphqlLogger(`[i] Publish event on ${sub.type} label`);
    publishEvent(sub.type, {
      apiObj,
      type: 'MODIFIED',
    });
  });
  informer.on('delete', apiObj => {
    graphqlLogger(`[i] Publish event on ${sub.type} label`);
    publishEvent(sub.type, {
      apiObj,
      type: 'DELETED',
    });
  });
  informer.on('error', err => {
    graphqlLogger(
      `[i] Error when watching, restart informer on api ${resourceApi} after 5sec`
    );
    console.error(err);
    setTimeout(() => {
      informer.start();
    }, 5000);
  });

  informer.start();
}

module.exports = { kinformer };
