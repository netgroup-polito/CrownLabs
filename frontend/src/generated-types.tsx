import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
import * as React from 'react';
import * as ApolloReactComponents from '@apollo/client/react/components';
export type Maybe<T> = T | null;
export type Exact<T extends { [key: string]: unknown }> = {
  [K in keyof T]: T[K];
};
export type MakeOptional<T, K extends keyof T> = Omit<T, K> &
  { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> &
  { [SubKey in K]: Maybe<T[SubKey]> };
export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
const defaultOptions = {};
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  JSON: any;
};

export type ContainerStartupOptions = {
  __typename?: 'ContainerStartupOptions';
  contentPath?: Maybe<Scalars['String']>;
  sourceArchiveURL?: Maybe<Scalars['String']>;
  startupArgs?: Maybe<Array<Maybe<Scalars['String']>>>;
};

export type ContainerStartupOptionsInput = {
  contentPath?: Maybe<Scalars['String']>;
  sourceArchiveURL?: Maybe<Scalars['String']>;
  startupArgs?: Maybe<Array<Maybe<Scalars['String']>>>;
};

export type EnvironmentListListItem = {
  __typename?: 'EnvironmentListListItem';
  containerStartupOptions?: Maybe<ContainerStartupOptions>;
  environmentType?: Maybe<EnvironmentType>;
  guiEnabled?: Maybe<Scalars['Boolean']>;
  image?: Maybe<Scalars['String']>;
  mode?: Maybe<Mode>;
  name?: Maybe<Scalars['String']>;
  persistent?: Maybe<Scalars['Boolean']>;
  resources?: Maybe<Resources>;
};

export type EnvironmentListListItemInput = {
  containerStartupOptions?: Maybe<ContainerStartupOptionsInput>;
  environmentType: EnvironmentType;
  guiEnabled?: Maybe<Scalars['Boolean']>;
  image: Scalars['String'];
  mode?: Maybe<Mode>;
  name: Scalars['String'];
  persistent?: Maybe<Scalars['Boolean']>;
  resources: ResourcesInput;
};

export type EnvironmentRef = {
  __typename?: 'EnvironmentRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
};

export type EnvironmentRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export enum EnvironmentType {
  VirtualMachine = 'VirtualMachine',
  Container = 'Container',
}

export type ImagesListItem = {
  __typename?: 'ImagesListItem';
  name?: Maybe<Scalars['String']>;
  versions?: Maybe<Array<Maybe<Scalars['String']>>>;
};

export type ImagesListItemInput = {
  name: Scalars['String'];
  versions: Array<Maybe<Scalars['String']>>;
};

export type InstanceRef = {
  __typename?: 'InstanceRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
};

export type InstanceRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input = {
  apiVersion?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Array<Maybe<Scalars['String']>>>;
  gracePeriodSeconds?: Maybe<Scalars['Float']>;
  kind?: Maybe<Scalars['String']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  preconditions?: Maybe<IoK8sApimachineryPkgApisMetaV1PreconditionsInput>;
  propagationPolicy?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1ListMeta = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ListMeta';
  continue?: Maybe<Scalars['String']>;
  remainingItemCount?: Maybe<Scalars['Float']>;
  resourceVersion?: Maybe<Scalars['String']>;
  selfLink?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry';
  apiVersion?: Maybe<Scalars['String']>;
  fieldsType?: Maybe<Scalars['String']>;
  fieldsV1?: Maybe<Scalars['String']>;
  manager?: Maybe<Scalars['String']>;
  operation?: Maybe<Scalars['String']>;
  time?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput = {
  apiVersion?: Maybe<Scalars['String']>;
  fieldsType?: Maybe<Scalars['String']>;
  fieldsV1?: Maybe<Scalars['String']>;
  manager?: Maybe<Scalars['String']>;
  operation?: Maybe<Scalars['String']>;
  time?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1ObjectMetaV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMetaV2';
  annotations?: Maybe<Scalars['JSON']>;
  clusterName?: Maybe<Scalars['String']>;
  creationTimestamp?: Maybe<Scalars['String']>;
  deletionGracePeriodSeconds?: Maybe<Scalars['Float']>;
  deletionTimestamp?: Maybe<Scalars['String']>;
  finalizers?: Maybe<Array<Maybe<Scalars['String']>>>;
  generateName?: Maybe<Scalars['String']>;
  generation?: Maybe<Scalars['Float']>;
  labels?: Maybe<Scalars['JSON']>;
  managedFields?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry>>
  >;
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
  ownerReferences?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2>>
  >;
  resourceVersion?: Maybe<Scalars['String']>;
  selfLink?: Maybe<Scalars['String']>;
  uid?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input = {
  annotations?: Maybe<Scalars['JSON']>;
  clusterName?: Maybe<Scalars['String']>;
  creationTimestamp?: Maybe<Scalars['String']>;
  deletionGracePeriodSeconds?: Maybe<Scalars['Float']>;
  deletionTimestamp?: Maybe<Scalars['String']>;
  finalizers?: Maybe<Array<Maybe<Scalars['String']>>>;
  generateName?: Maybe<Scalars['String']>;
  generation?: Maybe<Scalars['Float']>;
  labels?: Maybe<Scalars['JSON']>;
  managedFields?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput>>
  >;
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
  ownerReferences?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2Input>>
  >;
  resourceVersion?: Maybe<Scalars['String']>;
  selfLink?: Maybe<Scalars['String']>;
  uid?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2';
  apiVersion?: Maybe<Scalars['String']>;
  blockOwnerDeletion?: Maybe<Scalars['Boolean']>;
  controller?: Maybe<Scalars['Boolean']>;
  kind?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  uid?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2Input = {
  apiVersion: Scalars['String'];
  blockOwnerDeletion?: Maybe<Scalars['Boolean']>;
  controller?: Maybe<Scalars['Boolean']>;
  kind: Scalars['String'];
  name: Scalars['String'];
  uid: Scalars['String'];
};

export type IoK8sApimachineryPkgApisMetaV1PreconditionsInput = {
  resourceVersion?: Maybe<Scalars['String']>;
  uid?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1StatusCause = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusCause';
  field?: Maybe<Scalars['String']>;
  message?: Maybe<Scalars['String']>;
  reason?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1StatusDetailsV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusDetailsV2';
  causes?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1StatusCause>>>;
  group?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  name?: Maybe<Scalars['String']>;
  retryAfterSeconds?: Maybe<Scalars['Int']>;
  uid?: Maybe<Scalars['String']>;
};

export type IoK8sApimachineryPkgApisMetaV1StatusV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusV2';
  apiVersion?: Maybe<Scalars['String']>;
  code?: Maybe<Scalars['Int']>;
  details?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusDetailsV2>;
  kind?: Maybe<Scalars['String']>;
  message?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
  reason?: Maybe<Scalars['String']>;
  status?: Maybe<Scalars['String']>;
};

export type ItPolitoCrownlabsV1alpha1ImageList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageList';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec>;
  status?: Maybe<Scalars['String']>;
};

export type ItPolitoCrownlabsV1alpha1ImageListInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<SpecInput>;
  status?: Maybe<Scalars['String']>;
};

export type ItPolitoCrownlabsV1alpha1ImageListList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1ImageList>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1ImageListUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
};

export type ItPolitoCrownlabsV1alpha1Tenant = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec2>;
  status?: Maybe<Status>;
};

export type ItPolitoCrownlabsV1alpha1TenantInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<Spec2Input>;
  status?: Maybe<StatusInput>;
};

export type ItPolitoCrownlabsV1alpha1TenantList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1TenantList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1Tenant>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1TenantUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1TenantUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
};

