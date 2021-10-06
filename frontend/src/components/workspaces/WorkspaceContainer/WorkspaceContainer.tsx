import { FC } from 'react';
import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import Box from '../../common/Box';
import { TemplatesTableLogic } from '../Templates/TemplatesTableLogic';
import { WorkspaceRole } from '../../../utils';
import { Tooltip } from 'antd';

export interface IWorkspaceContainerProps {
  tenantNamespace: string;
  workspace: {
    id: number;
    title: string;
    role: WorkspaceRole;
    workspaceNamespace: string;
  };
}

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const {
    tenantNamespace,
    workspace: { role, title, workspaceNamespace },
  } = props;

  return (
    <>
      <Box
        header={{
          size: 'large',
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>{title}</b>
              </p>
            </div>
          ),
          left: role === 'manager' && (
            <div className="h-full flex justify-center items-center pl-10">
              <Tooltip title="Manage users">
                <Button
                  type="primary"
                  shape="circle"
                  size="large"
                  icon={<UserSwitchOutlined />}
                />
              </Tooltip>
            </div>
          ),
          right: role === 'manager' && (
            <div className="h-full flex justify-center items-center pr-10">
              <Tooltip title="Create template">
                <Button
                  type="lightdark"
                  shape="circle"
                  size="large"
                  icon={<PlusOutlined />}
                />
              </Tooltip>
            </div>
          ),
        }}
      >
        <TemplatesTableLogic
          tenantNamespace={tenantNamespace}
          role={role}
          workspaceNamespace={workspaceNamespace}
        />
      </Box>
    </>
  );
};

export default WorkspaceContainer;
