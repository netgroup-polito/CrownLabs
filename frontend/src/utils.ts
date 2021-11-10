export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export enum WorkspaceRole {
  user = 'user',
  manager = 'manager',
}
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
  name: string;
  ip: string;
  status: VmStatus;
  url: string | null;
  timeStamp?: string;
  workspaceId?: string;
  running?: boolean;
};
