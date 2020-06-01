import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import withStyles from '@material-ui/core/styles/withStyles';
import Slider from '@material-ui/core/Slider';

const iOSBoxShadow =
  '0 3px 1px rgba(0,0,0,0.1),0 4px 8px rgba(0,0,0,0.13),0 0 0 1px rgba(0,0,0,0.02)';

const ram = [
  { value: 0.5 },
  { value: 1 },
  { value: 1.5 },
  { value: 2 },
  { value: 2.5 },
  { value: 3 },
  { value: 3.5 },
  { value: 4 },
  { value: 4.5 },
  { value: 5 },
  { value: 5.5 },
  { value: 6 },
  { value: 6.5 },
  { value: 7 },
  { value: 7.5 },
  { value: 8 }
];

const cpu = [{ value: 1 }, { value: 2 }, { value: 3 }, { value: 4 }];

const IOSSlider = withStyles({
  root: {
    color: '#3880ff',
    height: 2,
    padding: '15px 0'
  },
  thumb: {
    height: 28,
    width: 28,
    backgroundColor: '#fff',
    boxShadow: iOSBoxShadow,
    marginTop: -14,
    marginLeft: -14,
    '&:focus, &:hover, &$active': {
      boxShadow:
        '0 3px 1px rgba(0,0,0,0.1),0 4px 8px rgba(0,0,0,0.3),0 0 0 1px rgba(0,0,0,0.02)',
      // Reset on touch devices, it doesn't add specificity
      '@media (hover: none)': {
        boxShadow: iOSBoxShadow
      }
    }
  },
  active: {},
  valueLabel: {
    left: 'calc(-50% + 11px)',
    top: -22,
    '& *': {
      background: 'transparent',
      color: '#ff0000'
    }
  },
  track: {
    height: 2
  },
  rail: {
    height: 2,
    opacity: 0.5,
    backgroundColor: '#bfbfbf'
  },
  mark: {
    backgroundColor: '#bfbfbf',
    height: 8,
    width: 1,
    marginTop: -3
  },
  markActive: {
    opacity: 1,
    backgroundColor: 'currentColor'
  }
})(Slider);

export default function TemplateForm(props) {
  const handleChangeVersion = event => {
    props.setVersion(event.target.value);
  };
  const handleChangeLabid = event => {
    props.setLabid(event.target.value);
  };
  const handleChangeNamespace = event => {
    props.setNamespace(event.target.value);
  };
  const handleChangeImage = event => {
    props.setImage(event.target.value);
    const len = props.imageList.get(event.target.value).length;
    const array = props.imageList.get(event.target.value);
    console.log(array[len - 1]);
    props.setVersion(array[len - 1]);
  };
  return (
    <Grid
      container
      spacing={0}
      alignItems="center"
      justify="center"
      direction="row"
      noValidate
      autoComplete="off"
    >
      <TextField
        style={{ margin: 10, width: '40%' }}
        name="courseCode"
        select
        label="Course Code"
        value={props.namespace}
        onChange={handleChangeNamespace}
        variant="outlined"
        helperText={props.errorcode === 1 ? 'Select a courseCode' : ' '}
      >
        {props.adminGroups.map(x => (
          <MenuItem key={x} value={x}>
            {x.split('course-')[1]}
          </MenuItem>
        ))}
      </TextField>
      <TextField
        required
        type="number"
        placeholder="insert lab id"
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic-image"
        label="Lab ID"
        name="labid"
        variant="outlined"
        value={props.labid}
        onChange={handleChangeLabid}
        helperText={props.errorcode === 2 ? 'Insert a valid labID!' : ' '}
      />
      <TextField
        required
        select
        placeholder="insert image name"
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Image name"
        name="image"
        value={props.image}
        onChange={handleChangeImage}
        variant="outlined"
        helperText={props.errorcode === 3 ? 'Select a valid Image!' : ' '}
      >
        {Array.from(props.imageList.keys()).map(x => (
          <MenuItem key={x} value={x}>
            {x}
          </MenuItem>
        ))}
      </TextField>

      <TextField
        required
        select
        InputLabelProps={{ shrink: true }}
        placeholder="insert image Version"
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Image Version"
        name="image"
        value={props.version}
        onChange={handleChangeVersion}
        variant="outlined"
        helperText={props.errorcode === 4 ? 'Select a valid Version!' : ' '}
      >
        {props.image !== null
          ? props.imageList.get(props.image).map(x => (
              <MenuItem key={x} value={x}>
                {x}
              </MenuItem>
            ))
          : []}
      </TextField>

      <Typography
        style={{ margin: 10, marginTop: 20, width: '35%' }}
        gutterBottom
      >
        Select memory (GB)
      </Typography>
      <IOSSlider
        style={{ margin: 10, marginTop: 20, width: '40%' }}
        aria-label="ios slider"
        defaultValue={2}
        marks={ram}
        valueLabelDisplay="on"
        aria-labelledby="discrete-slider"
        step={0.5}
        min={0.5}
        max={8}
        name="memory"
      />

      <Typography
        style={{ margin: 10, marginTop: 20, width: '35%' }}
        gutterBottom
      >
        Select CPU (Cores)
      </Typography>
      <IOSSlider
        style={{ margin: 10, marginTop: 20, width: '40%' }}
        aria-label="ios slider"
        defaultValue={2}
        marks={cpu}
        valueLabelDisplay="on"
        aria-labelledby="discrete-slider"
        step={1}
        min={1}
        max={4}
        name="cpu"
      />
    </Grid>
  );
}
