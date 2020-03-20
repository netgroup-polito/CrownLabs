import {Config, CustomObjectsApi, watch} from '@kubernetes/client-node';

export default class ApiManager {
    constructor(token, type) {
        this.kc = new Config(APISERVER_URL, token, type);
        this.apiCRD = this.kc.makeApiClient(CustomObjectsApi);
        this.templateGroup = "template.crown.team.com";
        this.instanceGroup = "instance.crown.team.com";
        this.version = "v1";
        this.templatePlural = "labtemplates";
        this.instancePlural = "labinstances";
        let parsedToken = this.parseJWTtoken(token);
        this.studentID = parsedToken.preferred_username;
        this.templateNamespace = parsedToken.groups;
        this.instanceNamespace = parsedToken.namespace[0];
    }

    parseJWTtoken(token) {
        let base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(decodeURIComponent(atob(base64).split('').map(function(c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join('')));
    }

    async retrieveSingleCRDtemplate(course) {
        if (course.includes("kubernetes:")) {
            let courseNs = course.replace("kubernetes:", '');
            let ret = await this.apiCRD.listNamespacedCustomObject(this.templateGroup, this.version, courseNs, this.templatePlural)
                .then(nodesResponse => {
                    return nodesResponse.body.items.map(x => {
                        return x.metadata.name;
                    });
                })
                .catch(error => {
                    console.log(error);
                    return null;
                });
            if(ret !== null) {
                return {course: courseNs, labs: ret};
            }
        }
        return null;
    }

    async getCRDtemplates() {
        return await Promise.all(this.templateNamespace.map(async x => await this.retrieveSingleCRDtemplate(x)));
    }

    getCRDinstance() {
        return this.apiCRD.listNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural);
    }

    createCRD(labTemplateName, labTemplateNamespace) {
        return this.apiCRD.createNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, {
            apiVersion: this.instanceGroup + "/" + this.version,
            kind: "LabInstance",
            metadata: {
                name: labTemplateName,
                namespace: this.instanceNamespace,
            },
            spec: {
                labTemplateName: labTemplateName,
                labTemplateNamespace: labTemplateNamespace,
                studentId: this.studentID
            }
        });
    }

    deleteCRD(name) {
        return this.apiCRD.deleteNamespacedCustomObject(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, name, {});
    }

    startWatching(func) {
        watch(this.kc, '/apis/' + this.instanceGroup + '/' + this.version + '/namespaces/' + this.instanceNamespace + '/' + this.instancePlural, {},
            function(type, object) {
                let msg = "[" + object.metadata.creationTimestamp + "] " + type + " " + object.metadata.name;
                func(msg);
            },
            function(e) {
                console.log('Stream ended', e);
            }
        );
    }

    getCRDstatus(name) {
        return this.apiCRD.getNamespacedCustomObjectStatus(this.instanceGroup, this.version, this.instanceNamespace, this.instancePlural, name);
    }
}