export type ItPolitoCrownlabsV1alpha1Workspace = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec3>;
  status?: Maybe<Status2>;
};

export type ItPolitoCrownlabsV1alpha1WorkspaceInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<Spec3Input>;
  status?: Maybe<Status2Input>;
};

export type ItPolitoCrownlabsV1alpha1WorkspaceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1Workspace>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1WorkspaceUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

export type ItPolitoCrownlabsV1alpha2Instance = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec4>;
  status?: Maybe<Status3>;
};

export type ItPolitoCrownlabsV1alpha2InstanceInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<Spec4Input>;
  status?: Maybe<Status3Input>;
};

export type ItPolitoCrownlabsV1alpha2InstanceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2Instance>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2InstanceSnapshot = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshot';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec5>;
  status?: Maybe<Status4>;
};

export type ItPolitoCrownlabsV1alpha2InstanceSnapshotInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<Spec5Input>;
  status?: Maybe<Status4Input>;
};

export type ItPolitoCrownlabsV1alpha2InstanceSnapshotList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshotList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
};

export type ItPolitoCrownlabsV1alpha2InstanceUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
};

export type ItPolitoCrownlabsV1alpha2Template = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Template';
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  spec?: Maybe<Spec6>;
  status?: Maybe<Scalars['String']>;
};

export type ItPolitoCrownlabsV1alpha2TemplateInput = {
  apiVersion?: Maybe<Scalars['String']>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  spec?: Maybe<Spec6Input>;
  status?: Maybe<Scalars['String']>;
};

export type ItPolitoCrownlabsV1alpha2TemplateList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList';
  apiVersion?: Maybe<Scalars['String']>;
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2Template>>>;
  kind?: Maybe<Scalars['String']>;
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2TemplateUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

export enum Mode {
  Standard = 'Standard',
  Exam = 'Exam',
  Exercise = 'Exercise',
}

export type Mutation = {
  __typename?: 'Mutation';
  createCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  createCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  createCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  createCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  createCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  createCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  deleteCrownlabsPolitoItV1alpha1CollectionImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha1CollectionTenant?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha1CollectionWorkspace?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha1ImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha1Tenant?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha1Workspace?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  deleteCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  patchCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  patchCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  patchCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  patchCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  patchCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  patchCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  patchCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  patchCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  patchCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  replaceCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  replaceCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  replaceCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  replaceCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  replaceCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  replaceCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  replaceCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  replaceCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  replaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

export type MutationCreateCrownlabsPolitoItV1alpha1ImageListArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};

export type MutationCreateCrownlabsPolitoItV1alpha1TenantArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

export type MutationCreateCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1CollectionImageListArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1CollectionTenantArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1CollectionWorkspaceArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceSnapshotArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedTemplateArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

export type MutationPatchCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha1TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha1TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};

export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};

export type Namespace = {
  __typename?: 'Namespace';
  created?: Maybe<Scalars['Boolean']>;
  name?: Maybe<Scalars['String']>;
};

export type NamespaceInput = {
  created: Scalars['Boolean'];
  name?: Maybe<Scalars['String']>;
};

export type PersonalNamespace = {
  __typename?: 'PersonalNamespace';
  created?: Maybe<Scalars['Boolean']>;
  name?: Maybe<Scalars['String']>;
};

export type PersonalNamespaceInput = {
  created: Scalars['Boolean'];
  name?: Maybe<Scalars['String']>;
};

