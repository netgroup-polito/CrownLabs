import { FC } from 'react';
import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import { TemplatesTable } from '../Templates/TemplatesTable';
import { TemplatesEmpty } from '../Templates/TemplatesEmpty';
import Box from '../../common/Box';
import { WorkspaceRole } from '../../../utils';

export interface IWorkspaceContainerProps {
  workspace: {
    id: number;
    title: string;
    role: WorkspaceRole;
    templates: Array<{
      id: string;
      name: string;
      gui: boolean;
      instances: Array<{
        id: number;
        name: string;
        ip: string;
        status: boolean;
      }>;
    }>;
  };
}

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const {
    workspace: { role, title, templates },
  } = props;

  const editTemplate = (id: string) => {
    return;
  };

  const deleteTemplate = (id: string) => {
    return;
  };

  return (
    <>
      <Box
        header={{
          size: 'large',
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>{title}</b>
            </p>
          ),
          left: role === 'manager' && (
            <Button
              type="primary"
              shape="circle"
              size="large"
              icon={<UserSwitchOutlined />}
            />
          ),
          right: role === 'manager' && (
            <Button
              type="lightdark"
              shape="circle"
              size="large"
              icon={<PlusOutlined />}
            />
          ),
        }}
      >
        {templates.length ? (
          <TemplatesTable
            templates={templates}
            role={role}
            editTemplate={editTemplate}
            deleteTemplate={deleteTemplate}
          />
        ) : (
          <TemplatesEmpty role={role} />
        )}
      </Box>
    </>
  );
};

export default WorkspaceContainer;
