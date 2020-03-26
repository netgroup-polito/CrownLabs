import React from "react";
import LabTemplatesList from "../components/LabTemplatesList";
import LabInstancesList from "../components/LabInstancesList";
import {Button, Col, Row} from "react-bootstrap";
import StatusArea from "../components/StatusArea";
import "./admin.css"

class ProfessorView extends React.Component {
    state = {showForm: false};

    constructor(props) {
        super(props);
    }

    showForm = () => {
        return (
            <div className="w3-panel w3-white w3-card w3-display-container">
                <form id="add-app">
                    <p className="w3-text-blue"><b>Server:</b></p>
                    <input type="text" id="fname" name="firstname" placeholder="Your name.."/>

                    <div className="divider"/>
                    <p className="w3-text-blue"><b>ID:</b></p>
                    <input type="text"/>


                    <p className="w3-text-blue"><b>Server details:</b></p>
                    <input type="text"/>

                    <Button variant="dark" className="nav_new"
                            onClick={() => this.setState({showForm: false})}>Create</Button>

                </form>
            </div>
        );
    };

    render() {
        return <div style={{minHeight: '100vh'}}>
            <Row >
                <Col >
                    <LabTemplatesList labs={this.props.templateLabs} func={this.props.funcTemplate}
                                      start={this.props.start}/>
                    <Button variant="dark" className="text-success"
                            onClick={() => {
                            }}> Enable/Disable</Button>
                </Col>
                <Col >
                    <LabInstancesList runningLabs={this.props.instanceLabs}
                                      func={this.props.funcInstance} connect={this.props.connect}
                                      stop={this.props.stop}
                                      showStatus={this.props.showStatus}/>
                </Col>
            </Row>
            <Row>
                <Col>
                    <div className="divider">
                    <Button variant="dark" className="text-success" onClick={() => this.setState({showForm: true})}> Create Template</Button>
                    </div>
                </Col>
                <Col>
                    <div className="divider">
                    <Button variant="dark" className="text-success"> Create Instance</Button>
                    </div>
                </Col>
                <Col>
                    <div className="divider">
                    <Button variant="dark" className="text-success"> Delete Template</Button>
                    </div>
                </Col>
                <Col>
                    <div className="divider">
                    <Button variant="dark" className="text-success"> Delete Instance</Button>
                    </div>
                </Col>
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