export type Query = {
  __typename?: 'Query';
  itPolitoCrownlabsV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  itPolitoCrownlabsV1alpha1ImageListList?: Maybe<ItPolitoCrownlabsV1alpha1ImageListList>;
  itPolitoCrownlabsV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  itPolitoCrownlabsV1alpha1TenantList?: Maybe<ItPolitoCrownlabsV1alpha1TenantList>;
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  itPolitoCrownlabsV1alpha1WorkspaceList?: Maybe<ItPolitoCrownlabsV1alpha1WorkspaceList>;
  itPolitoCrownlabsV1alpha2Instance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  itPolitoCrownlabsV1alpha2InstanceList?: Maybe<ItPolitoCrownlabsV1alpha2InstanceList>;
  itPolitoCrownlabsV1alpha2InstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotList?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotList>;
  itPolitoCrownlabsV1alpha2Template?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  itPolitoCrownlabsV1alpha2TemplateList?: Maybe<ItPolitoCrownlabsV1alpha2TemplateList>;
  listCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2InstanceList>;
  listCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotList>;
  listCrownlabsPolitoItV1alpha2TemplateForAllNamespaces?: Maybe<ItPolitoCrownlabsV1alpha2TemplateList>;
  readCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  readCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  readCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  readCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  readCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  readCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

export type QueryItPolitoCrownlabsV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha1ImageListListArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryItPolitoCrownlabsV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha1TenantListArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryItPolitoCrownlabsV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha1WorkspaceListArgs = {
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryItPolitoCrownlabsV1alpha2InstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha2InstanceListArgs = {
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryItPolitoCrownlabsV1alpha2InstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha2InstanceSnapshotListArgs = {
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryItPolitoCrownlabsV1alpha2TemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryItPolitoCrownlabsV1alpha2TemplateListArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryListCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryListCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryListCrownlabsPolitoItV1alpha2TemplateForAllNamespacesArgs = {
  allowWatchBookmarks?: Maybe<Scalars['Boolean']>;
  continue?: Maybe<Scalars['String']>;
  fieldSelector?: Maybe<Scalars['String']>;
  labelSelector?: Maybe<Scalars['String']>;
  limit?: Maybe<Scalars['Int']>;
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
  resourceVersionMatch?: Maybe<Scalars['String']>;
  timeoutSeconds?: Maybe<Scalars['Int']>;
  watch?: Maybe<Scalars['Boolean']>;
};

export type QueryReadCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryReadCrownlabsPolitoItV1alpha1TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryReadCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type QueryReadCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

export type Resources = {
  __typename?: 'Resources';
  cpu?: Maybe<Scalars['Int']>;
  disk?: Maybe<Scalars['String']>;
  memory?: Maybe<Scalars['String']>;
  reservedCPUPercentage?: Maybe<Scalars['Int']>;
};

export type ResourcesInput = {
  cpu: Scalars['Int'];
  disk?: Maybe<Scalars['String']>;
  memory: Scalars['String'];
  reservedCPUPercentage: Scalars['Int'];
};

export enum Role {
  Manager = 'manager',
  User = 'user',
}

export type SandboxNamespace = {
  __typename?: 'SandboxNamespace';
  created?: Maybe<Scalars['Boolean']>;
  name?: Maybe<Scalars['String']>;
};

export type SandboxNamespaceInput = {
  created: Scalars['Boolean'];
  name?: Maybe<Scalars['String']>;
};

export type Spec = {
  __typename?: 'Spec';
  images?: Maybe<Array<Maybe<ImagesListItem>>>;
  registryName?: Maybe<Scalars['String']>;
};

export type Spec2 = {
  __typename?: 'Spec2';
  createSandbox?: Maybe<Scalars['Boolean']>;
  email?: Maybe<Scalars['String']>;
  firstName?: Maybe<Scalars['String']>;
  lastName?: Maybe<Scalars['String']>;
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  workspaces?: Maybe<Array<Maybe<WorkspacesListItem>>>;
};

export type Spec2Input = {
  createSandbox?: Maybe<Scalars['Boolean']>;
  email: Scalars['String'];
  firstName: Scalars['String'];
  lastName: Scalars['String'];
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  workspaces?: Maybe<Array<Maybe<WorkspacesListItemInput>>>;
};

export type Spec3 = {
  __typename?: 'Spec3';
  prettyName?: Maybe<Scalars['String']>;
};

export type Spec3Input = {
  prettyName: Scalars['String'];
};

export type Spec4 = {
  __typename?: 'Spec4';
  running?: Maybe<Scalars['Boolean']>;
  templateCrownlabsPolitoItTemplateRef?: Maybe<TemplateCrownlabsPolitoItTemplateRef>;
  tenantCrownlabsPolitoItTenantRef?: Maybe<TenantCrownlabsPolitoItTenantRef>;
};

export type Spec4Input = {
  running?: Maybe<Scalars['Boolean']>;
  templateCrownlabsPolitoItTemplateRef: TemplateCrownlabsPolitoItTemplateRefInput;
  tenantCrownlabsPolitoItTenantRef: TenantCrownlabsPolitoItTenantRefInput;
};

export type Spec5 = {
  __typename?: 'Spec5';
  environmentRef?: Maybe<EnvironmentRef>;
  imageName?: Maybe<Scalars['String']>;
  instanceRef?: Maybe<InstanceRef>;
};

export type Spec5Input = {
  environmentRef?: Maybe<EnvironmentRefInput>;
  imageName: Scalars['String'];
  instanceRef: InstanceRefInput;
};

export type Spec6 = {
  __typename?: 'Spec6';
  deleteAfter?: Maybe<Scalars['String']>;
  description?: Maybe<Scalars['String']>;
  environmentList?: Maybe<Array<Maybe<EnvironmentListListItem>>>;
  prettyName?: Maybe<Scalars['String']>;
  workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<WorkspaceCrownlabsPolitoItWorkspaceRef>;
};

export type Spec6Input = {
  deleteAfter?: Maybe<Scalars['String']>;
  description: Scalars['String'];
  environmentList: Array<Maybe<EnvironmentListListItemInput>>;
  prettyName: Scalars['String'];
  workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<WorkspaceCrownlabsPolitoItWorkspaceRefInput>;
};

export type SpecInput = {
  images: Array<Maybe<ImagesListItemInput>>;
  registryName: Scalars['String'];
};

export type Status = {
  __typename?: 'Status';
  failingWorkspaces?: Maybe<Array<Maybe<Scalars['String']>>>;
  personalNamespace?: Maybe<PersonalNamespace>;
  ready?: Maybe<Scalars['Boolean']>;
  sandboxNamespace?: Maybe<SandboxNamespace>;
  subscriptions?: Maybe<Scalars['JSON']>;
};

export type Status2 = {
  __typename?: 'Status2';
  namespace?: Maybe<Namespace>;
  ready?: Maybe<Scalars['Boolean']>;
  subscription?: Maybe<Scalars['JSON']>;
};

export type Status2Input = {
  namespace?: Maybe<NamespaceInput>;
  ready?: Maybe<Scalars['Boolean']>;
  subscription?: Maybe<Scalars['JSON']>;
};

export type Status3 = {
  __typename?: 'Status3';
  initialReadyTime?: Maybe<Scalars['String']>;
  ip?: Maybe<Scalars['String']>;
  myDriveUrl?: Maybe<Scalars['String']>;
  phase?: Maybe<Scalars['String']>;
  url?: Maybe<Scalars['String']>;
};

export type Status3Input = {
  initialReadyTime?: Maybe<Scalars['String']>;
  ip?: Maybe<Scalars['String']>;
  myDriveUrl?: Maybe<Scalars['String']>;
  phase?: Maybe<Scalars['String']>;
  url?: Maybe<Scalars['String']>;
};

export type Status4 = {
  __typename?: 'Status4';
  phase?: Maybe<Scalars['String']>;
};

export type Status4Input = {
  phase: Scalars['String'];
};

export type StatusInput = {
  failingWorkspaces: Array<Maybe<Scalars['String']>>;
  personalNamespace: PersonalNamespaceInput;
  ready: Scalars['Boolean'];
  sandboxNamespace: SandboxNamespaceInput;
  subscriptions: Scalars['JSON'];
};

export type Subscription = {
  __typename?: 'Subscription';
  itPolitoCrownlabsV1alpha2InstanceUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate>;
  itPolitoCrownlabsV1alpha2TemplateUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TemplateUpdate>;
  itPolitoCrownlabsV1alpha1TenantUpdate?: Maybe<ItPolitoCrownlabsV1alpha1TenantUpdate>;
  itPolitoCrownlabsV1alpha1WorkspaceUpdate?: Maybe<ItPolitoCrownlabsV1alpha1WorkspaceUpdate>;
  itPolitoCrownlabsV1alpha1ImageListUpdate?: Maybe<ItPolitoCrownlabsV1alpha1ImageListUpdate>;
};

export type SubscriptionItPolitoCrownlabsV1alpha2InstanceUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha2InstanceSnapshotUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha2TemplateUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha1TenantUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha1WorkspaceUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha1ImageListUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type TemplateCrownlabsPolitoItTemplateRef = {
  __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
  templateWrapper?: Maybe<TemplateWrapper>;
};

export type TemplateCrownlabsPolitoItTemplateRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export type TemplateWrapper = {
  __typename?: 'TemplateWrapper';
  itPolitoCrownlabsV1alpha2Template?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

export type TenantCrownlabsPolitoItTenantRef = {
  __typename?: 'TenantCrownlabsPolitoItTenantRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
  tenantWrapper?: Maybe<TenantWrapper>;
};

export type TenantCrownlabsPolitoItTenantRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export type TenantWrapper = {
  __typename?: 'TenantWrapper';
  itPolitoCrownlabsV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
};

export enum UpdateType {
  Added = 'ADDED',
  Modified = 'MODIFIED',
  Deleted = 'DELETED',
}

export type WorkspaceCrownlabsPolitoItWorkspaceRef = {
  __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
};

export type WorkspaceCrownlabsPolitoItWorkspaceRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export type WorkspaceRef = {
  __typename?: 'WorkspaceRef';
  name?: Maybe<Scalars['String']>;
  namespace?: Maybe<Scalars['String']>;
  workspaceWrapper?: Maybe<WorkspaceWrapper>;
};

export type WorkspaceRefInput = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};

export type WorkspaceWrapper = {
  __typename?: 'WorkspaceWrapper';
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

export type WorkspacesListItem = {
  __typename?: 'WorkspacesListItem';
  groupNumber?: Maybe<Scalars['Int']>;
  role?: Maybe<Role>;
  workspaceRef?: Maybe<WorkspaceRef>;
};

export type WorkspacesListItemInput = {
  groupNumber?: Maybe<Scalars['Int']>;
  role: Role;
  workspaceRef: WorkspaceRefInput;
};

export type CreateInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String'];
  templateName: Scalars['String'];
  workspaceNamespace: Scalars['String'];
  tenantId: Scalars['String'];
  generateName?: Maybe<Scalars['String']>;
}>;

export type CreateInstanceMutation = {
  __typename?: 'Mutation';
  createdInstance?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
    status?: Maybe<{
      __typename?: 'Status3';
      ip?: Maybe<string>;
      phase?: Maybe<string>;
      url?: Maybe<string>;
    }>;
    spec?: Maybe<{
      __typename?: 'Spec4';
      running?: Maybe<boolean>;
      templateCrownlabsPolitoItTemplateRef?: Maybe<{
        __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
        name?: Maybe<string>;
        namespace?: Maybe<string>;
        templateWrapper?: Maybe<{
          __typename?: 'TemplateWrapper';
          itPolitoCrownlabsV1alpha2Template?: Maybe<{
            __typename?: 'ItPolitoCrownlabsV1alpha2Template';
            spec?: Maybe<{
              __typename?: 'Spec6';
              templateName?: Maybe<string>;
              templateDescription?: Maybe<string>;
              environmentList?: Maybe<
                Array<
                  Maybe<{
                    __typename?: 'EnvironmentListListItem';
                    guiEnabled?: Maybe<boolean>;
                    persistent?: Maybe<boolean>;
                  }>
                >
              >;
            }>;
          }>;
        }>;
      }>;
    }>;
  }>;
};

export type CreateTemplateMutationVariables = Exact<{
  workspaceName: Scalars['String'];
  workspaceNamespace: Scalars['String'];
  templateName: Scalars['String'];
  descriptionTemplate: Scalars['String'];
  image: Scalars['String'];
  guiEnabled: Scalars['Boolean'];
  persistent: Scalars['Boolean'];
  resources: ResourcesInput;
  templateId?: Maybe<Scalars['String']>;
  environmentType: EnvironmentType;
}>;

export type CreateTemplateMutation = {
  __typename?: 'Mutation';
  createdTemplate?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2Template';
    spec?: Maybe<{
      __typename?: 'Spec6';
      description?: Maybe<string>;
      name?: Maybe<string>;
      environmentList?: Maybe<
        Array<
          Maybe<{
            __typename?: 'EnvironmentListListItem';
            guiEnabled?: Maybe<boolean>;
            persistent?: Maybe<boolean>;
            resources?: Maybe<{
              __typename?: 'Resources';
              cpu?: Maybe<number>;
              disk?: Maybe<string>;
              memory?: Maybe<string>;
            }>;
          }>
        >
      >;
    }>;
    metadata?: Maybe<{
      __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMetaV2';
      id?: Maybe<string>;
    }>;
  }>;
};

export type DeleteInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String'];
  instanceName: Scalars['String'];
}>;

export type DeleteInstanceMutation = {
  __typename?: 'Mutation';
  deletedInstance?: Maybe<{
    __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusV2';
    kind?: Maybe<string>;
  }>;
};

export type DeleteTemplateMutationVariables = Exact<{
  workspaceNamespace: Scalars['String'];
  templateId: Scalars['String'];
}>;

export type DeleteTemplateMutation = {
  __typename?: 'Mutation';
  deletedTemplate?: Maybe<{
    __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusV2';
    kind?: Maybe<string>;
  }>;
};

export type OwnedInstancesQueryVariables = Exact<{
  tenantNamespace: Scalars['String'];
}>;

export type OwnedInstancesQuery = {
  __typename?: 'Query';
  instanceList?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList';
    instances?: Maybe<
      Array<
        Maybe<{
          __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
          status?: Maybe<{
            __typename?: 'Status3';
            ip?: Maybe<string>;
            phase?: Maybe<string>;
            url?: Maybe<string>;
          }>;
          spec?: Maybe<{
            __typename?: 'Spec4';
            running?: Maybe<boolean>;
            templateCrownlabsPolitoItTemplateRef?: Maybe<{
              __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
              name?: Maybe<string>;
              namespace?: Maybe<string>;
              templateWrapper?: Maybe<{
                __typename?: 'TemplateWrapper';
                itPolitoCrownlabsV1alpha2Template?: Maybe<{
                  __typename?: 'ItPolitoCrownlabsV1alpha2Template';
                  spec?: Maybe<{
                    __typename?: 'Spec6';
                    templateName?: Maybe<string>;
                    templateDescription?: Maybe<string>;
                    environmentList?: Maybe<
                      Array<
                        Maybe<{
                          __typename?: 'EnvironmentListListItem';
                          guiEnabled?: Maybe<boolean>;
                          persistent?: Maybe<boolean>;
                        }>
                      >
                    >;
                  }>;
                }>;
              }>;
            }>;
          }>;
        }>
      >
    >;
  }>;
};

