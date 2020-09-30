import React from 'react';
import Menu from '@material-ui/core/Menu';
import MenuItem from '@material-ui/core/MenuItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import IconButton from '@material-ui/core/IconButton';

export default function Selector(props) {
  const [anchorEl, setAnchorEl] = React.useState(null);
  const { selectors, value, setValue } = props;
  const handleClick = event => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };

  return (
    <>
      {selectors.length > 1 && (
        <IconButton color="secondary" onClick={handleClick}>
          {selectors.find(sel => sel.value === value).icon}
        </IconButton>
      )}
      <Menu
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={handleClose}
      >
        {selectors.map(({ text, icon, value: newValue }) => (
          <MenuItem
            onClick={() => {
              setValue(newValue);
              setAnchorEl(null);
            }}
            key={text}
          >
            <ListItemIcon>{icon}</ListItemIcon>
            {text}
          </MenuItem>
        ))}
      </Menu>
    </>
  );
}
