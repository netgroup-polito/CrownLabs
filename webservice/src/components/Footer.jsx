import React from 'react';
import GitHubLogo from '../assets/github-logo.png';
/**
 * Function to draw the document footer
 * @return the object to be drawn
 */

export default function Footer() {
  return (
    <div
      id="footer"
      style={{
        height: '70px',
        background: '#032364',
        fontSize: '1.2rem',
        color: 'white',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center'
      }}
    >
      <p>
        This software has been proudly developed at Politecnico di Torino. For
        info visit our&nbsp;
      </p>
      <a
        href="https://github.com/netgroup-polito/CrownLabs"
        style={{ color: 'inherit' }}
      >
        <p href="https://www.google.it">GitHub Page</p>
      </a>
      &nbsp;
      <a href="https://github.com/netgroup-polito/CrownLabs">
        <img height="25px" src={GitHubLogo} alt="GitHub logo" />
      </a>
    </div>
  );
}
