import React from 'react';
import TableRow from '@material-ui/core/TableRow';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';
import { labPapersStyle } from './StudentView';
import ProfessorFunctionalities from '../components/ProfessorFunctionalities';

export default function ProfessorView(props) {
  return (
    <>
      <TableRow style={labPapersStyle}>
        <LabTemplatesList
          delete={props.delete}
          labs={props.templateLabs}
          func={props.funcTemplate}
          start={props.start}
          isAdmin={true}
        />
        <LabInstancesList
          runningLabs={props.instanceLabs}
          func={props.funcInstance}
          connect={props.connect}
          stop={props.stop}
          showStatus={props.showStatus}
          isAdmin={true}
        />
      </TableRow>
      <TableRow style={labPapersStyle}>
        <ProfessorFunctionalities
          funcNewTemplate={props.createTemplate}
          adminGroups={props.adminGroups}
        />
      </TableRow>
    </>
  );
}