export type SshKeysQueryVariables = Exact<{
  tenantId: Scalars['String'];
}>;

export type SshKeysQuery = {
  __typename?: 'Query';
  tenant?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
    spec?: Maybe<{
      __typename?: 'Spec2';
      email?: Maybe<string>;
      firstName?: Maybe<string>;
      lastName?: Maybe<string>;
      publicKeys?: Maybe<Array<Maybe<string>>>;
    }>;
  }>;
};

export type WorkspaceTemplatesQueryVariables = Exact<{
  workspaceNamespace: Scalars['String'];
}>;

export type WorkspaceTemplatesQuery = {
  __typename?: 'Query';
  templateList?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList';
    templates?: Maybe<
      Array<
        Maybe<{
          __typename?: 'ItPolitoCrownlabsV1alpha2Template';
          spec?: Maybe<{
            __typename?: 'Spec6';
            description?: Maybe<string>;
            name?: Maybe<string>;
            environmentList?: Maybe<
              Array<
                Maybe<{
                  __typename?: 'EnvironmentListListItem';
                  guiEnabled?: Maybe<boolean>;
                  persistent?: Maybe<boolean>;
                  resources?: Maybe<{
                    __typename?: 'Resources';
                    cpu?: Maybe<number>;
                    disk?: Maybe<string>;
                    memory?: Maybe<string>;
                  }>;
                }>
              >
            >;
          }>;
          metadata?: Maybe<{
            __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMetaV2';
            id?: Maybe<string>;
          }>;
        }>
      >
    >;
  }>;
};

export type TenantQueryVariables = Exact<{
  tenantId: Scalars['String'];
}>;

export type TenantQuery = {
  __typename?: 'Query';
  tenant?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
    spec?: Maybe<{
      __typename?: 'Spec2';
      email?: Maybe<string>;
      firstName?: Maybe<string>;
      lastName?: Maybe<string>;
      workspaces?: Maybe<
        Array<
          Maybe<{
            __typename?: 'WorkspacesListItem';
            role?: Maybe<Role>;
            workspaceRef?: Maybe<{
              __typename?: 'WorkspaceRef';
              workspaceId?: Maybe<string>;
              workspaceWrapper?: Maybe<{
                __typename?: 'WorkspaceWrapper';
                itPolitoCrownlabsV1alpha1Workspace?: Maybe<{
                  __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
                  spec?: Maybe<{
                    __typename?: 'Spec3';
                    workspaceName?: Maybe<string>;
                  }>;
                  status?: Maybe<{
                    __typename?: 'Status2';
                    namespace?: Maybe<{
                      __typename?: 'Namespace';
                      workspaceNamespace?: Maybe<string>;
                    }>;
                  }>;
                }>;
              }>;
            }>;
          }>
        >
      >;
    }>;
  }>;
};

export type UpdatedOwnedInstancesSubscriptionVariables = Exact<{
  tenantNamespace: Scalars['String'];
  instanceName: Scalars['String'];
}>;

export type UpdatedOwnedInstancesSubscription = {
  __typename?: 'Subscription';
  updateInstance?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate';
    instance?: Maybe<{
      __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
      status?: Maybe<{
        __typename?: 'Status3';
        ip?: Maybe<string>;
        phase?: Maybe<string>;
        url?: Maybe<string>;
      }>;
      spec?: Maybe<{
        __typename?: 'Spec4';
        running?: Maybe<boolean>;
        templateCrownlabsPolitoItTemplateRef?: Maybe<{
          __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
          name?: Maybe<string>;
          namespace?: Maybe<string>;
          templateWrapper?: Maybe<{
            __typename?: 'TemplateWrapper';
            itPolitoCrownlabsV1alpha2Template?: Maybe<{
              __typename?: 'ItPolitoCrownlabsV1alpha2Template';
              spec?: Maybe<{
                __typename?: 'Spec6';
                templateName?: Maybe<string>;
                templateDescription?: Maybe<string>;
                environmentList?: Maybe<
                  Array<
                    Maybe<{
                      __typename?: 'EnvironmentListListItem';
                      guiEnabled?: Maybe<boolean>;
                      persistent?: Maybe<boolean>;
                    }>
                  >
                >;
              }>;
            }>;
          }>;
        }>;
      }>;
    }>;
  }>;
};

