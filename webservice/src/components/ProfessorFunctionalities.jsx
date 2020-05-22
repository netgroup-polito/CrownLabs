import React from 'react';
import Button from '@material-ui/core/Button';
import CloudUploadIcon from '@material-ui/icons/CloudUpload';
import TextField from '@material-ui/core/TextField';
import Grid from '@material-ui/core/Grid';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import Slider from '@material-ui/core/Slider';
import withStyles from '@material-ui/core/styles/withStyles';
import Typography from '@material-ui/core/Typography';
import DeleteIcon from '@material-ui/icons/Delete';
import Slide from '@material-ui/core/Slide';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import MenuItem from '@material-ui/core/MenuItem';

export default function ProfessorFunc(props) {
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
      <NewTemplateSlider
        funcNewTemplate={props.funcNewTemplate}
        adminGroups={props.adminGroups}
      />
    </Grid>
  );
}

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const iOSBoxShadow =
  '0 3px 1px rgba(0,0,0,0.1),0 4px 8px rgba(0,0,0,0.13),0 0 0 1px rgba(0,0,0,0.02)';

const ram = [
  { value: 1 },
  { value: 2 },
  { value: 4 },
  { value: 6 },
  { value: 8 },
  { value: 10 },
  { value: 12 },
  { value: 14 },
  { value: 16 }
];

const cpu = [
  { value: 1 },
  { value: 2 },
  { value: 3 },
  { value: 4 },
  { value: 5 },
  { value: 6 },
  { value: 7 },
  { value: 8 }
];

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
      color: '#000'
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

const NewTemplateSlider = props => {
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => {
    setOpen(true);
  };
  const handleClose = () => {
    let namespace = document.getElementsByName('courseCode')[0].value;
    let lab_number = document.getElementsByName('labNumber')[0].value;
    let description =
      'namespace: ' + namespace + ' laboratory number: ' + lab_number;
    let cpu = document.getElementsByName('cpu')[0].value;
    let memory = document.getElementsByName('memory')[0].value;
    let image = document.getElementsByName('image')[0].value;
    if (namespace === '' || lab_number === '' || image === '') {
      alert('Please fill required text boxes!');
      return;
    }
    props.funcNewTemplate(
      namespace,
      lab_number,
      description,
      Number(cpu),
      Number(memory),
      image
    );

    document.getElementsByName('courseCode')[0].value = '';
    document.getElementsByName('labNumber')[0].value = '';
    document.getElementsByName('image')[0].value = '';

    setOpen(false);
  };

  const handleAbort = () => {
    document.getElementsByName('courseCode')[0].value = '';
    document.getElementsByName('labNumber')[0].value = '';
    document.getElementsByName('image')[0].value = '';
    setOpen(false);
  };

  return (
    <Grid item>
      <Button
        variant="contained"
        color="primary"
        onClick={handleClickOpen}
        startIcon={<AddCircleIcon />}
        style={{ margin: '10px' }}
      >
        Create new Template
      </Button>
      <Dialog
        open={open}
        TransitionComponent={Transition}
        keepMounted
        aria-labelledby="alert-dialog-slide-title"
        aria-describedby="alert-dialog-slide-description"
      >
        <DialogTitle id="alert-dialog-slide-title">
          {'Create new template'}
        </DialogTitle>
        <DialogContent>
          <TemplateForm close={handleClose} adminGroups={props.adminGroups} />
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="primary"
            startIcon={<CloudUploadIcon />}
            onClick={handleClose}
            type="submit"
          >
            Upload
          </Button>

          <Button
            variant="contained"
            color="secondary"
            onClick={handleAbort}
            startIcon={<DeleteIcon />}
          >
            Abort
          </Button>
        </DialogActions>
      </Dialog>
    </Grid>
  );
};

const TemplateForm = props => {
  let map = new Map();
  props.adminGroups.forEach(x => map.set(x, x));

  const [namespace, setNamespace] = React.useState('');

  const handleChange = event => {
    setNamespace(event.target.value);
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
        value={namespace}
        onChange={handleChange}
        variant="outlined"
      >
        {props.adminGroups.map(x => (
          <MenuItem key={x} value={x}>
            {x}
          </MenuItem>
        ))}
      </TextField>
      <TextField
        required
        placeholder="insert lab number"
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Lab Number"
        name="labNumber"
        variant="outlined"
      />
      <TextField
        required
        placeholder="insert image"
        style={{ margin: 10, width: '83%' }}
        id="outlined-basic"
        label="Image"
        name="image"
        variant="outlined"
      />

      <Typography
        style={{ margin: 10, marginTop: 20, width: '35%' }}
        gutterBottom
      >
        Select memory
      </Typography>
      <IOSSlider
        style={{ margin: 10, marginTop: 20, width: '40%' }}
        aria-label="ios slider"
        defaultValue={8}
        marks={ram}
        valueLabelDisplay="on"
        aria-labelledby="discrete-slider"
        step={1}
        min={1}
        max={16}
        name="memory"
      />

      <Typography
        style={{ margin: 10, marginTop: 20, width: '35%' }}
        gutterBottom
      >
        Select CPU
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
        max={8}
        name="cpu"
      />
    </Grid>
  );
};
