import { FetchResult, MutationFunctionOptions } from '@apollo/client';
import { notification } from 'antd';
import Button from 'antd-button-color';
import { Dispatch, SetStateAction } from 'react';
import {
  ApplyInstanceMutation,
  Exact,
  ItPolitoCrownlabsV1alpha2Instance,
  ItPolitoCrownlabsV1alpha2Template,
  Maybe,
  OwnedInstancesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  UpdatedWorkspaceTemplatesSubscriptionResult,
  UpdateType,
  WorkspacesListItem2,
  WorkspaceTemplatesQuery,
} from './generated-types';
import { getInstancePatchJson } from './graphql-components/utils';
import { Instance, Template, WorkspaceRole } from './utils';

interface ItPolitoCrownlabsV1alpha2TemplateAlias {
  original: ItPolitoCrownlabsV1alpha2Template;
  alias: {
    name: string;
    id: string;
  };
}
export const makeGuiTemplate = (
  tq: ItPolitoCrownlabsV1alpha2TemplateAlias
): Template => {
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
              original: template ?? {},
              alias: {
                id: template?.metadata?.name ?? '',
                name: template?.spec?.prettyName ?? '',
              },
            }) ?? {},
          ].sort((a, b) => a.id.localeCompare(b.id))
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
  instance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>,
  userId?: string,
  tenantNamespace?: string,
  optional?: {
    workspaceName: string;
    templateName: string;
  }
) => {
  if (!instance || !userId || !tenantNamespace) {
    throw new Error('getInstances() error: a required parameter is undefined');
  }
  const { metadata, spec, status } = instance;
  const { environmentList, prettyName } = spec
    ?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
    ?.itPolitoCrownlabsV1alpha2Template?.spec ?? {
    environmentList: [],
    prettyName: '',
  };
  const { guiEnabled, persistent, environmentType } =
    (environmentList ?? [])[0] ?? {};
  return {
    name: metadata?.name,
    prettyName: spec?.prettyName,
    gui: guiEnabled,
    persistent: persistent,
    templatePrettyName: prettyName,
    templateName: spec?.templateCrownlabsPolitoItTemplateRef?.name ?? '',
    templateId: makeTemplateKey(
      getInstanceLabels(instance)?.crownlabsPolitoItTemplate ??
        optional?.templateName ??
        '',
      getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ??
        optional?.workspaceName ??
        ''
    ),
    workspaceId: getInstanceLabels(instance)?.crownlabsPolitoItWorkspace ?? '',
    environmentType: environmentType,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: userId,
    tenantNamespace: tenantNamespace,
    running: spec?.running,
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

    const { instance } = data?.updateInstance!;
    let { updateType } = data?.updateInstance!;

    let isPrettyNameUpdate = false;
    if (prev.instanceList?.instances) {
      if (updateType === UpdateType.Deleted) {
        setDataInstances(old =>
          old?.filter(i => i?.name !== instance?.metadata?.name)
        );
      } else if (updateType === UpdateType.Modified) {
        isPrettyNameUpdate = true;
        setDataInstances(old =>
          old?.map(i => {
            if (
              i.prettyName === instance?.spec?.prettyName &&
              i.name === instance?.metadata?.name
            )
              isPrettyNameUpdate = false;
            return i.name === instance?.metadata?.name ?? ''
              ? makeGuiInstance(instance, userId, tenantNamespace)
              : i;
          })
        );
      } else if (updateType === UpdateType.Added) {
        setDataInstances(old =>
          !old.find(i => i.name === instance?.metadata?.name)
            ? [...old, makeGuiInstance(instance ?? {}, userId, tenantNamespace)]
            : old
        );
      }
    }

    !isPrettyNameUpdate &&
      notifyStatus(
        instance?.status?.phase ?? '',
        instance ?? {},
        updateType ?? UpdateType.Added,
        tenantNamespace,
        WorkspaceRole.user
      );

    return prev;
  };
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
  instance: ItPolitoCrownlabsV1alpha2Instance | null,
  index: number
) => {
  const { metadata, spec, status } = instance ?? {};
  const { environmentList, prettyName } =
    spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
      ?.itPolitoCrownlabsV1alpha2Template?.spec ?? {};
  const { guiEnabled, persistent, environmentType } =
    (environmentList ?? [])[0] ?? {};
  const { firstName, lastName } =
    spec?.tenantCrownlabsPolitoItTenantRef?.tenantV1alpha2Wrapper
      ?.itPolitoCrownlabsV1alpha2Tenant?.spec ?? {};
  const { name: tenantName } = spec?.tenantCrownlabsPolitoItTenantRef ?? {};
  const { name: templateName, namespace: templateNamespace } =
    spec?.templateCrownlabsPolitoItTemplateRef ?? {};
  return {
    id: index,
    name: metadata?.name,
    prettyName: spec?.prettyName,
    gui: guiEnabled,
    persistent: persistent,
    templatePrettyName: prettyName,
    templateName: templateName,
    //TODO: use lables
    templateId: makeTemplateKey(
      templateName ?? '',
      (templateNamespace ?? '').replace(/^workspace-/, '')
    ),
    environmentType: environmentType,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: tenantName,
    tenantNamespace: metadata?.namespace,
    tenantDisplayName: `${firstName}\n${lastName}`,
    workspaceId: (templateNamespace ?? '').replace(/^workspace-/, ''),
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
  //const { sorting, sortingType, sortingTemplate } = sortingData;
  return Array.from(new Set(instances?.map(i => i.templatePrettyName))).map(
    t => {
      // Find all instances with Template Name === t
      const instancesFiltered = instances?.filter(
        ({ templatePrettyName: tpn }) => tpn === t
      );

      // Find sorting data for Template Name === t
      const sortDataTmp = sortingData.find(s => s.sortingTemplate === t);

      // If sorting data exist fot Template Name = t => sort instances
      let instancesSorted;
      if (sortDataTmp) {
        const { sorting, sortingType } = sortDataTmp;
        instancesSorted = instancesFiltered.sort((a, b) =>
          sorter(a, b, sortingType as keyof Instance, sorting)
        );
      }

      const [{ templateName, gui, persistent, workspaceId }] =
        instancesFiltered;
      return {
        id: templateName,
        name: t,
        gui,
        persistent,
        resources: { cpu: 0, memory: '', disk: '' },
        instances: instancesSorted || instancesFiltered,
        workspaceId,
      } as Template;
    }
  );
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

export const makeTemplateKey = (tid: string, wid: string) => `${tid}-${wid}`;

export const notifyStatus = (
  status: string,
  instance: ItPolitoCrownlabsV1alpha2Instance,
  updateType: UpdateType,
  tenantNamespace: string,
  role: WorkspaceRole
) => {
  if (updateType === UpdateType.Deleted) {
    const { tenantNamespace: tnm } = instance.metadata as any;
    if (tnm === tenantNamespace || role === WorkspaceRole.user) {
      notification.warning({
        message: instance.spec?.prettyName || instance.metadata?.name,
        description: `Instance deleted`,
      });
    }
  } else {
    const { tenantNamespace: tnm } = instance.metadata as any;
    const { templateName } = instance.spec?.templateCrownlabsPolitoItTemplateRef
      ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec as any;
    const content = (status: string) => {
      return {
        message: templateName,
        description: (
          <>
            <div>
              Instance Name:
              <i> {instance.spec?.prettyName || instance.metadata?.name}</i>
            </div>
            <div>
              Status:
              <i>
                {status === 'VmiReady'
                  ? ' started'
                  : status === 'VmiOff' && ' stopped'}
              </i>
            </div>
          </>
        ),
      };
    };
    switch (status) {
      case 'VmiOff':
        if (
          !instance.spec?.running &&
          (tnm === tenantNamespace || role === WorkspaceRole.user)
        ) {
          notification.warning({
            ...content(status),
          });
        }
        break;
      case 'VmiReady':
        if (
          instance.status?.url &&
          instance.spec?.running &&
          (tnm === tenantNamespace || role === WorkspaceRole.user)
        ) {
          notification.success({
            ...content(status),
            btn: instance.status?.url && (
              <Button
                type="success"
                size="small"
                href={instance.status?.url}
                target="_blank"
              >
                Connect
              </Button>
            ),
          });
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
  instance: Instance,
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
  try {
    return await instanceMutation({
      variables: {
        instanceId: instance.name,
        tenantNamespace: instance.tenantNamespace!,
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
  instance: Instance,
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
  try {
    return await instanceMutation({
      variables: {
        instanceId: instance.name,
        tenantNamespace: instance.tenantNamespace!,
        patchJson: getInstancePatchJson({ prettyName }),
        manager: 'frontend-instance-pretty-name',
      },
    });
  } catch {
    return false;
  }
};

export const workspaceGetName = (ws: WorkspacesListItem2): string =>
  ws?.workspaceWrapperTenantV1alpha2?.itPolitoCrownlabsV1alpha1Workspace?.spec
    ?.prettyName ?? '';