export type UpdatedSshKeysSubscriptionVariables = Exact<{
  tenantId: Scalars['String'];
}>;

export type UpdatedSshKeysSubscription = {
  __typename?: 'Subscription';
  updatedTenant?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha1TenantUpdate';
    updatedKeys?: Maybe<{
      __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
      spec?: Maybe<{
        __typename?: 'Spec2';
        email?: Maybe<string>;
        firstName?: Maybe<string>;
        lastName?: Maybe<string>;
        publicKeys?: Maybe<Array<Maybe<string>>>;
      }>;
    }>;
  }>;
};

export type UpdatedWorkspaceTemplatesSubscriptionVariables = Exact<{
  workspaceNamespace: Scalars['String'];
  templateName: Scalars['String'];
}>;

export type UpdatedWorkspaceTemplatesSubscription = {
  __typename?: 'Subscription';
  updatedTemplate?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate';
    template?: Maybe<{
      __typename?: 'ItPolitoCrownlabsV1alpha2Template';
      spec?: Maybe<{
        __typename?: 'Spec6';
        description?: Maybe<string>;
        name?: Maybe<string>;
        environmentList?: Maybe<
          Array<
            Maybe<{
              __typename?: 'EnvironmentListListItem';
              guiEnabled?: Maybe<boolean>;
              persistent?: Maybe<boolean>;
              resources?: Maybe<{
                __typename?: 'Resources';
                cpu?: Maybe<number>;
                disk?: Maybe<string>;
                memory?: Maybe<string>;
              }>;
            }>
          >
        >;
      }>;
      metadata?: Maybe<{
        __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMetaV2';
        id?: Maybe<string>;
      }>;
    }>;
  }>;
};

export type UpdatedTenantSubscriptionVariables = Exact<{
  tenantId: Scalars['String'];
}>;

export type UpdatedTenantSubscription = {
  __typename?: 'Subscription';
  updatedTenant?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha1TenantUpdate';
    tenant?: Maybe<{
      __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
      spec?: Maybe<{
        __typename?: 'Spec2';
        email?: Maybe<string>;
        firstName?: Maybe<string>;
        lastName?: Maybe<string>;
        workspaces?: Maybe<
          Array<
            Maybe<{
              __typename?: 'WorkspacesListItem';
              role?: Maybe<Role>;
              workspaceRef?: Maybe<{
                __typename?: 'WorkspaceRef';
                workspaceId?: Maybe<string>;
                workspaceWrapper?: Maybe<{
                  __typename?: 'WorkspaceWrapper';
                  itPolitoCrownlabsV1alpha1Workspace?: Maybe<{
                    __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
                    spec?: Maybe<{
                      __typename?: 'Spec3';
                      workspaceName?: Maybe<string>;
                    }>;
                    status?: Maybe<{
                      __typename?: 'Status2';
                      namespace?: Maybe<{
                        __typename?: 'Namespace';
                        workspaceNamespace?: Maybe<string>;
                      }>;
                    }>;
                  }>;
                }>;
              }>;
            }>
          >
        >;
      }>;
    }>;
  }>;
};

export const CreateInstanceDocument = gql`
  mutation createInstance(
    $tenantNamespace: String!
    $templateName: String!
    $workspaceNamespace: String!
    $tenantId: String!
    $generateName: String = "instance-"
  ) {
    createdInstance: createCrownlabsPolitoItV1alpha2NamespacedInstance(
      namespace: $tenantNamespace
      itPolitoCrownlabsV1alpha2InstanceInput: {
        kind: "Instance"
        apiVersion: "crownlabs.polito.it/v1alpha2"
        metadata: { generateName: $generateName }
        spec: {
          templateCrownlabsPolitoItTemplateRef: {
            name: $templateName
            namespace: $workspaceNamespace
          }
          tenantCrownlabsPolitoItTenantRef: {
            name: $tenantId
            namespace: $tenantNamespace
          }
        }
      }
    ) {
      status {
        ip
        phase
        url
      }
      spec {
        running
        templateCrownlabsPolitoItTemplateRef {
          name
          namespace
          templateWrapper {
            itPolitoCrownlabsV1alpha2Template {
              spec {
                templateName: prettyName
                templateDescription: description
                environmentList {
                  guiEnabled
                  persistent
                }
              }
            }
          }
        }
      }
    }
  }
`;
export type CreateInstanceMutationFn = Apollo.MutationFunction<
  CreateInstanceMutation,
  CreateInstanceMutationVariables
>;
export type CreateInstanceComponentProps = Omit<
  ApolloReactComponents.MutationComponentOptions<
    CreateInstanceMutation,
    CreateInstanceMutationVariables
  >,
  'mutation'
>;

export const CreateInstanceComponent = (
  props: CreateInstanceComponentProps
) => (
  <ApolloReactComponents.Mutation<
    CreateInstanceMutation,
    CreateInstanceMutationVariables
  >
    mutation={CreateInstanceDocument}
    {...props}
  />
);

/**
 * __useCreateInstanceMutation__
 *
 * To run a mutation, you first call `useCreateInstanceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateInstanceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createInstanceMutation, { data, loading, error }] = useCreateInstanceMutation({
 *   variables: {
 *      tenantNamespace: // value for 'tenantNamespace'
 *      templateName: // value for 'templateName'
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      tenantId: // value for 'tenantId'
 *      generateName: // value for 'generateName'
 *   },
 * });
 */
export function useCreateInstanceMutation(
  baseOptions?: Apollo.MutationHookOptions<
    CreateInstanceMutation,
    CreateInstanceMutationVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useMutation<
    CreateInstanceMutation,
    CreateInstanceMutationVariables
  >(CreateInstanceDocument, options);
}
export type CreateInstanceMutationHookResult = ReturnType<
  typeof useCreateInstanceMutation
>;
export type CreateInstanceMutationResult = Apollo.MutationResult<CreateInstanceMutation>;
export type CreateInstanceMutationOptions = Apollo.BaseMutationOptions<
  CreateInstanceMutation,
  CreateInstanceMutationVariables
>;
export const CreateTemplateDocument = gql`
  mutation createTemplate(
    $workspaceName: String!
    $workspaceNamespace: String!
    $templateName: String!
    $descriptionTemplate: String!
    $image: String!
    $guiEnabled: Boolean!
    $persistent: Boolean!
    $resources: ResourcesInput!
    $templateId: String = "template-"
    $environmentType: EnvironmentType!
  ) {
    createdTemplate: createCrownlabsPolitoItV1alpha2NamespacedTemplate(
      namespace: $workspaceNamespace
      itPolitoCrownlabsV1alpha2TemplateInput: {
        kind: "Template"
        apiVersion: "crownlabs.polito.it/v1alpha2"
        spec: {
          prettyName: $templateName
          description: $descriptionTemplate
          environmentList: [
            {
              name: "environmentName"
              environmentType: $environmentType
              image: $image
              guiEnabled: $guiEnabled
              persistent: $persistent
              resources: $resources
            }
          ]
          workspaceCrownlabsPolitoItWorkspaceRef: { name: $workspaceName }
        }
        metadata: { generateName: $templateId, namespace: $workspaceNamespace }
      }
    ) {
      spec {
        name: prettyName
        description
        environmentList {
          guiEnabled
          persistent
          resources {
            cpu
            disk
            memory
          }
        }
      }
      metadata {
        id: name
      }
    }
  }
`;
export type CreateTemplateMutationFn = Apollo.MutationFunction<
  CreateTemplateMutation,
  CreateTemplateMutationVariables
