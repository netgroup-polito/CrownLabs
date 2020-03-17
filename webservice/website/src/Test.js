import React, { Component } from 'react';
import { Container, Card, Button, Row, Col } from 'react-bootstrap';
import { makeTest } from './k8sApi/index';

class Test extends Component {
    constructor(props) {
        super(props);
        this.state = {
            msg: ""
        };
        this.test = this.test.bind(this);
    }

    test() {
        makeTest().then(response => {
            console.log(response);
            this.setState({msg: response});
        }).catch(error => {console.log(error)});
    }

    render() {
        return(
            <div>
                <Container>
                    <Row className="my-5">
                        <Col className="col-4"></Col>
                        <Col className="col-4">
                            <Card className="my-5 p-2" bg="light">
                                <Card.Body>
                                    <Card.Text as="h5">{this.state.msg}</Card.Text>
                                    <Button className="btn-block mt-5" variant="primary" onClick={this.test}>Make test</Button>
                                </Card.Body>
                            </Card>
                        </Col>
                        <Col className="col-4"></Col>
                    </Row>
                </Container>
            </div>
        );
    }
}

export default Test;