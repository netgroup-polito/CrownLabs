import React from 'react';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import IconButton from '@material-ui/core/IconButton';
import makeStyles from '@material-ui/core/styles/makeStyles';
import ArrowUpwardIcon from '@material-ui/icons/ArrowUpward';

const useStyles = makeStyles(() => ({
  flipOrderBtn: {
    transition: '0.5s ease-out',
    marginRight: 10
  },
  rotate0: {
    transform: 'rotateZ(0deg)'
  },
  rotate180: {
    transform: 'rotateZ(180deg)'
  }
}));

export default function OrderSelector(props) {
  const classes = useStyles();
  const [anchorEl, setAnchorEl] = React.useState(null);
  const { selectors, orderData, setOrderData } = props;
  const { order, isDirUp } = orderData;
  const handleClick = event => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <div>
      {selectors.length > 1 && (
        <IconButton color="secondary" onClick={handleClick}>
          {selectors.find(sel => sel.value === order).icon}
        </IconButton>
      )}
      <Menu
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {selectors.map(({ text, icon, value }) => (
          <MenuItem
            onClick={() => {
              setOrderData({ ...orderData, order: value });
              setAnchorEl(null);
            }}
            key={text}
          >
            <ListItemIcon>{icon}</ListItemIcon>
            {text}
          </MenuItem>
        ))}
      </Menu>
      <IconButton
        color="secondary"
        onClick={() => setOrderData({ ...orderData, isDirUp: !isDirUp })}
        className={`${classes.flipOrderBtn} ${
          isDirUp ? classes.rotate180 : classes.rotate0
        }`}
      >
        <ArrowUpwardIcon fontSize="large" />
      </IconButton>
    </div>
  );
}
