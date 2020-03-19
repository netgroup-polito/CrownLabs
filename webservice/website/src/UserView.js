import React from 'react';
import {Container, Card, ButtonGroup, Button, Row, Col, Navbar, Nav} from 'react-bootstrap';
import SideBar from './components/SideBar';
import Footer from './components/Footer';
import './App.css';
import ApiManager from "./services/ApiManager";
import Toastr from 'toastr';
import 'toastr/build/toastr.min.css'

export default class UserView extends React.Component {
    constructor(props) {
        super(props);
        this.connect = this.connect.bind(this);
        this.changeSelectedCRD = this.changeSelectedCRD.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRD = this.stopCRD.bind(this);
        this.notifyEvent = this.notifyEvent.bind(this);
        this.startIntervalCRDstatus = this.startIntervalCRDstatus.bind(this);
        this.state = {templateLabs: [], instanceLabs: new Map(), selectedCRD: null, events: ""};
        if (localStorage.getItem('token')) {
            this.apiManager = new ApiManager(localStorage.getItem('token'), localStorage.getItem('token_type'));
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
            this.apiManager.getCRDinstance()
                .then((nodesResponse) => {
                    const nodes = nodesResponse.body.items;
                    let toAdd = new Map(this.state.instanceLabs);
                    for (const x in nodes) {
                        toAdd.set(nodes[x].metadata.name, "http://example.com/");
                    }
                    this.setState({instanceLabs: toAdd});
                })
                .catch((error) => {
                    console.error(error);
                });
            //this.startIntervalCRDstatus();
        }
    }

    startIntervalCRDstatus() {
        // TODO: to be completed in next PR
        setInterval(() => {
            alert("ci siamo");
            const keys = Array.from(this.state.instanceLabs.keys());
            for(const lab in keys) {
                this.apiManager.getCRDstatus(lab)
                    .then(response => {
                        console.log(response);
                    })
                    .catch(error => {
                        console.log(error);
                    });
            }
        }, 10000);
    }

    notifyEvent(msg, obj) {
        this.setState({events: this.state.events += msg += "\n"});
    }

    changeSelectedCRD(name) {
        this.setState({selectedCRD: name});
    }

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
                    const newMap = new Map(this.state.instanceLabs);
                    newMap.set(this.state.selectedCRD, "http://example.com");
                    this.setState({instanceLabs: newMap});
                },
                (error) => {
                    let code = error.response._fetchResponse.status;
                    let msg = "";
                    switch (code) {
                        case 409: {
                            msg += "";
                            break;
                        }
                        case 100 : {
                            break;
                        }
                        default : {
                            msg += "An unusual error occurred (" + code + "), please try again.";
                        }
                    }
                    Toastr.error(msg);
                }
            )
            .finally(() => {
                this.changeSelectedCRD(null);
            });
    }

    stopCRD() {
        if (!this.state.selectedCRD) {
            Toastr.info("No lab to stop has been selected");
            return;
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The `" + this.state.selectedCRD + '` lab is not running');
            return;
        }
        if (this.state.selectedCRD !== "") {
            this.apiManager.deleteCRD(this.state.selectedCRD)
                .then(
                    (response) => {
                        Toastr.success("Successfully stopped `" + this.state.selectedCRD + "`");
                        const newMap = new Map(this.state.instanceLabs);
                        newMap.delete(this.state.selectedCRD);
                        this.setState({instanceLabs: newMap});
                    },
                    (error) => {
                        let code = error.response._fetchResponse.status;
                        let msg = "";
                        switch (code) {
                            case 409: {
                                msg += "";
                                break;
                            }
                            case 100 : {
                                break;
                            }
                            default : {
                                msg += "An unusual error occurred (" + code + "), please try again.";
                            }
                        }
                        Toastr.error(msg);
                    }
                )
                .finally(() => {
                    this.changeSelectedCRD(null);
                });
        }
    }

    connect() {
        if (!this.state.selectedCRD) {
            Toastr.info("No lab selected to connect to");
            return
        }
        if (!this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The lab `" + this.state.selectedCRD + "` is not running");
            return;
        }
        window.open(this.state.instanceLabs.get(this.state.selectedCRD), '_blank');
        this.changeSelectedCRD(null);
    }

    render() {
        const keys = Array.from(this.state.instanceLabs.keys());
        const runningLabs = keys.map(x => {
            return <Button key={x} variant="link" onClick={() => this.changeSelectedCRD(x)}>{x}</Button>;
        });
        return (
            <div style={{minHeight: '100vh'}}>
                <header>
                    <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                        <Container>
                            <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                            <Nav className="ml-auto" as="ul">
                                <Nav.Item as="li">
                                    <Button variant="outline-light"
                                            onClick={this.props.authManager.logout}>Logout</Button>
                                </Nav.Item>
                            </Nav>
                        </Container>
                    </Navbar>
                </header>
                <Container fluid className="cover" style={{backgroundColor: '#F2F2F2'}}>
                    <Row className="mt-5">
                        <Col className="col-3">
                            <SideBar labs={this.state.templateLabs} func={this.changeSelectedCRD}/>
                        </Col>
                        <Col className="col-6">
                            <Row className="my-5">
                                <Col className="col-1"/>
                                <Col className="col-10">
                                    <Card className="text-center headerstyle">
                                        <Card.Body>
                                            <Card.Text as="h6">Status information</Card.Text>
                                            <textarea readOnly align="center" className="textareastyle"
                                                      value={this.state.events}/>
                                        </Card.Body>
                                        <Card.Footer className="headerstyle">
                                            <ButtonGroup aria-label="Basic example">
                                                <Button variant="dark" className="text-success"
                                                        onClick={this.startCRD}>Start</Button>
                                                <Button variant="dark" className="text-danger"
                                                        onClick={this.stopCRD}>Stop</Button>
                                                <Button variant="dark" onClick={this.connect}>Connect</Button>
                                            </ButtonGroup>
                                        </Card.Footer>
                                    </Card>
                                </Col>
                                <Col className="col-1"/>
                            </Row>
                        </Col>
                        <Col className="col-2 text-center">
                            <Card className="my-5 p-2 text-center text-dark" border="dark"
                                  style={{backgroundColor: 'transparent'}}>
                                <Card.Body>
                                    <Card.Title className="p-2">Details</Card.Title>
                                    <p>Selected Lab</p>
                                    <p className="text-success">{this.state.selectedCRD || "-"}</p>
                                    <p>Running Labs</p>
                                    <p className="text-success">{runningLabs.length > 0 ? "" : "-"}</p>
                                    {runningLabs}
                                </Card.Body>
                            </Card>
                        </Col>
                        <Col className="col-1"/>
                    </Row>
                    <Footer/>
                </Container>
            </div>
        );
    }
}
