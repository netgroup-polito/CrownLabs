import { Button, Nav, Navbar } from 'react-bootstrap';
import React from 'react';
import NavItem from 'react-bootstrap/NavItem';
import FolderSharedIcon from '@material-ui/icons/FolderShared';
import ToolTip from '@material-ui/core/Tooltip';

/**
 * Function to draw the page header
 * @param props the property to check whether it is logged or not, to draw the apposite component
 * @return the component to be drawn
 */
export default function Header(props) {
  const toDraw = props.logged ? (
    <Button variant="outline-light" onClick={props.logout}>
      Logout
    </Button>
  ) : (
    <img src={require('../assets/logo_poli3.png')} height="50px" alt="" />
  );
  const name = props.adminHidden ? 'Professor Area' : 'Student Area';
  const adminBtn = props.renderAdminBtn ? (
    <Button variant="outline-light" onClick={props.switchAdminView}>
      {name}
    </Button>
  ) : (
    <div />
  );
  return (
    <header
      style={{
        position: 'sticky',
        top: 0,
        display: 'flex',
        justifyContent: 'center',
        backgroundColor: '#032364',
        alignContent: 'center',
        height: 70,
        padding: '0 20px'
      }}
    >
      <Navbar className="nav_new" variant="dark">
        <img src={require('../assets/crown.png')} height="40px" alt="" />;
        <Navbar.Brand className="navText" href="">
          CrownLabs
        </Navbar.Brand>
        <Nav className="ml-auto" as="ul">
          <Navbar.Text className="navText" href="" style={{ marginRight: 20 }}>
            {props.logged && props.name ? 'Welcome back, ' + props.name : ''}
          </Navbar.Text>
          <a href="https://crownlabs.polito.it/cloud" target="_blank">
            <ToolTip title="My drive">
              <FolderSharedIcon
                style={{
                  marginRight: 25,
                  color: 'white',
                  fontSize: '2.6rem'
                }}
              />
            </ToolTip>
          </a>
          <NavItem as="li" className="mr-2">
            {adminBtn}
          </NavItem>
          <NavItem as="li">{toDraw}</NavItem>
        </Nav>
      </Navbar>
    </header>
  );
}
