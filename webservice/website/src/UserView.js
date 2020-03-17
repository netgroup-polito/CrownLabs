import React from 'react';
import { Container, Card, ButtonGroup, Button, Row, Col } from 'react-bootstrap';
import SideBar from './components/SideBar';
import './App.css';

function UserView() {
    return(
        <div style={{backgroundColor: '#F2F2F2'}}>
            <Container className="cover">
                <Row>
                    <Col className="col-2">
                        <SideBar />
                    </Col>
                    <Col className="col-10">
                        <Row className="my-5">
                        <Col className="col-2"></Col>
                        <Col className="col-8">
                            <Card className="my-5 text-center headerstyle">
                                <Card.Body>
                                    <Card.Text as="h6">Status information</Card.Text>
                                    <textarea align="center" className="textareastyle"></textarea>
                                </Card.Body>
                                <Card.Footer className="headerstyle">
                                    <ButtonGroup aria-label="Basic example">
                                        <Button variant="dark" className="text-success">Start</Button>
                                        <Button variant="dark" className="text-danger">Stop</Button>
                                        <Button variant="dark">Reset</Button>
                                    </ButtonGroup>
                                </Card.Footer>  
                            </Card>
                        </Col>
                        <Col className="col-2"></Col>
                    </Row>
                    </Col>
                </Row>
            </Container>
        </div>
    );
}

export default UserView;