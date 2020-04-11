import React from 'react';
import {Container, Button, Row, Col, Navbar, Nav} from 'react-bootstrap';
import SideBar from './components/SideBar';
import Footer from './components/Footer';
import InfoCard from "./components/InfoCard";
import CentralView from "./components/CentralView";
import ApiManager from "./services/ApiManager";
import Toastr from 'toastr';

import './App.css';
import 'toastr/build/toastr.min.css'

export default class UserView extends React.Component {
    constructor(props) {
        super(props);
        let retrievedSessionToken = JSON.parse(sessionStorage.getItem('oidc.user:' + OIDC_PROVIDER_URL + ":" + OIDC_CLIENT_ID));
        if (!retrievedSessionToken) {
            document.location.href = '/logout';
        }
        this.connect = this.connect.bind(this);
        this.changeSelectedCRD = this.changeSelectedCRD.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRDinstance = this.stopCRDinstance.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);
        this.retrieveCRDinstanceStatus = this.retrieveCRDinstanceStatus.bind(this);
        this.state = {
            templateLabs: new Map(),
            instanceLabs: new Map(),
            selectedCRD: {name: null, namespace: null},
            events: ""
        };
        this.apiManager = new ApiManager(retrievedSessionToken.id_token, retrievedSessionToken.token_type || "Bearer");
        this.retrieveCRDtemplates();
        this.retrieveCRDinstance();
    }

    /**
     * Function to retrieve all CRD instances running
     */
    retrieveCRDinstance() {
        this.apiManager.getCRDinstance()
            .then((nodesResponse) => {
                const nodes = nodesResponse.body.items;
                let newMap = new Map();
                nodes.forEach(x => {
                    newMap.set(x.metadata.name, null);
                });
                this.setState({instanceLabs: newMap});
                this.retrieveCRDinstanceStatus();
                setInterval(() => {
                    this.retrieveCRDinstanceStatus();
                }, 10000);
            })
            .catch((error) => {
                this.handleErrors(error);
            });
    }

    /**
     * Function to retrieve all CRD templates available
     */
    retrieveCRDtemplates() {
        this.apiManager.getCRDtemplates()
            .then(res => {
                let newMap = this.state.templateLabs;
                res.forEach(x => {
                    x ? newMap.set(x.course, x.labs) : null;
                });
                this.setState({templateLabs: newMap});
                this.apiManager.startWatching(this.notifyEvent);
            })
            .catch((error) => {
                this.handleErrors(error);
            });
    }

    /**
     * Function to retrieve all CRD instances status
     */
    retrieveCRDinstanceStatus() {
        const keys = Array.from(this.state.instanceLabs.keys());
        keys.forEach(lab => {
            this.apiManager.getCRDstatus(lab)
                .then(response => {
                    if (response.body.status && response.body.status.url) {
                        const newMap = this.state.instanceLabs;
                        const status = response.body.status.url;
                        if (this.state.instanceLabs.get(lab) !== status) {
                            this.notifyEvent("[" + response.body.metadata.creationTimestamp + "] " + response.body.status.phase);
                            newMap.set(lab, status);
                        }
                        this.setState({instanceLabs: newMap});
                    }
                })
                .catch(error => {
                    this.handleErrors(error);
                });
        });
    }

    /**
     *Function to notify a Kubernetes Event related to your resources
     * @param msg the message received
     * @param obj the resource of interest
     */
    notifyEvent(msg) {
        this.setState({events: this.state.events + msg + "\n"});
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
        this.apiManager.createCRD(this.state.selectedCRD.name, this.state.selectedCRD.namespace)
            .then(
                (response) => {
                    Toastr.success("Successfully started lab `" + this.state.selectedCRD.name + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.set(this.state.selectedCRD.name, null);
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

    handleErrors(error) {
        let msg = "";
        switch (error.response._fetchResponse.status) {
            case 401 :
                msg += "Cluster expired, please authenticate again";
                setInterval(() => {
                    document.location.href = '/logout'
                }, 2000);
                break;
            case 403 :
                msg += "It seems you do not have the right permissions to perform this operation";
                break;
            case 409 :
                msg += "Resource already present";
                break;
            default :
                msg += "An error occurred(" + error.response._fetchResponse.status + "), please try again";
                setInterval(() => {
                    document.location.href = '/logout'
                }, 2000);
        }
        Toastr.error(msg);
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
        this.apiManager.deleteCRD(this.state.selectedCRD.name)
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
        if (this.state.instanceLabs.get(this.state.selectedCRD.name) === null) {
            Toastr.info("The lab `" + this.state.selectedCRD.name + "` is still starting");
            return;
        }
        window.open(this.state.instanceLabs.get(this.state.selectedCRD.name));
        this.changeSelectedCRD(null);
    }

    /**
     * Function to render this component,
     * It automatically updates every new change in the state variable
     * @returns the component to be drawn
     */
    render() {
        /*Retrieving instance labs to be drawn in the right bar and foreach one draw a button*/
        const keys = Array.from(this.state.instanceLabs.keys());
        const runningLabs = keys.map(x => {
            let color = this.state.instanceLabs.get(x) === null ? 'red' : 'green';
            return <Button key={x} variant="link" style={{color: color}}
                           onClick={() => this.changeSelectedCRD(x, null)}>{x}</Button>;
        });
        return (
            <div style={{minHeight: '100vh'}}>
                <header>
                    <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                        <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                        <Nav className="ml-auto" as="ul">
                            <Nav.Item as="li">
                                <Button variant="outline-light"
                                        onClick={this.props.logout}>Logout</Button>
                            </Nav.Item>
                        </Nav>
                    </Navbar>
                </header>
                <Container fluid className="cover" style={{backgroundColor: '#F2F2F2'}}>
                    <Row className="mt-5">
                        <Col className="col-3">
                            <SideBar labs={this.state.templateLabs} func={this.changeSelectedCRD}/>
                        </Col>
                        <Col className="col-6">
                            <CentralView start={this.startCRD} stop={this.stopCRDinstance} connect={this.connect}
                                         events={this.state.events}/>
                        </Col>
                        <InfoCard runningLabs={runningLabs} selectedCRD={this.state.selectedCRD.name}/>
                        <Col className="col-1"/>
                    </Row>
                    <Footer/>
                </Container>
            </div>
        );
    }
}
