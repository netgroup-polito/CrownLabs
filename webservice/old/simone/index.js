// CoreV1Api is used to perform all basic operations available here
// https://github.com/scality/kubernetes-client-javascript/blob/browser/src/gen/api/coreV1Api.ts

// CustomObjectsApi is used to perform all operations on CRD available here
// https://github.com/scality/kubernetes-client-javascript/blob/browser/src/gen/api/customObjectsApi.ts
import { Config, CoreV1Api, CustomObjectsApi, watch } from '@kubernetes/client-node';
import { UserManager } from 'oidc-client';

/*
*
* UserManager section, handles the oidc authZ
*
* uncomment the `document.body.append(loginButton);` to see the button and test oidc
* (need to provide a working OIDC_PROVIDER_URL and OIDC_CLIENT_ID)
*
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

if (window.location.pathname === '/callback') {
    userManager.signinRedirectCallback()
        .then((user) => {
            window.location = '/';
        })
        .catch((e) => {
            console.error(e);
            alert('Error oidc callback, see console');
        });
} else if (window.location.pathname === '/logout') {
    userManager.signoutRedirectCallback().then(function() {
        const h1 = document.createElement('h1');
        h1.innerText = 'Logged out';
        document.body.append(h1);

        const link = document.createElement('a');
        link.href = '/';
        link.innerText = 'Return to homepage';
        document.body.append(link);
    });
} else {
    userManager.getUser()
        .then((user) => {
            if (!user) {
                const loginButton = document.createElement('button');
                loginButton.innerText = 'Login';
                loginButton.addEventListener('click', function() {
                    userManager.signinRedirect();
                }, false);
                //document.body.append(loginButton);
            } else {
                render(user);
            }
        });
}

//Manually Bypassing authN :)
const user = userManager.getUser();
user.token = "eyJhbGciOiJSUzI1NiIsImtpZCI6ImVFa0U1Q2FjVjhMRk9VT3NpdzNGRGR1aTAtcmUxbl84OVVoOGhWTTIyLUUifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRlZmF1bHQtdG9rZW4tZndnMjUiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoiZGVmYXVsdCIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6ImU0ZmJmNTA1LTQxYjItNGM2OS05OTU3LTNhMGYxY2I1ZDZiMyIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OmRlZmF1bHQifQ.FtWSBY7SHTIq7HS9IL-7ePyJkq6br35P5iiTryAZ2T2COoW3AH_b6ONvwe535v-TyPtp0CbAei-dLb0GCJ4hYhZwQOkRs7NK2ffTWzalrRdHyzhvS1oTeSWLfMZEPGV8_14vUVs2zVYa5hdHahfYKO69UgWFcCrNeNcGtcpx8pqvfrGQse109WKadtUygaIBmlNuq1FyVVniGrf1dYtHWT0FTHPzVYCAs_YQwjIoq4cQDMl4U8uEO5pnSS8-JDEawPneXw7fvk9z-7yUgazTNVddOzUWrr9GR2z6WCjiH7b1FZpAGjjoO9dP54_Kke-ksm9vQv9TK-env6dCLCMQLQ";
user.token_type = "Bearer";
render(user);

/*
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
*/


/*
*
* KubernetesJavascript client section
*
*
* */

