import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import IconButton from '@material-ui/core/IconButton';
import InputAdornment from '@material-ui/core/InputAdornment';
import TextField from '@material-ui/core/TextField';
import ClearRoundedIcon from '@material-ui/icons/ClearRounded';

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
    flexWrap: 'wrap',
    margin: theme.spacing(1)
  },
  textField: {
    maxWidth: 120
  }
}));

export default function TextSelector(props) {
  const classes = useStyles();

  const { value, setValue } = props;

  return (
    <div className={classes.root}>
      <TextField
        label="Filters"
        color="secondary"
        value={value}
        className={classes.textField}
        onChange={e => {
          setValue(e.target.value);
        }}
        variant="outlined"
        size="small"
        InputProps={{
          endAdornment: (
            <InputAdornment position="end">
              {value.length > 0 && (
                <IconButton
                  onClick={() => {
                    setValue('');
                  }}
                  style={{ padding: 0 }}
                >
                  <ClearRoundedIcon />
                </IconButton>
              )}
            </InputAdornment>
          )
        }}
      />
    </div>
  );
}
