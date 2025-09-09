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
  isPersonalWorkspace?: boolean;
}

const WorkspaceContainer: FC<IWorkspaceContainerProps> = ({ ...props }) => {
  const [showUserListModal, setShowUserListModal] = useState<boolean>(false);

  const { tenantNamespace, workspace, availableQuota } = props;
  console.log('WorkspaceContainer props:', props);
  console.log('WorkspaceContainer received:', {
    tenantNamespace,
    workspace,
    isPersonal: props.isPersonalWorkspace,
    calculatedWorkspaceNamespace: props.isPersonalWorkspace ? tenantNamespace : workspace.namespace
  });
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [createTemplateMutation, { loading }] = useCreateTemplateMutation({
    onError: apolloErrorCatcher,
  });

  const [show, setShow] = useState(false);

  const isPersonal = props.isPersonalWorkspace;

  console.log(
    'WorkspaceContainer val:',
    isPersonal ? tenantNamespace : workspace.namespace,
  );
  console.log('WorkspaceContainer tenantNamespace:', tenantNamespace);
  console.log('WorkspaceContainer workspace:', workspace);
  console.log('WorkspaceContainer isPersonal:', isPersonal);

  const submitHandler = (t: Template) => {
    const finalWorkspaceNamespace = isPersonal ? tenantNamespace : workspace.namespace;
    const templateIdValue = `${workspace.name}-`;
    
    console.log('WorkspaceContainer submitHandler called with:', {
      template: t,
      isPersonal,
      tenantNamespace,
      workspaceNamespace: finalWorkspaceNamespace,
      templateId: templateIdValue,
      workspaceName: workspace.name
    });
    
    return createTemplateMutation({
      variables: {
        workspaceId: workspace.name,
        workspaceNamespace: finalWorkspaceNamespace,
        templateId: templateIdValue,
        templateName: t.name?.trim() || '',
        descriptionTemplate: t.name?.trim() || '',
        image:
          t.image?.includes('/') || t.image?.includes(':')
            ? t.image // Already a full image reference
            : t.registry
              ? `${t.registry}/${t.image}`.trim()
              : t.image || '', // Ensure we always return a string
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
    }).then(result => {
      console.log('WorkspaceContainer createTemplateMutation result:', result);
      return result;
    }).catch(error => {
      console.error('WorkspaceContainer createTemplateMutation error:', error);
      throw error;
    });
  }

  return (
    <>
      {console.log('WorkspaceContainer rendering ModalCreateTemplate with workspaceNamespace:', isPersonal ? tenantNamespace : workspace.namespace)}
      <ModalCreateTemplate
        workspaceNamespace={isPersonal ? tenantNamespace : workspace.namespace}
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
        <TemplatesTableLogic
          tenantNamespace={tenantNamespace}
          role={workspace.role}
          workspaceNamespace={
            isPersonal ? tenantNamespace : workspace.namespace
          }
          workspaceName={workspace.name}
          availableQuota={availableQuota}
          isPersonal={isPersonal}
        />
        {console.log('WorkspaceContainer rendering TemplatesTableLogic with workspaceNamespace:', isPersonal ? tenantNamespace : workspace.namespace)}
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
      </Box>
    </>
  );
};

export default WorkspaceContainer;
