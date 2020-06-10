import React from 'react';
import GitHubButton from 'react-github-btn';

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
        This software has been proudly developed at Politecnico di Torino
        &nbsp;&nbsp;&nbsp;
      </p>
      <GitHubButton
        href="https://github.com/netgroup-polito/CrownLabs"
        data-size="large"
        data-show-count="true"
        aria-label="Star netgroup-polito/CrownLabs on GitHub"
      >
        Star
      </GitHubButton>
    </div>
  );
}
