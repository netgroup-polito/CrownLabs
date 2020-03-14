// CoreV1Api is used to perform all basic operations available here
// https://github.com/scality/kubernetes-client-javascript/blob/browser/src/gen/api/coreV1Api.ts

// CustomObjectsApi is used to perform all operations on CRD available here
// https://github.com/scality/kubernetes-client-javascript/blob/browser/src/gen/api/customObjectsApi.ts
import { Config, CustomObjectsApi, watch } from '@kubernetes/client-node';
import { UserManager } from 'oidc-client';

/*
* UserManager section, handles the oidc authZ
* */

const userManager = new UserManager({
    authority: OIDC_PROVIDER_URL,
    client_id: OIDC_CLIENT_ID,
    redirect_uri: 'http://localhost:8000/callback',
    post_logout_redirect_uri: '/logout',
    response_type: 'id_token',
    scope: 'openid email profile',
    loadUserInfo: false,
});

// TODO : implement signin silently
//userManager.signinSilent();

//Manually Bypassing authN :)
const user = userManager.getUser();
user.token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImVFa0U1Q2FjVjhMRk9VT3NpdzNGRGR1aTAtcmUxbl84OVVoOGhWTTIyLUUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tZndnMjUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImU0ZmJmNTA1LTQxYjItNGM2OS05OTU3LTNhMGYxY2I1ZDZiMyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.FtWSBY7SHTIq7HS9IL-7ePyJkq6br35P5iiTryAZ2T2COoW3AH_b6ONvwe535v-TyPtp0CbAei-dLb0GCJ4hYhZwQOkRs7NK2ffTWzalrRdHyzhvS1oTeSWLfMZEPGV8_14vUVs2zVYa5hdHahfYKO69UgWFcCrNeNcGtcpx8pqvfrGQse109WKadtUygaIBmlNuq1FyVVniGrf1dYtHWT0FTHPzVYCAs_YQwjIoq4cQDMl4U8uEO5pnSS8-JDEawPneXw7fvk9z-7yUgazTNVddOzUWrr9GR2z6WCjiH7b1FZpAGjjoO9dP54_Kke-ksm9vQv9TK-env6dCLCMQLQ";
user.token_type = "Bearer";

/*
* KubernetesJavascript client section
*/

// Loading config from cluster authenticating via token
const kc = new Config(APISERVER_URL, user.token, user.token_type);
// CRD api
const k8sApi_crd = kc.makeApiClient(CustomObjectsApi);

export function getCRD() {
    k8sApi_crd.listClusterCustomObject("template.crown.team.com", "v1", "labtemplates")
        .then((nodesResponse) => {
            return nodesResponse;
        })
        .catch((error) => {
            console.error('Error retrieving nodes', error.body ? error.body : error);
        });
}

export function createCRD(selectedCRD) {
    k8sApi_crd.createNamespacedCustomObject("instance.crown.team.com", "v1", "default", "labinstances", {
                apiVersion: "instance.crown.team.com/v1",
                kind: "LabInstance",
                metadata: {
                    "name": selectedCRD + "-instance",
                },
                spec: {
                    labTemplateName: selectedCRD,
                    studentId: "123456"
                }
            })
                .then(
                    (response) => {
                        return response;
                    },
                    (err) => {
                        console.log('Error!: ' + err);
                    }
                );
}

export function deleteCRD(selectedCRD) {
    k8sApi_crd.deleteNamespacedCustomObject("instance.crown.team.com", "v1", "default", "labinstances", selectedCRD + "-instance", {})
                .then(
                    (response) => {
                        return response;
                    },
                    (err) => {
                        console.log('Error!: ' + err);
                    }
                );
}

// TODO : da fare
export function watchEvents() {
    watch(kc, '/api/v1/watch/events', {},
        (type, object) => {
            const node = document.createTextNode("NEW EVENT:" + type + object.message + "\n");
        },
        (e) => {
            console.log('Stream ended', e);
        });
}
