import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';
import Paper from '@material-ui/core/Paper';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import DeleteIcon from '@material-ui/icons/Delete';

/* The style for the ListItem */
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
  const { labs } = props;

  const courseNames = Array.from(labs.keys());
  const labList = courseNames.reduce(
    (acc, courseName) => [
      ...acc,
      ...labs.get(courseName).map(labName => ({ labName, courseName }))
    ],
    []
  );

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
          {labList.map(({ labName, courseName }, i) => (
            <ListItem
              key={labName}
              button
              selected={selectedIndex === i}
              disableRipple={props.isAdmin}
              onClick={() => {
                setSelectedIndex(i);
                props.func(labName, courseName);
              }}
            >
              <Tooltip title="Select it">
                <ListItemText
                  inset
                  primary={
                    labName.charAt(0).toUpperCase() +
                    labName.slice(1).replace(/-/g, ' ')
                  }
                />
              </Tooltip>
              {selectedIndex === i && props.delete ? (
                <Tooltip title="Delete template">
                  <IconButton
                    style={{ color: 'red' }}
                    button="true"
                    onClick={e => {
                      props.delete();
                      setSelectedIndex(-1);
                      e.stopPropagation(); // avoid triggering onClick on ListItem
                    }}
                  >
                    <DeleteIcon fontSize="large" />
                  </IconButton>
                </Tooltip>
              ) : null}
              {selectedIndex === i && props.start ? (
                <Tooltip title="Create VM">
                  <IconButton
                    key={labName}
                    variant="dark"
                    style={{ color: 'green' }}
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
          ))}
        </List>
      </ClickAwayListener>
    </Paper>
  );
}
