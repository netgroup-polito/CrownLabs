const k8s = require('@kubernetes/client-node');
const { publishEvent } = require('./decorateSubscription.js');

const kc = new k8s.KubeConfig();
kc.loadFromDefault();
const watch = new k8s.Watch(kc);

function kwatch(api, label) {
  if (!api) throw 'Parameter api cannot be empty!';
  if (!label) throw 'Parameter label cannot be empty!';

  watch
    .watch(
      api,
      { allowWatchBookmarks: false },
      (type, apiObj, watchObj) => {
        publishEvent(label, {
          apiObj,
          type,
        });
      },
      err => {
        console.log(err);
      }
    )
    .then(req => {});
}

module.exports = { kwatch };
