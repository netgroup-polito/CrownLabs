import { createContext } from 'react';
import type { OwnedInstancesQuery } from '../generated-types';
import type { Instance } from '../utils';

// Type for a single instance from the GraphQL query
type RawInstance = NonNullable<
  NonNullable<OwnedInstancesQuery['instanceList']>['instances']
>[number];

export interface IQuota {
  instances: number;
  cpu: number;
  memory: number;
  disk: number;
}

interface IOwnedInstancesContext {
  data?: OwnedInstancesQuery;
  rawInstances: RawInstance[]; // Raw GraphQL instances for quota calculations
  instances: Instance[]; // Transformed GUI instances
  loading: boolean;
  error?: Error;
  refetch: () => Promise<void>;

  // Quotas
  /** Map of used quotas per workspace */
  consumedQuota: Record<string, IQuota>;
  /** Map of total quotas per workspace */
  totalQuota: Record<string, IQuota>;
  /** Map of available quotas per workspace */
  availableQuota: Record<string, IQuota>;
}

export const OwnedInstancesContext = createContext<IOwnedInstancesContext>({
  data: undefined,
  rawInstances: [],
  instances: [],
  loading: false,
  error: undefined,
  refetch: async () => {},
  consumedQuota: {},
  totalQuota: {},
  availableQuota: {},
});
