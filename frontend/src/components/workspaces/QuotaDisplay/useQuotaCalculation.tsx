import { useMemo } from 'react';
import type {
  OwnedInstancesQuery,
  TenantQuery,
} from '../../../generated-types';

interface QuotaCalculationResult {
  consumedQuota: {
    cpu: number;
    memory: string;
    instances: number;
  };
  availableQuota: {
    cpu: number;
    memory: string;
    instances: number;
  };
  workspaceQuota: {
    cpu: number;
    memory: string;
    instances: number;
  };
}

// Use the actual types returned by the queries
type InstanceFromQuery = NonNullable<
  NonNullable<OwnedInstancesQuery['instanceList']>['instances']
>[number];
type TenantFromQuery = NonNullable<TenantQuery['tenant']>;

export const useQuotaCalculations = (
  instances: NonNullable<InstanceFromQuery>[] | undefined,
  tenant: TenantFromQuery | undefined,
): QuotaCalculationResult => {
  return useMemo(() => {
    const parseMemoryToGi = (v: string | number | null | undefined): number => {
      if (v == null) return 0;
      if (typeof v === 'number') return v;
      const s = String(v).trim();
      const m = s.match(/^([\d.]+)\s*(Ki|Mi|Gi|Ti|K|M|G|T)?$/i);
      if (!m) return parseFloat(s.replace(/[^\d.]/g, '')) || 0;
      const val = parseFloat(m[1]);
      const unit = (m[2] || '').toLowerCase();
      const pow = (n: number) => Math.pow(1024, n);
      if (unit === 'ki') return (val * pow(1)) / pow(3);
      if (unit === 'mi') return (val * pow(2)) / pow(3);
      if (unit === 'gi') return val;
      if (unit === 'ti') return (val * pow(4)) / pow(3);
      if (unit === 'k') return (val * 1e3) / pow(3);
      if (unit === 'm') return (val * 1e6) / pow(3);
      if (unit === 'g') return (val * 1e9) / pow(3);
      if (unit === 't') return (val * 1e12) / pow(3);
      return val;
    };

    // Helper function to format memory with max 1 decimal place
    const formatMemory = (memoryGi: number): string => {
      return Number(memoryGi.toFixed(1)).toString();
    };

    const consumedQuota = { cpu: 0, memoryGi: 0, instances: 0 };
    const items = instances ?? [];

    for (const inst of items) {
      const resources =
        inst?.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
          ?.itPolitoCrownlabsV1alpha2Template?.spec?.environmentList?.[0]
          ?.resources;
      const cpu = Number(resources?.cpu ?? 0);
      const mem = resources?.memory ?? '0Gi';
      consumedQuota.cpu += cpu;
      consumedQuota.memoryGi += parseMemoryToGi(mem);
      consumedQuota.instances += 1;
    }

    // Calculate available resources from quota - consumed
    const totalQuota = tenant?.status?.quota;
    const totalCpu = totalQuota?.cpu ? parseFloat(String(totalQuota.cpu)) : 0;
    const totalMemoryGi = totalQuota?.memory
      ? parseMemoryToGi(totalQuota.memory)
      : 0;
    const totalInstances = totalQuota?.instances ?? 0;

    const availableMemoryGi = Math.max(
      0,
      totalMemoryGi - consumedQuota.memoryGi,
    );

    return {
      consumedQuota: {
        cpu: consumedQuota.cpu,
        memory: formatMemory(consumedQuota.memoryGi),
        instances: consumedQuota.instances,
      },
      availableQuota: {
        cpu: Math.max(0, totalCpu - consumedQuota.cpu),
        memory: formatMemory(availableMemoryGi),
        instances: Math.max(0, totalInstances - consumedQuota.instances),
      },
      workspaceQuota: {
        cpu: totalCpu,
        memory: formatMemory(totalMemoryGi),
        instances: totalInstances,
      },
    };
  }, [instances, tenant?.status?.quota]);
};
