import React, { useState, useEffect } from 'react';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import ListSubheader from '@material-ui/core/ListSubheader';
import CancelOutlinedIcon from '@material-ui/icons/CancelOutlined';
import OpenInBrowserIcon from '@material-ui/icons/OpenInBrowser';
import IconButton from '@material-ui/core/IconButton';
import HourglassEmptyIcon from '@material-ui/icons/HourglassEmpty';
import Tooltip from '@material-ui/core/Tooltip';
import { makeStyles } from '@material-ui/core/styles';
import ClickAwayListener from '@material-ui/core/ClickAwayListener';
import SortByAlphaIcon from '@material-ui/icons/SortByAlpha';
import AccessTimeIcon from '@material-ui/icons/AccessTime';
import UserIcon from '@material-ui/icons/Person';
import { utc } from 'moment';
import OrderSelector from './OrderSelector';
import TextSelector from './TextSelector';

const useStyles = makeStyles(theme => ({
  root: {
    width: '100%',
    height: '100%',
    maxHeight: '70vh',
    backgroundColor: theme.palette.background.paper,
    position: 'relative',
    overflow: 'auto',
    '& > svg': {
      margin: theme.spacing(2)
    }
  },
  titlebar: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    fontSize: '30px'
  },
  buttonGroup: {
    width: '100%',
    padding: '10px',
    position: 'fixed',
    bottom: '0%',
    left: '10%'
  },
  labColorTag: {
    width: 40,
    height: '100%',
    borderRadius: 5,
    margin: '0 10px'
  },
  activeLab: {
    backgroundColor: theme.palette.success.main
  },
  loadingLab: {
    backgroundColor: theme.palette.warning.light
  },
  errorLab: {
    backgroundColor: theme.palette.error.light
  },
  rotating: {
    animation: 'rotate 1.5s ease-in-out infinite'
  },
  stopIcon: {
    color: theme.palette.error.main
  },
  launchIcon: {
    color: theme.palette.info.main
  },
  loadIcon: {
    color: theme.palette.warning.light
  },
  listTitle: {
    fontSize: 30
  },
  titleActions: {
    display: 'flex',
    justifyContent: 'end',
    alignItems: 'center'
  }
}));

const studentSelectors = [
  { text: 'Name', icon: <SortByAlphaIcon />, value: 'az' },
  { text: 'Created', icon: <AccessTimeIcon />, value: 'time' }
];
const adminSelectors = [
  ...studentSelectors,
  {
    text: 'User',
    icon: <UserIcon />,
    value: 'user'
  }
];

const getLabCodeFromName = name => name.slice(name.length - 4);

const RunningLabList = props => {
  const { labList, stop, connect, title, isStudentView } = props;

  const classes = useStyles();
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [textMatch, setTextMatch] = useState('');
  const [orderData, setOrderData] = useState(() => {
    const prevOrderData = JSON.parse(
      localStorage.getItem(`orderData-${title}-${isStudentView}`)
    );
    return prevOrderData || { isDirUp: true, order: 'az' };
  });

  useEffect(() => {
    localStorage.setItem(`orderData-${title}`, JSON.stringify(orderData));
  }, [orderData]);

  return (
    <ClickAwayListener
      onClickAway={() => {
        setSelectedIndex(-1);
      }}
    >
      <List
        className={classes.root}
        subheader={
          <ListSubheader className={classes.titlebar}>
            {title}
            <div className={classes.titleActions}>
              <TextSelector value={textMatch} setValue={setTextMatch} />
              <OrderSelector
                selectors={isStudentView ? studentSelectors : adminSelectors}
                setOrderData={setOrderData}
                orderData={orderData}
              />
            </div>
          </ListSubheader>
        }
      >
        {labList
          .filter(({ labName, ip, description, studentId }) => {
            if (textMatch !== '') {
              const labCode = getLabCodeFromName(labName);
              const textMatchLower = textMatch.toLowerCase();
              // not using regex but lowercase and include since it sohuld be faster, could be changed easily
              return (
                (description &&
                  description.toLowerCase().includes(textMatchLower)) ||
                labCode.includes(textMatchLower) ||
                ip.includes(textMatchLower) ||
                (studentId &&
                  studentId.toLowerCase().includes(textMatchLower)) ||
                labName.toLowerCase().includes(textMatchLower)
              );
            }
            return true;
          })
          .sort((a, b) => {
            let sortResult;
            const { order, isDirUp } = orderData;
            if (order === 'time')
              sortResult = utc(b.creationTime).diff(a.creationTime, 's');
            else if (order === 'user') {
              sortResult =
                a.studentId && b.studentId
                  ? a.studentId.localeCompare(b.studentId)
                  : a.labName.localeCompare(b.labName);
            } else
              sortResult =
                a.description && b.description
                  ? a.description.localeCompare(b.description)
                  : a.labName.localeCompare(b.labName);

            return isDirUp ? sortResult : -sortResult;
          })
          .map(
            (
              { labName, status, ip, creationTime, description, studentId },
              i
            ) => {
              const labCode = getLabCodeFromName(labName);
              const statusClassName =
                status === 0
                  ? classes.loadingLab
                  : status === 1
                  ? classes.activeLab
                  : classes.errorLab;

              return (
                <ListItem
                  key={labName}
                  button
                  selected={selectedIndex === i}
                  onClick={() => {
                    setSelectedIndex(i);
                  }}
                >
                  <div className={`${classes.labColorTag} ${statusClassName}`}>
                    &nbsp;
                  </div>
                  <ListItemText
                    primary={
                      description
                        ? `${description} - ${labCode}`
                        : `${labName.charAt(0).toUpperCase()}${labName
                            .slice(1)
                            .replace(/-/g, ' ')}`
                    }
                    secondary={
                      <>
                        {studentId && (
                          <div>
                            <b>User: </b>
                            {studentId}
                          </div>
                        )}
                        <div>
                          <b>Created: </b>
                          {utc(creationTime).format('DD/MM/YY HH:MM')}
                        </div>
                        <div>
                          <b>IP: </b>
                          {ip}
                        </div>
                      </>
                    }
                  />
                  {selectedIndex === i && stop ? (
                    <Tooltip title="Stop VM">
                      <IconButton
                        className={classes.stopIcon}
                        button="true"
                        onClick={e => {
                          stop(labName);
                          setSelectedIndex(-1);
                          e.stopPropagation(); // avoid triggering onClick on ListItem
                        }}
                      >
                        <CancelOutlinedIcon fontSize="large" />
                      </IconButton>
                    </Tooltip>
                  ) : null}
                  {selectedIndex === i && status === 1 ? (
                    <Tooltip title="Connect VM">
                      <IconButton
                        className={classes.launchIcon}
                        button="true"
                        onClick={e => {
                          connect(labName);
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
                      <IconButton className={classes.loadIcon}>
                        <HourglassEmptyIcon
                          className={classes.rotating}
                          fontSize="large"
                        />
                      </IconButton>
                    </Tooltip>
                  ) : null}
                </ListItem>
              );
            }
          )}
      </List>
    </ClickAwayListener>
  );
};

export default RunningLabList;
