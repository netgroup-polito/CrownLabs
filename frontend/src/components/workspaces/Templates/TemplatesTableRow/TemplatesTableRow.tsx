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
  useOwnedInstancesQuery,
} from '../../../../generated-types';
import { TenantContext } from '../../../../contexts/TenantContext';
import type { Template } from '../../../../utils';
import { cleanupLabels, WorkspaceRole } from '../../../../utils';
import { ModalAlert } from '../../../common/ModalAlert';
import { TemplatesTableRowSettings } from '../TemplatesTableRowSettings';
import NodeSelectorIcon from '../../../common/NodeSelectorIcon/NodeSelectorIcon';
import ModalCreateTemplate, {
  type Template as TemplateType,
} from '../../ModalCreateTemplate/ModalCreateTemplate';
import { useApplyTemplateMutation } from '../../../../generated-types';

export interface ITemplatesTableRowProps {
  template: Template;
  role: WorkspaceRole;
  totalInstances: number;
  tenantNamespace: string;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
  isPersonal?: boolean;
  deleteTemplate: (
    id: string,
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
    labelSelector?: JSON,
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

// Helper function to parse memory string (e.g., "4Gi" -> 4)
const parseMemory = (memoryStr: string | number): number => {
  if (typeof memoryStr === 'number') return memoryStr;
  if (!memoryStr) return 0;

  const match = String(memoryStr).match(/^(\d+(?:\.\d+)?)(.*)?$/);
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

const canCreateInstance = (
  template: Template,
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  },
): boolean => {
  // If no quota defined, default to allowing creation
  if (!availableQuota) return true;

  const templateCpu = template.resources?.cpu || 0;
  const availableCpu =
    availableQuota.cpu !== undefined
      ? typeof availableQuota.cpu === 'string'
        ? parseFloat(availableQuota.cpu)
        : availableQuota.cpu
      : 0;

  const templateMemory = parseMemory(template.resources?.memory || '0');
  const availableMemory =
    availableQuota.memory !== undefined
      ? parseMemory(availableQuota.memory)
      : 0;

  const availableInstances =
    availableQuota.instances !== undefined ? availableQuota.instances : 0;

  return (
    templateCpu <= availableCpu &&
    templateMemory <= availableMemory &&
    1 <= availableInstances
  );
};

const TemplatesTableRow: FC<ITemplatesTableRowProps> = ({ ...props }) => {
  const {
    template,
    role,
    totalInstances,
    createInstance,
    deleteTemplate,
    deleteTemplateLoading,
    expandRow,
    tenantNamespace,
    availableQuota,
    isPersonal,
  } = props;

  // Check if we can create an instance based on resources
  const canCreate = canCreateInstance(template, availableQuota);

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

  const { refetch: refetchOwnedInstances } = useOwnedInstancesQuery({
    onError: apolloErrorCatcher,
    variables: {
      tenantNamespace: tenantNamespace || '',
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
  const [selectedTemplate, setSelectedTemplate] = useState<
    TemplateType | undefined
  >(undefined);

  // Update mutation
  const [applyTemplateMutation, { loading: updateLoading }] =
    useApplyTemplateMutation({
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
        description: updatedTemplate.name || '', // Use name as description if no description field
        environmentList: [
          {
            name: updatedTemplate.name,
            image: updatedTemplate.image,
            guiEnabled: updatedTemplate.gui,
            persistent: updatedTemplate.persistent,
            mountMyDriveVolume: updatedTemplate.mountMyDrive,
            environmentType: updatedTemplate.imageType, // Use imageType instead of vmorcontainer
            resources: {
              cpu: updatedTemplate.cpu,
              memory: `${updatedTemplate.ram}Gi`,
              disk: updatedTemplate.persistent
                ? `${updatedTemplate.disk}Gi`
                : undefined,
            },
            sharedVolumeMounts: updatedTemplate.sharedVolumeMountInfos?.map(
              sv => ({
                sharedVolume: {
                  namespace: sv.sharedVolume.namespace,
                  name: sv.sharedVolume.name,
                },
                mountPath: sv.mountPath,
                readOnly: sv.readOnly,
              }),
            ),
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
                .catch(error => {
                  // Handle 404 errors gracefully (template deletion succeeded but GraphQL can't return the deleted object)
                  const isNotFoundError =
                    error?.graphQLErrors?.[0]?.extensions?.statusCode === 404 ||
                    error?.graphQLErrors?.[0]?.extensions?.responseBody
                      ?.code === 404 ||
                    error?.graphQLErrors?.[0]?.message?.includes('not found');

                  if (isNotFoundError) {
                    // Template was successfully deleted, just close the modal
                    setShowDeleteModalConfirm(false);
                    console.info(
                      `Template ${template.id} was successfully deleted (404 is expected for DELETE operations)`,
                    );
                  } else {
                    // For other errors, use the error handler
                    console.error('Delete template error:', error);
                    apolloErrorCatcher(error);
                  }
                })
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
        cpuInterval={{ min: 1, max: 8 }}
        ramInterval={{ min: 1, max: 32 }}
        diskInterval={{ min: 10, max: 100 }}
        workspaceNamespace={
          isPersonal ? tenantNamespace : template.workspaceNamespace
        }
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
                  RAM:{' '}
                  {convertMemory(template.resources.memory) || 'unavailable'}B
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
              template={template}
              createInstance={createInstance}
              editTemplate={handleEditTemplate}
              deleteTemplate={() => {
                const refetchQuery = isPersonal
                  ? refetchOwnedInstances
                  : refetchInstancesLabelSelector;

                refetchQuery()
                  .then(ils => {
                    let instances;

                    if (isPersonal) {
                      // Filter instances by current template for personal workspaces
                      const allInstances =
                        ils.data.instanceList?.instances || [];
                      instances = allInstances.filter(
                        instance =>
                          instance?.spec?.templateCrownlabsPolitoItTemplateRef
                            ?.name === template.id,
                      );
                    } else {
                      // For non-personal workspaces, use all instances (already filtered by label selector)
                      instances = ils.data.instanceList?.instances || [];
                    }

                    if (!instances?.length && !ils.error)
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
                        <div>
                          You have <b>reached your limit</b> of {instancesLimit}{' '}
                          instances
                        </div>
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
                  disabled={
                    totalInstances === instancesLimit ||
                    createDisabled ||
                    !canCreate
                  }
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
              disabled={
                totalInstances === instancesLimit ||
                createDisabled ||
                !canCreate
              }
              type="primary"
              size={'middle'}
            >
              Create
            </Dropdown.Button>
          ) : (
            <Button
              onClick={createInstanceHandler}
              className="hidden xs:block"
              disabled={
                totalInstances === instancesLimit ||
                createDisabled ||
                !canCreate
              }
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
