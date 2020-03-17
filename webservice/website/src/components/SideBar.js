import React from 'react';
import {Nav} from 'react-bootstrap';

const logo = require('../assets/logo_poli2.png');

function SideBar() {
    return(
        <div className="p-3">
            <div className="ml-4"><img src={logo} height="60px"/></div>
            <h5 className="mt-3">Laboratories</h5>
            <Nav role="complementary" className="mt-4">
                <Nav.Item as="h7">Cloud Computing</Nav.Item>
                <Nav.Item as="ul">
                    <li><a href="">Lab 1</a></li>
                    <li><a href="">Lab 2</a></li>
                    <li><a href="">Lab 3</a></li>
                </Nav.Item>
            </Nav>
        </div>
    );
}

export default SideBar;