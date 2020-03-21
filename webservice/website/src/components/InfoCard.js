import {Button, Card} from "react-bootstrap";
import React from "react";

/**
 * Function to draw the InfoCard
 * @param props containing the selected CRD and all the running ones
 * @return the object to be drawn
 */
export default function InfoCard(props) {
    /*Retrieving instance labs to be drawn in the right bar and foreach one draw a button*/
    const toDraw = Array.from(props.runningLabs.keys()).map(x => {
        let status = props.runningLabs.get(x) ? props.runningLabs.get(x).status : -1;
        let color = status === 0 ? 'orange' : status === 1 ? 'green' : 'red';
        return <Button key={x} variant="link" style={{color: color}}
                       onClick={() => props.func(x, null)}>{x}</Button>;
    });
    return <Card className="my-5 p-2 text-center text-dark" border="dark"
              style={{backgroundColor: 'transparent'}}>
            <Card.Body>
                <Card.Title className="p-2">Details</Card.Title>
                <p>Selected Lab</p>
                <p className="text-primary">{props.selectedCRD || "-"}</p>
                <p>Running Labs</p>
                <p className="text-success">{props.runningLabs.size > 0 ? "" : "-"}</p>
                {toDraw}
            </Card.Body>
        </Card>;
}