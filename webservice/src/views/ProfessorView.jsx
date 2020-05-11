import React from 'react';
import TableRow from '@material-ui/core/TableRow';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';
import { labPapersStyle } from './StudentView';
import ListSubheader from "@material-ui/core/ListSubheader";

export default function ProfessorView(props) {
  return (
    <>
      <TableRow style={labPapersStyle}>
        <LabTemplatesList
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
    </>
  );
}
