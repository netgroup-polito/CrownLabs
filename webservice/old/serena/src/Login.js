import React from 'react';
import { Container, Card, Form, InputGroup, FormControl, Button, Row, Col } from 'react-bootstrap';
import MaterialIcon from 'material-icons-react';
import './App.css';

const logo = require('./assets/logo_poli.png');

function Login() {
    return(
        <div>
            <Container>
                <Row className="my-5">
                    <Col className="col-4"></Col>
                    <Col className="col-4">
                        <Card className="my-5 p-2" bg="light">
                            <Card.Header className="text-center headerstyle">
                                <Card.Img variant="top" src={logo} style={{width: '70%', height: '70%'}}/>
                            </Card.Header>
                            <Card.Body>
                                <Form>
                                    <Form.Group>
                                        <Form.Label style={{color: '#004990'}}>Email</Form.Label>
                                        <InputGroup className="mb-3">
                                            <FormControl placeholder="Email" type="email" name="email" id="email" required/>
                                            <InputGroup.Append>
                                                <InputGroup.Text>
                                                    <MaterialIcon icon="email" color='#004990'/>
                                                </InputGroup.Text>
                                            </InputGroup.Append>
                                        </InputGroup>
                                    </Form.Group>
                                    <Form.Group>
                                        <Form.Label style={{color: '#004990'}}>Password</Form.Label>
                                        <a className="text-primary float-right" href="#">Forgot password?</a>
                                        <InputGroup className="mb-3">
                                            <FormControl placeholder="********" type="password" name="password" id="password" required/>
                                            <InputGroup.Append>
                                                <InputGroup.Text>
                                                    <MaterialIcon icon="vpn_key" color='#004990'/>
                                                </InputGroup.Text>
                                            </InputGroup.Append>
                                        </InputGroup>
                                    </Form.Group>
                                    <Button className="btn-block login mt-5" variant="primary">Login</Button>
                                </Form>
                            </Card.Body>
                        </Card>
                    </Col>
                    <Col className="col-4"></Col>
                </Row>
            </Container>
        </div>
    );
}

export default Login;