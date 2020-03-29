import React from 'react';
import {makeStyles} from 'material-ui-core/styles';
import List from 'material-ui-core/List';
import ListItem from 'material-ui-core/ListItem';
import ListItemText from 'material-ui-core/ListItemText';
import ListSubheader from "material-ui-core/ListSubheader";
import PlayArrowIcon from '@material-ui/icons/PlayArrow';
import { IconButton } from 'material-ui-core';
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


const courses = Array.from(props.labs.keys()).map((courseName, index) => {
    let offset = index * (props.labs.get(courseName).length + 1);
return (
    <li key={courseName} className={classes.listSection}>
      <ul className={classes.ul}>
        { props.labs.get(courseName).map((courseLab, index2) => {
                    let finalIndex = offset + index2;
                    return (
                        <ListItem key={courseLab}
                        button
                        selected={selectedIndex === finalIndex}
                        onClick={event => {
                            handleListItemClick(event, finalIndex);
                            props.func(courseLab, courseName);
                       }}
                            >
                        <Tooltip title="Select it">
                        <ListItemText inset primary={courseLab.charAt(0).toUpperCase() + courseLab.slice(1).replace(/-/g, " ")}/>
                        </Tooltip>
                       {selectedIndex==finalIndex ?  <Tooltip title="Create VM">
                       <IconButton key={courseLab} variant="dark" className="text-success"
                                    button="true"
                                    onClick={() => {
                                      if(selectedIndex==finalIndex) {
                                        props.start();
                                        setSelectedIndex(-1)
                                      }
                                    }}
                    >
                      <PlayArrowIcon fontSize="large" />
                      </IconButton>
                       </Tooltip> : null}
                    </ListItem>
                )})}
      </ul>
    </li>
  )});

  return (
        <div className="text-center">
             <List component="nav" subheader={
                    <ListSubheader style={{fontSize:"30px"}} component="div" id="nested-list-subheader">
                        Available Laboratories
                    </ListSubheader>
                }>
                 </List>
            <div className={classes.root}>
                <List className={classes.root} subheader={<li />}>
                    {courses}
                </List>
            </div>
        </div>
    );
}

