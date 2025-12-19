import type { IQuota } from '../contexts/OwnedInstancesContext';
import type { TenantQuery } from '../generated-types';
import { convertToGB, type Instance } from '../utils';

export function calculateWorkspaceConsumedQuota(
  instances?: Instance[],
): Record<string, IQuota> {
  if (!instances) return {};

  const workspaceUsedResources: Record<string, IQuota> = {};

  instances.forEach(instance => {
    if (!workspaceUsedResources[instance.workspaceName]) {
      workspaceUsedResources[instance.workspaceName] = {
        instances: 0,
        cpu: 0,
        memory: 0,
        disk: 0,
      };
    }

    workspaceUsedResources[instance.workspaceName].instances += 1;
    workspaceUsedResources[instance.workspaceName].cpu +=
      instance.resources.cpu;
    workspaceUsedResources[instance.workspaceName].memory +=
      instance.resources.memory;
    workspaceUsedResources[instance.workspaceName].disk +=
      instance.resources.disk;
  });

  return workspaceUsedResources;
}

export function calculateWorkspaceTotalQuota(
  tenantData: TenantQuery | undefined,
): Record<string, IQuota> {
  if (!tenantData) return {};

  const quotas =
    tenantData?.tenant?.spec?.workspaces?.reduce(
      (map, workspace) => {
        const workspaceName = workspace?.name || '';
        const workspaceQuota =
          workspace?.workspaceWrapperTenantV1alpha2
            ?.itPolitoCrownlabsV1alpha1Workspace?.spec?.quota;

        return {
          ...map,
          [workspaceName]: {
            instances: workspaceQuota?.instances || 0,
            cpu: workspaceQuota?.cpu ? parseFloat(workspaceQuota.cpu) || 0 : 0,
            memory: workspaceQuota?.memory
              ? convertToGB(workspaceQuota.memory)
              : 0,
            disk: 0, // TODO: add disk quota when available
          },
        };
      },
      {} as Record<string, IQuota>,
    ) || {};

  // Add personal workspace quota (if enabled)
  const personalWorkspaceQuota = tenantData?.tenant?.spec?.quota;
  quotas['personal'] = {
    instances: personalWorkspaceQuota?.instances || 0,
    cpu: personalWorkspaceQuota?.cpu
      ? parseFloat(personalWorkspaceQuota?.cpu)
      : 0,
    memory: personalWorkspaceQuota?.memory
      ? convertToGB(personalWorkspaceQuota?.memory)
      : 0,
    disk: 0, // TODO: add disk quota when available
  };

  return quotas;
}

export function calculateAvailableQuota(
  totalQuota: Record<string, IQuota>,
  consumedQuota: Record<string, IQuota>,
): Record<string, IQuota> {
  const availableQuota: Record<string, IQuota> = {};

  for (const workspace in totalQuota) {
    availableQuota[workspace] = {
      instances:
        (totalQuota[workspace]?.instances || 0) -
        (consumedQuota[workspace]?.instances || 0),
      cpu:
        (totalQuota[workspace]?.cpu || 0) -
        (consumedQuota[workspace]?.cpu || 0),
      memory:
        (totalQuota[workspace]?.memory || 0) -
        (consumedQuota[workspace]?.memory || 0),
      disk:
        (totalQuota[workspace]?.disk || 0) -
        (consumedQuota[workspace]?.disk || 0),
    };
  }

  return availableQuota;
}
