import React from 'react';
import {makeStyles} from 'material-ui-core/styles';
import List from 'material-ui-core/List';
import ListItem from 'material-ui-core/ListItem';
import ListItemText from 'material-ui-core/ListItemText';
import {Button} from "react-bootstrap";
import ListSubheader from "material-ui-core/ListSubheader";

/*The style for the ListItem*/
const useStyles = makeStyles(theme => ({
    root: {
        width: '100%',
        backgroundColor: theme.palette.background.paper,
    },
}));

/**
 * Function to draw a list of available lab templates
 * @param props contains all the functions to be associated with the components (click => select new template)
 * @return the component to be drawn
 */
export default function LabTemplatesList(props) {
    const classes = useStyles();
    const [selectedIndex, setSelectedIndex] = React.useState(-1);

    const handleListItemClick = (event, index) => {
        setSelectedIndex(index);
    };

    /*Parse the template maps and foreach one draw a ListItem with the right events associated
    * the index are linearized to make the click events take the right row*/
    const courses = Array.from(props.labs.keys()).map((courseName, index) => {
        let offset = index * (props.labs.get(courseName).length + 1);
        return props.labs.get(courseName).map((courseLab, index2) => {
            let finalIndex = offset + index2;
            return <ListItem key={courseLab}
                             button
                             selected={selectedIndex === finalIndex}
                             onClick={event => {
                                 handleListItemClick(event, finalIndex);
                                 props.func(courseLab, courseName)
                             }}
            >
                <ListItemText>{courseLab}</ListItemText>
            </ListItem>;
        });
    });

    return (
        <div className="text-center">
            <div className={classes.root}>
                <List component="nav" subheader={
                    <ListSubheader component="div" id="nested-list-subheader">
                        Available Laboratories
                    </ListSubheader>
                }>
                    {courses}
                </List>
            </div>
            <Button variant="dark" className="text-success"
                    onClick={props.start}>Start</Button>
        </div>
    );
}