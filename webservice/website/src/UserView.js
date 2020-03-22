import React from 'react';
import {Row, Col} from 'react-bootstrap';
import SideBar from './components/SideBar';
import Footer from './components/Footer';
import InfoCard from "./components/InfoCard";
import CentralView from "./components/CentralView";
import Header from "./components/Header";
import ApiManager from "./services/ApiManager";
import Toastr from 'toastr';

import './App.css';
import 'toastr/build/toastr.min.css'

/**
 * Main window class, by now rendering only the unprivileged user view
 */
export default class UserView extends React.Component {
    constructor(props) {
        super(props);
        /*Attempting to retrieve the token stored in the sessionStorage by OIDC library, otherwise go back*/
        let retrievedSessionToken = JSON.parse(sessionStorage.getItem('oidc.user:' + OIDC_PROVIDER_URL + ":" + OIDC_CLIENT_ID));
        /*For future development: the parseToken function in the ApiManager could be moved here and check the token fields here,
        * allowing this class to understand if the user is unprivileged or not*/
        if (!retrievedSessionToken || !retrievedSessionToken.id_token) {
            Toastr.error("You received a non valid token, please check carefully its fields");
            sessionStorage.clear();
            document.location.reload();
        }
        this.connect = this.connect.bind(this);
        this.changeSelectedCRD = this.changeSelectedCRD.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRDinstance = this.stopCRDinstance.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);
        /*State variable which contains:
        * - all lab templates as a Map: (course_group => Array of available templates for that course)
        * - all lab instances as a Map: (instance_name => URL if running, null otherwise)
        * - current selected CRD as an object (name, namespace). Namespace set only when a lab instance is selected, not a template (needed by deletion)
        * - all namespaced events as a string
        * */
        this.state = {
            templateLabs: new Map(),
            instanceLabs: new Map(),
            selectedCRD: {name: null, namespace: null},
            events: ""
        };
        this.apiManager = new ApiManager(retrievedSessionToken.id_token, retrievedSessionToken.token_type || "Bearer");
        this.retrieveCRDtemplates();
        this.retrieveCRDinstances();
        /*Start watching for namespaced events*/
        this.apiManager.startWatching(this.notifyEvent);
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
        this.apiManager.getCRDinstances()
            .then((nodesResponse) => {
                const nodes = nodesResponse.body.items;
                let newMap = this.state.instanceLabs;
                nodes.forEach(x => {
                    if (!newMap.has(x.metadata.name)) {
                        newMap.set(x.metadata.name, null);
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
        if (!this.state.selectedCRD.name) {
            Toastr.info("Please select a lab before starting it");
            return;
        }
        if (this.state.instanceLabs.has(this.state.selectedCRD.name)) {
            Toastr.info("The `" + this.state.selectedCRD.name + '` lab is already running');
            return;
        }
        this.apiManager.createCRDinstance(this.state.selectedCRD.name, this.state.selectedCRD.namespace)
            .then(
                (response) => {
                    Toastr.success("Successfully started lab `" + this.state.selectedCRD.name + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.set(response.body.metadata.name, {status: 0, url: null});
                    this.setState({instanceLabs: newMap});
                },
                (error) => {
                    this.handleErrors(error);
                }
            )
            .finally(() => {
                this.changeSelectedCRD(null, null);
            });
    }

    /**
     * Function to stop and delete the current selected CRD instance
     */
    stopCRDinstance() {
        if (!this.state.selectedCRD.name) {
            Toastr.info("No lab to stop has been selected");
            return;
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD.name)) {
            Toastr.info("The `" + this.state.selectedCRD.name + '` lab is not running');
            return;
        }
        this.apiManager.deleteCRDinstance(this.state.selectedCRD.name)
            .then(
                (response) => {
                    Toastr.success("Successfully stopped `" + this.state.selectedCRD.name + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.delete(this.state.selectedCRD.name);
                    this.setState({instanceLabs: newMap});
                },
                (error) => {
                    this.handleErrors(error);
                }
            )
            .finally(() => {
                this.changeSelectedCRD(null, null);
            });
    }

    /**
     * Function to connect to the VM of the actual selected CRD instance
     */
    connect() {
        if (!this.state.selectedCRD.name) {
            Toastr.info("No lab selected to connect to");
            return
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD.name)) {
            Toastr.info("The lab `" + this.state.selectedCRD.name + "` is not running");
            return;
        }
        if (this.state.instanceLabs.get(this.state.selectedCRD.name).status !== 1) {
            Toastr.info("The lab `" + this.state.selectedCRD.name + "` is still starting");
            return;
        }
        window.open(this.state.instanceLabs.get(this.state.selectedCRD.name).url);
        this.changeSelectedCRD(null);
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
            return;
        }
        if (object && object.status) {
            let msg = "[" + object.metadata.creationTimestamp + "] " + object.metadata.name + "\n|===> Event Type: " + type + ", Status: " + object.status.phase;
            if (object.status.phase.match(/Fail|Not/g)) {
                /*Object creation failed*/
                const newMap = this.state.instanceLabs;
                newMap.set(object.metadata.name, {url: null, status: -1});
                this.setState({instanceLabs: newMap, events: msg + "\n" + this.state.events})
            } else if (object.status.phase.match(/VmiRunning/g) && (type === "ADDED" || type === "MODIFIED")) {
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
     * Function to change the user selected CRD
     * @param name the name/label of the new one
     * @param namespace the namespace in which the template should be retrieved (null if want to run an instance)
     */
    changeSelectedCRD(name, namespace) {
        this.setState({
            selectedCRD: {
                name: name, namespace: namespace
            }
        });
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
        /*For future development, a part from Footer and Header the other components could be moved into another UserView class
        (and renaming this one to MainWindow) and handle the rendering of the correct view(privileged or not) here in this method
        by checking the parsed token field*/
        return (
            <div style={{minHeight: '100vh'}}>
                <Header logged={true} logout={this.props.logout}/>
                <Row className="mt-5 p-3">
                    <Col className="col-3">
                        <SideBar labs={this.state.templateLabs} func={this.changeSelectedCRD}/>
                    </Col>
                    <Col className="col-6">
                        <CentralView start={this.startCRD} stop={this.stopCRDinstance} connect={this.connect}
                                     events={this.state.events}/>
                    </Col>
                    <Col className="col-3 text-center">
                        <InfoCard runningLabs={this.state.instanceLabs} selectedCRD={this.state.selectedCRD.name}
                                  func={this.changeSelectedCRD}/>
                    </Col>
                </Row>
                <Footer/>
            </div>
        );
    }
}
