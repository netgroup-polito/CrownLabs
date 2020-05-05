import React from 'react';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';
import TableRow from '@material-ui/core/TableRow';
import { labPapersStyle } from './StudentView';
import Container from '@material-ui/core/Container';
import ProfessorFunctionalities from '../components/ProfessorFunctionalities';

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
        <TableRow>
          <ProfessorFunctionalities
            funcNewTemplate={this.props.createTemplate}
          />
        </TableRow>
}
