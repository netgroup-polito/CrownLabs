import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import '../views/admin.css';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import { IconButton } from '@material-ui/core';
import Tooltip from '@material-ui/core/Tooltip';
import Paper from '@material-ui/core/Paper';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';

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
      margin: theme.spacing(2)
    }
  },
  listSection: {
    backgroundColor: 'inherit'
  },
  ul: {
    backgroundColor: 'inherit',
    padding: 0
  }
}));

/**
 * Function to draw a list of available lab templates
 * @param props contains all the functions to be associated with the components (click => select new template)
 * @return the component to be drawn
 */
export default function LabTemplatesList(props) {
  const classes = useStyles();
  const [selectedIndex, setSelectedIndex] = React.useState(-1);

  const courses = Array.from(props.labs.keys()).map((courseName, index) => {
    let offset = index * (props.labs.get(courseName).length + 1);
    return (
      <li key={courseName} className={classes.listSection}>
        <ul className={classes.ul}>
          {props.labs.get(courseName).map((courseLab, index2) => {
            let finalIndex = offset + index2;
            return (
              <ListItem
                key={courseLab}
                button
                selected={selectedIndex === finalIndex}
                disableRipple={props.isAdmin}
                onClick={() => {
                  if (!props.isAdmin) {
                    console.log('clicked');
                    setSelectedIndex(finalIndex);
                    props.func(courseLab, courseName);
                  }
                }}
              >
                <Tooltip title="Select it">
                  <ListItemText
                    inset
                    primary={
                      courseLab.charAt(0).toUpperCase() +
                      courseLab.slice(1).replace(/-/g, ' ')
                    }
                  />
                </Tooltip>
                {selectedIndex === finalIndex && props.start ? (
                  <Tooltip title="Create VM">
                    <IconButton
                      key={courseLab}
                      variant="dark"
                      className="text-success"
                      button="true"
                      onClick={e => {
                        props.start();
                        setSelectedIndex(-1);
                        e.stopPropagation(); // avoid triggering onClick of ListIstem
                      }}
                    >
                      <PlayCircleOutlineIcon fontSize="large" />
                    </IconButton>
                  </Tooltip>
                ) : null}
              </ListItem>
            );
          })}
        </ul>
      </li>
    );
  });

  return (
    <Paper
      elevation={6}
      style={{
        flex: 1,
        minWidth: 450,
        maxWidth: 600,
        padding: 10,
        margin: 10,
        maxHeight: 350
      }}
    >
      <ClickAwayListener
        onClickAway={() => {
          setSelectedIndex(-1);
        }}
      >
        <List
          className={classes.root}
          subheader={
            <ListSubheader style={{ fontSize: '30px' }}>
              Available Laboratories
            </ListSubheader>
          }
        >
          {courses}
        </List>
      </ClickAwayListener>
    </Paper>
  );
}
