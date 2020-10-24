import React, { useState, useEffect } from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import List from '@material-ui/core/List';
import ListSubheader from '@material-ui/core/ListSubheader';
import PlayCircleOutlineIcon from '@material-ui/icons/PlayCircleOutline';
import Paper from '@material-ui/core/Paper';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import DeleteIcon from '@material-ui/icons/Delete';
import SortByAlphaIcon from '@material-ui/icons/SortByAlpha';
import OrderSelector from './OrderSelector';
import TextSelector from './TextSelector';
import Selector from './Selector';
import { vmTypeSelectors, ALL_VM_TYPES } from './RunningLabList';
import ListItem from './ListItem/ListItem';

/* The style for the ListItem */
const useStyles = makeStyles(theme => ({
  paper: {
    flex: 1,
    minWidth: 575,
    maxWidth: 650,
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
  const { labs, start, isStudentView, deleteLabTemplate } = props;
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
  const title = 'Available images';
  const [textMatch, setTextMatch] = useState('');
  const [orderData, setOrderData] = useState(() => {
    const prevOrderData = JSON.parse(
      localStorage.getItem(`orderData-${title}-${isStudentView}`)
    );
    return prevOrderData || { isDirUp: true, order: 'az' };
  });

  useEffect(() => {
    localStorage.setItem(
      `orderData-${title}-${isStudentView}`,
      JSON.stringify(orderData)
    );
  }, [orderData]);

  const [vmType, setVmType] = useState(() => {
    const prevVmType = JSON.parse(localStorage.getItem(`vmType`));
    return prevVmType || ALL_VM_TYPES;
  });

  useEffect(() => {
    localStorage.setItem(`vmType`, JSON.stringify(vmType));
  }, [vmType]);

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
                <Selector
                  selectors={vmTypeSelectors}
                  value={vmType}
                  setValue={setVmType}
                />
              </div>
            </ListSubheader>
          }
        >
          {labList
            .filter(({ type }) => {
              if (vmType === ALL_VM_TYPES) return true;
              return type === vmType;
            })
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
            .map(({ labName, courseName, description, type }, i) => {
              const templateFields = {
                ID: labName
              };

              const templateIcons = [
                {
                  color: 'error',
                  condition: selectedIndex === i && deleteLabTemplate,
                  onClick: e => {
                    deleteLabTemplate(labName, courseName);
                    setSelectedIndex(-1);
                    e.stopPropagation(); // avoid triggering onClick on ListItem
                  },
                  title: 'Delete template',
                  icon: DeleteIcon
                },
                {
                  condition: selectedIndex === i,
                  title: 'Start template',
                  color: 'success',
                  onClick: e => {
                    start(labName, courseName);
                    setSelectedIndex(-1);
                    e.stopPropagation(); // avoid triggering onClick of ListIstem
                  },
                  icon: PlayCircleOutlineIcon
                }
              ];

              return (
                <ListItem
                  key={labName}
                  button
                  selected={selectedIndex === i}
                  disableRipple
                  onClick={() => {
                    setSelectedIndex(i);
                  }}
                  primary={
                    description ||
                    labName.charAt(0).toUpperCase() +
                      labName.slice(1).replace(/-/g, ' ')
                  }
                  fields={templateFields}
                  icons={templateIcons}
                  type={type}
                  isSelected={selectedIndex === i}
                  showType={vmType === ALL_VM_TYPES}
                  vmTypeSelectors={vmTypeSelectors}
                />
              );
            })}
        </List>
      </ClickAwayListener>
    </Paper>
  );
}
