import {Config, CustomObjectsApi, watch} from '@kubernetes/client-node';

export default class ApiManager {
    constructor(token, type) {
        this.kc = new Config(APISERVER_URL, token, type);
        this.api = this.kc.makeApiClient(CustomObjectsApi);
        this.templateGroup = "template.crown.team.com";
        this.instanceGroup = "instance.crown.team.com";
        this.version = "v1";
        this.templateNamespace = "cloud-computing";
        this.instanceNamespace = 'test-simone';
        this.templatePlural = "labtemplates";
        this.instancePlural = "labinstances";
        this.getCRDtemplate = this.getCRDtemplate.bind(this);
        this.getCRDinstance = this.getCRDinstance.bind(this);
        this.getCRDstatus = this.getCRDstatus.bind(this);
        this.createCRD = this.createCRD.bind(this);
        this.deleteCRD = this.deleteCRD.bind(this);
    }

    getCRDtemplate() {
        return this.api.listNamespacedCustomObject(this.templateGroup, this.version, this.templateNamespace, this.templatePlural);
    }

    getCRDinstance() {
        return this.api.listNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural);
    }

    createCRD(labTemplateName) {
        return this.api.createNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, {
            apiVersion: this.instanceGroup + "/" + this.version,
            kind: "LabInstance",
            metadata: {
                name: labTemplateName,
                namespace: this.instanceNamespace,
            },
            spec: {
                labTemplateName: labTemplateName,
                studentId: "123456"
            }
        });
    }

    deleteCRD(name) {
        return this.api.deleteNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, name, {});
    }

    startWatching(func) {
        /*
        watch(this.kc, '/apis/instance.crown.team.com/v1/namespaces/test-simone/labinstances/watch/events', {},
            function(type, object) {
                let msg = "";
                if (type === 'ADDED') {
                    msg += 'Created Object';
                } else if (type === 'MODIFIED') {
                    msg += 'Modified Object';
                } else if (type === 'DELETED') {
                    msg += 'Deleted Object';
                } else {
                    msg += 'Unknown Event on object';
                }
                func(msg, object);
            },
            function(e) {
                console.log('Stream ended', e);
            }
        );*/
    }

    getCRDstatus(name) {
        return this.api.getNamespacedCustomObjectStatus(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, name);
    }
}