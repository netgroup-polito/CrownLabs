import React, { useState } from 'react';
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
  root: {
    flexGrow: 1
  },
  menuButton: {
    marginRight: theme.spacing(2)
  },
  title: {
    fontWeight: 'bold',
    fontSize: '24px',
    color: 'white',
    marginTop: '10px',
    flexGrow: 1
  }
}));

export default function Header(props) {
  const classes = useStyles();
  const [theme, setTheme] = useState('light');
  const iconColor = '#FFFFFF';
  const {
    renderAdminBtn,
    switchAdminView,
    adminHidden,
    logged,
    name,
    logout
  } = props;
  const adminBtn = renderAdminBtn ? (
    <Tooltip title="Switch professor/student view">
      <IconButton onClick={switchAdminView}>
        {adminHidden ? (
          <ProfIcon style={{ color: iconColor }} />
        ) : (
          <HomeIcon style={{ color: iconColor }} />
        )}
      </IconButton>
    </Tooltip>
  ) : null;

  return (
    <div className={classes.root}>
      <AppBar id="toolbar" position="static" style={{ background: '#032364' }}>
        <Toolbar>
          <img
            src={CrownLogo}
            style={{ marginRight: '20px', height: '40px' }}
            alt=""
          />
          <Typography variant="h6" className={classes.title}>
            Crownlabs
          </Typography>
          <Typography
            style={{
              textAlign: 'right',
              color: '#FFFFFF',
              marginRight: '20px',
              fontStyle: 'italic'
            }}
          >
            {logged && name ? ` Welcome back, ${name}!` : ''}
          </Typography>
          {logged ? (
            <a
              href="https://crownlabs.polito.it/cloud"
              target="_blank"
              rel="noreferrer"
            >
              <Tooltip title="MyDrive">
                <IconButton aria-label="MyDrive">
                  <FolderSharedIcon style={{ color: iconColor }} />
                </IconButton>
              </Tooltip>
            </a>
          ) : null}
          {adminBtn}
          <Tooltip title="Toggle light/dark theme">
            <IconButton
              onClick={() => {
                document.getElementById('themeSwitch').click();
                if (theme === 'light') setTheme('dark');
                else setTheme('light');
              }}
            >
              {theme !== 'light' ? (
                <Light style={{ color: iconColor }} t />
              ) : (
                <Dark style={{ color: iconColor }} />
              )}
            </IconButton>
          </Tooltip>
          {logged && (
            <Tooltip title="Logout">
              <IconButton onClick={logout}>
                <LogoutIcon style={{ color: iconColor }} />
              </IconButton>
            </Tooltip>
          )}
        </Toolbar>
      </AppBar>
    </div>
  );
}
