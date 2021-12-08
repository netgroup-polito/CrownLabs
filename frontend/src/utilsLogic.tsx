import { FetchResult, MutationFunctionOptions } from '@apollo/client';
import { notification } from 'antd';
import Button from 'antd-button-color';
import { Dispatch, SetStateAction } from 'react';
import {
  ApplyInstanceMutation,
  Exact,
  ItPolitoCrownlabsV1alpha2Instance,
  ItPolitoCrownlabsV1alpha2Template,
  OwnedInstancesQuery,
  Phase,
  UpdatedOwnedInstancesSubscriptionResult,
  UpdatedWorkspaceTemplatesSubscriptionResult,
  UpdateType,
  WorkspacesListItem,
  WorkspaceTemplatesQuery,
} from './generated-types';
import { getInstancePatchJson } from './graphql-components/utils';
import { matchK8sObject, replaceK8sObject } from './k8sUtils';
import { Instance, Template, WorkspaceRole } from './utils';

type Nullable<T> = T | null | undefined;

export enum SubObjType {
  Deletion,
  UpdatedInfo,
  PrettyName,
  Addition,
  Drop,
}
interface ItPolitoCrownlabsV1alpha2TemplateAlias {
  original: Nullable<ItPolitoCrownlabsV1alpha2Template>;
  alias: {
    name: string;
    id: string;
  };
}
export const makeGuiTemplate = (
  tq: ItPolitoCrownlabsV1alpha2TemplateAlias
): Template => {
  if (!tq.original) {
    throw new Error(
      'makeGuiTemplate() error: a required parameter is undefined'
    );
  }
  const environment = (tq.original.spec?.environmentList ?? [])[0];
  return {
    id: tq.alias.id ?? '',
    name: tq.alias.name ?? '',
    gui: !!environment?.guiEnabled,
    persistent: !!environment?.persistent,
    resources: {
      cpu: environment?.resources?.cpu ?? 0,
      memory: environment?.resources?.memory ?? '',
      disk: environment?.resources?.disk ?? '',
    },
    workspaceId:
      tq.original.spec?.workspaceCrownlabsPolitoItWorkspaceRef?.name ?? '',
    instances: [],
  };
};
interface TemplatesSubscriptionData {
  subscriptionData: { data: WorkspaceTemplatesQuery };
}
export const updateQueryWorkspaceTemplatesQuery = (
  setDataTemplate: Dispatch<SetStateAction<Template[]>>
) => {
  return (
    prev: WorkspaceTemplatesQuery,
    subscriptionDataObject: TemplatesSubscriptionData
  ) => {
    const { data } =
      subscriptionDataObject.subscriptionData as UpdatedWorkspaceTemplatesSubscriptionResult;
    const template = data?.updatedTemplate?.template;
    const { updateType } = data?.updatedTemplate ?? {
      updateType: UpdateType.Added,
    };

    if (prev.templateList?.templates) {
      if (updateType === UpdateType.Deleted) {
        setDataTemplate(old =>
          old.filter(t => t.id !== template?.metadata?.name)
        );
      } else if (updateType === UpdateType.Modified) {
        setDataTemplate(old =>
          old.map(t =>
            t.id === template?.metadata?.name
              ? makeGuiTemplate({
                  original: template,
                  alias: {
                    id: template.metadata.name,
                    name: template.spec?.prettyName ?? '',
                  },
                })
              : t
          )
        );
      } else if (updateType === UpdateType.Added) {
        setDataTemplate(old =>
          [
            ...old,
            makeGuiTemplate({
              original: template,
              alias: {
                id: template?.metadata?.name ?? '',
                name: template?.spec?.prettyName ?? '',
              },
            }),
          ].sort((a, b) => a.name.localeCompare(b.name))
        );
      }
    }
    return prev;
  };
};

export type InstanceLabels = {
  crownlabsPolitoItManagedBy?: string;
  crownlabsPolitoItPersistent?: string;
  crownlabsPolitoItTemplate?: string;
  crownlabsPolitoItWorkspace?: string;
};

export const getInstanceLabels = (
  i: ItPolitoCrownlabsV1alpha2Instance
): InstanceLabels | undefined => i.metadata?.labels as InstanceLabels;