export function render(user){
    // Variable used to track which CRD the user has selected
    let selectedCRD = "";

    // Creating Table to render cluster events
    const textarea_label = document.createElement('h2');
    const textarea = document.createElement('textarea', "rows=\"4\" cols=\"50\"");
    textarea.innerText = "cluster events will be written here";
    textarea_label.innerText = "Cluster Events";
    document.body.append(textarea_label);
    document.body.append(textarea);

    //Creating Status field to display current used resource status
    const status = document.createElement('h2');
    const status_content = document.createElement('span');
    status_content.innerText = "No running CRD";
    status.innerText = 'Status';
    document.body.append(status);
    document.body.append(status_content);

    const course_selected = document.createElement('h2');
    const course_selected_content = document.createElement('span');
    course_selected_content.innerText = "No course selected";
    course_selected.innerText = "Selected Course";
    document.body.append(course_selected);
    document.body.append(course_selected_content);

    // Creating available courses section
    const courses = document.createElement('h2');
    courses.innerText = 'Available courses';
    document.body.append(courses);

    // Creating start btn
    const start = document.createElement('BUTTON');
    start.innerText = "Start";
    start.onclick = () => {
        if(selectedCRD !== "") {
            //Create the CRD object
            // TODO : fill studentID and the name correctly according to the assigned one
            k8sApi_crd.createNamespacedCustomObject("instance.crown.team.com", "v1", "default", "labinstances", {
                apiVersion: "instance.crown.team.com/v1",
                kind: "LabInstance",
                metadata: {
                    "name": "web-applications-course-instance",
                },
                spec: {
                    labTemplateName: selectedCRD,
                    studentId: "123456"
                }
            })
                .then(
                    (response) => {
                        status_content.innerText = "RUNNING " + selectedCRD;
                    },
                    (err) => {
                        alert('Error running CRD, check console');
                        console.log('Error!: ' + err);
                    }
                );
        } else {
            alert("No CRD selected, please choose one");
        }

    };
    document.body.append(start);

    // Creating stop btn
    const stop = document.createElement('BUTTON');
    stop.innerText = "Stop";
    stop.onclick = () => {
        // Dumb check
        if(status_content.textContent.includes("RUNNING")) {
            k8sApi_crd.deleteNamespacedCustomObject("instance.crown.team.com", "v1", "default", "labinstances", "web-applications-course-instance", {})
                .then(
                    () => {
                        status_content.innerText = "No Running CRDs";
                    },
                    (err) => {
                        alert('Error deleting CRD, check console');
                        console.log('Error!: ' + err);
                    }
                );
        } else {
            alert("There's nothing to stop");
        }
    };
    document.body.append(stop);

    // Creating connect btn
    // TODO : add window.open(IP:PORT), where the service is reachable
    const connect = document.createElement('BUTTON');
    connect.innerText = "Connect";
    connect.onclick = () => {
        if(status_content.textContent.includes("RUNNING")) {
            window.open("https://google.it");
        } else {
            alert('No running CRD to connect to');
        }
    };
    document.body.append(connect);

    // Loading config from cluster authenticating via token
    const kc = new Config(APISERVER_URL, user.token, user.token_type);
    // Base api, not used right now since we need to work with CRDs
    //const k8sApi_base = kc.makeApiClient(CoreV1Api);
    // CRD api
    const k8sApi_crd = kc.makeApiClient(CustomObjectsApi);

    // Install watcher for all Cluster Events
    watch(kc, '/api/v1/watch/events', {},
        (type, object) => {
            const node = document.createTextNode("NEW EVENT:" + type + object.message + "\n");
            textarea.appendChild(node);
        },
        (e) => {
            console.log('Stream ended', e);
        });

    // Retrieving CRD labtemplates
    // TODO : once updated courses, repeat this operation for each course
    k8sApi_crd.listClusterCustomObject("template.crown.team.com", "v1", "labtemplates")
        .then((nodesResponse) => {
            const nodes = nodesResponse.body.items;
            const ul = document.createElement('ul');
            for (const idx in nodes) {
                const li = document.createElement('li');
                const btn = document.createElement("BUTTON");
                btn.innerText = nodes[idx].metadata.name;
                btn.onclick = () => {
                    selectedCRD = nodes[idx].metadata.name;
                    course_selected_content.innerText = nodes[idx].metadata.name;
                };
                li.appendChild(btn);
                ul.append(li);
            }
            document.body.append(ul);
        })
        .catch((error) => {
            alert('Error retrieving nodes, check console');
            console.error('Error retrieving nodes', error.body ? error.body : error);
        });

}

/*k8sApi.listCustomResourceDefinition()
    .then(function(nodesResponse) {
        const header = document.createElement('h2');
        header.innerText = 'Custom Objects';
        document.body.append(header);

        const nodes = nodesResponse.body.items;
        const ul = document.createElement('ul');
        for(const idx in nodes) {
            const li = document.createElement('li');
            li.innerText = nodes[idx].metadata.name;
            ul.append(li);
        }
        document.body.append(ul);
    })
    .catch(function(error) {
        console.error('Error retrieving nodes', error.body ? error.body : error);
    });
    */

/*
k8sApi.listNode()
    .then(function(nodesResponse) {
        const header = document.createElement('h2');
        header.innerText = 'Cluster Nodes';
        document.body.append(header);

        const nodes = nodesResponse.body.items;
        const ul = document.createElement('ul');
        for(const idx in nodes) {
            const li = document.createElement('li');
            li.innerText = nodes[idx].metadata.name;
            ul.append(li);
        }
        document.body.append(ul);
    })
    .catch(function(error) {
        console.error('Error retrieving nodes', error.body ? error.body : error);
    });

k8sApi.listPodForAllNamespaces()
    .then(function(nodesResponse) {
        const header = document.createElement('h2');
        header.innerText = 'Cluster Pods';
        document.body.append(header);

        const nodes = nodesResponse.body.items;
        const ul = document.createElement('ul');
        for(const idx in nodes) {
            const li = document.createElement('li');
            li.innerText = nodes[idx].metadata.name;
            ul.append(li);
        }
        document.body.append(ul);
    })
    .catch(function(error) {
        console.error('Error retrieving nodes', error.body ? error.body : error);
    });
*/
