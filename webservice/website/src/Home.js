import React from 'react';
import {Container, Card, Row, Col, Button} from 'react-bootstrap';
import Footer from "./components/Footer";
import './App.css';
import Header from "./components/Header";

/**
 * Function to draw the Home
 * @param props the function passed by App to perform login
 * @return the component to be drawn
 */
export default function Home(props) {
    return <div style={{minHeight: '100vh'}}>
            <Header logged={false}/>
            <Container fluid className="cover">
                <Row className="mt-5">
                    <Col className="col-2"/>
                    <Col className="col-8 mt-5">
                        <Card className="mt-5 p-3 text-center headerstyle">
                            <Card.Title as="h1">Welcome to CrownLabs!</Card.Title>
                            <Card.Body>
                                <Button variant="link" onClick={props.login}>Log in</Button>
                                <p className="d-inline">to access your laboratories.</p>
                            </Card.Body>
                        </Card>
                    </Col>
                    <Col className="col-2"/>
                </Row>
                <Footer/>
            </Container>
        </div>;
}
