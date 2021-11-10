import { FetchResult, MutationFunctionOptions } from '@apollo/client';
import { notification } from 'antd';
import Button from 'antd-button-color';
import {
  ApplyInstanceMutation,
  Exact,
  ItPolitoCrownlabsV1alpha2Instance,
  UpdateType,
} from '../../generated-types';
import { getInstancePatchJson } from '../../graphql-components/utils';
import { Instance, Template, WorkspaceRole } from '../../utils';

const getInstances = (
  instance: ItPolitoCrownlabsV1alpha2Instance,
  index: number,
  userId: string,
  tenantNamespace: string
) => {
  const { metadata, spec, status } = instance!;
  const { environmentList, templateName } = spec
    ?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
    ?.itPolitoCrownlabsV1alpha2Template?.spec! as any;
  const [{ guiEnabled, persistent }] = environmentList;
  return {
    id: index,
    name: metadata?.name,
    gui: guiEnabled,
    persistent: persistent,
    idTemplate: spec?.templateCrownlabsPolitoItTemplateRef?.name!,
    templatePrettyName: templateName,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: userId,
    tenantNamespace: tenantNamespace,
    running: spec?.running,
  } as Instance;
};

const getManagerInstances = (
  instance: ItPolitoCrownlabsV1alpha2Instance | null,
  index: number
) => {
  const { metadata, spec, status } = instance!;
  const { tenantNamespace } = metadata! as any;
  const { environmentList, templateName } = spec
    ?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
    ?.itPolitoCrownlabsV1alpha2Template?.spec! as any;
  const [{ guiEnabled, persistent }] = environmentList;
  const {
    firstName,
    lastName,
  } = spec?.tenantCrownlabsPolitoItTenantRef?.tenantWrapper?.itPolitoCrownlabsV1alpha1Tenant?.spec!;
  const { tenantId } = spec?.tenantCrownlabsPolitoItTenantRef as any;
  const { name, namespace } = spec?.templateCrownlabsPolitoItTemplateRef as any;
  return {
    id: index,
    name: metadata?.name,
    gui: guiEnabled,
    persistent: persistent,
    idTemplate: name,
    templatePrettyName: templateName,
    ip: status?.ip,
    status: status?.phase,
    url: status?.url,
    timeStamp: metadata?.creationTimestamp,
    tenantId: tenantId,
    tenantNamespace: tenantNamespace,
    tenantDisplayName: `${firstName}\n${lastName}`,
    workspaceId: namespace.replace(/^workspace-/, ''),
    running: spec?.running,
  } as Instance;
};

const getTemplatesMapped = (instances: Instance[]) => {
  return Array.from(new Set(instances?.map(i => i.templatePrettyName))).map(
    t => {
      const instancesFiltered = instances?.filter(
        ({ templatePrettyName: tpn }) => tpn === t
      );
      const [{ idTemplate, gui, persistent, workspaceId }] = instancesFiltered!;
      return {
        id: idTemplate,
        name: t,
        gui,
        persistent,
        resources: { cpu: 0, memory: '', disk: '' },
        instances: instancesFiltered,
        workspaceId,
      } as Template;
    }
  );
};

const getWorkspacesMapped = (
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

const notifyStatus = (
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
        message: instance.metadata?.name,
        description: `Instance deleted`,
      });
    }
  } else {
    const { tenantNamespace: tnm } = instance.metadata as any;
    const { templateName } = instance.spec?.templateCrownlabsPolitoItTemplateRef
      ?.templateWrapper?.itPolitoCrownlabsV1alpha2Template?.spec as any;
    switch (status) {
      case 'VmiOff':
        if (
          !instance.spec?.running &&
          (tnm === tenantNamespace || role === WorkspaceRole.user)
        ) {
          notification.warning({
            message: templateName,
            description: (
              <>
                <div>
                  Instance Name: <i>{instance.metadata?.name}</i>
                </div>
                <div>
                  Status: <i>Stopped</i>
                </div>
              </>
            ),
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
            message: <strong>{templateName}</strong>,
            description: (
              <>
                <div>
                  Instance Name: <i>{instance.metadata?.name}</i>
                </div>
                <div>
                  Status: <i>Started</i>
                </div>
              </>
            ),
            btn: instance.status?.url && (
              <Button
                type="success"
                size="small"
                onClick={() => window.open(instance.status?.url!, '_blank')}
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

const filterId = (instance: Instance, search: string) => {
  if (!search) {
    return true;
  }
  return instance.tenantId && instance.tenantId.includes(search);
};

enum DropDownAction {
  start = 'start',
  stop = 'stop',
  destroy = 'destroy',
  connect = 'connect',
  ssh = 'ssh',
  upload = 'upload',
  destroy_all = 'destroy_all',
}

const setInstanceRunning = async (
  running: boolean,
  instance: Instance,
  instanceMutation: (
    options?: MutationFunctionOptions<
      ApplyInstanceMutation,
      Exact<{
        instanceId: string;
        tenantNamespace: string;
        patchJson: string;
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
      },
    });
  } catch {
    return false;
  }
};

export {
  getInstances,
  getManagerInstances,
  notifyStatus,
  filterId,
  getTemplatesMapped,
  getWorkspacesMapped,
  setInstanceRunning,
  DropDownAction,
};