export const makeGuiInstance = (
  instance?: Nullable<ItPolitoCrownlabsV1alpha2Instance>,
  userId?: string,
  optional?: {
    workspaceName: string;
    templateName: string;
  }
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
  const { name: templateName } =
    spec?.templateCrownlabsPolitoItTemplateRef as any;
  const { guiEnabled, persistent, environmentType } =
    (environmentList ?? [])[0] ?? {};

  const myDrive = status?.myDriveUrl ?? '';

  return {
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
        ''
    ),
    environmentType: environmentType,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    myDriveUrl: myDrive,
    timeStamp: metadata?.creationTimestamp,
    tenantId: userId,
    tenantNamespace: tenantNamespace,
    workspaceId: getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ?? '',
    running: running,
  } as Instance;
};
interface InstancesSubscriptionData {
  subscriptionData: { data: OwnedInstancesQuery };
}
export const updateQueryOwnedInstancesQuery = (
  setDataInstances: Dispatch<SetStateAction<Instance[]>>,
  userId: string,
  tenantNamespace: string
) => {
  return (
    prev: OwnedInstancesQuery,
    subscriptionDataObject: InstancesSubscriptionData
  ) => {
    const { data } =
      subscriptionDataObject.subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

    if (!data?.updateInstance?.instance) return prev;

    const { instance, updateType } = data?.updateInstance;
    let notify = false;

    if (prev.instanceList?.instances) {
      let instances = [...prev.instanceList.instances];
      const found = instances.find(matchK8sObject(instance, false));
      const objType = getSubObjType(found, instance, updateType);
      switch (objType) {
        case SubObjType.Deletion:
          instances = instances.filter(matchK8sObject(instance, true));
          setDataInstances(old =>
            old?.filter(i => i?.name !== instance?.metadata?.name)
          );
          notify = false;
          break;
        case SubObjType.Addition:
          instances = [...instances, instance];
          setDataInstances(old =>
            !old.find(i => i.name === instance?.metadata?.name)
              ? [...old, makeGuiInstance(instance, userId)]
              : old
          );
          notify = true;
          break;
        case SubObjType.PrettyName:
          instances = instances.map(replaceK8sObject(instance));
          setDataInstances(old =>
            old?.map(i =>
              i.name === instance?.metadata?.name ?? ''
                ? makeGuiInstance(instance, userId)
                : i
            )
          );
          notify = false;
          break;
        case SubObjType.UpdatedInfo:
          instances = instances.map(replaceK8sObject(instance));
          setDataInstances(old =>
            old?.map(i =>
              i.name === instance?.metadata?.name ?? ''
                ? makeGuiInstance(instance, userId)
                : i
            )
          );
          notify = true;
          break;
        case SubObjType.Drop:
          notify = false;
          break;
        default:
          break;
      }
      prev.instanceList.instances = [...instances];
    }

    if (notify) notifyStatus(instance.status?.phase, instance, updateType);

    return prev;
  };
};

export const getSubObjType = (
  oldObj: Nullable<ItPolitoCrownlabsV1alpha2Instance>,
  newObj: ItPolitoCrownlabsV1alpha2Instance,
  uType: Nullable<UpdateType>
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
  instances: Instance[]
) =>
  templates.map(t => ({
    ...t,
    instances: instances.filter(
      i => i.templateId === makeTemplateKey(t.id, t.workspaceId)
    ),
  }));

//Utilities for active page only

export const getManagerInstances = (
  instance: Nullable<ItPolitoCrownlabsV1alpha2Instance>,
  index: number
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
  const myDrive = status?.myDriveUrl ?? '';
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
    myDriveUrl: myDrive,
    timeStamp: metadata?.creationTimestamp,
    tenantId: tenantName,
    tenantNamespace: tenantNamespace,
    tenantDisplayName: `${firstName}\n${lastName}`,
    workspaceId: workspaceName,
    running: spec?.running,
  } as Instance;
};

