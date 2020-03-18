import {Config, CustomObjectsApi} from '@kubernetes/client-node';

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
        this.getCRD = this.getCRD.bind(this);
        this.createCRD = this.createCRD.bind(this);
        this.deleteCRD = this.deleteCRD.bind(this);
    }

    getCRD() {
        return this.api.listNamespacedCustomObject(this.templateGroup, this.version, this.templateNamespace, this.templatePlural);
    }
    createCRD(labTemplateName) {
        return this.api.createNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, {
            apiVersion: this.instanceGroup + "/" + this.version,
            kind: "LabInstance",
            metadata: {
                name: labTemplateName + "-instance",
                namespace: this.instanceNamespace,
            },
            spec: {
                labTemplateName: labTemplateName,
                studentId: "123456"
            }
        });
    }
    deleteCRD(labTemplateName) {
        return this.api.deleteNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, labTemplateName + "-instance", {});
    }
}