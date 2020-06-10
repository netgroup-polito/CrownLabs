import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import CancelOutlinedIcon from '@material-ui/icons/CancelOutlined';
import OpenInBrowserIcon from '@material-ui/icons/OpenInBrowser';
import IconButton from '@material-ui/core/IconButton';
import HourglassEmptyIcon from '@material-ui/icons/HourglassEmpty';
import Tooltip from '@material-ui/core/Tooltip';
import Paper from '@material-ui/core/Paper';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';

/* The style for the ListItem */
const useStyles = makeStyles(theme => ({
  root: {
    width: '100%',
    height: '100%',
    backgroundColor: theme.palette.background.paper,
    position: 'relative',
    overflow: 'auto',
    maxHeight: '70vh',
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
  },
  buttonGroup: {
    width: '100%',
    padding: '10px',
    position: 'fixed',
    bottom: '0%',
    left: '10%'
  },
  rotating: {
    animation: 'rotate 1.5s ease-in-out infinite'
  }
}));

/**
 * Function to draw a list of running lab instances
 * @param props contains all the function to be associated with the components (buttons click, etc.)
 * @return The component to be drawn
 */
export default function LabInstancesList(props) {
  const classes = useStyles();
  const [selectedIndex, setSelectedIndex] = React.useState(-1);

  /* Parsing the instances array and draw for each one a list item with the right coloration, according to its status */
  const { runningLabs } = props;

  return (
    <Paper
      elevation={6}
      style={{
        flex: 1,
        minWidth: 450,
        maxWidth: 600,
        padding: 10,
        margin: 10,
        maxHeight: '70vh'
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
              Running Laboratories
            </ListSubheader>
          }
        >
          {Array.from(runningLabs.keys()).map((x, index) => {
            const status = props.runningLabs.get(x)
              ? props.runningLabs.get(x).status
              : -1;
            const color =
              status === 0 ? 'orange' : status === 1 ? 'lime' : 'red';
            return (
              <li key={x} className={classes.listSection}>
                <ul className={classes.ul}>
                  <ListItem
                    key={x}
                    button
                    selected={selectedIndex === index}
                    onClick={() => {
                      setSelectedIndex(index);
                      props.func(x, null);
                    }}
                  >
                    <ListItemText
                      style={{ backgroundColor: color, color: 'black' }}
                      inset
                      primary={
                        x.charAt(0).toUpperCase() +
                        x.slice(1).replace(/-/g, ' ')
                      }
                    />
                    {selectedIndex === index && props.stop ? (
                      <Tooltip title="Stop VM">
                        <IconButton
                          style={{ color: 'red' }}
                          button="true"
                          onClick={e => {
                            props.stop();
                            setSelectedIndex(-1);
                            e.stopPropagation(); // avoid triggering onClick on ListItem
                          }}
                        >
                          <CancelOutlinedIcon fontSize="large" />
                        </IconButton>
                      </Tooltip>
                    ) : null}
                    {selectedIndex === index && status === 1 ? (
                      <Tooltip title="Connect VM">
                        <IconButton
                          style={{ color: 'black' }}
                          button="true"
                          onClick={e => {
                            props.connect();
                            setSelectedIndex(-1);
                            e.stopPropagation(); // avoid triggering onClick on ListItem
                          }}
                        >
                          <OpenInBrowserIcon fontSize="large" />
                        </IconButton>
                      </Tooltip>
                    ) : null}
                    {status === 0 ? (
                      <Tooltip title="Loading VM">
                        <IconButton style={{ color: 'orange' }}>
                          <HourglassEmptyIcon
                            className={classes.rotating}
                            fontSize="large"
                          />
                        </IconButton>
                      </Tooltip>
                    ) : null}
                  </ListItem>
                </ul>
              </li>
            );
          })}
        </List>
      </ClickAwayListener>
    </Paper>
  );
}
