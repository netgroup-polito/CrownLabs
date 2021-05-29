import { FC } from 'react';
import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import { TemplatesTable } from '../Templates/TemplatesTable';
import { TemplatesEmpty } from '../Templates/TemplatesEmpty';
import Box from '../../common/Box';
import { Template, WorkspaceRole } from '../../../utils';

export interface IWorkspaceContainerProps {
  workspace: {
    id: number;
    title: string;
    role: WorkspaceRole;
    templates: Array<Template>;
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
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>{title}</b>
              </p>
            </div>
          ),
          left: role === 'manager' && (
            <div className="h-full flex justify-center items-center pl-10">
              <Button
                type="primary"
                shape="circle"
                size="large"
                icon={<UserSwitchOutlined />}
              />
            </div>
          ),
          right: role === 'manager' && (
            <div className="h-full flex justify-center items-center pr-10">
              <Button
                type="lightdark"
                shape="circle"
                size="large"
                icon={<PlusOutlined />}
              />
            </div>
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
