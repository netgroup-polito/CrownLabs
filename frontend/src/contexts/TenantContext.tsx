import type { ApolloError } from '@apollo/client';
import type { TenantQuery } from '../generated-types';
import { createContext } from 'react';
import type { JointContent } from 'antd/lib/message/interface';

export type Notifier = (
  type: 'warning' | 'success',
  key: string,
  content: JointContent,
) => void;

interface ITenantContext {
  data?: TenantQuery;
  displayName: string;
  loading?: boolean;
  error?: ApolloError;
  hasSSHKeys: boolean;
  now: Date;
  refreshClock: () => void;
  notify: Notifier;
}

export const TenantContext = createContext<ITenantContext>({
  data: undefined,
  displayName: '',
  loading: undefined,
  error: undefined,
  hasSSHKeys: false,
  now: new Date(),
  refreshClock: () => null,
  notify: () => void 0,
});
