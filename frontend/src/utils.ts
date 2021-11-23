import { Dispatch, SetStateAction } from 'react';
import { EnvironmentType } from './generated-types';

export const AV_SSKey = 'prevViewActivePage';
export const AVID_SSKey = 'prevExpandedIdViewActivePage';
export const DV_SSKey = 'prevViewActivePage';
export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export enum WorkspaceRole {
  user = 'user',
  manager = 'manager',
}
export type User = { tenantId: string; tenantNamespace: string };
export type BadgeSize = 'small' | 'middle' | 'large';
export type BoxHeaderSize = 'small' | 'middle' | 'large';
export type Workspace = {
  id: string;
  title: string;
  role: WorkspaceRole;
  templates?: Array<Template>;
};
export type Resources = {
  cpu: number;
  disk: string;
  memory: string;
};
export type Template = {
  id: string;
  name: string;
  gui: boolean;
  persistent: boolean;
  resources: Resources;
  instances: Array<Instance>;
  workspaceId?: string;
};

export type VmStatus =
  | '' //the environment phase is unknown.
  | 'Importing' //the image of the environment is being imported.
  | 'Starting' //the environment is starting.
  | 'Running' //the environment is running, but not yet ready.
  | 'VmiReady' //the environment is ready to be accessed.
  | 'Stopping' //the environment is being stopped.
  | 'VmiOff' //the environment is currently shut down.
  | 'Failed' //the environment has failed, and cannot be restarted.
  | 'CreationLoopBackoff'; //the environment has encountered a temporary error during creation.
export type Instance = {
  id: number;
  gui?: boolean;
  idTemplate?: string;
  templatePrettyName?: string;
  persistent?: boolean;
  tenantId?: string;
  tenantDisplayName?: string;
  tenantNamespace?: string;
  environmentType?: EnvironmentType;
  name: string;
  prettyName?: string;
  ip: string;
  status: VmStatus;
  url: string | null;
  timeStamp?: string;
  workspaceId?: string;
  running?: boolean;
};

export function multiStringIncludes(needle: string, ...haystack: string[]) {
  for (const str of haystack)
    if (str.toLocaleLowerCase().includes(needle)) return true;
  return false;
}

/**
 * Create a callback that can be used to set a list state, by toggling the presence of the list of a given value.
 * @param setList the setter for the list
 * @param create specify if the returned function is used to create a new instance or not
 * @returns a callback which accepts a value and toggles the presence of that value in the list
 */
export function makeListToggler<T>(
  setList: Dispatch<SetStateAction<Array<T>>>
): (value: T, create: boolean) => void {
  return (value: T, create: boolean) => {
    setList(list =>
      list.includes(value)
        ? create
          ? list
          : list.filter(v => v !== value)
        : [...list, value]
    );
  };
}
