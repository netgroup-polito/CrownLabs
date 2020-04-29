import React from 'react';
import FolderSharedIcon from '@material-ui/icons/FolderShared';
import ToolTip from '@material-ui/core/Tooltip';
import makeStyles from '@material-ui/core/styles/makeStyles';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import Typography from '@material-ui/core/Typography';
import { AccountCircle } from '@material-ui/icons';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import { createMuiTheme, ThemeProvider } from '@material-ui/core/styles';

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
    fontSize: '18px',
    color: 'white',
    marginTop: '10px',
    flexGrow: 1
  }
}));

export default function Header(props) {
  const classes = useStyles();
  const [anchorEl, setAnchorEl] = React.useState(null);
  const open = Boolean(anchorEl);

  const handleMenu = event => {
    setAnchorEl(event.currentTarget);
  };

  const handleClose = () => {
    setAnchorEl(null);
  };

  const name = props.adminHidden ? 'Professor Area' : 'Student Area';
  const adminBtn = props.renderAdminBtn ? (
    <MenuItem onClick={props.switchAdminView}>{name}</MenuItem>
  ) : null;

  const darkTheme = React.useMemo(
    () =>
      createMuiTheme({
        palette: {
          type: 'dark'
        }
      }),
    []
  );

  return (
    <ThemeProvider theme={darkTheme}>
      <div className={classes.root}>
        <AppBar
          id="toolbar"
          position="static"
          style={{ background: '#032364' }}
        >
          <Toolbar>
            <img
              src={require('../assets/crown.png')}
              style={{ marginRight: '20px', height: '40px' }}
              alt=""
            />
            <Typography variant="h6" className={classes.title}>
              {props.logged && props.name ? 'Welcome back, ' + props.name : ''}
            </Typography>
            {props.logged ? (
              <a href="https://crownlabs.polito.it/cloud" target="_blank">
                <ToolTip title="My drive">
                  <FolderSharedIcon
                    style={{
                      marginRight: 25,
                      color: 'white',
                      fontSize: '2rem'
                    }}
                  />
                </ToolTip>
              </a>
            ) : null}
            {props.logged && (
              <div>
                <IconButton
                  aria-label="account of current user"
                  aria-controls="menu-appbar"
                  aria-haspopup="true"
                  onClick={handleMenu}
                  color="secondary"
                >
                  <AccountCircle />
                </IconButton>
                <Menu
                  id="menu-appbar"
                  anchorEl={anchorEl}
                  anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'right'
                  }}
                  keepMounted
                  transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right'
                  }}
                  open={open}
                  onClose={handleClose}
                >
                  {adminBtn}
                  <MenuItem onClick={handleClose}>Change password</MenuItem>
                  <MenuItem onClick={props.logout}>Logout</MenuItem>
                </Menu>
              </div>
            )}
          </Toolbar>
        </AppBar>
      </div>
    </ThemeProvider>
  );
}
