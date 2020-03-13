"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const informer_1 = require("./informer");
class ListWatch {
    constructor(path, watch, listFn, autoStart = true) {
        this.path = path;
        this.watch = watch;
        this.listFn = listFn;
        this.objects = [];
        this.indexCache = {};
        this.callbackCache = {};
        this.watch = watch;
        this.listFn = listFn;
        this.callbackCache[informer_1.ADD] = [];
        this.callbackCache[informer_1.UPDATE] = [];
        this.callbackCache[informer_1.DELETE] = [];
        if (autoStart) {
            this.doneHandler(null);
        }
    }
    start() {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            yield this.doneHandler(null);
        });
    }
    on(verb, cb) {
        if (this.callbackCache[verb] === undefined) {
            throw new Error(`Unknown verb: ${verb}`);
        }
        this.callbackCache[verb].push(cb);
    }
    off(verb, cb) {
        if (this.callbackCache[verb] === undefined) {
            throw new Error(`Unknown verb: ${verb}`);
        }
        const indexToRemove = this.callbackCache[verb].findIndex((cachedCb) => cachedCb === cb);
        if (indexToRemove === -1) {
            return;
        }
        this.callbackCache[verb].splice(indexToRemove, 1);
    }
    get(name, namespace) {
        return this.objects.find((obj) => {
            return obj.metadata.name === name && (!namespace || obj.metadata.namespace === namespace);
        });
    }
    list(namespace) {
        if (!namespace) {
            return this.objects;
        }
        return this.indexCache[namespace];
    }
    doneHandler(err) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const promise = this.listFn();
            const result = yield promise;
            const list = result.body;
            deleteItems(this.objects, list.items, this.callbackCache[informer_1.DELETE].slice());
            this.addOrUpdateItems(list.items);
            this.watch.watch(this.path, { resourceVersion: list.metadata.resourceVersion }, this.watchHandler.bind(this), this.doneHandler.bind(this));
        });
    }
    addOrUpdateItems(items) {
        items.forEach((obj) => {
            addOrUpdateObject(this.objects, obj, this.callbackCache[informer_1.ADD].slice(), this.callbackCache[informer_1.UPDATE].slice());
            if (obj.metadata.namespace) {
                this.indexObj(obj);
            }
        });
    }
    indexObj(obj) {
        let namespaceList = this.indexCache[obj.metadata.namespace];
        if (!namespaceList) {
            namespaceList = [];
            this.indexCache[obj.metadata.namespace] = namespaceList;
        }
        addOrUpdateObject(namespaceList, obj);
    }
    watchHandler(phase, obj) {
        switch (phase) {
            case 'ADDED':
            case 'MODIFIED':
                addOrUpdateObject(this.objects, obj, this.callbackCache[informer_1.ADD].slice(), this.callbackCache[informer_1.UPDATE].slice());
                if (obj.metadata.namespace) {
                    this.indexObj(obj);
                }
                break;
            case 'DELETED':
                deleteObject(this.objects, obj, this.callbackCache[informer_1.DELETE].slice());
                if (obj.metadata.namespace) {
                    const namespaceList = this.indexCache[obj.metadata.namespace];
                    if (namespaceList) {
                        deleteObject(namespaceList, obj);
                    }
                }
                break;
        }
    }
}
exports.ListWatch = ListWatch;
// external for testing
function deleteItems(oldObjects, newObjects, deleteCallback) {
    return oldObjects.filter((obj) => {
        if (findKubernetesObject(newObjects, obj) === -1) {
            if (deleteCallback) {
                deleteCallback.forEach((fn) => fn(obj));
            }
            return false;
        }
        return true;
    });
}
exports.deleteItems = deleteItems;
// Only public for testing.
function addOrUpdateObject(objects, obj, addCallback, updateCallback) {
    const ix = findKubernetesObject(objects, obj);
    if (ix === -1) {
        objects.push(obj);
        if (addCallback) {
            addCallback.forEach((elt) => elt(obj));
        }
    }
    else {
        objects[ix] = obj;
        if (updateCallback) {
            updateCallback.forEach((elt) => elt(obj));
        }
    }
}
exports.addOrUpdateObject = addOrUpdateObject;
function isSameObject(o1, o2) {
    return o1.metadata.name === o2.metadata.name && o1.metadata.namespace === o2.metadata.namespace;
}
function findKubernetesObject(objects, obj) {
    return objects.findIndex((elt) => {
        return isSameObject(elt, obj);
    });
}
// Public for testing.
function deleteObject(objects, obj, deleteCallback) {
    const ix = findKubernetesObject(objects, obj);
    if (ix !== -1) {
        objects.splice(ix, 1);
        if (deleteCallback) {
            deleteCallback.forEach((elt) => elt(obj));
        }
    }
}
exports.deleteObject = deleteObject;
//# sourceMappingURL=cache.js.map