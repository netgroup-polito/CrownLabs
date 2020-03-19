import React from 'react';
import {Container, Card, ButtonGroup, Button, Row, Col} from 'react-bootstrap';
import SideBar from './components/SideBar';
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
        }
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
        if(!this.state.selectedCRD) {
            Toastr.info("No lab selected to connect to");
            return
        }
        if(!this.state.instanceLabs.has(this.state.selectedCRD)) {
            Toastr.info("The lab `" + this.state.selectedCRD +  "` is not running");
            return;
        }
        window.open(this.state.instanceLabs.get(this.state.selectedCRD), '_blank');
        this.changeSelectedCRD(null);
    }

    render() {
        const keys = Array.from(this.state.instanceLabs.keys());
        const runningLabs = keys.map(x => {
            return <li key={x}><Button variant="link" onClick={() => this.changeSelectedCRD(x)}>{x}</Button></li>;
        });
        return (
            <div style={{backgroundColor: '#F2F2F2'}}>
                <Container className="cover">
                    <Row>
                        <Col className="col-3">
                            <SideBar labs={this.state.templateLabs} func={this.changeSelectedCRD}/>
                        </Col>
                        <Col className="col-9">
                            <Row className="my-5">
                                <Col className="col-6">
                                    <Card className="my-5 text-center headerstyle">
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
                                                <Button variant="light"
                                                        onClick={this.props.authManager.logout}>Logout</Button>
                                            </ButtonGroup>
                                        </Card.Footer>
                                    </Card>
                                </Col>
                                <Col className="col-6">
                                    <Card className="my-5 text-center headerstyle">
                                        <Card.Body>
                                            <Card.Text as="h6">Information</Card.Text>
                                            <p align="center">Selected lab: {this.state.selectedCRD}</p>
                                            <ul>
                                                {runningLabs}
                                            </ul>
                                        </Card.Body>
                                    </Card>
                                </Col>
                            </Row>
                        </Col>
                    </Row>
                </Container>
            </div>
        );
    }
}