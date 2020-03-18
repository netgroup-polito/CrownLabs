import React from 'react';
import {Container, Card, ButtonGroup, Button, Row, Col, Navbar, Nav} from 'react-bootstrap';
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
	   <div style={{minHeight: '100vh'}}>
                <header>
                    <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                        <Container>
                            <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                            <Nav className="ml-auto" as="ul">
                                <Nav.Item as="li">
                                    <Button variant="outline-light" onClick={this.props.authManager.logout}>Logout</Button>
                                </Nav.Item>
                            </Nav>
                        </Container>
                    </Navbar>
                </header>
                <Container fluid className="cover" style={{backgroundColor: '#F2F2F2'}}>
                    <Row className="mt-5">
                        <Col className="col-3">
                            <SideBar labs={this.state.labs} func={this.changeSelectedCRD}/>
                        </Col>
                        <Col className="col-6">
                            <Row className="my-5">
				<Col className="col-1"/>
                                <Col className="col-10">
                                    <Card className="text-center headerstyle">
                                        <Card.Body>
                                            <Card.Text as="h6">Status information</Card.Text>
                                            <textarea align="center" className="textareastyle"/>
                                        </Card.Body>
                                        <Card.Footer className="headerstyle">
                                            <ButtonGroup aria-label="Basic example">
                                                <Button variant="dark" className="text-success" onClick={this.startCRD}>Start</Button>
                                                <Button variant="dark" className="text-danger" onClick={this.stopCRD}>Stop</Button>
                                                <Button variant="dark">Connect</Button>
                                            </ButtonGroup>
                                        </Card.Footer>
                                    </Card>
                                </Col>
				<Col className="col-1"/>
                            </Row>
                        </Col>
			<Col className="col-2 text-center">
                            <Card className="my-5 p-2 text-center text-dark" border="dark" style={{backgroundColor: 'transparent'}}>
                                <Card.Body>
                                    <Card.Title className="p-2">Details</Card.Title>
				    <p>Selected Lab</p>
				    <p className="text-success">{this.state.selectedCRD}</p>
				    <p>Running Labs</p>
				    <ul>{this.state.runningCRD}</ul>
                                </Card.Body>
                            </Card>
                        </Col>
			<Col className="col-1"/>
                    </Row>
                </Container>
	   </div>
        );
    }
}
