import {Container} from "react-bootstrap";
import React from "react";

const githubIcon = require('../assets/github-logo.png');

export default function funtionFooter() {
    return <footer className="py-4 blockquote-footer footerstyle">
        <Container fluid className="m-0 text-center text-secondary">
            <p className="d-inline">This software has been proudly developed at Politecnico di Torino. </p>
            <p className="d-inline">For info visit our</p>
            <img className="d-inline" height="25px" src={githubIcon} alt="GitHub logo"/>
            <a className="d-inline" href="https://github.com/netgroup-polito/CrownLabs">Github project
                repository</a>
        </Container>
    </footer>;
}