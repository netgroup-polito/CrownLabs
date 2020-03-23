import {Col, Row} from "react-bootstrap";
import LabTemplatesList from "../components/LabTemplatesList";
import LabInstancesList from "../components/LabInstancesList";
import React from "react";
import StatusArea from "../components/StatusArea";

export default function StudentView(props) {
    return <div style={{minHeight: '100vh'}}>
        <Row className="mt-5 p-3">
            <Col className="col-2"/>
            <Col className="col-4">
                <LabTemplatesList labs={props.templateLabs} func={props.funcTemplate}
                                  start={props.start}/>
            </Col>
            <Col className="col-4">
                <LabInstancesList runningLabs={props.instanceLabs}
                                  func={props.funcInstance} connect={props.connect}
                                  stop={props.stop}
                                  showStatus={props.showStatus}/>
            </Col>
            <Col className="col-2"/>
        </Row>
        <Row>
            <Col className="col-2"/>
            <Col className="col-8">
                <StatusArea hidden={props.hidden} events={props.events}/>
            </Col>
            <Col className="col-2"/>
        </Row>
    </div>;
}