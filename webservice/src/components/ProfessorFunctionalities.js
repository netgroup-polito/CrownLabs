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
import CsvIcon from '@material-ui/icons/Description';
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
      <NewTemplateSlider funcNewTemplate={props.funcNewTemplate} />
      <NewStudentSlider />
    </Grid>
  );
}

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const iOSBoxShadow =
  '0 3px 1px rgba(0,0,0,0.1),0 4px 8px rgba(0,0,0,0.13),0 0 0 1px rgba(0,0,0,0.02)';

const currencies = [
  {
    value: 'DEF',
    label: '-----'
  },
  {
    value: 'CLD',
    label: 'Cloud Computing'
  },
  {
    value: 'RDC',
    label: 'Reti locali e Data Center'
  },
  {
    value: 'TSR',
    label: 'Tecnologie e servizi di rete'
  }
];

const marks = [
  {
    value: 1
  },
  {
    value: 2
  },
  {
    value: 4
  },
  {
    value: 16
  }
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
    setOpen(false);
    let namespace = document.getElementsByName('courseCode')[0].value;
    let lab_number = document.getElementsByName('labNumber')[0].value;
    let description =
      'namespace: ' + namespace + ' laboratory number: ' + lab_number;
    let cpu = document.getElementsByName('cpu')[0].value;
    let memory = document.getElementsByName('memory')[0].value;
    let image = document.getElementsByName('image')[0].value;
    console.log(
      namespace + ' ' + lab_number + ' ' + cpu + ' ' + memory + ' ' + image
    );
    props.funcNewTemplate(
      namespace,
      lab_number,
      description,
      cpu,
      memory,
      image
    );
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
          {'Add new student(s)'}
        </DialogTitle>
        <DialogContent>
          <TemplateForm close={handleClose} />
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="primary"
            startIcon={<CloudUploadIcon />}
            onClick={handleClose}
          >
            Upload
          </Button>

          <Button
            variant="contained"
            color="secondary"
            onClick={handleClose}
            startIcon={<DeleteIcon />}
          >
            Abort
          </Button>
        </DialogActions>
      </Dialog>
    </Grid>
  );
};

const NewStudentSlider = () => {
  const [open, setOpen] = React.useState(false);
  const handleClickOpen = () => {
    setOpen(true);
  };
  const handleClose = () => {
    setOpen(false);
  };

  return (
    <>
      <Button
        variant="contained"
        color="primary"
        onClick={handleClickOpen}
        startIcon={<AddCircleIcon />}
        style={{ margin: '10px' }}
      >
        Add student(s)
      </Button>
      <Dialog
        open={open}
        TransitionComponent={Transition}
        keepMounted
        aria-labelledby="alert-dialog-slide-title"
        aria-describedby="alert-dialog-slide-description"
      >
        <DialogTitle id="alert-dialog-slide-title">
          {'Add new student(s)'}
        </DialogTitle>
        <DialogContent>
          <StudentsForm close={handleClose} />
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="primary"
            startIcon={<CloudUploadIcon />}
            onClick={handleClose}
          >
            Upload
          </Button>

          <Button
            variant="contained"
            color="secondary"
            onClick={handleClose}
            startIcon={<DeleteIcon />}
          >
            Abort
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

const StudentsForm = props => {
  const [currency, setCurrency] = React.useState('DEF');

  const handleChange = event => {
    setCurrency(event.target.value);
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
        id="outlined-basic"
        label="Student ID"
        variant="outlined"
      />
      {/*<TextField style={{margin: 10,width: "40%"}} id="outlined-basic" label="Course name" variant="outlined"/>*/}

      <TextField
        style={{ margin: 10, width: '40%' }}
        id="outlined-select-currency"
        select
        label="Select course"
        value={currency}
        onChange={handleChange}
        variant="outlined"
      >
        {currencies.map(option => (
          <MenuItem key={option.value} value={option.value}>
            {option.label}
          </MenuItem>
        ))}
      </TextField>

      <TextField
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Lab number"
        variant="outlined"
      />
      <TextField
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="Other"
        variant="outlined"
      />

      {/*<input type="file" id="file"/>*/}

      <Button
        style={{ width: '80%' }}
        variant="contained"
        startIcon={<CsvIcon />}
      >
        Upload CSV File
      </Button>
    </Grid>
  );
};

const TemplateForm = props => {
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
        id="outlined-basic"
        label="course_code"
        name="courseCode"
        variant="outlined"
      />
      <TextField
        style={{ margin: 10, width: '40%' }}
        id="outlined-basic"
        label="lab_number"
        name="labNumber"
        variant="outlined"
      />
      <TextField
        style={{ margin: 10, width: '83%' }}
        id="outlined-basic"
        label="image"
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
        marks={marks}
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
        marks={marks}
        valueLabelDisplay="on"
        aria-labelledby="discrete-slider"
        step={1}
        min={1}
        max={4}
        name="cpu"
      />
    </Grid>
  );
};
