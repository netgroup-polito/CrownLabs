import React, { useEffect, useState, useMemo } from 'react';
import type { QuotaProviderProps } from './QuotaContext.types';
import { QuotaContext } from './QuotaContext.types';

export const QuotaProvider: React.FC<QuotaProviderProps> = ({
  children,
  refreshQuota: initialRefresh,
  consumedQuota: initialConsumed,
  workspaceQuota: initialWorkspace,
  availableQuota: initialAvailable,
}) => {
  const [refreshQuota, setRefreshQuota] = useState<(() => void) | undefined>(
    initialRefresh,
  );
  const [consumedQuota, setConsumedQuota] = useState(initialConsumed);
  const [workspaceQuota, setWorkspaceQuota] = useState(initialWorkspace);
  const [availableQuota, setAvailableQuota] = useState(initialAvailable);

  // keep internal state in sync if provider receives initial props later
  useEffect(() => {
    if (initialConsumed !== undefined) setConsumedQuota(initialConsumed);
  }, [initialConsumed]);

  useEffect(() => {
    if (initialWorkspace !== undefined) setWorkspaceQuota(initialWorkspace);
  }, [initialWorkspace]);

  useEffect(() => {
    if (initialAvailable !== undefined) setAvailableQuota(initialAvailable);
  }, [initialAvailable]);

  useEffect(() => {
    if (initialRefresh !== undefined) setRefreshQuota(() => initialRefresh);
  }, [initialRefresh]);

  // make provider value stable to avoid needless re-renders of consumers
  const providerValue = useMemo(
    () => ({
      refreshQuota,
      consumedQuota,
      workspaceQuota,
      availableQuota,
      setConsumedQuota,
      setWorkspaceQuota,
      setAvailableQuota,
      setRefreshQuota,
    }),
    [refreshQuota, consumedQuota, workspaceQuota, availableQuota],
  );

  return (
    <QuotaContext.Provider value={providerValue}>
      {children}
    </QuotaContext.Provider>
  );
};
