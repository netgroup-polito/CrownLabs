import { Container } from 'react-bootstrap';
import React from 'react';

/**
 * Function to draw the document footer
 * @return the object to be drawn
 */
export default function Footer() {
  return (
    <footer
      id="footer"
      className="py-4"
      style={{ height: '70px', background: '#032364', fontSize: '1.2rem' }}
    >
      <Container fluid className="m-0 text-center text-secondary">
        <p className="d-inline" style={{ color: 'white' }}>
          This software has been proudly developed at Politecnico di Torino.{' '}
        </p>
        <p className="d-inline" style={{ color: 'white' }}>
          For info visit our
        </p>
        <img
          className="d-inline"
          height="25px"
          src={require('../assets/github-logo.png')}
          alt="GitHub logo"
        />
        <a
          className="d-inline"
          href="https://github.com/netgroup-polito/CrownLabs"
          style={{ color: 'white' }}
        >
          Github project repository
        </a>
      </Container>
    </footer>
  );
}
