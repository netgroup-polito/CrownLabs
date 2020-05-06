import React from 'react';
import TableRow from '@material-ui/core/TableRow';
import Container from '@material-ui/core/Container';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';
import { labPapersStyle } from './StudentView';

export default function ProfessorView(props) {
  return (
    <>
      <TableRow style={labPapersStyle}>
        <LabTemplatesList
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
    </>
  );
}
