import { useMemo } from 'react';
import {
  Phase2,
  type OwnedInstancesQuery,
  type TenantQuery,
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
type TenantFromQuery = TenantQuery['tenant']; // Remove NonNullable to allow null

// Parse memory values to decimal GB (base 10)
export const parseMemoryToGB = (
  v: string | number | null | undefined,
): number => {
  if (v == null) return 0;
  if (typeof v === 'number') return v;
  const s = String(v).trim();
  const m = s.match(/^([\d.]+)\s*(Ki|Mi|Gi|Ti|K|M|G|T)?$/i);
  if (!m) return parseFloat(s.replace(/[^\d.]/g, '')) || 0;
  const val = parseFloat(m[1]);
  const unit = (m[2] || '').toLowerCase();

  // Binary units (base 1024) - convert to GB
  if (unit === 'ki' || unit === 'kib') return (val * Math.pow(1024, 1)) / 1e9;
  if (unit === 'mi' || unit === 'mib') return (val * Math.pow(1024, 2)) / 1e9;
  if (unit === 'gi' || unit === 'gib') return (val * Math.pow(1024, 3)) / 1e9;
  if (unit === 'ti' || unit === 'tib') return (val * Math.pow(1024, 4)) / 1e9;

  // Decimal units (base 1000) - already in decimal
  if (unit === 'k' || unit === 'kb') return val / 1e6;
  if (unit === 'm' || unit === 'mb') return val / 1e3;
  if (unit === 'g' || unit === 'gb') return val;
  if (unit === 't' || unit === 'tb') return val * 1e3;

  // Default: assume GB
  return val;
};

// Keep the old function for backward compatibility (parses to GiB)
export const parseMemoryToGi = (
  v: string | number | null | undefined,
): number => {
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
const formatMemory = (memoryGB: number): string => {
  return Number(memoryGB.toFixed(1)).toString();
};

export const useQuotaCalculations = (
  instances: NonNullable<InstanceFromQuery>[] | undefined,
  tenant: TenantFromQuery | null | undefined, // Allow null
): QuotaCalculationResult => {
  return useMemo(() => {
    const consumedQuota = { cpu: 0, memoryGB: 0, instances: 0 };
    const items = instances ?? [];

    for (const inst of items) {
      // Count all instances not in 'Off' phase
      if (inst.status?.phase !== Phase2.Off) {
        const { cpu, mem } = (
          inst?.spec?.templateCrownlabsPolitoItTemplateRef?.templateWrapper
            ?.itPolitoCrownlabsV1alpha2Template?.spec?.environmentList ?? []
        ).reduce(
          (acc, env) => ({
            cpu: acc.cpu + Number(env?.resources?.cpu ?? 0),
            mem: acc.mem + parseMemoryToGB(env?.resources?.memory ?? '0Gi'),
          }),
          { cpu: 0, mem: 0 },
        );

        consumedQuota.cpu += cpu;
        consumedQuota.memoryGB += mem;
        consumedQuota.instances += 1;
      }
    }

    // Calculate available resources from quota - consumed
    const totalQuota = tenant?.status?.quota;
    const totalCpu = totalQuota?.cpu ? parseFloat(String(totalQuota.cpu)) : 0;
    const totalMemoryGB = totalQuota?.memory
      ? parseMemoryToGB(totalQuota.memory)
      : 0;
    const totalInstances = totalQuota?.instances ?? 0;

    const availableMemoryGB = Math.max(
      0,
      totalMemoryGB - consumedQuota.memoryGB,
    );

    return {
      consumedQuota: {
        cpu: consumedQuota.cpu,
        memory: formatMemory(consumedQuota.memoryGB),
        instances: consumedQuota.instances,
      },
      availableQuota: {
        cpu: Math.max(0, totalCpu - consumedQuota.cpu),
        memory: formatMemory(availableMemoryGB),
        instances: Math.max(0, totalInstances - consumedQuota.instances),
      },
      workspaceQuota: {
        cpu: totalCpu,
        memory: formatMemory(totalMemoryGB),
        instances: totalInstances,
      },
    };
  }, [instances, tenant?.status?.quota]);
};
