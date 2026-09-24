import {
  type FC,
  type SetStateAction,
  useContext,
  useMemo,
  useState,
} from 'react';
import { Dropdown, Badge, Space, Tooltip } from 'antd';
import { Button } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import {
  SelectOutlined,
  CodeOutlined,
  DeleteOutlined,
  FolderOpenOutlined,
  MoreOutlined,
  PoweroffOutlined,
  CaretRightOutlined,
  ExclamationCircleOutlined,
  ExportOutlined,
  CameraOutlined,
} from '@ant-design/icons';
import type { Instance } from '../../../../utils';
import {
  EnvironmentType,
  Phase2,
  Role,
  useApplyInstanceMutation,
  useCreateWorkspaceImageMutation,
  useDeleteInstanceMutation,
  useWorkspacesQuery,
} from '../../../../generated-types';
import {
  makeImageNotificationContent,
  setInstanceRunning,
} from '../../../../utilsLogic';
import { ErrorContext } from '../../../../errorHandling/ErrorContext';
import {
  VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM,
  VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX,
  VITE_APP_CROWNLABS_PUBLIC_REGISTRY_NAME_DESTINATION,
  VITE_APP_MYDRIVE_WORKSPACE_NAME,
} from '../../../../env';
import { TenantContext } from '../../../../contexts/TenantContext';
import { AuthContext } from '../../../../contexts/AuthContext';
import {
  ImageCreationModal,
  type ImageDestinationSelection,
} from '../../../common/ImageCreationModal';
import { OwnedInstancesContext } from '../../../../contexts/OwnedInstancesContext';

export interface IRowInstanceActionsDropdownProps {
  instance: Instance;
  fileManager?: boolean;
  extended: boolean;
  setSshModal: React.Dispatch<SetStateAction<boolean>>;
  onEnablePublicExposure?: () => void;
}

