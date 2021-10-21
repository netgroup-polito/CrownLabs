import { notification } from 'antd';
import Button from 'antd-button-color';
import { Dispatch, SetStateAction } from 'react';
import {
  ItPolitoCrownlabsV1alpha2Instance,
  ItPolitoCrownlabsV1alpha2Template,
  OwnedInstancesQuery,
  UpdatedOwnedInstancesSubscriptionResult,
  UpdatedWorkspaceTemplatesSubscriptionResult,
  UpdateType,
  WorkspaceTemplatesQuery,
} from './generated-types';
import { Instance, Template, VmStatus } from './utils';

interface ItPolitoCrownlabsV1alpha2TemplateAlias {
  original: ItPolitoCrownlabsV1alpha2Template;
  alias: {
    name: string;
    id: string;
  };
}
export const getTemplate = (
  tq: ItPolitoCrownlabsV1alpha2TemplateAlias
): Template => {
  const environment = tq.original.spec?.environmentList![0];
  return {
    id: tq.alias.id!,
    name: tq.alias.name!,
    gui: !!environment?.guiEnabled,
    persistent: !!environment?.persistent,
    resources: {
      cpu: environment?.resources?.cpu!,
      memory: environment?.resources?.memory!,
      disk: environment?.resources?.disk!,
    },
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
    const template = data?.updatedTemplate?.template!;
    const { updateType } = data?.updatedTemplate!;

    if (prev.templateList?.templates) {
      if (updateType === UpdateType.Deleted) {
        setDataTemplate(old =>
          old.filter(t => t.id !== template.metadata?.id!)
        );
      } else if (updateType === UpdateType.Modified) {
        setDataTemplate(old =>
          old.map(t =>
            t.id === template.metadata?.id
              ? getTemplate({
                  original: template,
                  alias: {
                    id: template.metadata.id!,
                    name: template.spec?.name!,
                  },
                })
              : t
          )
        );
      } else if (updateType === UpdateType.Added) {
        setDataTemplate(old =>
          [
            ...old,
            getTemplate({
              original: template,
              alias: {
                id: template.metadata?.id!,
                name: template.spec?.name!,
              },
            })!,
          ].sort((a, b) => a.id.localeCompare(b.id))
        );
      }
    }
    return prev;
  };
};

export const getInstance = (
  oiq: ItPolitoCrownlabsV1alpha2Instance,
  id: number,
  tenantNamespace?: string,
  workspaceId?: string
): Instance => {
  const { metadata, spec, status } = oiq!;
  const { environmentList, templateName } = spec
    ?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
    ?.itPolitoCrownlabsV1alpha2Template?.spec! as any;
  const environmentListItem = environmentList[0]!;
  return {
    id: id,
    /* idName: oiq.metadata?.name!, */
    idTemplate: spec?.templateCrownlabsPolitoItTemplateRef?.name!,
    templatePrettyName: templateName!,
    /* name: oiq?.spec?.prettyName! || oiq.metadata?.name!, */
    name: metadata?.name!,
    ip: status?.ip!,
    status: status?.phase! as VmStatus,
    url: status?.url!,
    gui: environmentListItem?.guiEnabled,
    persistent: environmentListItem?.persistent,
    timeStamp: metadata?.creationTimestamp!,
    running: spec?.running!,
    tenantNamespace: tenantNamespace,
    workspaceId: workspaceId,
  };
};
interface InstancesSubscriptionData {
  subscriptionData: { data: OwnedInstancesQuery };
}
export const updateQueryOwnedInstancesQuery = (
  dataInstances: Instance[],
  setDataInstances: Dispatch<SetStateAction<Instance[]>>
) => {
  return (
    prev: OwnedInstancesQuery,
    subscriptionDataObject: InstancesSubscriptionData
  ) => {
    const { data } =
      subscriptionDataObject.subscriptionData as UpdatedOwnedInstancesSubscriptionResult;

    const { instance } = data?.updateInstance!;
    let { updateType } = data?.updateInstance!;

    if (prev.instanceList?.instances) {
      if (updateType === UpdateType.Deleted) {
        setDataInstances(old =>
          old?.filter(i => i?.name !== instance!.metadata?.name)
        );
      } else if (updateType === UpdateType.Modified) {
        setDataInstances(old =>
          old?.map((i, n) =>
            i.name === instance?.metadata?.name! ? getInstance(instance!, n) : i
          )
        );
      } else if (updateType === UpdateType.Added) {
        if (dataInstances.find(i => i.name === instance?.metadata?.name!)) {
          setDataInstances(old => [
            ...old,
            getInstance(instance!, old?.length),
          ]);
        }
      }
    }

    const instancePhase = instance!.status?.phase;

    if (
      instancePhase === 'VmiReady' &&
      instance?.spec?.running &&
      (updateType === UpdateType.Added || updateType === UpdateType.Modified)
    ) {
      notification.success({
        message:
          instance!.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
            ?.itPolitoCrownlabsV1alpha2Template?.spec?.templateName,
        description: `${instance!.metadata?.name} started`,
        btn: instance!.status?.url && (
          <Button
            type="success"
            size="small"
            onClick={() => window.open(instance!.status?.url!, '_blank')}
          >
            Connect
          </Button>
        ),
      });
    }

    return prev;
  };
};

export const joinInstancesAndTemplates = (
  templates: Template[],
  instances: Instance[]
) =>
  templates.map(t => ({
    ...t,
    instances: instances.filter(i => i.idTemplate === t.id),
  }));
