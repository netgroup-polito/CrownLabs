import { createContext, useContext } from 'react';

export interface QuotaContextType {
  refreshQuota?: () => void;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
}

export interface QuotaProviderProps {
  children: React.ReactNode;
  refreshQuota?: () => void;
  availableQuota?: {
    cpu?: string | number;
    memory?: string;
    instances?: number;
  };
}

export const QuotaContext = createContext<QuotaContextType>({});

export const useQuotaContext = () => useContext(QuotaContext);