const RowInstanceActionsDropdown: FC<IRowInstanceActionsDropdownProps> = ({
  ...props
}) => {
  const {
    instance,
    fileManager,
    extended,
    setSshModal,
    onEnablePublicExposure,
  } = props;

  const {
    status,
    persistent,
    url,
    name,
    tenantNamespace,
    environmentType,
    gui,
  } = instance;

  const font20px = { fontSize: '20px' };

  const createImageResourceName = (imageName: string) => {
    const now = new Date();
    const twoDigits = (value: number) => value.toString().padStart(2, '0');
    const timestamp = `${now.getFullYear()}${twoDigits(
      now.getMonth() + 1,
    )}${twoDigits(now.getDate())}-${twoDigits(now.getHours())}${twoDigits(
      now.getMinutes(),
    )}${twoDigits(now.getSeconds())}`;

    return `${imageName}-${timestamp}`;
  };

  const [disabled, setDisabled] = useState(false);
  const [imageModalOpen, setImageModalOpen] = useState(false);
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [deleteInstanceMutation] = useDeleteInstanceMutation({
    onError: apolloErrorCatcher,
  });
  const [applyInstanceMutation] = useApplyInstanceMutation({
    onError: apolloErrorCatcher,
  });
  const [createWorkspaceImage, { loading: creatingWorkspaceImage }] =
    useCreateWorkspaceImageMutation({
      onError: apolloErrorCatcher,
    });
  const navigate = useNavigate();
  const { data: tenantData, notify } = useContext(TenantContext);
  const tenantName = tenantData?.tenant?.metadata?.name;
  const { profile } = useContext(AuthContext);
  const { instances: ownedInstances } = useContext(OwnedInstancesContext);

  const imageCreationPermissions = useMemo(() => {
    const isPersonalWorkspace =
      !instance.workspaceName || instance.workspaceName === 'personal';
    const personalWorkspaceAvailable = Boolean(
      tenantData?.tenant?.spec?.personalWorkspace &&
        tenantData?.tenant?.status?.personalNamespace?.created,
    );
    const isWorkspaceManager = isPersonalWorkspace
      ? personalWorkspaceAvailable
      : Boolean(
          tenantData?.tenant?.spec?.workspaces?.some(
            workspace =>
              workspace?.name === instance.workspaceName &&
              workspace.role === Role.Manager,
          ),
        );
    const adminGroup = `${VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX}:${VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM}`;
    const userGroups = (profile?.groups || []) as string[];
    const isClusterAdmin = userGroups.includes(adminGroup);
    const canUseWorkspaceDestinations = isWorkspaceManager || isClusterAdmin;
    const canPublishToPublicRegistry = Boolean(profile);

    return {
      canCreateImages:
        canUseWorkspaceDestinations || canPublishToPublicRegistry,
      canUseWorkspaceDestinations,
      canPublishToPublicRegistry,
      isClusterAdmin,
      isPersonalWorkspace,
      personalWorkspaceAvailable,
    };
  }, [instance.workspaceName, profile, tenantData?.tenant]);

  const { data: allWorkspacesData, loading: loadingAllWorkspaces } =
    useWorkspacesQuery({
      skip: !imageCreationPermissions.isClusterAdmin || !imageModalOpen,
      onError: apolloErrorCatcher,
    });

  const imageCreationAvailability = useMemo(() => {
    const latestInstance =
      ownedInstances.find(ownedInstance => ownedInstance.id === instance.id) ??
      instance;
    const hasSingleEnvironment = latestInstance.environments?.length === 1;
    const supportsImages =
      latestInstance.persistent &&
      (latestInstance.environmentType === EnvironmentType.VirtualMachine ||
        latestInstance.environmentType === EnvironmentType.CloudVm);

    return {
      showAction:
        hasSingleEnvironment &&
        supportsImages &&
        imageCreationPermissions.canCreateImages,
      canCreate:
        Boolean(tenantName) &&
        hasSingleEnvironment &&
        imageCreationPermissions.canCreateImages &&
        supportsImages &&
        Boolean(latestInstance.environments?.[0]?.name) &&
        latestInstance.status === Phase2.Off,
    };
  }, [
    imageCreationPermissions.canCreateImages,
    instance,
    ownedInstances,
    tenantName,
  ]);

  // Check if user has access to utilities workspace
  const hasUtilitiesAccess = Boolean(
    tenantData?.tenant?.spec?.workspaces?.some(
      ws => ws?.name === VITE_APP_MYDRIVE_WORKSPACE_NAME,
    ),
  );

  const handleDriveClick = () => {
    navigate('/drive');
  };

  const mutateInstanceStatus = async (running: boolean) => {
    if (!disabled) {
      setDisabled(true);
      try {
        const result = await setInstanceRunning(
          running,
          instance,
          applyInstanceMutation,
        );
        if (result) setTimeout(setDisabled, 400, false);
      } catch {
        // TODO: do nothing at the moment
      }
    }
  };

  const statusComponents = {
    [Phase2.Ready]: {
      menuIcon: <PoweroffOutlined style={font20px} />,
      menuText: 'Stop',
      menuAction: () => mutateInstanceStatus(false),
    },
    [Phase2.Off]: {
      menuIcon: <CaretRightOutlined style={font20px} />,
      menuText: 'Start',
      menuAction: () => mutateInstanceStatus(true),
    },
    Other: {
      menuIcon: <ExclamationCircleOutlined style={font20px} />,
      menuText: '',
      menuAction: () => null,
    },
  } as const;

  const { menuIcon, menuText, menuAction } =
    status === Phase2.Ready
      ? statusComponents[Phase2.Ready]
      : status === Phase2.Off
        ? statusComponents[Phase2.Off]
        : statusComponents.Other;

  const isContainer =
    environmentType === EnvironmentType.Container ||
    environmentType === EnvironmentType.Standalone;

  const sshDisabled = status !== Phase2.Ready || isContainer;
  const fileManagerDisabled = status !== Phase2.Ready && isContainer;

  const connectDisabled = status !== Phase2.Ready || (isContainer && !gui);

  const getFirstEnvironmentName = () => {
    return instance.environments?.[0]?.name || 'env';
  };

  const buildSSHLink = (envName: string) => {
    if (envName) {
      return `/instance/${instance.tenantNamespace}/${instance.name}/${envName}/ssh`;
    }
    return `/instance/${instance.tenantNamespace}/${instance.name}/env/ssh`;
  };

  const defaultImageName = `image-of-${instance.name}`
    .slice(0, 47)
    .replace(/-+$/, '');

  const currentWorkspaceNamespace = useMemo(() => {
    if (imageCreationPermissions.isPersonalWorkspace) {
      return instance.tenantNamespace;
    }

    const currentWorkspace = tenantData?.tenant?.spec?.workspaces?.find(
      workspace => workspace?.name === instance.workspaceName,
    );

    return (
      currentWorkspace?.workspaceWrapperTenantV1alpha2
        ?.itPolitoCrownlabsV1alpha1Workspace?.status?.namespace?.name ??
      `workspace-${instance.workspaceName}`
    );
  }, [
    imageCreationPermissions.isPersonalWorkspace,
    instance.tenantNamespace,
    instance.workspaceName,
    tenantData?.tenant?.spec?.workspaces,
  ]);

  const otherWorkspaceOptions = useMemo(() => {
    if (imageCreationPermissions.isClusterAdmin) {
      return (allWorkspacesData?.workspaces?.items ?? [])
        .filter(workspace =>
          Boolean(
            workspace?.metadata?.name && !workspace.metadata.deletionTimestamp,
          ),
        )
        .map(workspace => ({
          value: `workspace-${workspace?.metadata?.name}`,
          label: workspace?.spec?.prettyName ?? workspace?.metadata?.name ?? '',
        }))
        .filter(option => option.label && option.value);
    }

    return (tenantData?.tenant?.spec?.workspaces ?? [])
      .filter(workspace => workspace?.role === Role.Manager)
      .map(workspace => {
        const workspaceResource =
          workspace?.workspaceWrapperTenantV1alpha2
            ?.itPolitoCrownlabsV1alpha1Workspace;
        return {
          value:
            workspaceResource?.status?.namespace?.name ??
            `workspace-${workspace?.name}`,
          label: workspaceResource?.spec?.prettyName ?? workspace?.name ?? '',
        };
      })
      .filter(option => option.label && option.value);
  }, [
    allWorkspacesData?.workspaces?.items,
    imageCreationPermissions.isClusterAdmin,
    tenantData?.tenant?.spec?.workspaces,
  ]);

  const createImage = async (selection: ImageDestinationSelection) => {
    const destinationNamespace =
      selection.destination === 'public-registry'
        ? VITE_APP_CROWNLABS_PUBLIC_REGISTRY_NAME_DESTINATION
        : selection.destination === 'this-workspace'
          ? currentWorkspaceNamespace
          : selection.workspace === 'personal'
            ? instance.tenantNamespace
            : selection.workspace;
    const environmentName = instance.environments?.[0]?.name;

    if (!destinationNamespace || !environmentName || !tenantName) return;

    await createWorkspaceImage({
      variables: {
        destinationNamespace,
        sourceInstanceName: instance.name,
        sourceInstanceNamespace: instance.tenantNamespace,
        environmentName,
        imageName: selection.imageName,
        resourceName: createImageResourceName(selection.imageName),
        tenantName,
        description: selection.description,
      },
    });

    const targetWorkspaceLabel =
      selection.workspace === 'personal'
        ? 'your Personal Workspace'
        : otherWorkspaceOptions.find(
            option => option.value === selection.workspace,
          )?.label;
    const notification =
      selection.destination === 'public-registry'
        ? 'The image will be published to the public registry shortly.'
        : selection.destination === 'this-workspace'
          ? 'Your new image will be available under “Images” in this workspace.'
          : `Your new image will be available under “Images” in ${
              targetWorkspaceLabel ?? 'the selected workspace'
            }.`;

    notify(
      'success',
      `image/${instance.id}/${selection.imageName}/${selection.destination}`,
      makeImageNotificationContent(selection.imageName, notification),
    );
    setImageModalOpen(false);
  };

  return (
    <>
      {extended && imageCreationAvailability.canCreate && (
        <Tooltip title="Create New Image">
          <Button
            className={`hidden ${
              !sshDisabled || fileManager ? 'xl:block' : 'sm:block'
            } py-0 border-0`}
            type="link"
            shape="circle"
            size="middle"
            aria-label="Create New Image"
            icon={
              <CameraOutlined
                className="flex justify-center items-center"
                style={{ fontSize: '22px', color: '#fff' }}
              />
            }
            disabled={creatingWorkspaceImage}
            onClick={() => setImageModalOpen(true)}
          />
        </Tooltip>
      )}
      <Dropdown
        trigger={['click']}
        menu={{
          items: [
            {
              key: 'connect',
              label: 'Connect',
              disabled: connectDisabled,
              icon: <SelectOutlined style={font20px} />,
              onClick: gui
                ? () => window.open(url!, '_blank')
                : () => setSshModal(true),
              className: `flex items-center sm:hidden ${
                !connectDisabled
                  ? extended
                    ? 'primary-color-fg'
                    : 'success-color-fg xs:hidden'
                  : 'pointer-events-none'
              }`,
            },
            persistent
              ? {
                  key: 'persistent',
                  label: menuText,
                  icon: menuIcon,
                  onClick: () => menuAction,
                  className: `flex items-center ${
                    extended ? ' sm:hidden' : 'xs:hidden'
                  }`,
                }
              : null,
            {
              type: 'divider',
              className: `${extended ? 'sm:hidden' : 'xs:hidden'}`,
            },
            ...(onEnablePublicExposure
              ? [
                  {
                    key: 'expose',
                    label: (
                      <Space align="center">
                        Port Exposure
                        {instance.publicExposure &&
                          (instance.publicExposure?.ports ?? []).length > 0 && (
                            <Badge
                              count={
                                (instance.publicExposure?.ports ?? []).length
                              }
                              showZero={false}
                              size="small"
                            />
                          )}
                      </Space>
                    ),
                    icon: <SelectOutlined style={font20px} />,
                    onClick: () => onEnablePublicExposure?.(),
                  },
                  {
                    type: 'divider' as const,
                  },
                ]
              : []),
            {
              key: 'ssh',
              icon: <CodeOutlined style={font20px} />,
              disabled: sshDisabled,
              label: (
                <>
                  SSH
                  {/* Only show direct link button if there's exactly one environment */}
                  {(!instance.environments ||
                    instance.environments.length === 1) && (
                    <Button
                      disabled={sshDisabled}
                      type="link"
                      className="ml-3"
                      color="primary"
                      variant="solid"
                      shape="circle"
                      size="small"
                      icon={
                        <Link
                          to={buildSSHLink(getFirstEnvironmentName())}
                          target="_blank"
                          rel="noopener noreferrer"
                          onClick={e => e.stopPropagation()}
                          style={{
                            color: 'inherit',
                            display: 'flex',
                            alignItems: 'center',
                          }}
                        >
                          <span style={{ filter: 'drop-shadow(0 0 0 black)' }}>
                            <ExportOutlined style={{ fontSize: 15 }} />
                          </span>
                        </Link>
                      }
                    ></Button>
                  )}
                </>
              ),
              onClick: () => setSshModal(true),
              className: `flex items-center ${extended ? 'xl:hidden' : ''} ${sshDisabled ? 'pointer-events-none' : ''}`,
            },
            imageCreationAvailability.showAction
              ? {
                  key: 'new-image',
                  icon: <CameraOutlined style={font20px} />,
                  disabled: !imageCreationAvailability.canCreate,
                  label: 'Create New Image',
                  onClick: () => setImageModalOpen(true),
                  className: `flex items-center ${
                    !imageCreationAvailability.canCreate
                      ? 'cursor-not-allowed'
                      : ''
                  }`,
                }
              : null,
            ...(hasUtilitiesAccess &&
            environmentType === EnvironmentType.VirtualMachine
              ? [
                  {
                    key: 'upload',
                    label: 'Drive',
                    icon: <FolderOpenOutlined style={font20px} />,
                    disabled: fileManagerDisabled,
                    className: `flex items-center ${extended ? 'xl:hidden' : ''} `,
                    onClick: handleDriveClick,
                  },
                ]
              : []),
            {
              type: 'divider',
              className: `${extended ? 'sm:hidden' : 'xs:hidden'}`,
            },
            {
              key: 'destroy',
              label: 'Destroy',
              danger: true,
              icon: <DeleteOutlined style={font20px} />,
              onClick: () =>
                deleteInstanceMutation({
                  variables: {
                    instanceId: name,
                    tenantNamespace: tenantNamespace!,
                  },
                }),
              className: `flex items-center ${
                extended ? ' sm:hidden' : 'xs:hidden'
              }`,
            },
          ],
        }}
      >
        <Button
          className={`${
            extended
              ? !sshDisabled || fileManager
                ? 'xl:hidden'
                : 'sm:hidden'
              : ''
          } flex justify-center items-center`}
          color="default"
          type="link"
          shape="circle"
          size="middle"
          icon={<MoreOutlined className="flex items-center" style={font20px} />}
        />
      </Dropdown>
      <ImageCreationModal
        open={imageModalOpen}
        defaultImageName={defaultImageName}
        breadcrumbItems={[
          instance.workspaceName || 'personal',
          instance.prettyName || instance.name,
          ...(instance.environments?.[0]?.name
            ? [instance.environments[0].name]
            : []),
        ]}
        currentWorkspaceName={currentWorkspaceNamespace}
        isPersonalWorkspace={imageCreationPermissions.isPersonalWorkspace}
        personalWorkspaceAvailable={
          imageCreationPermissions.personalWorkspaceAvailable
        }
        canUseWorkspaceDestinations={
          imageCreationPermissions.canUseWorkspaceDestinations
        }
        canPublishToPublicRegistry={
          imageCreationPermissions.canPublishToPublicRegistry
        }
        canCreateImage={imageCreationAvailability.canCreate}
        creating={creatingWorkspaceImage}
        loadingWorkspaces={loadingAllWorkspaces}
        otherWorkspaceOptions={otherWorkspaceOptions}
        onCancel={() => setImageModalOpen(false)}
        onCreate={selection =>
          void createImage(selection).catch(() => undefined)
        }
      />
    </>
  );
};

export default RowInstanceActionsDropdown;
