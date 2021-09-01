export type someKeysOf<T> = { [key in keyof T]?: T[key] };
export type WorkspaceRole = 'user' | 'manager';
export type BadgeSize = 'small' | 'middle' | 'large';
export type BoxHeaderSize = 'small' | 'middle' | 'large';
export type Resources = {
  cpu: number;
  disk: number;
  memory: number;
};
export type Template = {
  id: string;
  name: string;
  gui: boolean;
  persistent: boolean;
  resources: Resources;
  instances: Array<Instance>;
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
  name: string;
  ip: string;
  status: VmStatus;
  url: string | null;
};
