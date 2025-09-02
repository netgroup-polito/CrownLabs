import {
  CodeOutlined,
  DesktopOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { Space, Tooltip, Dropdown, Badge } from 'antd';
import { Button } from 'antd';
import type { FetchResult } from '@apollo/client';
import type { FC } from 'react';
import { useCallback, useContext, useState } from 'react';
import SvgInfinite from '../../../../assets/infinite.svg?react';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import type {
  CreateInstanceMutation,
  DeleteTemplateMutation,
} from '../../../../generated-types';
import {
  useInstancesLabelSelectorQuery,
  useNodesLabelsQuery,
} from '../../../../generated-types';
import { JSONDeepCopy } from '../../../../utils';
import { TenantContext } from '../../../../contexts/TenantContext';
import type { Template } from '../../../../utils';
import { cleanupLabels, WorkspaceRole } from '../../../../utils';
import { ModalAlert } from '../../../common/ModalAlert';
import { TemplatesTableRowSettings } from '../TemplatesTableRowSettings';
import NodeSelectorIcon from '../../../common/NodeSelectorIcon/NodeSelectorIcon';
import ModalCreateTemplate, { Template as TemplateType } from '../../ModalCreateTemplate/ModalCreateTemplate';
import { useApplyTemplateMutation, EnvironmentType } from '../../../../generated-types'; // <-- Use the correct hook

export interface ITemplatesTableRowProps {
  template: Template;
  role: WorkspaceRole;
  totalInstances: number;
  editTemplate: (id: string) => void;
  deleteTemplate: (
    id: string
  ) => Promise<
    FetchResult<
      DeleteTemplateMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  deleteTemplateLoading: boolean;
  createInstance: (
    id: string,
    labelSelector?: JSON
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  expandRow: (value: string, create: boolean) => void;
}

const convertMemory = (s: string): string =>
  s.includes('M') && Number(s.split('M')[0]) >= 1000
    ? `${Number(s.split('M')[0]) / 1000}G`
    : s;

const canCreateInstance = (template: Template, allTemplates: Template[], workspaceQuota?: {
  cpu?: string | number;
  memory?: string;
  instances?: number;
}): boolean => {
  // If no quota defined, default to allowing creation
  if (!workspaceQuota) return true;
  
  // Calculate current usage (similar to QuotaDisplay)
  let usedCpu = 0;
  let usedMemory = 0;
  let runningInstances = 0;
  
  // Sum resources from all running instances
  allTemplates.forEach(tmpl => {
    const count = tmpl.instances?.length || 0;
    runningInstances += count;
    if (tmpl.resources) {
      usedCpu += (tmpl.resources.cpu || 0) * count;
      usedMemory += parseMemory(tmpl.resources.memory || '0') * count;
    }
  });
  
  // Get the quota limits (with defaults)
  const quotaLimits = {
    cpu: workspaceQuota?.cpu ? parseInt(String(workspaceQuota.cpu)) : 8,
    memory: workspaceQuota?.memory ? parseMemory(workspaceQuota.memory) : 16,
    instances: workspaceQuota?.instances || 8,
  };
  
  // Add the new instance's resource requirements
  const newUsage = {
    cpu: usedCpu + (template.resources?.cpu || 0),
    memory: usedMemory + parseMemory(template.resources?.memory || '0'),
    instances: runningInstances + 1,
  };
  
  // Check if any quota would be exceeded
  return (
    newUsage.cpu <= quotaLimits.cpu &&
    newUsage.memory <= quotaLimits.memory &&
    newUsage.instances <= quotaLimits.instances
  );
};

// Helper function to parse memory (copy from QuotaDisplay.tsx)
const parseMemory = (memoryStr: string): number => {
  if (!memoryStr) return 0;

  const match = memoryStr.match(/^(\d+(?:\.\d+)?)(.*)?$/);
  if (!match) return 0;

  const value = parseFloat(match[1]);
  const unit = match[2]?.toLowerCase() || '';

  switch (unit) {
    case 'gi':
    case 'g':
      return value;
    case 'mi':
    case 'm':
      return value / 1024;
    case 'ki':
    case 'k':
      return value / (1024 * 1024);
    case 'ti':
    case 't':
      return value * 1024;
    default:
      // Assume GB if no unit
      return value;
  }
};

const TemplatesTableRow: FC<ITemplatesTableRowProps> = ({ ...props }) => {
  const {
    template,
    role,
    totalInstances,
    createInstance,
    editTemplate,
    deleteTemplate,
    deleteTemplateLoading,
    expandRow,
    templates,  // All templates in the workspace
    workspaceQuota, // Add this to props
    isPersonal,
  } = props;

  // Check if we can create an instance based on resources
  const canCreate = canCreateInstance(template, templates, workspaceQuota);
  
  const {
    data: labelsData,
    loading: loadingLabels,
    error: labelsError,
  } = useNodesLabelsQuery({ fetchPolicy: 'no-cache' });

  const { data, refreshClock } = useContext(TenantContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const { refetch: refetchInstancesLabelSelector } =
    useInstancesLabelSelectorQuery({
      onError: apolloErrorCatcher,
      variables: {
        labels: `crownlabs.polito.it/template=${template.id},crownlabs.polito.it/workspace=${template.workspaceName}`,
      },
      skip: true,
      fetchPolicy: 'network-only',
    });

  const [showDeleteModalNotPossible, setShowDeleteModalNotPossible] =
    useState(false);
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);
  const [createDisabled, setCreateDisabled] = useState(false);

  // Modal state
  const [showEditModal, setShowEditModal] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<TemplateType | undefined>(undefined);

  // Update mutation
  const [applyTemplateMutation, { loading: updateLoading }] = useApplyTemplateMutation({
    onError: apolloErrorCatcher,
  });

  const createInstanceHandler = useCallback(() => {
    setCreateDisabled(true);
    createInstance(template.id)
      .then(() => {
        refreshClock();
        setTimeout(setCreateDisabled, 400, false);
        expandRow(template.id, true);
      })
      .catch(() => setCreateDisabled(false));
  }, [createInstance, expandRow, refreshClock, template.id]);

  const handleEditTemplate = (template: TemplateType) => {
    setSelectedTemplate(template);
    setShowEditModal(true);
  };

  // Handler to submit the update mutation
  const handleUpdateTemplate = async (updatedTemplate: TemplateType) => {
    // Build the patch JSON for the template update
    const patch = {
      spec: {
        prettyName: updatedTemplate.name,
        description: updatedTemplate.description ?? '', // Add description if needed
        environmentList: [
          {
            name: updatedTemplate.name,
            image: updatedTemplate.image,
            guiEnabled: updatedTemplate.gui,
            persistent: updatedTemplate.persistent,
            mountMyDriveVolume: updatedTemplate.mountMyDrive,
            environmentType: updatedTemplate.vmorcontainer,
            resources: {
              cpu: updatedTemplate.cpu,
              memory: `${updatedTemplate.ram}Gi`,
              disk: updatedTemplate.persistent ? `${updatedTemplate.disk}Gi` : undefined,
            },
            sharedVolumeMounts: updatedTemplate.sharedVolumeMountInfos?.map(sv => ({
              sharedVolume: {
                namespace: sv.sharedVolume.namespace,
                name: sv.sharedVolume.name,
              },
              mountPath: sv.mountPath,
              readOnly: sv.readOnly,
            })),
          },
        ],
        workspaceCrownlabsPolitoItWorkspaceRef: {
          name: template.workspaceName,
          namespace: template.workspaceNamespace,
        },
      },
    };

    return applyTemplateMutation({
      variables: {
        templateId: template.id,
        workspaceNamespace: template.workspaceNamespace,
        patchJson: JSON.stringify(patch),
        manager: 'web-frontend', // or your manager string
      },
    });
  };

  const instancesLimit = data?.tenant?.status?.quota?.instances ?? 1;

  return (
    <>
      <ModalAlert
        headTitle={template.name}
        message="Cannot delete this template"
        description="A template with active instances cannot be deleted. Please delete al the instances associated with this template."
        type="warning"
        buttons={[
          <Button
            key={0}
            shape="round"
            className="w-24"
            type="primary"
            onClick={() => setShowDeleteModalNotPossible(false)}
          >
            Close
          </Button>,
        ]}
        show={showDeleteModalNotPossible}
        setShow={setShowDeleteModalNotPossible}
      />
      <ModalAlert
        headTitle={template.name}
        message="Delete template"
        description="Do you really want to delete this template?"
        type="warning"
        buttons={[
          <Button
            key={0}
            shape="round"
            className="mr-2 w-24"
            type="primary"
            onClick={() => setShowDeleteModalConfirm(false)}
          >
            Close
          </Button>,
          <Button
            key={1}
            shape="round"
            className="ml-2 w-24"
            color="danger"
            loading={deleteTemplateLoading}
            onClick={() =>
              deleteTemplate(template.id)
                .then(() => setShowDeleteModalConfirm(false))
                .catch(console.warn)
            }
          >
            {!deleteTemplateLoading && 'Delete'}
          </Button>,
        ]}
        show={showDeleteModalConfirm}
        setShow={setShowDeleteModalConfirm}
      />
      <ModalCreateTemplate
        show={showEditModal}
        setShow={setShowEditModal}
        template={selectedTemplate}
        cpuInterval={{ min: 1, max: 8 }} // <-- Pass actual intervals
        ramInterval={{ min: 1, max: 32 }}
        diskInterval={{ min: 10, max: 100 }}
        workspaceNamespace={template.workspaceNamespace}
        submitHandler={handleUpdateTemplate}
        loading={updateLoading}
        isPersonal={isPersonal}
      />
      <div className="w-full flex justify-between py-0">
        <div
          className="flex w-full items-center cursor-pointer"
          onClick={() => expandRow(template.id, false)}
        >
          <Space size="middle">
            <div className="flex items-center">
              {template.gui ? (
                <DesktopOutlined
                  style={{ fontSize: '24px', color: '#1c7afd' }}
                />
              ) : (
                <CodeOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
              )}
              <label className="ml-3 cursor-pointer">{template.name}</label>
              {template.persistent && (
                <Tooltip
                  title={
                    <>
                      <div className="text-center">
                        These Instances can be stopped and restarted without
                        being deleted.
                      </div>
                      <div className="text-center">
                        Your files won't be deleted in case of an internal
                        disservice of CrownLabs.
                      </div>
                    </>
                  }
                >
                  <div className="success-color-fg ml-3 flex items-center">
                    <SvgInfinite width="22px" />
                  </div>
                </Tooltip>
              )}
              {template.nodeSelector && (
                <div className="ml-3 flex items-center">
                  <NodeSelectorIcon
                    isOnWorkspace={true}
                    nodeSelector={template.nodeSelector}
                  />
                </div>
              )}
            </div>
          </Space>
        </div>
        <Space size="small">
          {template.instances.length ? (
            <Badge
              count={template.instances.length}
              color="blue"
              className="mx-2"
            />
          ) : (
            ''
          )}
          <Tooltip
            placement="left"
            title={
              <>
                <div>
                  CPU: {template.resources.cpu || 'unavailable'} core(s)
                </div>
                <div>
                  RAM: {convertMemory(template.resources.memory) || 'unavailable'}B
                </div>
                <div>
                  {template.persistent
                    ? ` DISK: ${
                        convertMemory(template.resources.disk) || 'unavailable'
                      }B`
                    : ``}
                </div>
              </>
            }
          >
            <Button type="link" color="orange" size="middle" className="px-0">
              Info
            </Button>
          </Tooltip>
          {role === WorkspaceRole.manager ? (
            <TemplatesTableRowSettings
              id={template.id}
              template={template} // <-- Pass the template object
              createInstance={createInstance}
              editTemplate={handleEditTemplate} // <-- Pass handler
              deleteTemplate={() => {
                refetchInstancesLabelSelector()
                  .then(ils => {
                    if (!ils.data.instanceList?.instances!.length && !ils.error)
                      setShowDeleteModalConfirm(true);
                    else setShowDeleteModalNotPossible(true);
                  })
                  .catch(console.warn);
              }}
            />
          ) : (
            <Tooltip placement="top" title="Create Instance">
              <Button
                onClick={createInstanceHandler}
                className="xs:hidden block"
                type="link"
                color="primary"
                size="large"
                icon={<PlayCircleOutlined style={{ fontSize: '22px' }} />}
              />
            </Tooltip>
          )}
          {instancesLimit === totalInstances || !canCreate ? (
            <Tooltip
              overlayClassName="w-44"
              title={
                <>
                  <div className="text-center">
                    {!canCreate ? (
                      <>
                        <div>Not enough resources available.</div>
                        <div>Check your workspace quota usage.</div>
                      </>
                    ) : (
                      <>
                        <div>You have <b>reached your limit</b> of {instancesLimit} instances</div>
                        <div className="text-center mt-2">
                          Please <b>delete</b> an instance to create a new one
                        </div>
                      </>
                    )}
                  </div>
                </>
              }
            >
              <span className="cursor-not-allowed">
                <Button
                  onClick={createInstanceHandler}
                  className="hidden xs:block pointer-events-none"
                  disabled={totalInstances === instancesLimit || createDisabled || !canCreate}
                  type="primary"
                  shape="round"
                  size={'middle'}
                >
                  Create
                </Button>
              </span>
            </Tooltip>
          ) : template.nodeSelector &&
            JSON.stringify(template.nodeSelector) === '{}' ? (
            <Dropdown.Button
              menu={{
                items:
                  loadingLabels || labelsError
                    ? [
                        {
                          key: 'error',
                          label: loadingLabels
                            ? 'Loading labels...'
                            : 'Error loading labels',
                          disabled: true,
                        },
                      ]
                    : labelsData?.labels?.map(({ key, value }) => ({
                        key: JSON.stringify({ [key]: value }),
                        label: `${cleanupLabels(key)}=${value}`,
                        disabled: loadingLabels,
                        onClick: () => {
                          createInstance(template.id, JSON.parse(key))
                            .then(() => {
                              refreshClock();
                              setTimeout(setCreateDisabled, 400, false);
                              expandRow(template.id, true);
                            })
                            .catch(() => setCreateDisabled(false));
                        },
                      })) || [],
              }}
              onClick={createInstanceHandler}
              disabled={totalInstances === instancesLimit || createDisabled || !canCreate}
              type="primary"
              size={'middle'}
            >
              Create
            </Dropdown.Button>
          ) : (
            <Button
              onClick={createInstanceHandler}
              className="hidden xs:block"
              disabled={totalInstances === instancesLimit || createDisabled || !canCreate}
              type="primary"
              shape="round"
              size={'middle'}
            >
              Create
            </Button>
          )}
        </Space>
      </div>
    </>
  );
};

export default TemplatesTableRow;
