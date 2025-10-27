import { PlusOutlined, UserSwitchOutlined } from '@ant-design/icons';
import { Badge, Modal, Tooltip } from 'antd';
import { Button, Card } from 'antd';
import type { FC } from 'react';
import { useContext, useEffect, useState } from 'react';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import {
  useCreateTemplateMutation,
  useApplyTemplateMutation,
  EnvironmentType,
} from '../../../generated-types';
import { getTemplatePatchJson } from '../../../graphql-components/utils';
import { AuthContext } from '../../../contexts/AuthContext';
import { makeRandomDigits } from '../../../utils';
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
  const { userId } = useContext(AuthContext);
  const getManager = () =>
    `${workspace.name}-${userId || makeRandomDigits(10)}`;
  const [showUserListModal, setShowUserListModal] = useState<boolean>(false);

  const { tenantNamespace, workspace, availableQuota, refreshQuota } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [createTemplateMutation, { loading }] = useCreateTemplateMutation({
    onError: apolloErrorCatcher,
  });
  const [applyTemplateMutation] = useApplyTemplateMutation({
    onError: apolloErrorCatcher,
  });

  const [show, setShow] = useState(false);
  // Template currently being edited (undefined => create new)
  const [editingTemplate, setEditingTemplate] = useState<Template | undefined>(
    undefined,
  );

  useEffect(() => {
    const handler = (e: Event) => {
      console.debug('[openTemplateModal] event received', e);
      const detail = (e as CustomEvent).detail;
      const t = detail as Template;
      if (t) {
        t.imageType = detail.environmentType || null;
        if (t.imageType === EnvironmentType.VirtualMachine) {
          t.image = detail.image || '';
        } else {
          t.registry = detail.image || '';
        }
        t.cpu = detail.resources.cpu || 1;
        t.ram = detail.resources.memory
          ? parseInt(detail.resources.memory) / 1000
          : 1;
        t.disk = detail.resources.disk
          ? parseInt(detail.resources.disk) / 1000
          : 10;
        setEditingTemplate(t);
        setShow(true);
      }
    };
    window.addEventListener('openTemplateModal', handler as EventListener);
    return () =>
      window.removeEventListener('openTemplateModal', handler as EventListener);
  }, []);

  // clear editingTemplate when modal hidden
  useEffect(() => {
    if (!show) setEditingTemplate(undefined);
  }, [show]);

  const isPersonal = props.isPersonalWorkspace;

  const submitHandler = (t: Template) => {
    const finalWorkspaceNamespace = isPersonal
      ? tenantNamespace
      : workspace.namespace;

    // normalize final image
    let finalImage = t.image || '';
    if (finalImage && !finalImage.includes('/') && !finalImage.includes('.')) {
      finalImage = `registry.internal.crownlabs.polito.it/${finalImage}`;
    }

    // If editing an existing template (has id) -> apply patch
    if (t && t.id) {
      const templateId = t.id;
      const patchJson = getTemplatePatchJson({
        metadata: { name: templateId, namespace: finalWorkspaceNamespace },
        spec: {
          prettyName: t.name?.trim() || '',
          description: t.name?.trim() || '',
          environmentList: [
            {
              name: t.environmentName || t.name?.trim() || 'env-0',
              environmentType: t.imageType || EnvironmentType.Container,
              image: finalImage,
              guiEnabled: !!t.gui,
              persistent: !!t.persistent,
              rewriteURL: !!t.rewriteUrl,
              resources: {
                cpu: t.cpu,
                memory: `${t.ram * 1000}M`,
                disk: t.disk ? `${t.disk * 1000}M` : undefined,
                reservedCPUPercentage: t.reservedCPUPercentage ?? 50,
              },
              sharedVolumeMounts: t.sharedVolumeMountInfos ?? [],
            },
          ],
        },
      });

      return applyTemplateMutation({
        variables: {
          templateId,
          workspaceNamespace: finalWorkspaceNamespace,
          patchJson,
          manager: getManager(),
        },
      })
        .then(result => {
          refreshQuota?.();
          // notify other components to refresh templates data
          window.dispatchEvent(
            new CustomEvent('templatesChanged', {
              detail: {
                workspaceNamespace: finalWorkspaceNamespace,
                templateId,
              },
            }),
          );
          return result;
        })
        .catch(error => {
          console.error(
            'WorkspaceContainer applyTemplateMutation error:',
            error,
          );
          throw error;
        });
    }

    // Create new template flow (keep existing behavior)
    const templateIdValue = `${workspace.name}-`;
    return createTemplateMutation({
      variables: {
        workspaceId: workspace.name,
        workspaceNamespace: finalWorkspaceNamespace,
        templateId: templateIdValue,
        templateName: t.name?.trim() || '',
        descriptionTemplate: t.name?.trim() || '',
        image: finalImage,
        guiEnabled: t.gui,
        rewriteURL: t.rewriteUrl,
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
        refreshQuota?.();
        // notify other components to refresh templates data (new template created)
        window.dispatchEvent(
          new CustomEvent('templatesChanged', {
            detail: { workspaceNamespace: finalWorkspaceNamespace },
          }),
        );
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
    <Card
      className="cl-card-box"
      bordered={false}
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        background: 'transparent',
        boxShadow: 'none',
      }} // force bounded card + transparent background
      bodyStyle={{ background: 'transparent', padding: 0 }} // make inner body transparent
    >
      {/* keep header / controls as-is inside the card */}
      <div className="flex flex-col flex-1 min-h-0">
        <ModalCreateTemplate
          workspaceNamespace={
            isPersonal ? tenantNamespace : workspace.namespace
          }
          {...(editingTemplate ? { template: editingTemplate } : {})}
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
                      // open modal in "create" mode
                      setEditingTemplate(undefined);
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
              // do not clip overflow here — let the inner scrollable element handle scrolling
            }}
          >
            {/* use the box helper class so the CSS (.cl-table-instance) takes effect; keeps scroll surface correct */}
            <div
              className="cl-table-instance"
              style={{
                flex: '1 1 auto',
                minHeight: 0,
                overflowY: 'auto',
                overflowX: 'hidden',
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
      {/* Listen for global edit requests from row settings if parent wiring isn't present */}
      {/* This ensures clicking "Edit" always opens the modal with the selected template */}
      {/* Add listener once per container */}
    </Card>
  );
};

export default WorkspaceContainer;
