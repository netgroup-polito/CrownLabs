import React from 'react';
import type { QuotaProviderProps } from './QuotaContext.types';
import { QuotaContext } from './QuotaContext.types';

export const QuotaProvider: React.FC<QuotaProviderProps> = ({
  children,
  refreshQuota,
  consumedQuota,
  workspaceQuota,
  availableQuota,
}) => {
  return (
    <QuotaContext.Provider
      value={{
        refreshQuota,
        consumedQuota,
        workspaceQuota,
        availableQuota,
      }}
    >
      {children}
    </QuotaContext.Provider>
  );
};
