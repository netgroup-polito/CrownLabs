import React from 'react';
import {Navbar, Nav, NavItem, Container, Card, Row, Col, Button, Image} from 'react-bootstrap';
import Footer from "./components/Footer";
import './App.css';

const logo = require('./assets/logo_poli3.png');
const logo2 = require('./assets/logo_poli.png');

export default function Home(props) {
    return (
        <div style={{minHeight: '100vh'}}>
            <header>
                <Navbar bg="dark" variant="dark" expand="lg" fixed="top">
                    <Navbar.Brand href="">CrownLabs</Navbar.Brand>
                    <Nav className="ml-auto" as="ul">
                        <NavItem as="li">
                            <img src={logo} height="50px" alt=""/>
                        </NavItem>
                    </Nav>
                </Navbar>
            </header>
            <Container fluid className="cover">
                <Row className="mt-5">
                    <Col className="col-2"/>
                    <Col className="col-8 mt-5">
                        <Card className="mt-5 p-3 text-center headerstyle">
                            <Card.Title as="h1">Welcome to CrownLabs!</Card.Title>
                            <Card.Body>
                                <Button variant="link" onClick={props.authManager.login}>Log in</Button>
                                <p className="d-inline">to access your laboratories.</p>
                            </Card.Body>
                        </Card>
                    </Col>
                    <Col className="col-2"/>
                </Row>
                <Footer/>
            </Container>

        </div>
    );
}
