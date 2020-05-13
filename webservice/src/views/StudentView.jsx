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
const useStyles = makeStyles(theme => ({
  labPapers: labPapersStyle
}));

export default function StudentView(props) {
  const classes = useStyles();

  return (
    <>
      <div className={classes.labPapers}>
        <LabTemplatesList
          labs={props.templateLabs}
          func={props.funcTemplate}
          start={props.start}
        />
        <LabInstancesList
          runningLabs={props.instanceLabs}
          func={props.funcInstance}
          connect={props.connect}
          stop={props.stop}
          showStatus={props.showStatus}
        />
      </div>
    </>
  );
}
