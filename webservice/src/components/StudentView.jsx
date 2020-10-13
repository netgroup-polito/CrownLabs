import React from 'react';
import { makeStyles } from '@material-ui/core/styles';
import LabTemplatesList from './LabTemplatesList';
import LabInstancesList from './LabInstancesList';

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
  const { templateLabs, start, instanceLabs, connect, stop } = props;
  return (
    <div className={classes.labPapers}>
      <LabTemplatesList labs={templateLabs} start={start} isStudentView />
      <LabInstancesList
        runningLabs={instanceLabs}
        connect={connect}
        stop={stop}
        isStudentView
      />
    </div>
  );
}
