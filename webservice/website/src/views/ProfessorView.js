import React from "react";
import LabTemplatesList from "../components/LabTemplatesList";
import LabInstancesList from "../components/LabInstancesList";
import {Button, Col, Row} from "react-bootstrap";
import { MDBContainer, MDBScrollbar } from "mdbreact";
import StatusArea from "../components/StatusArea";
import "./admin.css"

class ProfessorView extends React.Component{
    constructor(props) {
        super(props);
    }

    outerContainerStyle = { width: "800px", height: "400px" };

    state = {showForm: false};
    showForm = () => {
        return (
            <div>
                <form id="add-app">
                    <Col className="center-col">

                            <label> Server: </label>
                            <input type="text" id="fname" name="firstname" placeholder="Your name.."/>

                            <div className="divider"/>
                            <label> id : </label>
                            <input type="text"/>


                            <label>Server details : </label>
                            <input type="text"/>

                            <button>Create</button>


                    </Col>
                </form>
            </div>
        );
    };

    render() {
        return <div style={{minHeight: '100vh'}}>
            <Row className="mt-5 p-3">
                <Col className="col-2"/>
                <Col className="col-4">
                    <LabTemplatesList labs={this.props.templateLabs} func={this.props.funcTemplate}
                                      start={this.props.start}/>
                    <Button variant="dark" className="text-success"
                            onClick={() => {}}> Enable/Disable</Button>
                </Col>
                <Col className="col-4">
                    <LabInstancesList runningLabs={this.props.instanceLabs}
                                      func={this.props.funcInstance} connect={this.props.connect}
                                      stop={this.props.stop}
                                      showStatus={this.props.showStatus}/>
                </Col>
                <Col className="col-2"/>
            </Row>
            <Row>
                <Button variant="dark" className="text-success" onClick={() => this.setState({showForm: true})}> Create Template</Button>
                <div className="divider"/>
                <Button variant="dark" className="text-success"> Create Instance</Button>
                <div className="divider"/>
                <Button variant="dark" className="text-success"> Delete Template</Button>
                <div className="divider"/>
                <Button variant="dark" className="text-success"> Delete Instance</Button>
            </Row>

            {this.state.showForm ? this.showForm() : null}

            <Row>
                <Col className="col-2"/>
                <Col className="col-8">
                    <StatusArea hidden={this.props.hidden} events={this.props.events}/>
                </Col>
                <Col className="col-2"/>
            </Row>



        </div>;
    }


}

export default ProfessorView