>;
export type CreateTemplateComponentProps = Omit<
  ApolloReactComponents.MutationComponentOptions<
    CreateTemplateMutation,
    CreateTemplateMutationVariables
  >,
  'mutation'
>;

export const CreateTemplateComponent = (
  props: CreateTemplateComponentProps
) => (
  <ApolloReactComponents.Mutation<
    CreateTemplateMutation,
    CreateTemplateMutationVariables
  >
    mutation={CreateTemplateDocument}
    {...props}
  />
);

/**
 * __useCreateTemplateMutation__
 *
 * To run a mutation, you first call `useCreateTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useCreateTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [createTemplateMutation, { data, loading, error }] = useCreateTemplateMutation({
 *   variables: {
 *      workspaceName: // value for 'workspaceName'
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      templateName: // value for 'templateName'
 *      descriptionTemplate: // value for 'descriptionTemplate'
 *      image: // value for 'image'
 *      guiEnabled: // value for 'guiEnabled'
 *      persistent: // value for 'persistent'
 *      resources: // value for 'resources'
 *      templateId: // value for 'templateId'
 *      environmentType: // value for 'environmentType'
 *   },
 * });
 */
export function useCreateTemplateMutation(
  baseOptions?: Apollo.MutationHookOptions<
    CreateTemplateMutation,
    CreateTemplateMutationVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useMutation<
    CreateTemplateMutation,
    CreateTemplateMutationVariables
  >(CreateTemplateDocument, options);
}
export type CreateTemplateMutationHookResult = ReturnType<
  typeof useCreateTemplateMutation
>;
export type CreateTemplateMutationResult = Apollo.MutationResult<CreateTemplateMutation>;
export type CreateTemplateMutationOptions = Apollo.BaseMutationOptions<
  CreateTemplateMutation,
  CreateTemplateMutationVariables
>;
export const DeleteInstanceDocument = gql`
  mutation deleteInstance($tenantNamespace: String!, $instanceName: String!) {
    deletedInstance: deleteCrownlabsPolitoItV1alpha2NamespacedInstance(
      namespace: $tenantNamespace
      name: $instanceName
    ) {
      kind
    }
  }
`;
export type DeleteInstanceMutationFn = Apollo.MutationFunction<
  DeleteInstanceMutation,
  DeleteInstanceMutationVariables
>;
export type DeleteInstanceComponentProps = Omit<
  ApolloReactComponents.MutationComponentOptions<
    DeleteInstanceMutation,
    DeleteInstanceMutationVariables
  >,
  'mutation'
>;

export const DeleteInstanceComponent = (
  props: DeleteInstanceComponentProps
) => (
  <ApolloReactComponents.Mutation<
    DeleteInstanceMutation,
    DeleteInstanceMutationVariables
  >
    mutation={DeleteInstanceDocument}
    {...props}
  />
);

/**
 * __useDeleteInstanceMutation__
 *
 * To run a mutation, you first call `useDeleteInstanceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteInstanceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteInstanceMutation, { data, loading, error }] = useDeleteInstanceMutation({
 *   variables: {
 *      tenantNamespace: // value for 'tenantNamespace'
 *      instanceName: // value for 'instanceName'
 *   },
 * });
 */
export function useDeleteInstanceMutation(
  baseOptions?: Apollo.MutationHookOptions<
    DeleteInstanceMutation,
    DeleteInstanceMutationVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useMutation<
    DeleteInstanceMutation,
    DeleteInstanceMutationVariables
  >(DeleteInstanceDocument, options);
}
export type DeleteInstanceMutationHookResult = ReturnType<
  typeof useDeleteInstanceMutation
>;
export type DeleteInstanceMutationResult = Apollo.MutationResult<DeleteInstanceMutation>;
export type DeleteInstanceMutationOptions = Apollo.BaseMutationOptions<
  DeleteInstanceMutation,
  DeleteInstanceMutationVariables
>;
export const DeleteTemplateDocument = gql`
  mutation deleteTemplate($workspaceNamespace: String!, $templateId: String!) {
    deletedTemplate: deleteCrownlabsPolitoItV1alpha2NamespacedTemplate(
      namespace: $workspaceNamespace
      name: $templateId
    ) {
      kind
    }
  }
`;
export type DeleteTemplateMutationFn = Apollo.MutationFunction<
  DeleteTemplateMutation,
  DeleteTemplateMutationVariables
>;
export type DeleteTemplateComponentProps = Omit<
  ApolloReactComponents.MutationComponentOptions<
    DeleteTemplateMutation,
    DeleteTemplateMutationVariables
  >,
  'mutation'
>;

export const DeleteTemplateComponent = (
  props: DeleteTemplateComponentProps
) => (
  <ApolloReactComponents.Mutation<
    DeleteTemplateMutation,
    DeleteTemplateMutationVariables
  >
    mutation={DeleteTemplateDocument}
    {...props}
  />
);

/**
 * __useDeleteTemplateMutation__
 *
 * To run a mutation, you first call `useDeleteTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteTemplateMutation, { data, loading, error }] = useDeleteTemplateMutation({
 *   variables: {
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      templateId: // value for 'templateId'
 *   },
 * });
 */
export function useDeleteTemplateMutation(
  baseOptions?: Apollo.MutationHookOptions<
    DeleteTemplateMutation,
    DeleteTemplateMutationVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useMutation<
    DeleteTemplateMutation,
    DeleteTemplateMutationVariables
  >(DeleteTemplateDocument, options);
}
export type DeleteTemplateMutationHookResult = ReturnType<
  typeof useDeleteTemplateMutation
>;
export type DeleteTemplateMutationResult = Apollo.MutationResult<DeleteTemplateMutation>;
export type DeleteTemplateMutationOptions = Apollo.BaseMutationOptions<
  DeleteTemplateMutation,
  DeleteTemplateMutationVariables
>;
export const OwnedInstancesDocument = gql`
  query ownedInstances($tenantNamespace: String!) {
    instanceList: listCrownlabsPolitoItV1alpha2NamespacedInstance(
      namespace: $tenantNamespace
    ) {
      instances: items {
        status {
          ip
          phase
          url
        }
        spec {
          running
          templateCrownlabsPolitoItTemplateRef {
            name
            namespace
            templateWrapper {
              itPolitoCrownlabsV1alpha2Template {
                spec {
                  templateName: prettyName
                  templateDescription: description
                  environmentList {
                    guiEnabled
                    persistent
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;
export type OwnedInstancesComponentProps = Omit<
  ApolloReactComponents.QueryComponentOptions<
    OwnedInstancesQuery,
    OwnedInstancesQueryVariables
  >,
  'query'
> &
  (
    | { variables: OwnedInstancesQueryVariables; skip?: boolean }
    | { skip: boolean }
  );

export const OwnedInstancesComponent = (
  props: OwnedInstancesComponentProps
) => (
  <ApolloReactComponents.Query<
    OwnedInstancesQuery,
    OwnedInstancesQueryVariables
  >
    query={OwnedInstancesDocument}
    {...props}
  />
);

/**
 * __useOwnedInstancesQuery__
 *
 * To run a query within a React component, call `useOwnedInstancesQuery` and pass it any options that fit your needs.
 * When your component renders, `useOwnedInstancesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useOwnedInstancesQuery({
 *   variables: {
 *      tenantNamespace: // value for 'tenantNamespace'
 *   },
 * });
 */
