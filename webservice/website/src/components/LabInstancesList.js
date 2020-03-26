import React from 'react';
import {makeStyles} from 'material-ui-core/styles';
import List from 'material-ui-core/List';
import ListItem from 'material-ui-core/ListItem';
import ListItemText from 'material-ui-core/ListItemText';
import {Button, ButtonGroup} from "react-bootstrap";
import ListSubheader from "material-ui-core/ListSubheader";
import ListItemIcon from "material-ui-core/ListItemIcon";
import Icon from "material-ui-core/Icon";
import "../views/admin.css"

/*The style for the ListItem*/
const useStyles = makeStyles(theme => ({
    root: {
        width: '100%',
        backgroundColor: theme.palette.background.paper,
    },
}));

/**
 * Function to draw a list of running lab instances
 * @param props contains all the function to be associated with the components (buttons click, etc.)
 * @return The component to be drawn
 */
export default function LabInstancesList(props) {
    const classes = useStyles();
    const [selectedIndex, setSelectedIndex] = React.useState(-1);

    const handleListItemClick = (event, index) => {
        setSelectedIndex(index);
    };

    /*Parsing the instances array and draw for each one a list item with the right coloration, according to its status*/
    const courses = Array.from(props.runningLabs.keys()).map((x, index) => {
        let status = props.runningLabs.get(x) ? props.runningLabs.get(x).status : -1;
        let color = status === 0 ? 'orange' : status === 1 ? 'green' : 'red';
        return <ListItem key={x}
                         button
                         selected={selectedIndex === index}
                         onClick={event => {
                             handleListItemClick(event, index);
                             props.func(x, null)
                         }}
        >
            <ListItemText>{x}</ListItemText>
            <ListItemIcon>
                <Icon style={{backgroundColor: color}}/>
            </ListItemIcon>
        </ListItem>;
    });

    return (
        <div className="w3-panel w3-white w3-card w3-display-container">
            <div className={classes.root}>
                <List component="nav" subheader={
                    <ListSubheader component="div" id="nested-list-subheader">
                        Running Laboratories
                    </ListSubheader>
                }>
                    {courses}
                </List>
            </div>
            <ButtonGroup aria-label="Basic example">
                <Button variant="dark" className="nav_new" onClick={() =>{props.connect(); setSelectedIndex(-1);}}>Connect</Button>
                <Button variant="dark" className="nav_new"
                        onClick={() => {props.stop(); setSelectedIndex(-1)}}>Stop</Button>
                <Button variant="dark" className="nav_new"
                        onClick={props.showStatus}>Show status</Button>
            </ButtonGroup>
        </div>
    );
}