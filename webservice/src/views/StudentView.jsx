import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';

export const labPapersStyle = {
  display: 'flex',
  justifyContent: 'space-around',
  width: '100%',
  flexWrap: 'wrap',
  marginTop: 30
};
const useStyles = makeStyles(() => ({
  labPapers: labPapersStyle
}));

export default function StudentView(props) {
  const classes = useStyles();
  const {
    templateLabs,
    selectTemplate,
    start,
    instanceLabs,
    selectInstance,
    connect,
    stop,
    showStatus
  } = props;

  return (
    <div className={classes.labPapers}>
      <LabTemplatesList
        labs={templateLabs}
        selectTemplate={selectTemplate}
        start={start}
      />
      <LabInstancesList
        runningLabs={instanceLabs}
        selectInstance={selectInstance}
        connect={connect}
        stop={stop}
        showStatus={showStatus}
      />
    </div>
  );
}
