import React from 'react';
import {Nav} from 'react-bootstrap';
import {Button} from "react-bootstrap";

export default function SideBar(props) {
    const labs = props.labs.map(name => {
        return <li key={name}><Button variant="link" onClick={() => props.func(name)}>{name}</Button></li>;
    });
    return (
        <div className="p-3">
            <h5 className="mt-3">Laboratories</h5>
            <Nav role="complementary" className="mt-4">
                <Nav.Item as="h6">Cloud Computing
                    <Nav.Item as="ul">
                        {labs}
                    </Nav.Item>
                </Nav.Item>
            </Nav>
        </div>
    );
}
