import React from 'react';
import LabTemplatesList from '../components/LabTemplatesList';
import LabInstancesList from '../components/LabInstancesList';
import TableRow from '@material-ui/core/TableRow';
import { labPapersStyle } from './StudentView';
import Container from '@material-ui/core/Container';

class ProfessorView extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
    return (
      <Container>
        <TableRow style={labPapersStyle}>
          <LabTemplatesList
            labs={this.props.templateLabs}
            func={this.props.funcTemplate}
            start={this.props.start}
            isAdmin={true}
          />

          <LabInstancesList
            runningLabs={this.props.instanceLabs}
            func={this.props.funcInstance}
            connect={this.props.connect}
            stop={this.props.stop}
            showStatus={this.props.showStatus}
            isAdmin={true}
          />
        </TableRow>
      </Container>
    );
  }
}
export default ProfessorView;
