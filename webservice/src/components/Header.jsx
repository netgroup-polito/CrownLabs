import React from 'react';
import FolderSharedIcon from '@material-ui/icons/FolderShared';
import makeStyles from '@material-ui/core/styles/makeStyles';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import Typography from '@material-ui/core/Typography';
import Light from '@material-ui/icons/Brightness7';
import Dark from '@material-ui/icons/Brightness4';
import ProfIcon from '@material-ui/icons/School';
import HomeIcon from '@material-ui/icons/Home';
import LogoutIcon from '@material-ui/icons/MeetingRoom';
import { Tooltip } from '@material-ui/core';
import * as CrownLogo from '../assets/crown.png';
/**
 * Function to draw the page header
 * @param props the property to check whether it is logged or not, to draw the apposite component
 * @return the component to be drawn
 */

const useStyles = makeStyles(theme => ({
  header: {
    flexGrow: 1
  },
  menuButton: {
    marginRight: theme.spacing(2)
  },
  title: {
    fontWeight: 600,
    fontSize: '1.5rem',
    flexGrow: 1
  }
}));

export default function Header(props) {
  const classes = useStyles();
  const {
    switchAdminView,
    isStudentView,
    name,
    logout,
    setIsLightTheme,
    isLightTheme,
    isAlsoAdmin
  } = props;

  return (
    <div className={classes.header}>
      <AppBar position="static" color="primary">
        <Toolbar>
          <img
            src={CrownLogo}
            style={{ marginRight: '20px', height: '40px' }}
            alt="Crownlabs"
          />
          <Typography variant="h6" className={classes.title}>
            Crownlabs
          </Typography>
          <Typography
            style={{
              textAlign: 'right',
              marginRight: '20px',
              fontStyle: 'italic'
            }}
          >
            {name && ` Welcome back ${name.substring(0, name.indexOf(' '))}!`}
          </Typography>
          <a
            href="https://crownlabs.polito.it/cloud"
            target="_blank"
            rel="noreferrer"
          >
            <Tooltip title="MyDrive">
              <IconButton color="secondary">
                <FolderSharedIcon />
              </IconButton>
            </Tooltip>
          </a>
          {isAlsoAdmin && (
            <Tooltip
              title={
                isStudentView
                  ? 'Switch to professor view'
                  : 'Switch to student view'
              }
            >
              <IconButton onClick={switchAdminView} color="secondary">
                {isStudentView ? <ProfIcon /> : <HomeIcon />}
              </IconButton>
            </Tooltip>
          )}
          <Tooltip title="Toggle light/dark theme">
            <IconButton
              color="secondary"
              onClick={() => {
                setIsLightTheme(!isLightTheme);
              }}
            >
              {isLightTheme ? <Dark /> : <Light />}
            </IconButton>
          </Tooltip>
          <Tooltip title="Logout">
            <IconButton onClick={logout} color="secondary">
              <LogoutIcon />
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>
    </div>
  );
}
