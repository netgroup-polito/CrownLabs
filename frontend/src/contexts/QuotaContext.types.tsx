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

  // setters so other parts of the app can update the context reactively
  setConsumedQuota?: (q?: {
    cpu: number;
    memory: string;
    instances: number;
  }) => void;
  setWorkspaceQuota?: (q?: {
    cpu: number;
    memory: string;
    instances: number;
  }) => void;
  setAvailableQuota?: (q?: {
    cpu: number;
    memory: string;
    instances: number;
  }) => void;

  // allow registering a refresh function (e.g. DashboardLogic.refetch)
  setRefreshQuota?: (fn?: () => void) => void;
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
