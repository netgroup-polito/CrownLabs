import {Card, Col} from "react-bootstrap";
import React from "react";

export default function InfoCard(props) {
    return <Col className="col-2 text-center">
        <Card className="my-5 p-2 text-center text-dark" border="dark"
              style={{backgroundColor: 'transparent'}}>
            <Card.Body>
                <Card.Title className="p-2">Details</Card.Title>
                <p>Selected Lab</p>
                <p className="text-primary">{props.selectedCRD || "-"}</p>
                <p>Running Labs</p>
                <p className="text-success">{props.runningLabs.length > 0 ? "" : "-"}</p>
                {props.runningLabs}
            </Card.Body>
        </Card>
    </Col>;
}