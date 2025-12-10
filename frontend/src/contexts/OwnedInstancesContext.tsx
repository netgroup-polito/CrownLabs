import { createContext } from 'react';
import type { OwnedInstancesQuery } from '../generated-types';
import type { Instance } from '../utils';

// Type for a single instance from the GraphQL query
type RawInstance = NonNullable<
  NonNullable<OwnedInstancesQuery['instanceList']>['instances']
>[number];

interface IOwnedInstancesContext {
  data?: OwnedInstancesQuery;
  rawInstances: RawInstance[]; // Raw GraphQL instances for quota calculations
  instances: Instance[]; // Transformed GUI instances
  loading: boolean;
  error?: Error;
  refetch: () => Promise<void>;
}

export const OwnedInstancesContext = createContext<IOwnedInstancesContext>({
  data: undefined,
  rawInstances: [],
  instances: [],
  loading: false,
  error: undefined,
  refetch: async () => {},
});
