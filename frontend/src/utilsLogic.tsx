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
import { AutoEnroll, Phase, UpdateType } from './generated-types';
import { getInstancePatchJson } from './graphql-components/utils';
import type {
  Instance,
  SharedVolume,
  Template,
  Workspace,
  WorkspacesAvailable,
} from './utils';
import { WorkspaceRole, WorkspacesAvailableAction } from './utils';
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
  const environment = (tq.original.spec?.environmentList ?? [])[0];
  return {
    id: tq.alias.id ?? '',
    name: tq.alias.name ?? '',
    gui: !!environment?.guiEnabled,
    persistent: !!environment?.persistent,
    nodeSelector: environment?.nodeSelector,
    resources: {
      cpu: environment?.resources?.cpu ?? 0,
      memory: environment?.resources?.memory ?? '',
      disk: environment?.resources?.disk ?? '',
    },
    workspaceName:
      tq.original.spec?.workspaceCrownlabsPolitoItWorkspaceRef?.name ?? '',
    instances: [],
    workspaceNamespace:
      'workspace-' +
      (tq.original.spec?.workspaceCrownlabsPolitoItWorkspaceRef?.name ?? ''),
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
  const { running, prettyName } = spec ?? {};
  const { environmentList, prettyName: templatePrettyName } = spec
    ?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
    ?.itPolitoCrownlabsV1alpha2Template?.spec ?? {
    environmentList: [],
    prettyName: '',
  };
  const templateName = spec?.templateCrownlabsPolitoItTemplateRef?.name;
  const { guiEnabled, persistent, environmentType } =
    (environmentList ?? [])[0] ?? {};

  const instanceID = tenantNamespace + '/' + metadata?.name;

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
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: userId,
    tenantNamespace: tenantNamespace,
    workspaceName:
      getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ?? '',
    running: running,
    nodeName: status?.nodeName,
    nodeSelector: status?.nodeSelector,
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
    role: role! as unknown as WorkspaceRole,
    templates: [],
  } as Workspace;
};
interface InstancesSubscriptionData {
  subscriptionData: { data: OwnedInstancesQuery };
}
export const updateQueryOwnedInstancesQuery = (
  setDataInstances: Dispatch<SetStateAction<Instance[]>>,
  userId: string,
  notifier: Notifier,
) => {
  return (
    prev: OwnedInstancesQuery,
    subscriptionDataObject: InstancesSubscriptionData,
  ) => {
    const { data } =
      subscriptionDataObject.subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

    if (!data?.updateInstance?.instance) return prev;

    const { instance: instanceK8s, updateType } = data.updateInstance;
    let notify = false;

    setDataInstances(old => {
      const instance = makeGuiInstance(instanceK8s, userId);
      const found = old.find(i => i.id === instance.id);
      const objType = getSubObjTypeCustom(found, instance, updateType);
      switch (objType) {
        case SubObjType.Deletion:
          old = old.filter(i => i.id !== instance.id);
          notify = false;
          break;
        case SubObjType.Addition:
          old = !old.find(i => i.id === instance.id) ? [...old, instance] : old;
          notify = true;
          break;
        case SubObjType.PrettyName:
          old = old?.map(i => (i.id === instance.id ? instance : i));
          notify = false;
          break;
        case SubObjType.UpdatedInfo:
          old = old?.map(i => (i.id === instance.id ? instance : i));
          notify = true;
          break;
        case SubObjType.Drop:
          notify = false;
          break;
        default:
          break;
      }

      if (notify)
        notifyStatus(
          instanceK8s.status?.phase,
          instanceK8s,
          updateType,
          notifier,
        );

      return old;
    });

    return prev;
  };
};

export const getSubObjTypeCustom = (
  oldObj: Nullable<Instance>,
  newObj: Instance,
  uType: Nullable<UpdateType>,
) => {
  if (uType === UpdateType.Deleted) return SubObjType.Deletion;
  const { running: oldRunning, status: oldStatus } = oldObj ?? {};
  const { running: newRunning, status: newStatus } = newObj;
  if (oldObj) {
    if (oldObj.prettyName !== newObj.prettyName) return SubObjType.PrettyName;
    if (oldStatus !== newStatus || oldRunning !== newRunning) {
      return SubObjType.UpdatedInfo;
    }
    return SubObjType.Drop;
  }
  return SubObjType.Addition;
};

export const getSubObjTypeK8s = (
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
    return SubObjType.Drop;
  }
  return SubObjType.Addition;
};

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

  // Template Info
  const {
    templateWrapper,
    name: templateName,
    namespace: templateNamespace,
  } = spec?.templateCrownlabsPolitoItTemplateRef ?? {};
  const { prettyName: templatePrettyname, environmentList } =
    templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec ?? {};
  const { guiEnabled, persistent, environmentType } =
    (environmentList ?? [])[0] ?? {};

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
    templatePrettyName: templatePrettyname,
    environmentType: environmentType,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: tenantName,
    tenantNamespace: tenantNamespace,
    tenantDisplayName: `${firstName}\n${lastName}`,
    workspaceName: workspaceName,
    running: spec?.running,
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

    const [{ templateId, gui, persistent, workspaceName, templatePrettyName }] =
      instancesFiltered;
    return {
      id: templateId,
      name: templatePrettyName,
      gui,
      persistent,
      resources: { cpu: 0, memory: '', disk: '' },
      instances: instancesSorted || instancesFiltered,
      workspaceName,
      workspaceNamespace: 'workspace-' + workspaceName,
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
          {status === Phase.Ready ? (
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
                {status === Phase.Ready
                  ? ' started'
                  : status === Phase.Off && ' stopped'}
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
    const { url } = instance.status ?? {};
    const { prettyName: templateName } =
      instance.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
        ?.itPolitoCrownlabsV1alpha2Template?.spec ?? {};

    switch (status) {
      case Phase.Off:
        if (!instance.spec?.running) {
          notify(
            'warning',
            `${namespace}/${name}/stopped`,
            makeNotificationContent(templateName, prettyName || name, status),
          );
        }
        break;
      case Phase.Ready:
        if (instance.spec?.running) {
          notify(
            'success',
            `${namespace}/${name}/ready`,
            makeNotificationContent(
              templateName,
              prettyName || name,
              status,
              url,
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
    status: status?.phase,
    timeStamp: metadata?.creationTimestamp,
    namespace: metadata?.namespace,
  } as SharedVolume;
};
