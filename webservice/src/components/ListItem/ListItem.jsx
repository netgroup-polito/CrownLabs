import React from 'react';
import ListItem from '@material-ui/core/ListItem';
import makeStyles from '@material-ui/core/styles/makeStyles';
import ListItemText from '@material-ui/core/ListItemText';
import ListItemIcons from './ListItemIcons';
import ListItemFields from './ListItemFields';

const useStyles = makeStyles(theme => ({
  instanceInfo: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    color: theme.palette.info.main,
    width: theme.spacing(7) // same as inset fo MUI List item prop
  }
}));

function List(props) {
  const classes = useStyles();

  const {
    primary,
    fields,
    icons,
    onClick,
    isSelected,
    showType,
    type,
    customInfo,
    vmTypeSelectors
  } = props;

  return (
    <ListItem button selected={isSelected} disableRipple onClick={onClick}>
      <div className={classes.instanceInfo}>
        {customInfo}
        {showType && type && (
          <>{vmTypeSelectors.find(sel => sel.value === type).icon}</>
        )}
      </div>
      <ListItemText
        primary={primary}
        secondary={<ListItemFields fields={fields} />}
      />
      <ListItemIcons icons={icons} />
    </ListItem>
  );
}

export default List;
