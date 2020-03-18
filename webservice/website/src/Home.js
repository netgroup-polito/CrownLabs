import React from 'react';
import {Navbar, Nav, NavItem, Container, Card, Row, Col, Button} from 'react-bootstrap';
import './App.css';

const logo = require('./assets/logo_poli3.png');
const githubIcon = require('./assets/github-logo.png');

export default function Home(props) {
    return (
            <div style={{minHeight: '100vh'}}>
                <header>
                    <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                        <Container>
                            <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                            <Nav className="ml-auto" as="ul">
                                <NavItem as="li">
                                    <img src={logo} height="50px" alt=""/>
                                </NavItem>
                            </Nav>
                        </Container>
                    </Navbar>
                </header>
                <Container fluid className="cover">
                    <Row className="mt-5">
                        <Col className="col-2"/>
                        <Col className="col-8 mt-5">
                            <Card className="mt-5 p-3 text-center headerstyle">
                                <Card.Title as="h1">Welcome to CrownLabs web site!</Card.Title>
                                <Card.Body>
                                    <Button variant="link" onClick={props.authManager.login}>Log in</Button>
                                    <p className="d-inline">to access laboratories.</p>
                                </Card.Body>
                            </Card>
                        </Col>
                        <Col className="col-2"/>
                    </Row>
		<footer className="py-4 blockquote-footer footerstyle">
                    <Container fluid className="m-0 text-center text-secondary">
                        <p className="d-inline">This software has been proudly developed at Politecnico di Torino. </p>
                        <p className="d-inline">For info visit our</p>
                        <img className="d-inline" height="25px" src={githubIcon} alt="GitHub logo"/>
                        <a className="d-inline" href="https://github.com/netgroup-polito/CrownLabs">Github project
                            repository</a>
                    </Container>
                </footer>
                </Container>
                
            </div>
        );
}
