import { createContext } from 'react';
import type { WorkspaceImagesQuery } from '../generated-types';

interface IWorkspaceImagesContext {
  data?: WorkspaceImagesQuery;
  loading: boolean;
  error?: Error;
  workspaceNamespace?: string;
}

export const WorkspaceImagesContext = createContext<IWorkspaceImagesContext>({
  data: undefined,
  loading: false,
  error: undefined,
  workspaceNamespace: undefined,
});
