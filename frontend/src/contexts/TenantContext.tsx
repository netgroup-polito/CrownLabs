import type { ApolloError } from '@apollo/client';
import type { TenantQuery } from '../generated-types';
import { createContext } from 'react';

interface ITenantContext {
  data?: TenantQuery;
  displayName: string;
  loading?: boolean;
  error?: ApolloError;
  hasSSHKeys: boolean;
  now: Date;
  refreshClock: () => void;
}

export const TenantContext = createContext<ITenantContext>({
  data: undefined,
  displayName: '',
  loading: undefined,
  error: undefined,
  hasSSHKeys: false,
  now: new Date(),
  refreshClock: () => null,
});
