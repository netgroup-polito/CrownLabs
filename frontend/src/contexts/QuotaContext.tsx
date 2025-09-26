import React from 'react';
import type { QuotaProviderProps } from './QuotaContext.types';
import { QuotaContext } from './QuotaContext.types';

export const QuotaProvider: React.FC<QuotaProviderProps> = ({
  children,
  refreshQuota,
  availableQuota,
}) => {
  return (
    <QuotaContext.Provider value={{ refreshQuota, availableQuota }}>
      {children}
    </QuotaContext.Provider>
  );
};
