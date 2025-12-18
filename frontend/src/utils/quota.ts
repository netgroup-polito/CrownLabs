import type { IQuota } from '../contexts/OwnedInstancesContext';
import { type Instance } from '../utils';

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
