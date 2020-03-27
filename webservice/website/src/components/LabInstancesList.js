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
import StopIcon from '@material-ui/icons/Stop';
import DesktopWindowsIcon from '@material-ui/icons/DesktopWindows';
import { IconButton } from 'material-ui-core';
import VisibilityIcon from '@material-ui/icons/Visibility';
import HourglassEmptyIcon from '@material-ui/icons/HourglassEmpty';
import Tooltip from '@material-ui/core/Tooltip';


/*The style for the ListItem*/
const useStyles = makeStyles(theme => ({
    root: {
        width: '100%',
        height: '100%',
        backgroundColor: theme.palette.background.paper,
        position: 'relative',
        overflow: 'auto',
        maxHeight: '44vh',
        '& > svg': {
            margin: theme.spacing(2),
          },
      },
    listSection: {
      backgroundColor: 'inherit',
    },
    ul: {
      backgroundColor: 'inherit',
      padding: 0,
    },
    buttonGroup: {
        width: '100%',
        padding: '10px',
        position: 'fixed',
        bottom: '0%',
        left: '10%',
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
        return (
            <li key={x} className={classes.listSection}>
              <ul className={classes.ul}>
              <ListItem key={x}
                         button
                         selected={selectedIndex === index}
                         onClick={event => {
                             handleListItemClick(event, index);
                             props.func(x, null)
                         }}
        >
        <ListItemText style={{backgroundColor: color}} inset primary={x.charAt(0).toUpperCase() + x.slice(1).replace(/-/g, " ")}/>        
            {selectedIndex==index ?  
            <Tooltip title="Stop VM">
                <IconButton style={{color: "red"}} button="true" onClick={() => {props.stop(); setSelectedIndex(-1)}}>
                <StopIcon fontSize="large"/>              
            </IconButton>
            </Tooltip> : null}
            {(selectedIndex==index && status ===1) ? 
            <Tooltip title="Connect VM">
                <IconButton style={{color: "black"}} button="true" onClick={() =>{props.connect(); setSelectedIndex(-1);}}>
                <DesktopWindowsIcon fontSize="large"/>
            </IconButton>
            </Tooltip> : null}
            {(selectedIndex==index && status ===0) ? 
            <Tooltip title="Loading VM">
                <IconButton style={{color: "orange"}}>
                <HourglassEmptyIcon fontSize="large"/>
            </IconButton>
            </Tooltip> : null}
        </ListItem>
        </ul>
        </li>
    )});

    return (
        <div className="w3-panel w3-white w3-card w3-display-container">
                <List component="nav" subheader={
                    <ListSubheader style={{fontSize:"30px"}} component="div" id="nested-list-subheader">
                        Running Laboratories
                    </ListSubheader>
                }>
                </List>
                <List className={classes.root} subheader={<li />}>
                {courses}
                </List>
                {/* maybe it will be deleted */}
                <Tooltip title="Show status">
                <IconButton style={{color: "green"}} button="true" onClick={props.showStatus}>
                <VisibilityIcon fontSize="large"/>
                </IconButton>
                </Tooltip>
        </div>
    );
}