export function useOwnedInstancesQuery(
  baseOptions: Apollo.QueryHookOptions<
    OwnedInstancesQuery,
    OwnedInstancesQueryVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(
    OwnedInstancesDocument,
    options
  );
}
export function useOwnedInstancesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    OwnedInstancesQuery,
    OwnedInstancesQueryVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useLazyQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(
    OwnedInstancesDocument,
    options
  );
}
export type OwnedInstancesQueryHookResult = ReturnType<
  typeof useOwnedInstancesQuery
>;
export type OwnedInstancesLazyQueryHookResult = ReturnType<
  typeof useOwnedInstancesLazyQuery
>;
export type OwnedInstancesQueryResult = Apollo.QueryResult<
  OwnedInstancesQuery,
  OwnedInstancesQueryVariables
>;
export const SshKeysDocument = gql`
  query sshKeys($tenantId: String!) {
    tenant: itPolitoCrownlabsV1alpha1Tenant(name: $tenantId) {
      spec {
        email
        firstName
        lastName
        publicKeys
      }
    }
  }
`;
export type SshKeysComponentProps = Omit<
  ApolloReactComponents.QueryComponentOptions<
    SshKeysQuery,
    SshKeysQueryVariables
  >,
  'query'
> &
  ({ variables: SshKeysQueryVariables; skip?: boolean } | { skip: boolean });

export const SshKeysComponent = (props: SshKeysComponentProps) => (
  <ApolloReactComponents.Query<SshKeysQuery, SshKeysQueryVariables>
    query={SshKeysDocument}
    {...props}
  />
);

/**
 * __useSshKeysQuery__
 *
 * To run a query within a React component, call `useSshKeysQuery` and pass it any options that fit your needs.
 * When your component renders, `useSshKeysQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useSshKeysQuery({
 *   variables: {
 *      tenantId: // value for 'tenantId'
 *   },
 * });
 */
export function useSshKeysQuery(
  baseOptions: Apollo.QueryHookOptions<SshKeysQuery, SshKeysQueryVariables>
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useQuery<SshKeysQuery, SshKeysQueryVariables>(
    SshKeysDocument,
    options
  );
}
export function useSshKeysLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<SshKeysQuery, SshKeysQueryVariables>
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useLazyQuery<SshKeysQuery, SshKeysQueryVariables>(
    SshKeysDocument,
    options
  );
}
export type SshKeysQueryHookResult = ReturnType<typeof useSshKeysQuery>;
export type SshKeysLazyQueryHookResult = ReturnType<typeof useSshKeysLazyQuery>;
export type SshKeysQueryResult = Apollo.QueryResult<
  SshKeysQuery,
  SshKeysQueryVariables
>;
export const WorkspaceTemplatesDocument = gql`
  query workspaceTemplates($workspaceNamespace: String!) {
    templateList: itPolitoCrownlabsV1alpha2TemplateList(
      namespace: $workspaceNamespace
    ) {
      templates: items {
        spec {
          name: prettyName
          description
          environmentList {
            guiEnabled
            persistent
            resources {
              cpu
              disk
              memory
            }
          }
        }
        metadata {
          id: name
        }
      }
    }
  }
`;
export type WorkspaceTemplatesComponentProps = Omit<
  ApolloReactComponents.QueryComponentOptions<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >,
  'query'
> &
  (
    | { variables: WorkspaceTemplatesQueryVariables; skip?: boolean }
    | { skip: boolean }
  );

export const WorkspaceTemplatesComponent = (
  props: WorkspaceTemplatesComponentProps
) => (
  <ApolloReactComponents.Query<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >
    query={WorkspaceTemplatesDocument}
    {...props}
  />
);

/**
 * __useWorkspaceTemplatesQuery__
 *
 * To run a query within a React component, call `useWorkspaceTemplatesQuery` and pass it any options that fit your needs.
 * When your component renders, `useWorkspaceTemplatesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useWorkspaceTemplatesQuery({
 *   variables: {
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *   },
 * });
 */
export function useWorkspaceTemplatesQuery(
  baseOptions: Apollo.QueryHookOptions<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useQuery<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >(WorkspaceTemplatesDocument, options);
}
export function useWorkspaceTemplatesLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useLazyQuery<
    WorkspaceTemplatesQuery,
    WorkspaceTemplatesQueryVariables
  >(WorkspaceTemplatesDocument, options);
}
export type WorkspaceTemplatesQueryHookResult = ReturnType<
  typeof useWorkspaceTemplatesQuery
>;
export type WorkspaceTemplatesLazyQueryHookResult = ReturnType<
  typeof useWorkspaceTemplatesLazyQuery
>;
export type WorkspaceTemplatesQueryResult = Apollo.QueryResult<
  WorkspaceTemplatesQuery,
  WorkspaceTemplatesQueryVariables
>;
export const TenantDocument = gql`
  query tenant($tenantId: String!) {
    tenant: itPolitoCrownlabsV1alpha1Tenant(name: $tenantId) {
      spec {
        email
        firstName
        lastName
        workspaces {
          role
          workspaceRef {
            workspaceId: name
            workspaceWrapper {
              itPolitoCrownlabsV1alpha1Workspace {
                spec {
                  workspaceName: prettyName
                }
                status {
                  namespace {
                    workspaceNamespace: name
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;
export type TenantComponentProps = Omit<
  ApolloReactComponents.QueryComponentOptions<
    TenantQuery,
    TenantQueryVariables
  >,
  'query'
> &
  ({ variables: TenantQueryVariables; skip?: boolean } | { skip: boolean });

export const TenantComponent = (props: TenantComponentProps) => (
  <ApolloReactComponents.Query<TenantQuery, TenantQueryVariables>
    query={TenantDocument}
    {...props}
  />
);

/**
 * __useTenantQuery__
 *
 * To run a query within a React component, call `useTenantQuery` and pass it any options that fit your needs.
 * When your component renders, `useTenantQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTenantQuery({
 *   variables: {
 *      tenantId: // value for 'tenantId'
 *   },
 * });
 */
export function useTenantQuery(
  baseOptions: Apollo.QueryHookOptions<TenantQuery, TenantQueryVariables>
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useQuery<TenantQuery, TenantQueryVariables>(
    TenantDocument,
    options
  );
}
export function useTenantLazyQuery(
  baseOptions?: Apollo.LazyQueryHookOptions<TenantQuery, TenantQueryVariables>
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useLazyQuery<TenantQuery, TenantQueryVariables>(
    TenantDocument,
    options
  );
}
export type TenantQueryHookResult = ReturnType<typeof useTenantQuery>;
export type TenantLazyQueryHookResult = ReturnType<typeof useTenantLazyQuery>;
export type TenantQueryResult = Apollo.QueryResult<
  TenantQuery,
  TenantQueryVariables
>;
export const UpdatedOwnedInstancesDocument = gql`
  subscription updatedOwnedInstances(
    $tenantNamespace: String!
    $instanceName: String!
  ) {
    updateInstance: itPolitoCrownlabsV1alpha2InstanceUpdate(
      namespace: $tenantNamespace
      name: $instanceName
    ) {
      instance: payload {
        status {
          ip
          phase
          url
        }
        spec {
          running
          templateCrownlabsPolitoItTemplateRef {
            name
            namespace
            templateWrapper {
              itPolitoCrownlabsV1alpha2Template {
                spec {
                  templateName: prettyName
                  templateDescription: description
                  environmentList {
                    guiEnabled
                    persistent
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;
export type UpdatedOwnedInstancesComponentProps = Omit<
  ApolloReactComponents.SubscriptionComponentOptions<
    UpdatedOwnedInstancesSubscription,
    UpdatedOwnedInstancesSubscriptionVariables
  >,
  'subscription'
>;

export const UpdatedOwnedInstancesComponent = (
  props: UpdatedOwnedInstancesComponentProps
) => (
  <ApolloReactComponents.Subscription<
    UpdatedOwnedInstancesSubscription,
    UpdatedOwnedInstancesSubscriptionVariables
  >
    subscription={UpdatedOwnedInstancesDocument}
    {...props}
  />
);

/**
 * __useUpdatedOwnedInstancesSubscription__
 *
 * To run a query within a React component, call `useUpdatedOwnedInstancesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useUpdatedOwnedInstancesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUpdatedOwnedInstancesSubscription({
 *   variables: {
 *      tenantNamespace: // value for 'tenantNamespace'
 *      instanceName: // value for 'instanceName'
 *   },
 * });
 */
export function useUpdatedOwnedInstancesSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    UpdatedOwnedInstancesSubscription,
    UpdatedOwnedInstancesSubscriptionVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useSubscription<
    UpdatedOwnedInstancesSubscription,
    UpdatedOwnedInstancesSubscriptionVariables
  >(UpdatedOwnedInstancesDocument, options);
}
export type UpdatedOwnedInstancesSubscriptionHookResult = ReturnType<
  typeof useUpdatedOwnedInstancesSubscription
>;
export type UpdatedOwnedInstancesSubscriptionResult = Apollo.SubscriptionResult<UpdatedOwnedInstancesSubscription>;
export const UpdatedSshKeysDocument = gql`
  subscription updatedSshKeys($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha1TenantUpdate(namespace: $tenantId) {
      updatedKeys: payload {
        spec {
          email
          firstName
          lastName
          publicKeys
        }
      }
    }
  }
`;
export type UpdatedSshKeysComponentProps = Omit<
  ApolloReactComponents.SubscriptionComponentOptions<
    UpdatedSshKeysSubscription,
    UpdatedSshKeysSubscriptionVariables
  >,
  'subscription'
>;

export const UpdatedSshKeysComponent = (
  props: UpdatedSshKeysComponentProps
) => (
  <ApolloReactComponents.Subscription<
    UpdatedSshKeysSubscription,
    UpdatedSshKeysSubscriptionVariables
  >
    subscription={UpdatedSshKeysDocument}
    {...props}
  />
);

/**
 * __useUpdatedSshKeysSubscription__
 *
 * To run a query within a React component, call `useUpdatedSshKeysSubscription` and pass it any options that fit your needs.
 * When your component renders, `useUpdatedSshKeysSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUpdatedSshKeysSubscription({
 *   variables: {
 *      tenantId: // value for 'tenantId'
 *   },
 * });
 */
export function useUpdatedSshKeysSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    UpdatedSshKeysSubscription,
    UpdatedSshKeysSubscriptionVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useSubscription<
    UpdatedSshKeysSubscription,
    UpdatedSshKeysSubscriptionVariables
  >(UpdatedSshKeysDocument, options);
}
export type UpdatedSshKeysSubscriptionHookResult = ReturnType<
  typeof useUpdatedSshKeysSubscription
>;
export type UpdatedSshKeysSubscriptionResult = Apollo.SubscriptionResult<UpdatedSshKeysSubscription>;
export const UpdatedWorkspaceTemplatesDocument = gql`
  subscription updatedWorkspaceTemplates(
    $workspaceNamespace: String!
    $templateName: String!
  ) {
    updatedTemplate: itPolitoCrownlabsV1alpha2TemplateUpdate(
      namespace: $workspaceNamespace
      name: $templateName
    ) {
      template: payload {
        spec {
          name: prettyName
          description
          environmentList {
            guiEnabled
            persistent
            resources {
              cpu
              disk
              memory
            }
          }
        }
        metadata {
          id: name
        }
      }
    }
  }
`;
export type UpdatedWorkspaceTemplatesComponentProps = Omit<
  ApolloReactComponents.SubscriptionComponentOptions<
    UpdatedWorkspaceTemplatesSubscription,
    UpdatedWorkspaceTemplatesSubscriptionVariables
  >,
  'subscription'
>;

export const UpdatedWorkspaceTemplatesComponent = (
  props: UpdatedWorkspaceTemplatesComponentProps
) => (
  <ApolloReactComponents.Subscription<
    UpdatedWorkspaceTemplatesSubscription,
    UpdatedWorkspaceTemplatesSubscriptionVariables
  >
    subscription={UpdatedWorkspaceTemplatesDocument}
    {...props}
  />
);

/**
 * __useUpdatedWorkspaceTemplatesSubscription__
 *
 * To run a query within a React component, call `useUpdatedWorkspaceTemplatesSubscription` and pass it any options that fit your needs.
 * When your component renders, `useUpdatedWorkspaceTemplatesSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUpdatedWorkspaceTemplatesSubscription({
 *   variables: {
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      templateName: // value for 'templateName'
 *   },
 * });
 */
export function useUpdatedWorkspaceTemplatesSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    UpdatedWorkspaceTemplatesSubscription,
    UpdatedWorkspaceTemplatesSubscriptionVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useSubscription<
    UpdatedWorkspaceTemplatesSubscription,
    UpdatedWorkspaceTemplatesSubscriptionVariables
  >(UpdatedWorkspaceTemplatesDocument, options);
}
export type UpdatedWorkspaceTemplatesSubscriptionHookResult = ReturnType<
  typeof useUpdatedWorkspaceTemplatesSubscription
>;
export type UpdatedWorkspaceTemplatesSubscriptionResult = Apollo.SubscriptionResult<UpdatedWorkspaceTemplatesSubscription>;
export const UpdatedTenantDocument = gql`
  subscription updatedTenant($tenantId: String!) {
    updatedTenant: itPolitoCrownlabsV1alpha1TenantUpdate(namespace: $tenantId) {
      tenant: payload {
        spec {
          email
          firstName
          lastName
          workspaces {
            role
            workspaceRef {
              workspaceId: name
              workspaceWrapper {
                itPolitoCrownlabsV1alpha1Workspace {
                  spec {
                    workspaceName: prettyName
                  }
                  status {
                    namespace {
                      workspaceNamespace: name
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
`;
export type UpdatedTenantComponentProps = Omit<
  ApolloReactComponents.SubscriptionComponentOptions<
    UpdatedTenantSubscription,
    UpdatedTenantSubscriptionVariables
  >,
  'subscription'
>;

export const UpdatedTenantComponent = (props: UpdatedTenantComponentProps) => (
  <ApolloReactComponents.Subscription<
    UpdatedTenantSubscription,
    UpdatedTenantSubscriptionVariables
  >
    subscription={UpdatedTenantDocument}
    {...props}
  />
);

/**
 * __useUpdatedTenantSubscription__
 *
 * To run a query within a React component, call `useUpdatedTenantSubscription` and pass it any options that fit your needs.
 * When your component renders, `useUpdatedTenantSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUpdatedTenantSubscription({
 *   variables: {
 *      tenantId: // value for 'tenantId'
 *   },
 * });
 */
export function useUpdatedTenantSubscription(
  baseOptions: Apollo.SubscriptionHookOptions<
    UpdatedTenantSubscription,
    UpdatedTenantSubscriptionVariables
  >
) {
  const options = { ...defaultOptions, ...baseOptions };
  return Apollo.useSubscription<
    UpdatedTenantSubscription,
    UpdatedTenantSubscriptionVariables
  >(UpdatedTenantDocument, options);
}
export type UpdatedTenantSubscriptionHookResult = ReturnType<
  typeof useUpdatedTenantSubscription
>;
export type UpdatedTenantSubscriptionResult = Apollo.SubscriptionResult<UpdatedTenantSubscription>;
