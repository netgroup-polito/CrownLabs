import {Button, Nav, Navbar} from "react-bootstrap";
import React from "react";
import NavItem from "react-bootstrap/NavItem";

export default function Header(props) {
    const toDraw = props.logged? <Button variant="outline-light" onClick={props.logout}>Logout</Button> : <img src={require('../assets/logo_poli3.png')} height="50px" alt=""/>;
    return <header>
        <Navbar bg="dark" variant="dark" expand="lg">
            <Navbar.Brand href="">CrownLabs</Navbar.Brand>
            <Nav className="ml-auto" as="ul">
                <NavItem as="li">
                    {toDraw}
                </NavItem>
            </Nav>
        </Navbar>
    </header>
}
