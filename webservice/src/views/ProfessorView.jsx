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
import DeleteIcon from '@material-ui/icons/Cancel';
import TemplateForm from '../components/NewTemplateForm';
import { labPapersStyle } from './StudentView';
import LabInstancesList from '../components/LabInstancesList';
import LabTemplatesList from '../components/LabTemplatesList';

const Transition = React.forwardRef(function Transition(props, ref) {
  return <Slide direction="up" ref={ref} {...props} />;
});

// next disable is to avoid to create a single file for the trantision component
// eslint-disable-next-line react/no-multi-comp
export default function ProfessorView(props) {
  const {
    deleteLabTemplate,
    templateLabs,
    start,
    instanceLabs,
    connect,
    stop,
    showStatus,
    registryName,
    imageList,
    createNewTemplate,
    adminGroups
  } = props;

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
    if (!adminGroups.includes(namespace)) {
      setErrorcode(1);
      return;
    }
    if (labid === null || !labid.length) {
      setErrorcode(2);
      return;
    }
    if (!imageList.has(image)) {
      setErrorcode(3);
      return;
    }
    if (!imageList.get(image).includes(version)) {
      setErrorcode(4);
      return;
    }

    createNewTemplate(
      namespace,
      labid,
      `namespace: ${namespace} laboratory number: ${labid}`,
      Number(document.getElementsByName('cpu')[0].value),
      Number(document.getElementsByName('memory')[0].value),
      `${registryName}/${image}:${version}`
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
    <>
      <TableRow style={labPapersStyle}>
        <LabTemplatesList
          deleteLabTemplate={deleteLabTemplate}
          labs={templateLabs}
          start={start}
        />
        <LabInstancesList
          runningLabs={instanceLabs}
          connect={connect}
          stop={stop}
          showStatus={showStatus}
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
            <Dialog open={open} TransitionComponent={Transition} keepMounted>
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
                  imageList={imageList}
                  adminGroups={adminGroups}
                  errorcode={errorcode}
                />
              </DialogContent>
              <DialogActions>
                <Button
                  variant="contained"
                  color="primary"
                  startIcon={<AddCircleIcon />}
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
        </Grid>
      </TableRow>
    </>
  );
}
