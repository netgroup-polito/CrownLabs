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
import { utc } from 'moment';
import OrderSelector from './OrderSelector';

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
    color: theme.palette.success.dark
  },
  loadIcon: {
    color: theme.palette.warning.light
  },
  listTitle: {
    fontSize: 30
  }
}));

const RunningLabList = props => {
  const { labList, selectInstance, stop, connect, title } = props;
  const classes = useStyles();
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const selectors = [
    { text: 'Name', icon: <SortByAlphaIcon />, value: 'az' },
    { text: 'Created', icon: <AccessTimeIcon />, value: 'time' }
  ];
  const [orderData, setOrderData] = useState(() => {
    const prevOrderData = JSON.parse(localStorage.getItem('orderData'));
    return prevOrderData || { isDirUp: true, order: 'az' };
  });

  useEffect(() => {
    localStorage.setItem('orderData', JSON.stringify(orderData));
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
            <OrderSelector
              selectors={selectors}
              setOrderData={setOrderData}
              orderData={orderData}
            />
          </ListSubheader>
        }
      >
        {labList
          .sort((a, b) => {
            let sortResult;
            const { order, isDirUp } = orderData;
            if (order === 'time') {
              sortResult = utc(b.creationTime).diff(a.creationTime, 's');
            } else sortResult = a.labName.localeCompare(b.labName);
            return isDirUp ? sortResult : -sortResult;
          })
          .map(({ labName, status, ip, creationTime, description }, i) => {
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
                  selectInstance(labName, null);
                }}
              >
                <div className={`${classes.labColorTag} ${statusClassName}`}>
                  &nbsp;
                </div>
                <ListItemText
                  color="primary"
                  primary={
                    description ||
                    `${labName.charAt(0).toUpperCase()}${labName
                      .slice(1)
                      .replace(/-/g, ' ')}`
                  }
                  secondary={
                    <>
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
                        stop();
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
                        connect();
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
          })}
      </List>
    </ClickAwayListener>
  );
};

export default RunningLabList;
