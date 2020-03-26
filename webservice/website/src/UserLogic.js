import React from 'react';
import Footer from './components/Footer';
import Header from "./components/Header";
import StudentView from "./views/StudentView";
import ProfessorView from "./views/ProfessorView";
import ApiManager from "./services/ApiManager";
import Toastr from 'toastr';

import 'toastr/build/toastr.min.css'
import {Container} from "react-bootstrap";

/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserLogic extends React.Component {
    /*The State variable contains:
    * - all lab templates as a Map: (course_group => Array of available templates for that course)
    * - all lab instances as a Map: (instance_name => URL if running, null otherwise)
    * - current selected CRD template as an object (name, namespace).
    * - current selected CRD instance
    * - all namespaced events as a string
    * - boolean variable whether to show the status info area
    * - adminHidden whether to render or not the admin page (changed by the button in the StudentView IF adminGroups is not false)
    * - isAdminSomewhere which is false if the user is not admin in any course, otherwise true
    * */

    constructor(props) {
        super(props);
        this.connect = this.connect.bind(this);
        this.changeSelectedCRDtemplate = this.changeSelectedCRDtemplate.bind(this);
        this.changeSelectedCRDinstance = this.changeSelectedCRDinstance.bind(this);
        this.startCRDinstance = this.startCRDinstance.bind(this);
        this.stopCRDinstance = this.stopCRDinstance.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);


        /*Retrieve, decode and check the token received. If errors immediately logout after a msg prompt*/
        let retrievedSessionToken = JSON.parse(sessionStorage.getItem('oidc.user:' + OIDC_PROVIDER_URL + ":" + OIDC_CLIENT_ID));
        let parsedToken = this.parseJWTtoken(retrievedSessionToken.id_token);
        if (!this.checkToken(parsedToken, retrievedSessionToken)) {
            this.logoutInterval();
        }
        /*Differentiate the two different kind of group: where the user is admin (professor or PhD) and the one where he is just a student*/
        let adminGroups = parsedToken.groups.filter(x => x.match(/kubernetes:\S+admin/g)).map(x => x.replace('kubernetes:', '').replace('-admin', ''));
        let userGroups = parsedToken.groups.filter(x => x.includes('kubernetes:') && !x.includes('-admin')).map(x => x.replace('kubernetes:', ''));
        this.apiManager = new ApiManager(this.props.id_token, this.props.token_type, parsedToken.preferred_username, userGroups, parsedToken.namespace[0], adminGroups);
        this.state = {
            templateLabs: new Map(),
            instanceLabs: new Map(),
            selectedTemplate: {name: null, namespace: null},
            selectedInstance: null,
            events: "",
            statusHidden: true,
            adminHidden: true,
            isAdminSomewhere: adminGroups.length >= 0
        };
        this.retrieveCRDtemplates();
        this.retrieveCRDinstances()
            .catch((error) => {
                this.handleErrors(error);
            })
            .finally(() => {
                /*Start watching for namespaced events*/
                this.apiManager.startWatching(this.notifyEvent);

                /* @@@@@@@@@@@ TO BE USED ONLY IF WATCHER IS BROKEN
                this.retrieveCRDinstanceStatus();
                setInterval(() => {this.retrieveCRDinstanceStatus()}, 10000);
                */
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
     * Function to check the token, but encoded and decoded
     * @param parsed the decoded one
     * @return {boolean} true or false whether the token satisfies the constraints
     */
    checkToken(parsed, retrievedSessionToken) {
        if (!parsed.groups || !parsed.groups.length) {
            Toastr.error("You do not belong to any namespace to see laboratories");
            return false;
        }
        if (!parsed.namespace || !parsed.namespace[0]) {
            Toastr.error("You do not have your own namespace where to run laboratories");
            return false;
        }
        return true;
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
    startCRDinstance() {
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
            .then(() => {
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
                this.changeSelectedCRDinstance(null);
            });
    }

    /**
     * Function to perform logout when session with cluster expires
     */
    logoutInterval() {
        setInterval(() => {
            /*A reload probably is sufficient to re-authN the token*/
            this.props.logout();
        }, 2000);
    }

    /**
     * Function to connect to the VM of the actual selected CRD instance
     */
    connect() {
        if (!this.state.selectedInstance) {
            Toastr.info("No running lab selected to connect to");
            return;
        } else if (!this.state.instanceLabs.has(this.state.selectedInstance)) {
            Toastr.info("The lab `" + this.state.selectedInstance + "` is not running");
            return;
        } else {
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
        this.changeSelectedCRDinstance(null);
    }

    /**
     * * @@@@ UNUSED (since watcher has been patched and works)
     *
     * Function to retrieve all CRD instances status
     */
    retrieveCRDinstanceStatus() {
        const keys = Array.from(this.state.instanceLabs.keys());
        keys.forEach(lab => {
            this.apiManager.getCRDstatus(lab)
                .then(response => {
                    if (response.body.status && response.body.status.phase) {
                        let msg = "[" + response.body.metadata.creationTimestamp + "] " + lab + " => " + response.body.status.phase;
                        const newMap = this.state.instanceLabs;
                        if (response.body.status.phase.match(/Fail|Not/g)) {
                            /*Object creation failed*/
                            newMap.set(lab, {url: null, status: -1});
                        } else if (response.body.status.phase.match(/VmiReady/g)) {
                            /*Object creation succeeded*/
                            newMap.set(lab, {url: response.body.status.url, status: 1});
                        }
                        this.setState({instanceLabs: newMap, events: msg + "\n" + this.state.events})
                    }
                })
                .catch(error => {
                    this.handleErrors(error);
                });
        });
    }

    /**
     *Function to notify a Kubernetes Event related to your resources
     * @param type the type of the event
     * @param object the object of the event
     */
    notifyEvent(type, object) {
        if (!type) {
            /*Watch session ended, restart it*/
            this.apiManager.startWatching(this.notifyEvent);
            this.setState({events: ""});
            return;
        }
        if (object && object.status) {
            let msg = "[" + object.metadata.creationTimestamp + "] " + object.metadata.name + " {type: " + type + ", status: " + object.status.phase + "}";
            const newMap = this.state.instanceLabs;
            if (object.status.phase.match(/Fail|Not/g)) {
                /*Object creation failed*/
                newMap.set(object.metadata.name, {url: null, status: -1});
            } else if (object.status.phase.match(/VmiReady/g) && (type === "ADDED" || type === "MODIFIED")) {
                /*Object creation succeeded*/
                newMap.set(object.metadata.name, {url: object.status.url, status: 1});
            }
            this.setState({instanceLabs: newMap, events: msg + "\n" + this.state.events})
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
                msg += "Forbidden, something in the ticket renewal failed";
                this.logoutInterval();
                break;
            case 403 :
                msg += "It seems you do not have the right permissions to perform this operation";
                break;
            case 409 :
                msg += "The resource is already present";
                break;
            default :
                msg += "An error occurred(" + error.response._fetchResponse.status + "), please login again";
                this.logoutInterval()
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
                <Header logged={true} logout={this.props.logout} adminHidden={this.state.adminHidden}
                        renderAdminBtn={this.state.isAdminSomewhere}
                        switchAdminView={() => this.setState({adminHidden: !this.state.adminHidden})}/>
                <Container fluid className="cover mt-5">
                    {!this.state.adminHidden ? <ProfessorView
                            templateLabs={this.state.templateLabs}
                            instanceLabs={this.state.instanceLabs}
                            events={this.state.events}
                            showStatus={() => this.setState({statusHidden: !this.state.statusHidden})}
                            hidden={this.state.statusHidden}
                            // createTemplate={this.apiManager.createCRDtemplate(this.MyName,this.MyNamespace)}
                            // createLab={this.apiManager.createCRDinstance(this.MyName,this.MyNamespace)}
                            // deleteLab={this.apiManager.deleteCRDinstance(this.MyName)}
                            // enableOdisable={this.apiManager.setCRDinstanceStatus(CRDinstanceStatus)}

                        /> :
                        <StudentView templateLabs={this.state.templateLabs}
                                     funcTemplate={this.changeSelectedCRDtemplate}
                                     funcInstance={this.changeSelectedCRDinstance}
                                     start={this.startCRDinstance}
                                     instanceLabs={this.state.instanceLabs}
                                     connect={this.connect}
                                     stop={this.stopCRDinstance}
                                     events={this.state.events}
                                     showStatus={() => this.setState({statusHidden: !this.state.statusHidden})}
                                     hidden={this.state.statusHidden}/>}
                </Container>
                <Footer/>
            </div>
        );
    }
}
