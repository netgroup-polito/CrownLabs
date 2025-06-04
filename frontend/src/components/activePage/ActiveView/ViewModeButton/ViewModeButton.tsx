import type { FC, Dispatch, SetStateAction } from 'react';
import { Radio, Tooltip } from 'antd';
import { TeamOutlined, UserOutlined } from '@ant-design/icons';
import { WorkspaceRole } from '../../../../utils';

export interface IViewModeButtonProps {
  currentView: WorkspaceRole;
  setCurrentView: Dispatch<SetStateAction<WorkspaceRole>>;
}

const ViewModeButton: FC<IViewModeButtonProps> = ({ ...props }) => {
  const { currentView, setCurrentView } = props;

  return (
    <Radio.Group
      value={currentView}
      onChange={e => setCurrentView(e.target.value)}
    >
      <Radio.Button
        className="hidden lg:inline-block"
        value={WorkspaceRole.user}
      >
        Personal
      </Radio.Button>
      <Radio.Button
        className="hidden lg:inline-block"
        value={WorkspaceRole.manager}
      >
        Managed
      </Radio.Button>
      <Radio.Button className="lg:hidden" value="user">
        <Tooltip placement="top" title="Personal">
          <UserOutlined />
        </Tooltip>
      </Radio.Button>
      <Radio.Button className="lg:hidden" value="manager">
        <Tooltip placement="top" title="Managed">
          <TeamOutlined />
        </Tooltip>
      </Radio.Button>
    </Radio.Group>
  );
};

export default ViewModeButton;
