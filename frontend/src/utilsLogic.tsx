import {
  CheckCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import type { FetchResult, MutationFunctionOptions } from '@apollo/client';
import { Button } from 'antd';
import { type Dispatch, type SetStateAction } from 'react';
import type {
  ApplyInstanceMutation,
  Exact,
  ItPolitoCrownlabsV1alpha1Workspace,
  ItPolitoCrownlabsV1alpha2Instance,
  ItPolitoCrownlabsV1alpha2SharedVolume,
  ItPolitoCrownlabsV1alpha2Template,
  Maybe,
  OwnedInstancesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  WorkspacesListItem,
} from './generated-types';
import {
  AutoEnroll,
  Phase,
  Phase2,
  Phase5,
  UpdateType,
} from './generated-types';
import { getInstancePatchJson } from './graphql-components/utils';
import type {
  Instance,
  InstanceEnvironment,
  SharedVolume,
  Template,
  Workspace,
  WorkspacesAvailable,
} from './utils';
import { convertToGB, WorkspaceRole, WorkspacesAvailableAction } from './utils';
import type { DeepPartial } from '@apollo/client/utilities';
import type { JointContent } from 'antd/lib/message/interface';
import type { Notifier } from './contexts/TenantContext';

type Nullable<T> = T | null | undefined;

export enum SubObjType {
  Deletion,
  UpdatedInfo,
  PrettyName,
  Addition,
  Drop,
  PublicExposureChange,
}
interface ItPolitoCrownlabsV1alpha2TemplateAlias {
  original: Nullable<DeepPartial<ItPolitoCrownlabsV1alpha2Template>>;
  alias: {
    name: string;
    id: string;
  };
}
export const makeGuiTemplate = (
  tq: ItPolitoCrownlabsV1alpha2TemplateAlias,
): Template => {
  if (!tq.original) {
    throw new Error(
      'makeGuiTemplate() error: a required parameter is undefined',
    );
  }
  const environmentList = tq.original.spec?.environmentList ?? [];
  const hasMultipleEnvironments = environmentList.length > 1;

  // For backwards compatibility use the first environment for main properties
  const primaryEnvironment = environmentList[0];

  const hasGUI = environmentList.some(env => env?.guiEnabled);
  const hasPersistent = environmentList.some(env => env?.persistent);

  const aggregatedResources = environmentList.reduce(
    (acc, env) => {
      if (env?.resources) {
        acc.cpu += env.resources.cpu ?? 0;
        const memoryStr = env.resources.memory || '0';
        let memoryGB = 0;
        if (memoryStr.includes('G')) {
          memoryGB = parseInt(memoryStr.replace(/[^\d]/g, '')) || 0;
        } else if (memoryStr.includes('M')) {
          memoryGB = (parseInt(memoryStr.replace(/[^\d]/g, '')) || 0) / 1000;
        }

        const diskStr = env.resources.disk || '0';
        let diskGB = 0;
        if (diskStr.includes('G')) {
          diskGB = parseInt(diskStr.replace(/[^\d]/g, '')) || 0;
        } else if (diskStr.includes('M')) {
          diskGB = (parseInt(diskStr.replace(/[^\d]/g, '')) || 0) / 1000;
        }

        acc.memorySum += memoryGB;
        acc.diskSum += diskGB;
      }
      return acc;
    },
    { cpu: 0, memorySum: 0, diskSum: 0 },
  );

  return {
    id: tq.alias.id ?? '',
    name: tq.alias.name ?? '',
    gui: hasGUI,
    persistent: hasPersistent,
    nodeSelector: primaryEnvironment?.nodeSelector,
    resources: {
      cpu: aggregatedResources.cpu,
      memory:
        aggregatedResources.memorySum > 0
          ? `${aggregatedResources.memorySum}G`
          : '',
      disk:
        aggregatedResources.diskSum > 0
          ? `${aggregatedResources.diskSum}G`
          : '',
    },
    environmentList: environmentList.map(env => ({
      name: env?.name ?? '',
      guiEnabled: !!env?.guiEnabled,
      persistent: !!env?.persistent,
      environmentType: env?.environmentType,
      resources: {
        cpu: env?.resources?.cpu ?? 0,
        memory: env?.resources?.memory ?? '',
        disk: env?.resources?.disk ?? '',
      },
    })),
    hasMultipleEnvironments,
    workspaceName:
      tq.original.spec?.workspaceCrownlabsPolitoItWorkspaceRef?.name ?? '',
    instances: [],
    workspaceNamespace:
      'workspace-' +
      (tq.original.spec?.workspaceCrownlabsPolitoItWorkspaceRef?.name ?? ''),
    allowPublicExposure: tq.original.spec?.allowPublicExposure ?? false,
  } as Template;
};

export type InstanceLabels = {
  crownlabsPolitoItManagedBy?: string;
  crownlabsPolitoItPersistent?: string;
  crownlabsPolitoItTemplate?: string;
  crownlabsPolitoItWorkspace?: string;
};

export const getInstanceLabels = (
  i: DeepPartial<ItPolitoCrownlabsV1alpha2Instance>,
): InstanceLabels | undefined => i.metadata?.labels as InstanceLabels;

// Helper functions for type conversions
const safePhaseConversion = (phase: unknown): Phase => {
  return (phase as Phase) || Phase.Starting;
};

const safePhase2Conversion = (phase: unknown): Phase2 => {
  return (phase as Phase2) || Phase2.Running;
};

const safePhase5Conversion = (phase: unknown): Phase5 => {
  return (phase as Phase5) || Phase5.Pending;
};

const safeWorkspaceRoleConversion = (role: unknown): WorkspaceRole => {
  return (role as WorkspaceRole) || WorkspaceRole.user;
};

// Helper functions for public exposure logic
interface PublicExposurePort {
  name?: string;
  port?: number;
  targetPort?: number;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
}

interface PublicExposureSpec {
  ports?: PublicExposurePort[];
}

interface PublicExposureStatus {
  externalIP?: string;
  phase?: Phase;
  ports?: PublicExposurePort[];
}

const hasActivePublicExposure = (
  publicExposure: unknown,
  publicExposureStatus: unknown,
): boolean => {
  const spec = publicExposure as PublicExposureSpec;
  const status = publicExposureStatus as PublicExposureStatus;

  return Boolean(
    (spec?.ports && spec.ports.length > 0) ||
      (status?.ports &&
        status.ports.length > 0 &&
        safePhaseConversion(status.phase) !== Phase.Off),
  );
};

const mapPortToPortListItem = (p: unknown, specPort?: unknown) => {
  const port = p as PublicExposurePort;
  const spec = specPort as PublicExposurePort;

  // Use spec port information to preserve original user request
  // If specPort is 0, it means "Auto" was requested, so we preserve that info
  const isAutoPort = spec?.port === 0;

  return {
    name: port.name || '',
    port: port.port && port.port > 0 ? String(port.port) : '',
    targetPort: port.targetPort || 0,
    protocol: port.protocol || 'TCP',
    // Add metadata to distinguish between auto and manually requested ports
    isAutoPort: isAutoPort,
    specPort: spec?.port || 0, // Original spec.port value
  };
};

const buildPublicExposureObject = (
  publicExposure: unknown,
  publicExposureStatus: unknown,
) => {
  if (!hasActivePublicExposure(publicExposure, publicExposureStatus)) {
    return undefined;
  }

  const spec = publicExposure as PublicExposureSpec;
  const status = publicExposureStatus as PublicExposureStatus;

  // Normalize ports to [] if null/undefined
  const statusPorts = Array.isArray(status?.ports) ? status.ports : [];
  const specPorts = Array.isArray(spec?.ports) ? spec.ports : [];

  // Create a map for matching status ports with spec ports by targetPort
  const specPortsByTarget = new Map<number, PublicExposurePort>();
  specPorts.forEach(sp => {
    if (sp?.targetPort) {
      specPortsByTarget.set(sp.targetPort, sp);
    }
  });

  const portsToUse = statusPorts.length > 0 ? statusPorts : specPorts;

  return {
    externalIP: status?.externalIP || '',
    phase: safePhaseConversion(status?.phase),
    ports: portsToUse
      .filter((p): p is PublicExposurePort => p != null)
      .map(p => {
        // Find matching spec port by targetPort to preserve original spec.port info
        const matchingSpecPort = p.targetPort
          ? specPortsByTarget.get(p.targetPort)
          : undefined;
        return mapPortToPortListItem(p, matchingSpecPort);
      }),
  };
};

export const makeGuiInstance = (
  instance?: Nullable<DeepPartial<ItPolitoCrownlabsV1alpha2Instance>>,
  userId?: string,
  optional?: {
    workspaceName: string;
    templateName: string;
  },
) => {
  if (!instance || !userId) {
    throw new Error('getInstances() error: a required parameter is undefined');
  }

  const { metadata, spec, status } = instance;
  const { name, namespace: tenantNamespace } = metadata ?? {};
  const { running, prettyName, publicExposure } = spec ?? {};
  const { publicExposure: publicExposureStatus } = status ?? {};

  const templateName = spec?.templateCrownlabsPolitoItTemplateRef?.name;
  const templateSpec = spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec;
  const templatePrettyName = templateSpec?.prettyName || '';

  const templateEnvironmentList = templateSpec?.environmentList || [];
  const instanceStatusEnvironmentList = status?.environments || [];

  const environments = templateEnvironmentList.map(templateEnv => {
      const envStatus = instanceStatusEnvironmentList.find(
        env => env?.name === templateEnv?.name,
      );

      return {
        name: envStatus?.name ?? '',
        phase: envStatus?.phase,
        ip: envStatus?.ip,
        guiEnabled: templateEnv?.guiEnabled ?? false,
        persistent: templateEnv?.persistent ?? false,
        environmentType: templateEnv?.environmentType,
        quota: {
          cpu: templateEnv?.resources?.cpu || 0,
          memory: templateEnv?.resources?.memory
            ? convertToGB(templateEnv?.resources?.memory)
            : 0,
          disk: templateEnv?.resources?.disk
            ? convertToGB(templateEnv?.resources?.disk)
            : 0,
        },
      } as InstanceEnvironment;
    }) ?? [];

  const hasMultipleEnvironments = environments.length > 1;

  // For backwards compatibility, use the first environment for main properties
  const primaryEnvironment = (templateEnvironmentList ?? [])[0] ?? {};
  const primaryStatus = environments[0];

  const { guiEnabled, persistent, environmentType } = primaryEnvironment;

  // determine if public exposure is allowed from template spec
  const allowPublicExposure =
    spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
      ?.itPolitoCrownlabsV1alpha2Template?.spec?.allowPublicExposure ?? false;

  const instanceID = tenantNamespace + '/' + metadata?.name;

  const publicExposureObj = buildPublicExposureObject(
    publicExposure,
    publicExposureStatus,
  );
  // Normalize ports to [] if null/undefined
  if (publicExposureObj && !publicExposureObj.ports) {
    publicExposureObj.ports = [];
  }

  console.log("makeGuiInstance", prettyName, environments)

  return {
    id: instanceID,
    name: name,
    prettyName: prettyName,
    gui: guiEnabled,
    persistent: persistent,
    templatePrettyName: templatePrettyName,
    templateName: templateName ?? '',
    templateId: makeTemplateKey(
      getInstanceLabels(instance)?.crownlabsPolitoItTemplate ??
        optional?.templateName ??
        '',
      getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ??
        optional?.workspaceName ??
        '',
    ),
    environmentType: environmentType,
    ip: primaryStatus?.ip ?? status?.ip ?? '',
    status: safePhase2Conversion(primaryStatus?.phase ?? status?.phase),
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: userId,
    tenantNamespace: tenantNamespace,
    workspaceName:
      getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ?? '',
    running: running,
    nodeName: status?.nodeName,
    nodeSelector: status?.nodeSelector,
    allowPublicExposure,
    tenantDisplayName: userId, // Using userId as display name since tenant info is not available
    myDriveUrl: '',
    publicExposure: publicExposureObj,
    environments: environments,
    hasMultipleEnvironments,
    resources: {
      cpu: environments.reduce((acc, env) => acc + env.quota.cpu, 0),
      memory: environments.reduce((acc, env) => acc + env.quota.memory, 0),
      disk: environments.reduce((acc, env) => acc + env.quota.disk, 0),
    },
  } as Instance;
};

export const makeWorkspace = (
  workspace: Nullable<DeepPartial<WorkspacesListItem>>,
) => {
  if (!workspace) {
    throw new Error('getInstances() error: a required parameter is undefined');
  }

  const { name, role, workspaceWrapperTenantV1alpha2 } = workspace;
  const { spec, status } =
    workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace ?? {};

  return {
    name: name,
    namespace: status?.namespace?.name,
    prettyName: spec?.prettyName,
    role: safeWorkspaceRoleConversion(role),
    templates: [],
  } as Workspace;
};
interface InstancesSubscriptionData {
  subscriptionData: { data: OwnedInstancesQuery };
}
export const updateQueryOwnedInstancesQuery = (
  setDataInstances: Dispatch<SetStateAction<Instance[]>>,
  userId: string,
  _notifier: Notifier,
) => {
  return (
    prev: OwnedInstancesQuery,
    subscriptionDataObject: InstancesSubscriptionData,
  ) => {
    const { data } =
      subscriptionDataObject.subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

    if (!data?.updateInstance?.instance) return prev;

    const { instance: instanceK8s, updateType } = data.updateInstance;
    let shouldNotify = false;

    setDataInstances(old => {
      const instanceGui = makeGuiInstance(instanceK8s, userId);
      const objType = getSubObjTypeCustom(
        old.find(i => i.id === instanceGui.id),
        instanceGui,
        updateType,
      );

      switch (objType) {
        case SubObjType.Addition:
          shouldNotify = true;
          return [...old, instanceGui];
        case SubObjType.Deletion:
          shouldNotify = true;
          return old.filter(i => i.id !== instanceGui.id);
        case SubObjType.UpdatedInfo:
          // Don't notify for publicExposure-only changes
          if (
            JSON.stringify(
              old.find(i => i.id === instanceGui.id)?.publicExposure,
            ) !== JSON.stringify(instanceGui.publicExposure)
          ) {
            // This is a publicExposure change, update silently
            return old.map(i => (i.id === instanceGui.id ? instanceGui : i));
          } else {
            shouldNotify = true;
            return old.map(i => (i.id === instanceGui.id ? instanceGui : i));
          }
        case SubObjType.PrettyName:
          return old.map(i => (i.id === instanceGui.id ? instanceGui : i));
        case SubObjType.Drop:
        default:
          // Always apply updates to ensure real-time sync
          return old.map(i => (i.id === instanceGui.id ? instanceGui : i));
      }
    });

    // Send notification if needed
    if (shouldNotify) {
      notifyStatus(
        instanceK8s.status?.phase,
        instanceK8s,
        updateType,
        _notifier,
      );
    }

    return prev;
  };
};

export const getSubObjTypeCustom = (
  oldObj: Nullable<Instance>,
  newObj: Instance,
  uType: Nullable<UpdateType>,
) => {
  if (uType === UpdateType.Deleted) return SubObjType.Deletion;
  const {
    running: oldRunning,
    status: oldStatus,
    publicExposure: oldPublicExposure,
    environments: oldEnvironments,
  } = oldObj ?? {};
  const {
    running: newRunning,
    status: newStatus,
    publicExposure: newPublicExposure,
    environments: newEnvironments,
  } = newObj;
  if (oldObj) {
    if (oldObj.prettyName !== newObj.prettyName) return SubObjType.PrettyName;

    if (oldStatus !== newStatus || oldRunning !== newRunning) {
      return SubObjType.UpdatedInfo;
    }

    // Check for publicExposure changes
    const oldPEString = JSON.stringify(oldPublicExposure);
    const newPEString = JSON.stringify(newPublicExposure);
    if (oldPEString !== newPEString) {
      // PublicExposure changed - force UI update without notification
      return SubObjType.UpdatedInfo;
    }

    if (oldEnvironments && newEnvironments) {
      const environmentPhaseChanged = oldEnvironments.some((oldEnv, index) => {
        const newEnv = newEnvironments[index];
        return newEnv && oldEnv?.phase !== newEnv?.phase;
      });

      if (environmentPhaseChanged) {
        return SubObjType.UpdatedInfo;
      }
    }
    return SubObjType.Drop;
  }
  return SubObjType.Addition;
};

// Enhanced version of getSubObjTypeK8s to detect publicExposure changes
const getSubObjTypeK8sEnhanced = (
  oldObj: Nullable<DeepPartial<ItPolitoCrownlabsV1alpha2Instance>>,
  newObj: DeepPartial<ItPolitoCrownlabsV1alpha2Instance>,
  uType: Nullable<UpdateType>,
) => {
  if (uType === UpdateType.Deleted) return SubObjType.Deletion;
  const { spec: oldSpec, status: oldStatus } = oldObj ?? {};
  const { spec: newSpec, status: newStatus } = newObj;

  if (oldObj) {
    if (oldSpec?.prettyName !== newSpec?.prettyName)
      return SubObjType.PrettyName;

    if (
      oldStatus?.phase !== newStatus?.phase ||
      oldSpec?.running !== newSpec?.running
    ) {
      return SubObjType.UpdatedInfo;
    }

    // Check for publicExposure changes in both spec and status
    const oldSpecPE = JSON.stringify(oldSpec?.publicExposure);
    const newSpecPE = JSON.stringify(newSpec?.publicExposure);
    const oldStatusPE = JSON.stringify(oldStatus?.publicExposure);
    const newStatusPE = JSON.stringify(newStatus?.publicExposure);

    if (oldSpecPE !== newSpecPE || oldStatusPE !== newStatusPE) {
      // PublicExposure changed - treat as PublicExposureChange without notification
      return SubObjType.PublicExposureChange;
    }

    // Check if any environment phase changed
    const oldEnvironments = oldStatus?.environments || [];
    const newEnvironments = newStatus?.environments || [];

    const environmentPhaseChanged = oldEnvironments.some((oldEnv, index) => {
      const newEnv = newEnvironments[index];
      return newEnv && oldEnv?.phase !== newEnv?.phase;
    });

    if (environmentPhaseChanged) {
      return SubObjType.UpdatedInfo;
    }
    return SubObjType.Drop;
  }
  return SubObjType.Addition;
};

// Override the original function
export { getSubObjTypeK8sEnhanced as getSubObjTypeK8s };

export const joinInstancesAndTemplates = (
  templates: Template[],
  instances: Instance[],
) =>
  templates.map(t => ({
    ...t,
    instances: instances.filter(
      i => i.templateId === makeTemplateKey(t.id, t.workspaceName),
    ),
  }));

export const availableWorkspaces = (
  workspaces: Maybe<DeepPartial<ItPolitoCrownlabsV1alpha1Workspace>>[],
  userWorkspaces: Workspace[],
) =>
  workspaces
    .map(w => {
      const wa: WorkspacesAvailable = {
        name: w?.metadata?.name ?? '',
        prettyName: w?.spec?.prettyName ?? '',
        role:
          userWorkspaces.find(uw => uw.name === w?.metadata?.name)?.role ??
          null,
      };
      if (wa.role === null) {
        // user is not enrolled and has not candidate status
        if (w?.spec?.autoEnroll === AutoEnroll.Immediate) {
          wa.action = WorkspacesAvailableAction.Join;
        } else if (w?.spec?.autoEnroll === AutoEnroll.WithApproval) {
          wa.action = WorkspacesAvailableAction.AskToJoin;
        }
      } else if (wa.role === WorkspaceRole.candidate) {
        // user has candidate status
        wa.action = WorkspacesAvailableAction.Waiting;
      } else {
        // user is enrolled
        wa.action = WorkspacesAvailableAction.None;
      }
      return wa;
    })
    .filter(w => w.action !== WorkspacesAvailableAction.None)
    .sort((a, b) =>
      a.prettyName?.toLowerCase() < b.prettyName?.toLowerCase() ? -1 : 1,
    );

//Utilities for active page only

export const getManagerInstances = (
  instance: Nullable<DeepPartial<ItPolitoCrownlabsV1alpha2Instance>>,
  _index: number,
) => {
  if (!instance) {
    throw new Error('getInstances() error: a required parameter is undefined');
  }
  const { metadata, spec, status } = instance;
  const { publicExposure } = spec ?? {};
  const { publicExposure: publicExposureStatus } = status ?? {};

  // Template Info
  const {
    templateWrapper,
    name: templateName,
    namespace: templateNamespace,
  } = spec?.templateCrownlabsPolitoItTemplateRef ?? {};
  const { prettyName: templatePrettyname, environmentList } =
    templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec ?? {};

  const environments =
    status?.environments?.map(envStatus => {
      const templateEnv = environmentList?.find(
        env => env?.name === envStatus?.name,
      );
      return {
        name: envStatus?.name ?? '',
        phase: envStatus?.phase,
        ip: envStatus?.ip,
        guiEnabled: templateEnv?.guiEnabled ?? false,
        persistent: templateEnv?.persistent ?? false,
        environmentType: templateEnv?.environmentType,
      };
    }) ?? [];

  const hasMultipleEnvironments = environments.length > 1;

  // For backwards compatibility, use the first environment for main properties
  const primaryEnvironment = (environmentList ?? [])[0] ?? {};
  const primaryStatus = environments[0];

  const { guiEnabled, persistent, environmentType } = primaryEnvironment;

  // determine if public exposure allowed by template
  const allowPublicExposure =
    spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
      ?.itPolitoCrownlabsV1alpha2Template?.spec?.allowPublicExposure ?? false;

  // Tenant Info
  const { namespace: tenantNamespace } = metadata ?? {};
  const { name: tenantName, tenantV1alpha2Wrapper } =
    spec?.tenantCrownlabsPolitoItTenantRef ?? {};
  const { firstName, lastName } =
    tenantV1alpha2Wrapper?.itPolitoCrownlabsV1alpha2Tenant?.spec ?? {};
  const workspaceName = (templateNamespace ?? '').replace(/^workspace-/, '');
  const instanceID = tenantNamespace + '/' + metadata?.name;

  return {
    id: instanceID,
    name: metadata?.name,
    prettyName: spec?.prettyName,
    gui: guiEnabled,
    persistent: persistent,
    templateId: makeTemplateKey(templateName, workspaceName),
    templateName: templateName,
    templatePrettyName: templatePrettyname,
    environmentType: environmentType,
    ip: primaryStatus?.ip ?? status?.ip,
    status: safePhase2Conversion(primaryStatus?.phase ?? status?.phase),
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: tenantName,
    tenantNamespace: tenantNamespace,
    tenantDisplayName: `${firstName}\n${lastName}`,
    workspaceName: workspaceName,
    running: spec?.running,
    allowPublicExposure,
    myDriveUrl: '',
    publicExposure: buildPublicExposureObject(
      publicExposure,
      publicExposureStatus,
    ),
    nodeSelector: status?.nodeSelector,
    nodeName: status?.nodeName,
    environments: environments,
    hasMultipleEnvironments: hasMultipleEnvironments,
  } as Instance;
};

export const getTemplatesMapped = (
  instances: Instance[],
  sortingData: Array<{
    sortingType: string;
    sorting: number;
    sortingTemplate: string;
  }>,
) => {
  return Array.from(new Set(instances?.map(i => i.templateId))).map(t => {
    // Find all instances with KEY[Template ID + Workspace ID] === t
    const instancesFiltered = instances?.filter(
      ({ templateId: tid }) => tid === t,
    );

    // Find sorting data for instances with KEY[Template ID + Workspace ID] === t
    const sortDataTmp = sortingData.find(s => s.sortingTemplate === t);

    // If sorting data exist for instances with KEY[Template ID + Workspace ID] === t => sort instances
    let instancesSorted;
    if (sortDataTmp) {
      const { sorting, sortingType } = sortDataTmp;
      instancesSorted = instancesFiltered.sort((a, b) =>
        sorter(a, b, sortingType as keyof Instance, sorting),
      );
    }

    const [
      {
        templateId,
        gui,
        persistent,
        workspaceName,
        templatePrettyName,
        allowPublicExposure,
        environments,
        hasMultipleEnvironments,
      },
    ] = instancesFiltered;

    const environmentList =
      environments?.map(env => ({
        name: env.name,
        guiEnabled: env.guiEnabled || false,
        persistent: env.persistent || false,
        environmentType: env.environmentType,
        resources: { cpu: 0, disk: '', memory: '' },
      })) || [];

    return {
      id: templateId,
      name: templatePrettyName,
      gui,
      persistent,
      resources: { cpu: 0, memory: '', disk: '' },
      instances: instancesSorted || instancesFiltered,
      workspaceName,
      workspaceNamespace: 'workspace-' + workspaceName,
      allowPublicExposure,
      environmentList: environmentList,
      hasMultipleEnvironments: hasMultipleEnvironments ?? false,
    };
  });
};

export const getWorkspacesMapped = (
  templates: Template[],
  workspaces: Workspace[],
) => {
  return workspaces
    .map(ws => ({
      ...ws,
      templates: templates.filter(t => t.workspaceName === ws.name),
    }))
    .filter(ws => ws.templates.length);
};

export const makeTemplateKey = (tid: Nullable<string>, wid: string) =>
  `${tid}-${wid}`;

const makeNotificationContent = (
  templateName: Nullable<string>,
  instanceName: Nullable<string>,
  status: Nullable<string>,
  instanceUrl?: Nullable<string>,
) => {
  const font20px = { fontSize: '20px' };
  return {
    content: (
      <div className="flex justify-between items-start gap-1 p-1 w-72">
        <div className="flex flex-none items-start ">
          {status === Phase2.Ready ? (
            <CheckCircleOutlined
              className="success-color-fg mr-3"
              style={font20px}
            />
          ) : (
            <ExclamationCircleOutlined
              className="warning-color-fg mr-3"
              style={font20px}
            />
          )}
        </div>
        <div className="flex flex-grow flex-col items-start gap-1">
          <div className="pr-1 flex justify-start">
            <b> {templateName}</b>
          </div>
          <div className="pr-1 flex justify-start">
            Instance Name:
            <i> {instanceName}</i>
          </div>
          <div className="pr-1 flex justify-between w-full">
            <div>
              Status:
              <i>
                {status === Phase2.Ready
                  ? ' running'
                  : status === Phase2.Off && ' stopped'}
              </i>
            </div>
            {instanceUrl && (
              <Button
                color="green"
                variant="solid"
                size="small"
                href={instanceUrl}
                target="_blank"
              >
                Connect
              </Button>
            )}
          </div>
        </div>
      </div>
    ),
    icon: <></>,
    className: 'mr-6 flex justify-end',
    duration: 5,
  } as JointContent;
};

export const notifyStatus = (
  status: Nullable<string>,
  instance: Nullable<DeepPartial<ItPolitoCrownlabsV1alpha2Instance>>,
  updateType: Nullable<UpdateType>,
  notify: Notifier,
) => {
  if (!instance) {
    throw new Error('notifyStatus error: instance parameter is undefined');
  }
  if (updateType !== UpdateType.Deleted) {
    const { name, namespace } = instance.metadata ?? {};
    const { prettyName } = instance.spec ?? {};
    const { url, environments } = instance.status ?? {};
    const { prettyName: templateName } =
      instance.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
        ?.itPolitoCrownlabsV1alpha2Template?.spec ?? {};

    // Only set URL for single-environment instances
    let iUrl;
    if (url && environments && environments.length == 1) {
      const firstEnvName = environments[0]?.name;
      if (firstEnvName) {
        const baseUrl = url.endsWith('/') ? url.slice(0, -1) : url;
        iUrl = `${baseUrl}/${firstEnvName}/`;
      }
    }

    switch (status) {
      case Phase2.Off:
        if (!instance.spec?.running) {
          notify(
            'warning',
            `${namespace}/${name}/stopped`,
            makeNotificationContent(templateName, prettyName || name, status),
          );
        }
        break;
      case Phase2.Ready:
        if (instance.spec?.running) {
          notify(
            'success',
            `${namespace}/${name}/ready`,
            makeNotificationContent(
              templateName,
              prettyName || name,
              status,
              iUrl,
            ),
          );
        }
        break;
    }
  }
};

export const filterUser = (instance: Instance, search: string) => {
  if (!search) {
    return true;
  }
  const composedString = `${
    instance.tenantId
  }${instance.tenantDisplayName?.replace(/\s+/g, '')}`.toLowerCase();
  return composedString.includes(search);
};

export function sorter<T>(a: T, b: T, key: keyof T, value: number): number {
  const valA = a[key];
  const valB = b[key];
  let result = 1;
  if (typeof valA === 'string' && typeof valB === 'string') {
    result = valA?.toLowerCase() < valB?.toLowerCase() ? 1 : -1;
  }
  return value === 1 ? result : result * -1;
}

export const setInstanceRunning = async (
  running: boolean,
  instance: Nullable<Instance>,
  instanceMutation: (
    options?: MutationFunctionOptions<
      ApplyInstanceMutation,
      Exact<{
        instanceId: string;
        tenantNamespace: string;
        patchJson: string;
        manager: string;
      }>
    >,
  ) => Promise<
    FetchResult<
      ApplyInstanceMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >,
) => {
  if (!instance) {
    throw new Error(
      'setInstanceRunning error: instance parameter is undefined',
    );
  }
  try {
    return await instanceMutation({
      variables: {
        instanceId: instance.name,
        tenantNamespace: instance.tenantNamespace,
        patchJson: getInstancePatchJson({ running }),
        manager: 'frontend-instance-running',
      },
    });
  } catch {
    return false;
  }
};

export const setInstancePrettyname = async (
  prettyName: string,
  instance: Nullable<Instance>,
  instanceMutation: (
    options?: MutationFunctionOptions<
      ApplyInstanceMutation,
      Exact<{
        instanceId: string;
        tenantNamespace: string;
        patchJson: string;
        manager: string;
      }>
    >,
  ) => Promise<
    FetchResult<
      ApplyInstanceMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >,
) => {
  if (!instance) {
    throw new Error(
      'setInstancePrettyname error: instance parameter is undefined',
    );
  }
  try {
    return await instanceMutation({
      variables: {
        instanceId: instance.name,
        tenantNamespace: instance.tenantNamespace,
        patchJson: getInstancePatchJson({ prettyName }),
        manager: 'frontend-instance-pretty-name',
      },
    });
  } catch {
    return false;
  }
};

export const workspaceGetName = (
  ws: Nullable<DeepPartial<WorkspacesListItem>>,
): string =>
  ws?.workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace?.spec
    ?.prettyName ?? '';

export const makeGuiSharedVolume = (
  shVol?: Nullable<ItPolitoCrownlabsV1alpha2SharedVolume>,
) => {
  if (!shVol) {
    throw new Error(
      'getSharedVolumes() error: a required parameter is undefined',
    );
  }

  const { metadata, spec, status } = shVol;

  return {
    id: metadata?.namespace + '/' + metadata?.name,
    name: metadata?.name,
    prettyName: spec?.prettyName,
    size: spec?.size,
    status: safePhase5Conversion(status?.phase),
    timeStamp: metadata?.creationTimestamp,
    namespace: metadata?.namespace,
  } as SharedVolume;
};
