import React from 'react';
import {Nav} from 'react-bootstrap';
import {Button} from "react-bootstrap";

export default function SideBar(props) {
    const courses = Array.from(props.labs.keys()).map(courseName => {
        let camelizedName = courseName.replace(/(^.)|\W+(.)/g, (match) => {
            return match.toUpperCase();
        });
        let labInCourse = props.labs.get(courseName).map(courseLab => {
            return <li key={courseLab}><Button variant="link" onClick={() => props.func(courseLab, courseName)}>{courseLab}</Button>
            </li>
        });
        return <Nav.Item key={courseName} as="h6">{camelizedName}<Nav.Item as="ul">{labInCourse}</Nav.Item></Nav.Item>
    });
    return (
        <div className="p-3">
            <h5 className="mt-3">Laboratories</h5>
            <Nav role="complementary" className="mt-4">
                {courses}
            </Nav>
        </div>
    );
}
