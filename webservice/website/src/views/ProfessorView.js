import React from "react";
import LabTemplatesList from "../components/LabTemplatesList";
import LabInstancesList from "../components/LabInstancesList";
import {Button, Col, Row} from "react-bootstrap";
import StatusArea from "../components/StatusArea";
import "./admin.css"


export default function ProfessorView(props) {

    return <div style={{minHeight: '100vh'}}>
        <Row className="mt-5 p-3">
            <Col className="col-2"/>
            <Col className="col-4">
                <LabTemplatesList labs={props.templateLabs} func={props.funcTemplate}
                                  start={props.start}/>
                <Button variant="dark" className="text-success"
                        onClick={() => {}}> Enable/Disable</Button>
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
                <Button variant="dark" className="text-success"> Create Template</Button>
            <div className="divider"/>
                <Button variant="dark" className="text-success"> Create Instance</Button>
            <div className="divider"/>
                <Button variant="dark" className="text-success"> Delete Template</Button>
            <div className="divider"/>
            <Button variant="dark" className="text-success"> Delete Instance</Button>

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