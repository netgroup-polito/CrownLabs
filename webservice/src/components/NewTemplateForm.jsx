import Grid from '@material-ui/core/Grid';
import TextField from '@material-ui/core/TextField';
import MenuItem from '@material-ui/core/MenuItem';
import Typography from '@material-ui/core/Typography';
import React from 'react';
import withStyles from '@material-ui/core/styles/withStyles';
import Slider from '@material-ui/core/Slider';
import { VM_TYPES } from '../services/ApiManager';

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
  const {
    namespace,
    errorcode,
    adminGroups,
    image,
    imageList,
    description,
    type,
    version,
    setVersion,
    setImage,
    setNamespace,
    setType,
    setDescription
  } = props;

  const handleChangeVersion = event => {
    setVersion(event.target.value);
  };
  const handleChangeDescription = event => {
    setDescription(event.target.value);
  };
  const handleChangeType = event => {
    setType(event.target.value);
  };

  const handleChangeNamespace = event => {
    setNamespace(event.target.value);
  };
  const handleChangeImage = event => {
    setImage(event.target.value);
    const len = imageList.get(event.target.value).length;
    const array = imageList.get(event.target.value);
    setVersion(array[len - 1]);
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
        style={{ margin: 10, width: '84%' }}
        name="courseCode"
        select
        label="Course Code"
        value={namespace || ''}
        onChange={handleChangeNamespace}
        variant="outlined"
        helperText={errorcode === 1 ? 'Select a courseCode' : ' '}
      >
        {adminGroups.map(x => (
          <MenuItem key={x} value={x}>
            {x.split('course-')[1]}
          </MenuItem>
        ))}
      </TextField>
      <TextField
        required
        select
        placeholder="insert image name"
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Image name"
        name="image"
        value={image || ''}
        onChange={handleChangeImage}
        variant="outlined"
        helperText={errorcode === 3 ? 'Select a valid Image!' : ' '}
      >
        {Array.from(imageList.keys()).map(imageName => (
          <MenuItem key={imageName} value={imageName}>
            {imageName}
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
        name="version"
        disabled={!image || imageList.get(image).length === 1}
        value={version || ''}
        onChange={handleChangeVersion}
        variant="outlined"
        helperText={errorcode === 4 ? 'Select a valid Version!' : ' '}
      >
        {image !== null
          ? imageList.get(image).map(imageVersion => (
              <MenuItem key={imageVersion} value={imageVersion}>
                {imageVersion}
              </MenuItem>
            ))
          : []}
      </TextField>
      <TextField
        required
        type="text"
        InputLabelProps={{ shrink: true }}
        placeholder=""
        style={{
          margin: 10,
          width: '60%'
        }}
        id="outlined-basic"
        label="VM description"
        name="description"
        value={description || ''}
        onChange={handleChangeDescription}
        variant="outlined"
        helperText={errorcode === 4 ? 'Select a valid Version!' : ' '}
      />
      <TextField
        required
        select
        InputLabelProps={{ shrink: true }}
        style={{
          margin: 10,
          width: '20%'
        }}
        id="outlined-basic"
        label="VM type"
        name="type"
        value={type || ''}
        onChange={handleChangeType}
        variant="outlined"
        helperText={errorcode === 4 ? 'Select a valid Version!' : ' '}
      >
        {Object.values(VM_TYPES).map(vmType => (
          <MenuItem key={vmType} value={vmType}>
            {vmType}
          </MenuItem>
        ))}
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
