import {Card, Col, Row} from "react-bootstrap";
import React from "react";

/**
 * Function to render the Status Area
 * @param props contains the parameter whether to check if this area has to be shown or not
 * @return the component to be drawn
 */
export default function StatusArea(props) {
    return <div style={{minHeight: '100vh'}}>
        {props.hidden ? <div/> : <Row className="my-5">
            <Col className="col-12">
                <Card className="text-center headerstyle">
                    <Card.Body>
                        <Card.Text as="h6">Status information</Card.Text>
                        <textarea readOnly align="center" className="textareastyle"
                                  value={props.events}/>
                    </Card.Body>
                </Card>
            </Col>
        </Row>}
    </div>;
}