export const getTemplatesMapped = (
  instances: Instance[],
  sortingData: Array<{
    sortingType: string;
    sorting: number;
    sortingTemplate: string;
  }>
) => {
  return Array.from(new Set(instances?.map(i => i.templateId))).map(t => {
    // Find all instances with KEY[Template ID + Workspace ID] === t
    const instancesFiltered = instances?.filter(
      ({ templateId: tid }) => tid === t
    );

    // Find sorting data for instances with KEY[Template ID + Workspace ID] === t
    const sortDataTmp = sortingData.find(s => s.sortingTemplate === t);

    // If sorting data exist for instances with KEY[Template ID + Workspace ID] === t => sort instances
    let instancesSorted;
    if (sortDataTmp) {
      const { sorting, sortingType } = sortDataTmp;
      instancesSorted = instancesFiltered.sort((a, b) =>
        sorter(a, b, sortingType as keyof Instance, sorting)
      );
    }

    const [{ templateId, gui, persistent, workspaceId, templatePrettyName }] =
      instancesFiltered;
    return {
      id: templateId,
      name: templatePrettyName,
      gui,
      persistent,
      resources: { cpu: 0, memory: '', disk: '' },
      instances: instancesSorted || instancesFiltered,
      workspaceId,
    };
  });
};

export const getWorkspacesMapped = (
  templates: Template[],
  workspaces: Array<{
    prettyName: string;
    role: WorkspaceRole;
    namespace: string;
    id: string;
  }>
) => {
  return workspaces
    .map(ws => ({
      id: ws.id,
      title: ws.prettyName,
      role: ws.role,
      templates: templates.filter(({ workspaceId: id }) => id === ws.id),
    }))
    .filter(ws => ws.templates.length);
};

export const makeTemplateKey = (tid: Nullable<string>, wid: string) =>
  `${tid}-${wid}`;

const makeNotificationContent = (
  templateName: Nullable<string>,
  instanceName: Nullable<string>,
  status: Nullable<string>,
  instanceUrl?: Nullable<string>
) => {
  return {
    message: templateName,
    description: (
      <>
        <div>
          Instance Name:
          <i> {instanceName}</i>
        </div>
        <div>
          Status:
          <i>
            {status === Phase.Ready
              ? ' started'
              : status === Phase.Off && ' stopped'}
          </i>
        </div>
      </>
    ),
    btn: instanceUrl && (
      <Button type="success" size="small" href={instanceUrl} target="_blank">
        Connect
      </Button>
    ),
  };
};

export const notifyStatus = (
  status: Nullable<string>,
  instance: Nullable<ItPolitoCrownlabsV1alpha2Instance>,
  updateType: Nullable<UpdateType>
) => {
  if (!instance) {
    throw new Error('notifyStatus error: instance parameter is undefined');
  }
  if (updateType === UpdateType.Deleted) {
    notification.warning({
      message: instance.spec?.prettyName || instance.metadata?.name,
      description: `Instance deleted`,
    });
  } else {
    const { name } = instance.metadata ?? {};
    const { prettyName } = instance.spec ?? {};
    const { url } = instance.status ?? {};
    const { name: templateName } =
      instance.spec?.templateCrownlabsPolitoItTemplateRef ?? {};

    switch (status) {
      case Phase.Off:
        !instance.spec?.running &&
          notification.warning(
            makeNotificationContent(templateName, prettyName || name, status)
          );
        break;
      case Phase.Ready:
        instance.spec?.running &&
          notification.success(
            makeNotificationContent(
              templateName,
              prettyName || name,
              status,
              url
            )
          );
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
    result = valA?.toLowerCase()! < valB?.toLowerCase()! ? 1 : -1;
  }
  return value === 1 ? result : result * -1;
}

export enum DropDownAction {
  start = 'start',
  stop = 'stop',
  destroy = 'destroy',
  connect = 'connect',
  ssh = 'ssh',
  upload = 'upload',
  destroy_all = 'destroy_all',
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
    >
  ) => Promise<
    FetchResult<ApplyInstanceMutation, Record<string, any>, Record<string, any>>
  >
) => {
  if (!instance) {
    throw new Error(
      'setInstanceRunning error: instance parameter is undefined'
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
    >
  ) => Promise<
    FetchResult<ApplyInstanceMutation, Record<string, any>, Record<string, any>>
  >
) => {
  if (!instance) {
    throw new Error(
      'setInstancePrettyname error: instance parameter is undefined'
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

export const workspaceGetName = (ws: WorkspacesListItem): string =>
  ws?.workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace?.spec
    ?.prettyName ?? '';
