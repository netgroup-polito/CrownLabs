import React from 'react';
import { Card, Col, Row } from 'react-bootstrap';
import Container from '@material-ui/core/Container';
import Footer from '../components/Footer';
import '../App.css';
import Header from '../components/Header';

/**
 * Function to draw the Home
 * @param props the function passed by App to perform login
 * @return the component to be drawn
 */
export default function Home(props) {
  return (
    <div style={{ minHeight: '100vh' }}>
      <Header logged={false} />
      <Container
        // the height of the container is viewport heigh - header height(70) - footer height(70)
        style={{
          height: 'calc(100vh - 140px)',
          overflow: 'auto'
        }}
      >
        <Row className="mt-5">
          <Col className="col-2" />
          <Col className="col-8 mt-5">
            <Card className="mt-5 p-3 text-center headerstyle">
              <Card.Title as="h1">Welcome to CrownLabs!</Card.Title>
              <Card.Body>
                <a
                  href="#"
                  onClick={props.login}
                  style={{ color: '#0000FF', background: '#ffffa0' }}
                >
                  Log in
                </a>
                <p className="d-inline"> to access your laboratories.</p>
              </Card.Body>
            </Card>
          </Col>
          <Col className="col-2" />
        </Row>
      </Container>
      <Footer />
    </div>
  );
}
