import {
  CodeOutlined,
  DesktopOutlined,
  PlayCircleOutlined,
  SelectOutlined,
  AppstoreAddOutlined,
  DockerOutlined,
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
import { TenantContext } from '../../../../contexts/TenantContext';
import {
  OwnedInstancesContext,
  type IQuota,
} from '../../../../contexts/OwnedInstancesContext';
import type { Template } from '../../../../utils';
import { cleanupLabels, convertToGB, WorkspaceRole } from '../../../../utils';
import { ModalAlert } from '../../../common/ModalAlert';
import { TemplatesTableRowSettings } from '../TemplatesTableRowSettings';
import NodeSelectorIcon from '../../../common/NodeSelectorIcon/NodeSelectorIcon';

export interface ITemplatesTableRowProps {
  template: Template;
  role: WorkspaceRole;
  totalInstances: number;
  tenantNamespace: string;
  isPersonal?: boolean;
  workspaceName: string;
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

const canCreateInstance = (
  template: Template,
  availableQuota: IQuota,
): boolean => {
  // If no quota defined, default to allowing creation
  if (!availableQuota) return true;

  const templateCpu = template.resources?.cpu || 0;
  const templateMemory = convertToGB(template.resources?.memory || '0');

  // TODO: check for disk quota when implemented
  return (
    availableQuota.instances >= 1 &&
    availableQuota.cpu >= templateCpu &&
    availableQuota.memory >= templateMemory
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
    isPersonal,
    workspaceName,
  } = props;

  // Get the available quota in the workspace from the OwnedInstancesContext
  const { availableQuota } = useContext(OwnedInstancesContext);
  const workspaceAvailableQuota: IQuota = availableQuota?.[workspaceName] || {
    instances: 0,
    cpu: 0,
    memory: 0,
    disk: 0,
  };

  const canCreate = canCreateInstance(template, workspaceAvailableQuota);

  const {
    data: labelsData,
    loading: loadingLabels,
    error: labelsError,
  } = useNodesLabelsQuery({ fetchPolicy: 'no-cache' });

  const { data, refreshClock } = useContext(TenantContext);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const { refetch: refetchOwnedInstances, data: instancesData } = useContext(
    OwnedInstancesContext,
  );

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

  const handleEditTemplate = () => {
    window.dispatchEvent(
      new CustomEvent('openTemplateModal', { detail: template }),
    );
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
                .then(() => {
                  setShowDeleteModalConfirm(false);
                })
                .catch(error => {
                  const isNotFoundError =
                    error?.graphQLErrors?.[0]?.extensions?.statusCode === 404 ||
                    error?.graphQLErrors?.[0]?.extensions?.responseBody
                      ?.code === 404 ||
                    error?.graphQLErrors?.[0]?.message?.includes('not found');

                  if (isNotFoundError) {
                    setShowDeleteModalConfirm(false);
                    console.info(
                      `Template ${template.id} was successfully deleted (404 is expected for DELETE operations)`,
                    );
                  } else {
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
      <div className="w-full flex justify-between py-0">
        <div
          className="flex w-full items-center cursor-pointer"
          onClick={() => expandRow(template.id, false)}
        >
          <Space size="middle">
            <div className="flex items-center">
              {template.hasMultipleEnvironments ? (
                <Tooltip
                  placement="right"
                  title={
                    <div className="p-2">
                      <div className="font-semibold mb-2 text-center">
                        Multiple Environments ({template.environmentList.length}
                        )
                      </div>
                      {template.environmentList.map((env, index) => (
                        <div key={index} className="p-1">
                          <div className="flex items-center gap-2 mb-1">
                            <span className="font-medium">{env.name}</span>
                            {env.guiEnabled ? (
                              <div className="flex items-center gap-1.5">
                                <DesktopOutlined
                                  style={{ fontSize: '14px', color: '#1c7afd' }}
                                />
                                <span className="text-xs">VM GUI</span>
                                {env.persistent && (
                                  <>
                                    <SvgInfinite
                                      width="14px"
                                      className="success-color-fg ml-1"
                                    />
                                    <span className="text-xs">Persistent</span>
                                  </>
                                )}
                              </div>
                            ) : env.environmentType === 'Container' ? (
                              <div className="flex items-center gap-1.5">
                                <DockerOutlined
                                  style={{ fontSize: '14px', color: '#1c7afd' }}
                                />
                                <span className="text-xs">Container SSH</span>
                                {env.persistent && (
                                  <>
                                    <SvgInfinite
                                      width="14px"
                                      className="success-color-fg ml-1"
                                    />
                                    <span className="text-xs">Persistent</span>
                                  </>
                                )}
                              </div>
                            ) : (
                              <div className="flex items-center gap-1.5">
                                <CodeOutlined
                                  style={{ fontSize: '14px', color: '#1c7afd' }}
                                />
                                <span className="text-xs">VM SSH</span>
                                {env.persistent && (
                                  <>
                                    <SvgInfinite
                                      width="14px"
                                      className="success-color-fg ml-1"
                                    />
                                    <span className="text-xs">Persistent</span>
                                  </>
                                )}
                              </div>
                            )}
                          </div>
                        </div>
                      ))}
                    </div>
                  }
                >
                  <Badge
                    count={template.environmentList.length}
                    size="small"
                    color="green"
                    offset={[3, -3]}
                  >
                    <AppstoreAddOutlined
                      style={{ fontSize: '24px', color: '#1c7afd' }}
                    />
                  </Badge>
                </Tooltip>
              ) : template.gui ? (
                <DesktopOutlined
                  style={{ fontSize: '24px', color: '#1c7afd' }}
                />
              ) : (
                <CodeOutlined style={{ fontSize: '24px', color: '#1c7afd' }} />
              )}
              <label className="ml-3 cursor-pointer">
                <Space>
                  {template.name}
                  {!template.hasMultipleEnvironments &&
                    template.allowPublicExposure && (
                      <Tooltip title="Public Port Exposure - This template allows exposing internal ports to external networks for remote access">
                        <SelectOutlined className="text-fuchsia-400" />
                      </Tooltip>
                    )}
                  {template.hasMultipleEnvironments && (
                    <span className="text-xs text-gray-500 ml-2">
                      ({template.environmentList.length} environments)
                    </span>
                  )}
                </Space>
              </label>
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
                {template.hasMultipleEnvironments ? (
                  <div>
                    <div className="font-semibold mb-2">
                      Multiple Environments ({template.environmentList.length}):
                    </div>
                    {template.environmentList.map((env, index) => (
                      <div
                        key={index}
                        className="mb-2 p-2 border-l-2 border-blue-300"
                      >
                        <div className="font-medium">Env: {env.name}</div>
                        <div>GUI: {env.guiEnabled ? 'Yes' : 'No'}</div>
                        <div>CPU: {env.resources.cpu} core(s)</div>
                        <div>
                          RAM:{' '}
                          {convertMemory(env.resources.memory) || 'unavailable'}
                          B
                        </div>
                        {env.persistent && (
                          <div>
                            DISK:{' '}
                            {convertMemory(env.resources.disk) || 'unavailable'}
                            B
                          </div>
                        )}
                      </div>
                    ))}
                    <div className="mt-2 pt-2 border-t border-gray-300">
                      <div className="font-medium">Total Resources:</div>
                      <div>Total CPU: {template.resources.cpu} core(s)</div>
                      <div>
                        Total RAM: {convertMemory(template.resources.memory)}B
                      </div>
                      {template.persistent && (
                        <div>
                          Total DISK: {convertMemory(template.resources.disk)}B
                        </div>
                      )}
                    </div>
                  </div>
                ) : (
                  <>
                    <div>
                      CPU: {template.resources.cpu || 'unavailable'}vCPU(s)
                    </div>
                    <div>
                      RAM:{' '}
                      {convertMemory(template.resources.memory) ||
                        'unavailable'}
                      B
                    </div>
                    <div>
                      {template.persistent
                        ? ` DISK: ${
                            convertMemory(template.resources.disk) ||
                            'unavailable'
                          }B`
                        : ``}
                    </div>
                  </>
                )}
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
                if (isPersonal) {
                  // For personal templates, use data from context
                  refetchOwnedInstances().then(() => {
                    const allInstances =
                      instancesData?.instanceList?.instances || [];
                    const instances = allInstances.filter(
                      (instance): instance is NonNullable<typeof instance> =>
                        instance != null &&
                        instance?.spec?.templateCrownlabsPolitoItTemplateRef
                          ?.name === template.id,
                    );

                    if (!instances?.length) setShowDeleteModalConfirm(true);
                    else setShowDeleteModalNotPossible(true);
                  });
                } else {
                  // For workspace templates, use label selector query
                  refetchInstancesLabelSelector()
                    .then(ils => {
                      const instances =
                        ils?.data?.instanceList?.instances || [];

                      if (!instances?.length && !ils?.error)
                        setShowDeleteModalConfirm(true);
                      else setShowDeleteModalNotPossible(true);
                    })
                    .catch(console.warn);
                }
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
              classNames={{ root: 'w-44' }}
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
                          setCreateDisabled(true);
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
