import React, { useState, useEffect } from 'react';
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
import SortByAlphaIcon from '@material-ui/icons/SortByAlpha';
import OrderSelector from './OrderSelector';
import TextSelector from './TextSelector';
import { vmTypeSelectors } from './RunningLabList';
/* The style for the ListItem */
const useStyles = makeStyles(theme => ({
  paper: {
    flex: 1,
    minWidth: 450,
    maxWidth: 600,
    padding: 10,
    margin: 10,
    maxHeight: 350
  },
  list: {
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
  },
  listSubHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    fontSize: '30px',
    padding: theme.spacing(0, 1)
  },
  titleActions: {
    display: 'flex',
    justifyContent: 'end',
    alignItems: 'center'
  },
  startIcon: {
    color: theme.palette.success.main
  },
  deleteIcon: {
    color: theme.palette.error.main
  },
  templateType: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    color: theme.palette.info.main,
    width: theme.spacing(7)
  }
}));

const selectors = [{ text: 'Name', icon: <SortByAlphaIcon />, value: 'az' }];

/**
 * Function to draw a list of available lab templates
 * @param props contains all the functions to be associated with the components (click => select new template)
 * @return the component to be drawn
 */
export default function LabTemplatesList(props) {
  const classes = useStyles();
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const { labs, start, isAdmin, deleteLabTemplate } = props;
  const courseNames = Array.from(labs.keys());
  const labList = courseNames.reduce(
    (acc, courseName) => [
      ...acc,
      ...labs.get(courseName).map(({ labName, description, type }) => ({
        labName,
        courseName,
        description,
        type
      }))
    ],
    []
  );
  const title = 'Available Labs';
  const [textMatch, setTextMatch] = useState('');
  const [orderData, setOrderData] = useState(() => {
    const prevOrderData = JSON.parse(
      localStorage.getItem(`orderData-${title}`)
    );
    return prevOrderData || { isDirUp: true, order: 'az' };
  });

  useEffect(() => {
    localStorage.setItem(`orderData-${title}`, JSON.stringify(orderData));
  }, [orderData]);

  return (
    <Paper elevation={6} className={classes.paper}>
      <ClickAwayListener
        onClickAway={() => {
          setSelectedIndex(-1);
        }}
      >
        <List
          className={classes.list}
          subheader={
            <ListSubheader className={classes.listSubHeader}>
              <div>{title}</div>
              <div className={classes.titleActions}>
                <OrderSelector
                  selectors={selectors}
                  setOrderData={setOrderData}
                  orderData={orderData}
                />
                <TextSelector value={textMatch} setValue={setTextMatch} />
              </div>
            </ListSubheader>
          }
        >
          {labList
            .filter(({ labName, description }) => {
              if (textMatch !== '') {
                const textMatchLower = textMatch.toLowerCase();
                return (
                  labName.toLowerCase().includes(textMatchLower) ||
                  (description &&
                    description.toLowerCase().includes(textMatchLower))
                );
              }
              return true;
            })
            .sort((a, b) => {
              const { isDirUp } = orderData;
              const sortResult =
                a.description && b.description
                  ? a.description.localeCompare(b.description)
                  : a.labName.localeCompare(b.labName);
              return isDirUp ? sortResult : -sortResult;
            })
            .map(({ labName, courseName, description, type }, i) => (
              <ListItem
                key={labName}
                button
                selected={selectedIndex === i}
                disableRipple={isAdmin}
                onClick={() => {
                  setSelectedIndex(i);
                }}
              >
                <Tooltip title="Select it">
                  <>
                    <div className={classes.templateType}>
                      {vmTypeSelectors.find(sel => sel.value === type).icon}
                    </div>
                    <ListItemText
                      primary={
                        description ||
                        labName.charAt(0).toUpperCase() +
                          labName.slice(1).replace(/-/g, ' ')
                      }
                      secondary={
                        <>
                          <b>ID: </b>
                          {labName}
                          <br />
                        </>
                      }
                    />
                  </>
                </Tooltip>
                {selectedIndex === i && deleteLabTemplate ? (
                  <Tooltip title="Delete template">
                    <IconButton
                      className={classes.deleteIcon}
                      onClick={e => {
                        deleteLabTemplate(labName, courseName);
                        setSelectedIndex(-1);
                        e.stopPropagation(); // avoid triggering onClick on ListItem
                      }}
                    >
                      <DeleteIcon fontSize="large" />
                    </IconButton>
                  </Tooltip>
                ) : null}
                {selectedIndex === i && start ? (
                  <Tooltip title="Create VM">
                    <IconButton
                      className={classes.startIcon}
                      key={labName}
                      variant="dark"
                      onClick={e => {
                        start(labName, courseName);
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
