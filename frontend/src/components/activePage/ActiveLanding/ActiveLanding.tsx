import { FC, useState } from 'react';
import { Col } from 'antd';
import NestedTables from '../NestedTables/NestedTables';
import ViewModeButton from '../ActiveLanding/ViewModeButton/ViewModeButton';
import { workspaces, templates } from '../tempData';
import Box from '../../common/Box';

export interface IActiveLandingProps {
  isTenantManager: boolean;
}

const ActiveLanding: FC<IActiveLandingProps> = ({ ...props }) => {
  const [managerView, setManagerView] = useState(false);
  const { isTenantManager } = props;
  return (
    <Box
      header={{
        size: 'small',
        right: isTenantManager && (
          <ViewModeButton
            setManagerView={setManagerView}
            managerView={managerView}
          />
        ),
      }}
    >
      <Col>
        <NestedTables
          workspaces={workspaces}
          templates={templates}
          isManager={managerView}
          destroyAll={() => alert('VMs deleted')}
        />
      </Col>
    </Box>
  );
};

export default ActiveLanding;
