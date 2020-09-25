import React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GitHubButton from 'react-github-btn';

const useStyles = makeStyles(theme => ({
  footer: {
    height: '70px',
    backgroundColor: theme.palette.primary.main,
    fontSize: '1.2rem',
    color: theme.palette.getContrastText(theme.palette.primary.main),
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center'
  }
}));

export default function Footer() {
  const classes = useStyles();
  return (
    <div className={classes.footer}>
      <p>
        This software has been proudly developed at Politecnico di Torino
        &nbsp;&nbsp;&nbsp;
      </p>
      <GitHubButton
        href="https://github.com/netgroup-polito/CrownLabs"
        data-size="large"
        data-show-count="true"
        aria-label="Star netgroup-polito/CrownLabs on GitHub"
      >
        Star
      </GitHubButton>
    </div>
  );
}
