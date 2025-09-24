import { PlusOutlined, UserSwitchOutlined } from '@ant-design/icons';
import { Badge, Modal, Tooltip } from 'antd';
import { Button } from 'antd';
import type { FC } from 'react';
import { useContext, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import {
  useCreateTemplateMutation,
  EnvironmentType,
} from '../../../generated-types';
import type { Workspace } from '../../../utils';
import { WorkspaceRole } from '../../../utils';
import UserListLogic from '../../accountPage/UserListLogic/UserListLogic';
import Box from '../../common/Box';
import ModalCreateTemplate from '../ModalCreateTemplate';
import type { Template } from '../ModalCreateTemplate/ModalCreateTemplate';
import { TemplatesTableLogic } from '../Templates/TemplatesTableLogic';

export interface IWorkspaceContainerProps {
  tenantNamespace: string;
  workspace: Workspace;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
  refreshQuota?: () => void; // Add refresh function
  isPersonalWorkspace?: boolean;
}

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const [showUserListModal, setShowUserListModal] = useState<boolean>(false);

  const { tenantNamespace, workspace, availableQuota, refreshQuota } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [createTemplateMutation, { loading }] = useCreateTemplateMutation({
    onError: apolloErrorCatcher,
  });

  const [show, setShow] = useState(false);

  const isPersonal = props.isPersonalWorkspace;

  const submitHandler = (t: Template) => {
    const finalWorkspaceNamespace = isPersonal
      ? tenantNamespace
      : workspace.namespace;
    const templateIdValue = `${workspace.name}-`;

    // The image should already be properly formatted from ModalCreateTemplate
    // But add a fallback just in case
    let finalImage = t.image || '';

    // Only apply fallback logic if the image doesn't already contain a registry
    if (finalImage && !finalImage.includes('/') && !finalImage.includes('.')) {
      finalImage = `registry.internal.crownlabs.polito.it/${finalImage}`;
    }

    return createTemplateMutation({
      variables: {
        workspaceId: workspace.name,
        workspaceNamespace: finalWorkspaceNamespace,
        templateId: templateIdValue,
        templateName: t.name?.trim() || '',
        descriptionTemplate: t.name?.trim() || '',
        image: finalImage,
        guiEnabled: t.gui,
        persistent: t.persistent,
        mountMyDriveVolume: t.mountMyDrive,
        environmentType: t.imageType || EnvironmentType.Container,
        resources: {
          cpu: t.cpu,
          memory: `${t.ram * 1000}M`,
          disk: t.disk ? `${t.disk * 1000}M` : undefined,
          reservedCPUPercentage: 50,
        },
        sharedVolumeMounts: t.sharedVolumeMountInfos ?? [],
      },
    })
      .then(result => {
        // Refresh quota after template creation
        refreshQuota?.();
        return result;
      })
      .catch(error => {
        console.error(
          'WorkspaceContainer createTemplateMutation error:',
          error,
        );
        throw error;
      });
  };

  return (
    <>
      {/* make Box fill available vertical space so its inner table can scroll */}
      <div className="flex flex-col flex-1 min-h-0">
        <ModalCreateTemplate
          workspaceNamespace={
            isPersonal ? tenantNamespace : workspace.namespace
          }
          cpuInterval={{ max: 8, min: 1 }}
          ramInterval={{ max: 32, min: 1 }}
          diskInterval={{ max: 50, min: 10 }}
          setShow={setShow}
          show={show}
          submitHandler={submitHandler}
          loading={loading}
          isPersonal={isPersonal}
        />
        <Box
          header={{
            size: 'large',
            center: (
              <div className="h-full flex justify-center items-center px-5">
                <p className="md:text-4xl text-2xl text-center mb-0">
                  <b>{workspace.prettyName}</b>
                </p>
              </div>
            ),
            left: workspace.role === WorkspaceRole.manager && (
              <div className="h-full flex justify-center items-center pl-10">
                <Tooltip title="Manage users">
                  <Button
                    type="primary"
                    shape="circle"
                    size="large"
                    icon={<UserSwitchOutlined />}
                    onClick={() => setShowUserListModal(true)}
                  >
                    {workspace.waitingTenants && (
                      <Badge
                        count={workspace.waitingTenants}
                        color="yellow"
                        className="absolute -top-2.5 -right-2.5"
                      />
                    )}
                  </Button>
                </Tooltip>
              </div>
            ),
            right: workspace.role === WorkspaceRole.manager && (
              <div className="h-full flex justify-center items-center pr-10">
                <Tooltip title="Create template">
                  <Button
                    onClick={() => {
                      setShow(true);
                    }}
                    type="primary"
                    shape="circle"
                    size="large"
                    icon={<PlusOutlined />}
                  />
                </Tooltip>
              </div>
            ),
          }}
        >
          {/* make the Box body a flex column and make the templates area scrollable */}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              flex: '1 1 auto',
              minHeight: 0,
              overflow: 'hidden',
            }}
          >
            <div
              style={{
                flex: '1 1 auto',
                minHeight: 0,
                overflow: 'auto',
              }}
            >
              <TemplatesTableLogic
                tenantNamespace={tenantNamespace}
                role={workspace.role}
                workspaceNamespace={
                  isPersonal ? tenantNamespace : workspace.namespace
                }
                workspaceName={workspace.name}
                availableQuota={availableQuota}
                refreshQuota={refreshQuota}
                isPersonal={isPersonal}
              />
            </div>
          </div>
        </Box>
      </div>

      <Modal
        destroyOnHidden={true}
        title={`Users in ${workspace.prettyName} `}
        width="800px"
        open={showUserListModal}
        footer={null}
        onCancel={() => setShowUserListModal(false)}
      >
        <UserListLogic workspace={workspace} />
      </Modal>
    </>
  );
};

export default WorkspaceContainer;
