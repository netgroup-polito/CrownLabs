import React from 'react';
import {Container, Card, ButtonGroup, Button, Row, Col} from 'react-bootstrap';
import SideBar from './components/SideBar';
import './App.css';
import ApiManager from "./services/ApiManager";

export default class UserView extends React.Component {
    constructor(props) {
        super(props);
        this.state = {labs: []};
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
        this.print = this.print.bind(this);
    }
    print(name) {
        alert(name);
    }
    render() {
        return (
            <div style={{backgroundColor: '#F2F2F2'}}>
                <Container className="cover">
                    <Row>
                        <Col className="col-3">
                            <SideBar labs={this.state.labs} func={this.print}/>
                        </Col>
                        <Col className="col-9">
                            <Row className="my-5">
                                <Col className="col-2"/>
                                <Col className="col-8">
                                    <Card className="my-5 text-center headerstyle">
                                        <Card.Body>
                                            <Card.Text as="h6">Status information</Card.Text>
                                            <textarea align="center" className="textareastyle"/>
                                        </Card.Body>
                                        <Card.Footer className="headerstyle">
                                            <ButtonGroup aria-label="Basic example">
                                                <Button variant="dark" className="text-success" onClick={this.apiManager.createCRD}>Start</Button>
                                                <Button variant="dark" className="text-danger" onClick={() => this.apiManager.deleteCRD()}>Stop</Button>
                                                <Button variant="dark">Reset</Button>
                                                <Button variant="light"
                                                        onClick={this.props.authManager.logout}>Logout</Button>
                                            </ButtonGroup>
                                        </Card.Footer>
                                    </Card>
                                </Col>
                                <Col className="col-2"/>
                            </Row>
                        </Col>
                    </Row>
                </Container>
            </div>
        );
    }
}