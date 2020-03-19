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
        this.connect = this.connect.bind(this);
        this.changeSelectedCRD = this.changeSelectedCRD.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRDinstance = this.stopCRDinstance.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);
        this.retrieveCRDinstanceStatus = this.retrieveCRDinstanceStatus.bind(this);
        this.state = {templateLabs: [], instanceLabs: new Map(), selectedCRD: null, events: ""};
        if (localStorage.getItem('token')) {
            this.apiManager = new ApiManager(localStorage.getItem('token'), localStorage.getItem('token_type'));
            this.retrieveCRDtemplate();
            this.retrieveCRDinstance();
        }
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
            })
            .catch((error) => {
                console.error(error);
            })
            .finally(() => {
                this.retrieveCRDinstanceStatus()
            });
        setInterval(() => {
            this.retrieveCRDinstanceStatus();
        }, 10000);
    }

    /**
     * Function to retrieve all CRD templates available
     */
    retrieveCRDtemplate() {
        this.apiManager.getCRDtemplate()
            .then((nodesResponse) => {
                const nodes = nodesResponse.body.items;
                this.setState({
                    templateLabs: nodes.map(x => {
                        return x.metadata.name;
                    })
                });
                this.apiManager.startWatching(this.notifyEvent);
            })
            .catch((error) => {
                console.error(error);
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
                        if (this.state.instanceLabs.get(lab) !== status) {
                            newMap.set(lab, status);
                        }
                        this.setState({instanceLabs: newMap});
                    }
                })
                .catch(error => {
                    console.log(error);
                });
        });
    }

    /**
     *Function to notify a Kubernetes Event related to your resources
     * @param msg the message received
     * @param obj the resource of interest
     */
    notifyEvent(msg, obj) {
        this.setState({events: msg + "\n" + this.state.events});
    }

    /**
     * Function to change the user selected CRD
     * @param name the name/label of the new one
     */
    changeSelectedCRD(name) {
        this.setState({selectedCRD: name});
    }

    /**
     * Function to start and create a CRD instance using the actual selected one
     */
    startCRD() {
        if (!this.state.selectedCRD) {
            Toastr.info("Please select a lab before starting it");
            return;
        }
        if (this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The `" + this.state.selectedCRD + '` lab is already running');
            return;
        }
        this.apiManager.createCRD(this.state.selectedCRD)
            .then(
                (response) => {
                    Toastr.success("Successfully started lab `" + this.state.selectedCRD + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.set(this.state.selectedCRD, null);
                    this.setState({instanceLabs: newMap});
                },
                (error) => {
                    let code = error.response._fetchResponse.status;
                    let msg = "";
                    switch (code) {
                        case 403:
                            msg += "Unable to create " + this.state.selectedCRD + ", lack of permissions";
                            break;
                        case 409 :
                            msg += "Resource " + this.state.selectedCRD + " already exists";
                            break;
                        default :
                            msg += "An unusual error occurred (" + code + "), please try again";
                    }
                    Toastr.error(msg);
                }
            )
            .finally(() => {
                this.changeSelectedCRD(null);
            });
    }

    /**
     * Function to stop and delete the current selected CRD instance
     */
    stopCRDinstance() {
        if (!this.state.selectedCRD) {
            Toastr.info("No lab to stop has been selected");
            return;
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The `" + this.state.selectedCRD + '` lab is not running');
            return;
        }
        this.apiManager.deleteCRD(this.state.selectedCRD)
            .then(
                (response) => {
                    Toastr.success("Successfully stopped `" + this.state.selectedCRD + "`");
                    const newMap = this.state.instanceLabs;
                    newMap.delete(this.state.selectedCRD);
                    this.setState({instanceLabs: newMap});
                },
                (error) => {
                    let code = error.response._fetchResponse.status;
                    let msg = code === 403 ? "Unable to delete " + this.state.selectedCRD + ", lack of permissions" : "An unusual error occurred (" + code + "), please try again";
                    Toastr.error(msg);
                }
            )
            .finally(() => {
                this.changeSelectedCRD(null);
            });
    }

    /**
     * Function to connect to the VM of the actual selected CRD instance
     */
    connect() {
        if (!this.state.selectedCRD) {
            Toastr.info("No lab selected to connect to");
            return
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The lab `" + this.state.selectedCRD + "` is not running");
            return;
        }
        if (this.state.instanceLabs.get(this.state.selectedCRD) === null) {
            Toastr.info("The lab `" + this.state.selectedCRD + "` is still starting");
            return;
        }
        window.open("https://" + this.state.instanceLabs.get(this.state.selectedCRD));
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
                           onClick={() => this.changeSelectedCRD(x)}>{x}</Button>;
        });
        return (
            <div style={{minHeight: '100vh'}}>
                <header>
                    <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                        <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                        <Nav className="ml-auto" as="ul">
                            <Nav.Item as="li">
                                <Button variant="outline-light"
                                        onClick={this.props.authManager.logout}>Logout</Button>
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
                        <InfoCard runningLabs={runningLabs} selectedCRD={this.state.selectedCRD}/>
                        <Col className="col-1"/>
                    </Row>
                    <Footer/>
                </Container>
            </div>
        );
    }
}
