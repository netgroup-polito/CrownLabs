import React from 'react';
import TableRow from '@material-ui/core/TableRow';
import Grid from '@material-ui/core/Grid';
import Slide from '@material-ui/core/Slide';
import Button from '@material-ui/core/Button';
import AddCircleIcon from '@material-ui/icons/AddCircle';
import Dialog from '@material-ui/core/Dialog';
import DialogTitle from '@material-ui/core/DialogTitle';
import DialogContent from '@material-ui/core/DialogContent';
import DialogActions from '@material-ui/core/DialogActions';
import CloudUploadIcon from '@material-ui/icons/AddCircle';
import DeleteIcon from '@material-ui/icons/Cancel';
import TemplateForm from '../components/NewTemplateForm';
import { labPapersStyle } from './StudentView';
import LabInstancesList from '../components/LabInstancesList';
import LabTemplatesList from '../components/LabTemplatesList';

export default function ProfessorView(props) {
  return (
    <>
      <TableRow style={labPapersStyle}>
        <LabTemplatesList
          delete={props.delete}
          labs={props.templateLabs}
          func={props.funcTemplate}
          start={props.start}
          isAdmin
        />
        <LabInstancesList
          runningLabs={props.instanceLabs}
          func={props.funcInstance}
          connect={props.connect}
          stop={props.stop}
          showStatus={props.showStatus}
          isAdmin
        />
      </TableRow>
      <TableRow style={labPapersStyle}>
        <Grid
          container
          spacing={0}
          alignItems="center"
          justify="center"
          direction="row"
          noValidate
          autoComplete="off"
        >
          <NewTemplateWrapper
            registryName={props.registryName}
            imageList={props.imageList}
            funcNewTemplate={props.funcNewTemplate}
            adminGroups={props.adminGroups}
          />
        </Grid>
      </TableRow>
    </>
  );
}

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

const NewTemplateWrapper = props => {
  const [open, setOpen] = React.useState(false);
  const [image, setImage] = React.useState(null);
  const [version, setVersion] = React.useState(null);
  const [labid, setLabid] = React.useState(null);
  const [namespace, setNamespace] = React.useState(null);
  const [errorcode, setErrorcode] = React.useState(0);

  const handleClickOpen = () => {
    setOpen(true);
  };
  const handleClose = () => {
    if (!props.adminGroups.includes(namespace)) {
      setErrorcode(1);
      return;
    }
    if (labid === null || !labid.length) {
      setErrorcode(2);
      return;
    }
    if (!props.imageList.has(image)) {
      setErrorcode(3);
      return;
    }
    if (!props.imageList.get(image).includes(version)) {
      setErrorcode(4);
      return;
    }

    props.funcNewTemplate(
      namespace,
      labid,
      `namespace: ${namespace} laboratory number: ${labid}`,
      Number(document.getElementsByName('cpu')[0].value),
      Number(document.getElementsByName('memory')[0].value),
      `${props.registryName}/${image}:${version}`
    );

    setErrorcode(0);
    setNamespace(null);
    if (image !== null) setVersion([]);
    setImage(null);
    setLabid(null);
    setOpen(false);
  };

  const handleAbort = () => {
    setErrorcode(0);
    setNamespace(null);
    if (image !== null) setVersion([]);
    setImage(null);
    setLabid(null);
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
          Create new template
        </DialogTitle>
        <DialogContent>
          <TemplateForm
            labid={labid}
            image={image}
            version={version}
            namespace={namespace}
            setLabid={setLabid}
            setImage={setImage}
            setVersion={setVersion}
            setNamespace={setNamespace}
            imageList={props.imageList}
            adminGroups={props.adminGroups}
            errorcode={errorcode}
          />
        </DialogContent>
        <DialogActions>
          <Button
            variant="contained"
            color="primary"
            startIcon={<CloudUploadIcon />}
            onClick={handleClose}
            type="submit"
          >
            Create
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
