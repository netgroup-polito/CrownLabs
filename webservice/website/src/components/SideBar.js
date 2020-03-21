import React from 'react';
import {Nav} from 'react-bootstrap';
import {Button} from "react-bootstrap";

/**
 * Private function to camelize strings (ex. cloud-computing => Cloud Computing)
 * @param string
 * @return {void|*}
 */
function camelize(string) {
    return string.replace(/(^.)|\W+(.)/g, (match) => {
        return match.toUpperCase();
    })
}

/**
 * Function to draw the SideBar
 * @param props containing the map (course_group => List of all course templates)
 * @return the object to be drawn
 */
export default function SideBar(props) {
    /*Mapping each course_group to a NavBar with a list of all the lab templates belonging to it as Buttons*/
    const courses = Array.from(props.labs.keys()).map(courseName => {
        let labInCourse = props.labs.get(courseName).map(courseLab => {
            return <li key={courseLab}><Button variant="link" onClick={() => props.func(courseLab, courseName)}>{courseLab}</Button></li>
        });
        return <Nav.Item key={courseName} as="h5">{camelize(courseName)}<Nav.Item as="ul">{labInCourse}</Nav.Item></Nav.Item>;
    });
    return (
        <div className="p-3">
            <h4 className="mt-3">Laboratories</h4>
            <Nav role="complementary" className="mt-4">
                {courses}
            </Nav>
        </div>
    );
}