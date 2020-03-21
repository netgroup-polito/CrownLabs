import {Button, ButtonGroup, Card, Col, Row} from "react-bootstrap";
import React from "react";

/**
 * Function to render the central part of the UserView
 * @param props the functions to be called when buttons are clicked
 * @return the component to be drawn
 */
export default function CentralView(props) {
    return <Row className="my-5">
        <Col className="col-12">
            <Card className="text-center headerstyle">
                <Card.Body>
                    <Card.Text as="h6">Status information</Card.Text>
                    <textarea readOnly align="center" className="textareastyle"
                              value={props.events}/>
                </Card.Body>
                <Card.Footer className="headerstyle">
                    <ButtonGroup aria-label="Basic example">
                        <Button variant="dark" className="text-success"
                                onClick={props.start}>Start</Button>
                        <Button variant="dark" className="text-danger"
                                onClick={props.stop}>Stop</Button>
                        <Button variant="dark" onClick={props.connect}>Connect</Button>
                    </ButtonGroup>
                </Card.Footer>
            </Card>
        </Col>
    </Row>;
}