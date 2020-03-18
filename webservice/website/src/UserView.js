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
        this.state = {labs: [], selectedCRD: "", runningCRD: ""};
        if (localStorage.getItem('token')) {
            this.apiManager = new ApiManager(localStorage.getItem('token'), localStorage.getItem('token_type'));
            this.apiManager.getCRD()
                .then((nodesResponse) => {
                    const nodes = nodesResponse.body.items;
                    this.setState({labs: nodes.map(x => {
                        return x.metadata.name;
                    })});
                })
                .catch((error) => {
                    console.error(error);
                });
        }
        this.changeSelectedCRD = this.changeSelectedCRD.bind(this);
        this.startCRD = this.startCRD.bind(this);
        this.stopCRD = this.stopCRD.bind(this);
    }
    changeSelectedCRD(name) {
        this.setState({selectedCRD: name});
    }
    startCRD() {
        if(this.state.selectedCRD !== "") {
            this.apiManager.createCRD(this.state.selectedCRD)
                .then(
                    (response) => {
                        Toastr.success("Successfully started lab `" + this.state.selectedCRD + "`");
                        this.setState({runningCRD: this.state.selectedCRD});
                    },
                    (err) => {
                        console.log(err);
                    }
                );
        }
    }
    stopCRD() {
        if(this.state.runningCRD !== "") {
            this.apiManager.deleteCRD(this.state.runningCRD)
                .then(
                    (response) => {
                        Toastr.success("Successfully stopped `" + this.state.runningCRD + "`");
                        this.setState({runningCRD: ""});
                        console.log(response);
                    },
                    (error) => {
                        console.log(error);
                    }
                );
        }
    }
    render() {
        return (
            <div style={{backgroundColor: '#F2F2F2'}}>
                <Container className="cover">
                    <Row>
                        <Col className="col-3">
                            <SideBar labs={this.state.labs} func={this.changeSelectedCRD}/>
                        </Col>
                        <Col className="col-9">
                            <Row className="my-5">
                                <Col className="col-6">
                                    <Card className="my-5 text-center headerstyle">
                                        <Card.Body>
                                            <Card.Text as="h6">Status information</Card.Text>
                                            <textarea align="center" className="textareastyle"/>
                                        </Card.Body>
                                        <Card.Footer className="headerstyle">
                                            <ButtonGroup aria-label="Basic example">
                                                <Button variant="dark" className="text-success" onClick={this.startCRD}>Start</Button>
                                                <Button variant="dark" className="text-danger" onClick={this.stopCRD}>Stop</Button>
                                                <Button variant="dark">Connect</Button>
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
                                            <p align="center">Running lab:  {this.state.runningCRD}</p>
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