import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';
import { TemplatesTable, TemplatesEmpty } from './';
import { Auth } from '../auth';
import { FC, useContext, useState, useEffect } from 'react';
import Box from '../common/Box';
import '../../../index.less'; //To delete, usefull only to storybook
import { ModalCreateTemplate } from '../Modal';

export interface IWorkspaceProps {
  workspace: {
    id: number;
    title: string;
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

const Workspace: FC<IWorkspaceProps> = ({ ...props }) => {
  const { workspace } = props;
  const auth = useContext(Auth);
  const [showCreateTemplate, setShowCreateTemplate] = useState(false);
  const [templates, setTemplates] = useState(workspace.templates);

  useEffect(() => {
    setTemplates(workspace.templates);
  }, [workspace.templates]);

  const createInstance = (id: string) => {
    setTemplates(templates => {
      return templates.map(template =>
        template.id === id
          ? Object.assign({}, template, {
              instances: [
                ...template.instances,
                {
                  id: template.instances.length
                    ? Math.max(
                        ...template.instances.map(instance => instance.id)
                      ) + 1
                    : 1,
                  name: template.name,
                  ip: '192.168.1.5',
                  status: true,
                },
              ],
            })
          : template
      );
    });
  };

  const destroyInstance = (idInstance: number, idTemplate: string) => {
    setTemplates(templates => {
      return templates.map(template =>
        template.id === idTemplate
          ? Object.assign({}, template, {
              instances: template.instances.filter(
                instance => instance.id !== idInstance
              ),
            })
          : template
      );
    });
  };

  const editTemplate = (id: string) => {
    return;
  };

  const deleteTemplate = (id: string) => {
    setTemplates(templates => {
      return templates.filter(template => template.id !== id);
    });
  };

  return (
    <>
      <Box
        headTitle={workspace.title}
        headLeft={
          auth ? (
            <Button
              type="primary"
              shape="circle"
              size="large"
              icon={<UserSwitchOutlined />}
            />
          ) : (
            ''
          )
        }
        headRight={
          auth ? (
            <>
              <Button
                type="lightdark"
                shape="circle"
                size="large"
                icon={<PlusOutlined />}
                onClick={() => setShowCreateTemplate(true)}
              />
              <ModalCreateTemplate
                showmodal={showCreateTemplate}
                setshowmodal={setShowCreateTemplate}
              />
            </>
          ) : (
            ''
          )
        }
      >
        {templates.length ? (
          <TemplatesTable
            templates={templates}
            createInstance={createInstance}
            destroyInstance={destroyInstance}
            editTemplate={editTemplate}
            deleteTemplate={deleteTemplate}
          />
        ) : (
          <TemplatesEmpty />
        )}
      </Box>
    </>
  );
};

export default Workspace;

/*footer={
  workspace?.templates.length ? (
    <Button
      type="success"
      shape="round"
      size={'large'}
      disabled={
        workspace.templates
          .map(template => template.instances.length)
          .reduce((accum, length) => accum + length)
          ? false
          : true
      }
    >
      Show Active
    </Button>
  ) : (
    ''
  )
}*/
