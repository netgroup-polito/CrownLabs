import React from 'react';
import Footer from './components/Footer';
import Header from "./components/Header";
import StudentView from "./views/StudentView";
import ProfessorView from "./views/ProfessorView";
import ApiManager from "./services/ApiManager";
import Toastr from 'toastr';

import 'toastr/build/toastr.min.css'

/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserLogic extends React.Component {
    constructor(props) {
        super(props);
        this.connect = this.connect.bind(this);
        this.changeSelectedCRDtemplate = this.changeSelectedCRDtemplate.bind(this);
        this.changeSelectedCRDinstance = this.changeSelectedCRDinstance.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRDinstance = this.stopCRDinstance.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);

        /*Attempting to retrieve the token stored in the sessionStorage by OIDC library, otherwise go back*/
        let retrievedSessionToken = JSON.parse(sessionStorage.getItem('oidc.user:' + OIDC_PROVIDER_URL + ":" + OIDC_CLIENT_ID));
        if (!retrievedSessionToken || !retrievedSessionToken.id_token) {
            Toastr.error("You received a non valid token, please check carefully its fields");
            sessionStorage.clear();
            document.location.reload();
        }
        /*State variable which contains:
        * - all lab templates as a Map: (course_group => Array of available templates for that course)
        * - all lab instances as a Map: (instance_name => URL if running, null otherwise)
        * - current selected CRD template as an object (name, namespace).
        * - current selected CRD instance
        * - all namespaced events as a string
        * */
        let parsedToken = this.parseJWTtoken(retrievedSessionToken.id_token);
        // TODO : check fields
        this.apiManager = new ApiManager(retrievedSessionToken.id_token, retrievedSessionToken.token_type || "Bearer", parsedToken.preferred_username, parsedToken.groups, parsedToken.namespace[0]);
        this.state = {
            templateLabs: new Map(),
            instanceLabs: new Map(),
            selectedTemplate: {name: null, namespace: null},
            selectedInstance: null,
            events: "",
            statusHidden: true,
            /*TODO : add this field to the access token*/
            privileged: parsedToken.privileged
        };
        this.retrieveCRDtemplates();
        this.retrieveCRDinstances()
            .then(() => {
                /*Start watching for namespaced events*/
                this.apiManager.startWatching(this.notifyEvent);
            })
            .catch((error) => {
                this.handleErrors(error);
            });
    }

    /**
     * Function to parse a JWT token
     * @param token the token received by keycloak
     * @returns {any} the decrypted token as a JSON object
     */
    parseJWTtoken(token) {
        let base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
        return JSON.parse(decodeURIComponent(atob(base64).split('').map(function (c) {
            return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
        }).join('')));
    }

    /**
     * Private function to retrieve all CRD templates available
     */
    retrieveCRDtemplates() {
        this.apiManager.getCRDtemplates()
            .then(res => {
                let newMap = this.state.templateLabs;
                res.forEach(x => {
                    x ? newMap.set(x.course, x.labs) : null;
                });
                this.setState({templateLabs: newMap});
            })
            .catch((error) => {
                this.handleErrors(error);
            });
    }

    /**
     * Private function to retrieve all CRD instances running
     */
    retrieveCRDinstances() {
        return this.apiManager.getCRDinstances()
            .then((nodesResponse) => {
                const nodes = nodesResponse.body.items;
                let newMap = this.state.instanceLabs;
                nodes.forEach(x => {
                    if (!newMap.has(x.metadata.name)) {
                        newMap.set(x.metadata.name, {status: 0, url: null});
                    }
                });
                this.setState({instanceLabs: newMap});
            })
            .catch((error) => {
                this.handleErrors(error);
            });
    }

    /**
     * Function to start and create a CRD instance using the actual selected one
     */
    startCRD() {
        if (!this.state.selectedTemplate.name) {
            Toastr.info("Please select a lab before starting it");
            return;
        }
        if (this.state.instanceLabs.has(this.state.selectedTemplate.name)) {
            Toastr.info("The `" + this.state.selectedTemplate.name + '` lab is already running');
            return;
        }
        this.apiManager.createCRDinstance(this.state.selectedTemplate.name, this.state.selectedTemplate.namespace)
            .then(
                (response) => {
                    Toastr.success("Successfully started lab `" + this.state.selectedTemplate.name + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.set(response.body.metadata.name, {status: 0, url: null});
                    this.setState({instanceLabs: newMap});
                }
            )
            .catch((error) => {
                this.handleErrors(error);
            })
            .finally(() => {
                this.changeSelectedCRDtemplate(null, null);
            });
    }

    /**
     * Function to stop and delete the current selected CRD instance
     */
    stopCRDinstance() {
        if (!this.state.selectedInstance) {
            Toastr.info("No lab to stop has been selected");
            return;
        }
        if (!this.state.instanceLabs.has(this.state.selectedInstance)) {
            Toastr.info("The `" + this.state.selectedInstance + '` lab is not running');
            return;
        }
        this.apiManager.deleteCRDinstance(this.state.selectedInstance)
            .then(
                (response) => {
                    Toastr.success("Successfully stopped `" + this.state.selectedInstance + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.delete(this.state.selectedInstance);
                    this.setState({instanceLabs: newMap});
                }
            )
            .catch((error) => {
                this.handleErrors(error);
            })
            .finally(() => {
                this.changeSelectedCRDtemplate(null);
            });
    }

    /**
     * Function to connect to the VM of the actual selected CRD instance
     */
    connect() {
        if (!this.state.selectedInstance) {
            Toastr.info("No lab selected to connect to");
            return
        }
        if (!this.state.instanceLabs.has(this.state.selectedInstance)) {
            Toastr.info("The lab `" + this.state.selectedInstance + "` is not running");
            return;
        }
        switch (this.state.instanceLabs.get(this.state.selectedInstance).status) {
            case 1 :
                window.open(this.state.instanceLabs.get(this.state.selectedInstance).url);
                break;
            case 0:
                Toastr.info("The lab `" + this.state.selectedInstance + "` is still starting");
                break;
            default:
                Toastr.info("An error has occurred with the lab `" + this.state.selectedInstance + "`");
                break;
        }
    }

    /**
     *Function to notify a Kubernetes Event related to your resources
     * @param type the type of the event
     * @param object the object of the event
     */
    notifyEvent(type, object) {
        if (!type) {
            /*Watch session ended, restart it*/
            document.location.reload();
            return;
        }
        if (object && object.status) {
            let msg = "[" + object.metadata.creationTimestamp + "] " + object.metadata.name + "\n|===> Event Type: " + type + ", Status: " + object.status.phase;
            if (object.status.phase.match(/Fail|Not/g)) {
                /*Object creation failed*/
                const newMap = this.state.instanceLabs;
                newMap.set(object.metadata.name, {url: null, status: -1});
                this.setState({instanceLabs: newMap, events: msg + "\n" + this.state.events})
            } else if (object.status.phase.match(/VmiReady/g) && (type === "ADDED" || type === "MODIFIED")) {
                /*Object creation succeeded*/
                const newMap = this.state.instanceLabs;
                newMap.set(object.metadata.name, {url: object.status.url, status: 1});
                this.setState({instanceLabs: newMap, events: msg + "\n" + this.state.events})
            } else {
                /*The object is still creating*/
                this.setState({events: msg + "\n" + this.state.events});
            }
        }
    }

    /**
     * Function to change the user selected CRD template
     * @param name the name/label of the new one
     * @param namespace the namespace in which the template should be retrieved
     */
    changeSelectedCRDtemplate(name, namespace) {
        this.setState({
            selectedTemplate: {name: name, namespace: namespace}
        });
    }

    /**
     * Function to change the user selected CRD instance
     * @param name the name/label of the new one
     */
    changeSelectedCRDinstance(name) {
        this.setState({selectedInstance: name});
    }

    /**
     * Function to handle all errors
     * @param error the error message received
     */
    handleErrors(error) {
        let msg = "";
        switch (error.response._fetchResponse.status) {
            case 401 :
                msg += "Token still valid but expired validity for the Cluster, please login again";
                setInterval(() => {
                    document.location.reload();
                }, 2000);
                break;
            case 403 :
                msg += "It seems you do not have the right permissions to perform this operation";
                break;
            case 409 :
                msg += "The resource is already present";
                break;
            default :
                msg += "An error occurred(" + error.response._fetchResponse.status + "), please login again";
                setInterval(() => {
                    document.location.reload();
                }, 2000);
        }
        Toastr.error(msg);
    }

    /**
     * Function to render this component,
     * It automatically updates every new change in the state variable
     * @returns the component to be drawn
     */
    render() {
        return (
            <div style={{minHeight: '100vh'}}>
                <Header logged={true} logout={this.props.logout}/>
                {this.state.privileged ? <ProfessorView/> :
                    <StudentView templateLabs={this.state.templateLabs} funcTemplate={this.changeSelectedCRDtemplate}
                                 funcInstance={this.changeSelectedCRDinstance}
                                 start={this.startCRD}
                                 instanceLabs={this.state.instanceLabs}
                                 connect={this.connect}
                                 stop={this.stopCRDinstance}
                                 events={this.state.events}
                                 showStatus={() => this.setState({statusHidden: !this.state.statusHidden})}
                                 hidden={this.state.statusHidden}/>}
                <Footer/>
            </div>
        );
    }
}
