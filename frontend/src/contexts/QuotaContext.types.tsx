import { createContext, useContext } from 'react';

export interface QuotaContextType {
  refreshQuota?: () => void;
  consumedQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
  workspaceQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
  availableQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
}

export interface QuotaProviderProps {
  children: React.ReactNode;
  refreshQuota?: () => void;
  consumedQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
  workspaceQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
  availableQuota?: {
    cpu: number;
    memory: string;
    instances: number;
  };
}

export const QuotaContext = createContext<QuotaContextType>({});

export const useQuotaContext = () => useContext(QuotaContext);
