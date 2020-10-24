import React, { Fragment } from 'react';
import IconButton from '@material-ui/core/IconButton';
import Tooltip from '@material-ui/core/Tooltip';
import { makeStyles } from '@material-ui/core/styles';

const useStyles = makeStyles(theme => ({
  success: {
    color: theme.palette.success.main
  },
  info: {
    color: theme.palette.info.main
  },
  warning: {
    color: theme.palette.warning.light
  },
  error: {
    color: theme.palette.error.main
  }
}));

function ListItemIcons(props) {
  const { icons } = props;
  const classes = useStyles();
  return (
    <>
      {icons.map(
        ({ title, condition, icon: Icon, color, onClick, iconClassName }) => (
          <Fragment key={title}>
            {condition && (
              <Tooltip title={title}>
                <IconButton
                  className={classes[color]}
                  variant="dark"
                  onClick={onClick}
                >
                  <Icon fontSize="large" className={iconClassName} />
                </IconButton>
              </Tooltip>
            )}
          </Fragment>
        )
      )}
    </>
  );
}

export default ListItemIcons;
