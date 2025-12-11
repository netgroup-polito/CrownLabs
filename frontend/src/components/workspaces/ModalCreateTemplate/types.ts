import type { EnvironmentType } from '../../../generated-types';

export type Interval = {
  max: number;
  min: number;
};

export type Resources = {
  cpu: Interval;
  disk: Interval;
  ram: Interval;
};

export type TemplateForm = {
  name: string;
  environments: TemplateFormEnv[];
  deleteAfter: string;
  inactivityTimeout: string;
};

export type TemplateFormEnv = {
  name: string;
  environmentType: EnvironmentType;
  image: string;
  registry: string;
  gui: boolean;
  cpu: number;
  ram: number;
  persistent: boolean;
  disk: number;
  sharedVolumeMounts: TemplateFormEnvShVol[];
  rewriteUrl: boolean;
};

export type ChildFormItem = {
  parentFormName: number;
  restField: {
    fieldKey?: number | undefined;
  };
};

export interface TemplateFormEnvShVol {
  sharedVolume: string; // workspace/shvol
  mountPath: string;
  readOnly: boolean;
}
export type Image = {
  name: string;
  type: Array<EnvironmentType>;
  registry: string;
};

export type ImageList = {
  name: string;
  registryName: string;
  images: Array<{
    name: string;
    versions: Array<string>;
  }>;
};
