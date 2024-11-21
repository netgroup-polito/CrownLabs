import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
const defaultOptions = {} as const;
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  /** The `BigInt` scalar type represents non-fractional signed whole numeric values. */
  BigInt: { input: any; output: any; }
  /** The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf). */
  JSON: { input: any; output: any; }
};

export type Access2 = {
  __typename?: 'Access2';
  /** The action of the access */
  action?: Maybe<Scalars['String']['output']>;
  /** The effect of the access */
  effect?: Maybe<Scalars['String']['output']>;
  /** The resource of the access */
  resource?: Maybe<Scalars['String']['output']>;
};

export type Access2Input = {
  /** The action of the access */
  action?: InputMaybe<Scalars['String']['input']>;
  /** The effect of the access */
  effect?: InputMaybe<Scalars['String']['input']>;
  /** The resource of the access */
  resource?: InputMaybe<Scalars['String']['input']>;
};

/** The accessory of the artifact */
export type Accessory = {
  __typename?: 'Accessory';
  /** The artifact id of the accessory */
  artifactId?: Maybe<Scalars['BigInt']['output']>;
  /** The creation time of the accessory */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The artifact digest of the accessory */
  digest?: Maybe<Scalars['String']['output']>;
  /** The icon of the accessory */
  icon?: Maybe<Scalars['String']['output']>;
  /** The ID of the accessory */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The artifact size of the accessory */
  size?: Maybe<Scalars['BigInt']['output']>;
  /** The subject artifact id of the accessory */
  subjectArtifactId?: Maybe<Scalars['BigInt']['output']>;
  /** The artifact size of the accessory */
  type?: Maybe<Scalars['String']['output']>;
};

export enum Addition {
  BuildHistory = 'BUILD_HISTORY',
  Dependencies = 'DEPENDENCIES',
  ReadmeMd = 'README_MD',
  ValuesYaml = 'VALUES_YAML'
}

export type Artifact = {
  __typename?: 'Artifact';
  accessories?: Maybe<Array<Maybe<Accessory>>>;
  additionLinks?: Maybe<Scalars['JSON']['output']>;
  annotations?: Maybe<Scalars['JSON']['output']>;
  /** The digest of the artifact */
  digest?: Maybe<Scalars['String']['output']>;
  extraAttrs?: Maybe<Scalars['JSON']['output']>;
  /** The digest of the icon */
  icon?: Maybe<Scalars['String']['output']>;
  /** The ID of the artifact */
  id?: Maybe<Scalars['BigInt']['output']>;
  labels?: Maybe<Array<Maybe<Label>>>;
  /** The manifest media type of the artifact */
  manifestMediaType?: Maybe<Scalars['String']['output']>;
  /** The media type of the artifact */
  mediaType?: Maybe<Scalars['String']['output']>;
  /** The ID of the project that the artifact belongs to */
  projectId?: Maybe<Scalars['BigInt']['output']>;
  /** The latest pull time of the artifact */
  pullTime?: Maybe<Scalars['String']['output']>;
  /** The push time of the artifact */
  pushTime?: Maybe<Scalars['String']['output']>;
  references?: Maybe<Array<Maybe<Reference>>>;
  /** The ID of the repository that the artifact belongs to */
  repositoryId?: Maybe<Scalars['BigInt']['output']>;
  /** The scan overview attached in the metadata of tag */
  scanOverview?: Maybe<Scalars['JSON']['output']>;
  /** The size of the artifact */
  size?: Maybe<Scalars['BigInt']['output']>;
  tags?: Maybe<Array<Maybe<Tag>>>;
  /** The type of the artifact, e.g. image, chart, etc */
  type?: Maybe<Scalars['String']['output']>;
};

export type AuditLog = {
  __typename?: 'AuditLog';
  /** The ID of the audit log entry. */
  id?: Maybe<Scalars['Int']['output']>;
  /** The time when this operation is triggered. */
  opTime?: Maybe<Scalars['String']['output']>;
  /** The operation against the repository in this log entry. */
  operation?: Maybe<Scalars['String']['output']>;
  /** Name of the repository in this log entry. */
  resource?: Maybe<Scalars['String']['output']>;
  /** Tag of the repository in this log entry. */
  resourceType?: Maybe<Scalars['String']['output']>;
  /** Username of the user in this log entry. */
  username?: Maybe<Scalars['String']['output']>;
};

export type AuthproxySetting = {
  __typename?: 'AuthproxySetting';
  /** The fully qualified URI of login endpoint of authproxy, such as 'https://192.168.1.2:8443/login' */
  endpoint?: Maybe<Scalars['String']['output']>;
  /** The certificate to be pinned when connecting auth proxy. */
  serverCertificate?: Maybe<Scalars['String']['output']>;
  /** The flag to determine whether Harbor can skip search the user/group when adding him as a member. */
  skipSearch?: Maybe<Scalars['Boolean']['output']>;
  /** The fully qualified URI of token review endpoint of authproxy, such as 'https://192.168.1.2:8443/tokenreview' */
  tokenreivewEndpoint?: Maybe<Scalars['String']['output']>;
  /** The flag to determine whether Harbor should verify the certificate when connecting to the auth proxy. */
  verifyCert?: Maybe<Scalars['Boolean']['output']>;
};

export enum AutoEnroll {
  Immediate = 'IMMEDIATE',
  WithApproval = 'WITH_APPROVAL',
  Empty = '_EMPTY_'
}

/** Timestamps of the Instance automation phases (check, termination and submission). */
export type Automation = {
  __typename?: 'Automation';
  /** The last time the Instance desired status was checked. */
  lastCheckTime?: Maybe<Scalars['String']['output']>;
  /** The time the Instance content submission has been completed. */
  submissionTime?: Maybe<Scalars['String']['output']>;
  /** The (possibly expected) termination time of the Instance. */
  terminationTime?: Maybe<Scalars['String']['output']>;
};

/** Timestamps of the Instance automation phases (check, termination and submission). */
export type AutomationInput = {
  /** The last time the Instance desired status was checked. */
  lastCheckTime?: InputMaybe<Scalars['String']['input']>;
  /** The time the Instance content submission has been completed. */
  submissionTime?: InputMaybe<Scalars['String']['input']>;
  /** The (possibly expected) termination time of the Instance. */
  terminationTime?: InputMaybe<Scalars['String']['input']>;
};

export type BoolConfigItem = {
  __typename?: 'BoolConfigItem';
  /** The configure item can be updated or not */
  editable?: Maybe<Scalars['Boolean']['output']>;
  /** The boolean value of current config item */
  value?: Maybe<Scalars['Boolean']['output']>;
};

/** The CVE Allowlist for system or project */
export type CveAllowlist = {
  __typename?: 'CVEAllowlist';
  /** The creation time of the allowlist. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** the time for expiration of the allowlist, in the form of seconds since epoch.  This is an optional attribute, if it's not set the CVE allowlist does not expire. */
  expiresAt?: Maybe<Scalars['Int']['output']>;
  /** ID of the allowlist */
  id?: Maybe<Scalars['Int']['output']>;
  items?: Maybe<Array<Maybe<CveAllowlistItem>>>;
  /** ID of the project which the allowlist belongs to.  For system level allowlist this attribute is zero. */
  projectId?: Maybe<Scalars['Int']['output']>;
  /** The update time of the allowlist. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The item in CVE allowlist */
export type CveAllowlistItem = {
  __typename?: 'CVEAllowlistItem';
  /** The ID of the CVE, such as "CVE-2019-10164" */
  cveId?: Maybe<Scalars['String']['output']>;
};

/** A specified chart entry */
export type ChartVersion = {
  __typename?: 'ChartVersion';
  /** The API version of this chart */
  apiVersion: Scalars['String']['output'];
  /** The version of the application enclosed in the chart */
  appVersion: Scalars['String']['output'];
  /** The created time of the chart entry */
  created?: Maybe<Scalars['String']['output']>;
  /** Whether or not this chart is deprecated */
  deprecated?: Maybe<Scalars['Boolean']['output']>;
  /** A one-sentence description of chart */
  description?: Maybe<Scalars['String']['output']>;
  /** The digest value of the chart entry */
  digest?: Maybe<Scalars['String']['output']>;
  /** The name of template engine */
  engine: Scalars['String']['output'];
  /** The URL to the relevant project page */
  home?: Maybe<Scalars['String']['output']>;
  /** The URL to an icon file */
  icon: Scalars['String']['output'];
  /** A list of string keywords */
  keywords?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** A list of label */
  labels?: Maybe<Array<Maybe<Label>>>;
  /** The name of the chart */
  name: Scalars['String']['output'];
  /** A flag to indicate if the chart entry is removed */
  removed?: Maybe<Scalars['Boolean']['output']>;
  /** The URL to the source code of chart */
  sources?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** The urls of the chart entry */
  urls?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** A SemVer 2 version of chart */
  version: Scalars['String']['output'];
};

/** The health status of component */
export type ComponentHealthStatus = {
  __typename?: 'ComponentHealthStatus';
  /** (optional) The error message when the status is "unhealthy" */
  error?: Maybe<Scalars['String']['output']>;
  /** The component name */
  name?: Maybe<Scalars['String']['output']>;
  /** The health status of component */
  status?: Maybe<Scalars['String']['output']>;
};

export type ConfigurationsResponse = {
  __typename?: 'ConfigurationsResponse';
  auditLogForwardEndpoint?: Maybe<StringConfigItem>;
  authMode?: Maybe<StringConfigItem>;
  httpAuthproxyAdminGroups?: Maybe<StringConfigItem>;
  httpAuthproxyAdminUsernames?: Maybe<StringConfigItem>;
  httpAuthproxyEndpoint?: Maybe<StringConfigItem>;
  httpAuthproxyServerCertificate?: Maybe<StringConfigItem>;
  httpAuthproxySkipSearch?: Maybe<BoolConfigItem>;
  httpAuthproxyTokenreviewEndpoint?: Maybe<StringConfigItem>;
  httpAuthproxyVerifyCert?: Maybe<BoolConfigItem>;
  ldapBaseDn?: Maybe<StringConfigItem>;
  ldapFilter?: Maybe<StringConfigItem>;
  ldapGroupAdminDn?: Maybe<StringConfigItem>;
  ldapGroupAttributeName?: Maybe<StringConfigItem>;
  ldapGroupBaseDn?: Maybe<StringConfigItem>;
  ldapGroupMembershipAttribute?: Maybe<StringConfigItem>;
  ldapGroupSearchFilter?: Maybe<StringConfigItem>;
  ldapGroupSearchScope?: Maybe<IntegerConfigItem>;
  ldapScope?: Maybe<IntegerConfigItem>;
  ldapSearchDn?: Maybe<StringConfigItem>;
  ldapTimeout?: Maybe<IntegerConfigItem>;
  ldapUid?: Maybe<StringConfigItem>;
  ldapUrl?: Maybe<StringConfigItem>;
  ldapVerifyCert?: Maybe<BoolConfigItem>;
  notificationEnable?: Maybe<BoolConfigItem>;
  oidcAdminGroup?: Maybe<StringConfigItem>;
  oidcAutoOnboard?: Maybe<BoolConfigItem>;
  oidcClientId?: Maybe<StringConfigItem>;
  oidcEndpoint?: Maybe<StringConfigItem>;
  oidcExtraRedirectParms?: Maybe<StringConfigItem>;
  oidcGroupFilter?: Maybe<StringConfigItem>;
  oidcGroupsClaim?: Maybe<StringConfigItem>;
  oidcName?: Maybe<StringConfigItem>;
  oidcScope?: Maybe<StringConfigItem>;
  oidcUserClaim?: Maybe<StringConfigItem>;
  oidcVerifyCert?: Maybe<BoolConfigItem>;
  projectCreationRestriction?: Maybe<StringConfigItem>;
  quotaPerProjectEnable?: Maybe<BoolConfigItem>;
  readOnly?: Maybe<BoolConfigItem>;
  robotNamePrefix?: Maybe<StringConfigItem>;
  robotTokenDuration?: Maybe<IntegerConfigItem>;
  scanAllPolicy?: Maybe<ScanAllPolicy>;
  selfRegistration?: Maybe<BoolConfigItem>;
  sessionTimeout?: Maybe<IntegerConfigItem>;
  skipAuditLogDatabase?: Maybe<BoolConfigItem>;
  storagePerProject?: Maybe<IntegerConfigItem>;
  tokenExpiration?: Maybe<IntegerConfigItem>;
  uaaClientId?: Maybe<StringConfigItem>;
  uaaClientSecret?: Maybe<StringConfigItem>;
  uaaEndpoint?: Maybe<StringConfigItem>;
  uaaVerifyCert?: Maybe<BoolConfigItem>;
};

/** Options to customize container startup */
export type ContainerStartupOptions = {
  __typename?: 'ContainerStartupOptions';
  /** Path on which storage (EmptyDir/Storage) will be mounted and into which, if given in SourceArchiveURL, will be extracted the archive */
  contentPath?: Maybe<Scalars['String']['output']>;
  /** Whether forcing the container working directory to be the same as the contentPath (or default mydrive path if not specified) */
  enforceWorkdir?: Maybe<Scalars['Boolean']['output']>;
  /** URL from which GET the archive to be extracted into ContentPath */
  sourceArchiveURL?: Maybe<Scalars['String']['output']>;
  /** Arguments to be passed to the application container on startup */
  startupArgs?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

/** Options to customize container startup */
export type ContainerStartupOptionsInput = {
  /** Path on which storage (EmptyDir/Storage) will be mounted and into which, if given in SourceArchiveURL, will be extracted the archive */
  contentPath?: InputMaybe<Scalars['String']['input']>;
  /** Whether forcing the container working directory to be the same as the contentPath (or default mydrive path if not specified) */
  enforceWorkdir?: InputMaybe<Scalars['Boolean']['input']>;
  /** URL from which GET the archive to be extracted into ContentPath */
  sourceArchiveURL?: InputMaybe<Scalars['String']['input']>;
  /** Arguments to be passed to the application container on startup */
  startupArgs?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};

/** Optional urls for advanced integration features. */
export type CustomizationUrls = {
  __typename?: 'CustomizationUrls';
  /** URL to which POST an archive with the contents found (at instance termination) in Template.ContainerStartupOptions.ContentPath. */
  contentDestination?: Maybe<Scalars['String']['output']>;
  /** URL from which GET the archive to be extracted into Template.ContainerStartupOptions.ContentPath. This field, if set, OVERRIDES Template.ContainerStartupOptions.SourceArchiveURL. */
  contentOrigin?: Maybe<Scalars['String']['output']>;
  /** URL which is periodically checked (with a GET request) to determine automatic instance shutdown. Should return any 2xx status code if the instance has to keep running, any 4xx otherwise. In case of 2xx response, it should output a JSON with a `deadline` field containing a ISO_8601 compliant date/time string of the expected instance termination time. See instautoctrl.StatusCheckResponse for exact definition. */
  statusCheck?: Maybe<Scalars['String']['output']>;
};

/** Optional urls for advanced integration features. */
export type CustomizationUrlsInput = {
  /** URL to which POST an archive with the contents found (at instance termination) in Template.ContainerStartupOptions.ContentPath. */
  contentDestination?: InputMaybe<Scalars['String']['input']>;
  /** URL from which GET the archive to be extracted into Template.ContainerStartupOptions.ContentPath. This field, if set, OVERRIDES Template.ContainerStartupOptions.SourceArchiveURL. */
  contentOrigin?: InputMaybe<Scalars['String']['input']>;
  /** URL which is periodically checked (with a GET request) to determine automatic instance shutdown. Should return any 2xx status code if the instance has to keep running, any 4xx otherwise. In case of 2xx response, it should output a JSON with a `deadline` field containing a ISO_8601 compliant date/time string of the expected instance termination time. See instautoctrl.StatusCheckResponse for exact definition. */
  statusCheck?: InputMaybe<Scalars['String']['input']>;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItem = {
  __typename?: 'EnvironmentListListItem';
  /** Options to customize container startup */
  containerStartupOptions?: Maybe<ContainerStartupOptions>;
  /** For VNC based containers, hide the noVNC control bar when true */
  disableControls?: Maybe<Scalars['Boolean']['output']>;
  /** The type of environment to be instantiated, among VirtualMachine, Container, CloudVM and Standalone. */
  environmentType: EnvironmentType;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: Maybe<Scalars['Boolean']['output']>;
  /** The VM or container to be started when instantiating the environment. */
  image: Scalars['String']['output'];
  /** The mode associated with the environment (Standard, Exam, Exercise) */
  mode?: Maybe<Mode>;
  /** Whether the instance has to have the user's MyDrive volume */
  mountMyDriveVolume: Scalars['Boolean']['output'];
  /** The name identifying the specific environment. */
  name: Scalars['String']['output'];
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: Maybe<Scalars['Boolean']['output']>;
  /** The amount of computational resources associated with the environment. */
  resources: Resources;
  /** Whether the environment needs the URL Rewrite or not. */
  rewriteURL?: Maybe<Scalars['Boolean']['output']>;
  /** Name of the storage class to be used for the persistent volume (when needed) */
  storageClassName?: Maybe<Scalars['String']['output']>;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItemInput = {
  /** Options to customize container startup */
  containerStartupOptions?: InputMaybe<ContainerStartupOptionsInput>;
  /** For VNC based containers, hide the noVNC control bar when true */
  disableControls?: InputMaybe<Scalars['Boolean']['input']>;
  /** The type of environment to be instantiated, among VirtualMachine, Container, CloudVM and Standalone. */
  environmentType: EnvironmentType;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: InputMaybe<Scalars['Boolean']['input']>;
  /** The VM or container to be started when instantiating the environment. */
  image: Scalars['String']['input'];
  /** The mode associated with the environment (Standard, Exam, Exercise) */
  mode?: InputMaybe<Mode>;
  /** Whether the instance has to have the user's MyDrive volume */
  mountMyDriveVolume: Scalars['Boolean']['input'];
  /** The name identifying the specific environment. */
  name: Scalars['String']['input'];
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: InputMaybe<Scalars['Boolean']['input']>;
  /** The amount of computational resources associated with the environment. */
  resources: ResourcesInput;
  /** Whether the environment needs the URL Rewrite or not. */
  rewriteURL?: InputMaybe<Scalars['Boolean']['input']>;
  /** Name of the storage class to be used for the persistent volume (when needed) */
  storageClassName?: InputMaybe<Scalars['String']['input']>;
};

/**
 * Environment represents the reference to the environment to be snapshotted, in case more are
 * associated with the same Instance. If not specified, the first available environment is considered.
 */
export type EnvironmentRef = {
  __typename?: 'EnvironmentRef';
  /** The name of the resource to be referenced. */
  name: Scalars['String']['output'];
  /**
   * The namespace containing the resource to be referenced. It should be left
   * empty in case of cluster-wide resources.
   */
  namespace?: Maybe<Scalars['String']['output']>;
};

/**
 * Environment represents the reference to the environment to be snapshotted, in case more are
 * associated with the same Instance. If not specified, the first available environment is considered.
 */
export type EnvironmentRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String']['input'];
  /**
   * The namespace containing the resource to be referenced. It should be left
   * empty in case of cluster-wide resources.
   */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export enum EnvironmentType {
  CloudVm = 'CLOUD_VM',
  Container = 'CONTAINER',
  Standalone = 'STANDALONE',
  VirtualMachine = 'VIRTUAL_MACHINE'
}

export type ExecHistory = {
  __typename?: 'ExecHistory';
  /** the creation time of purge job. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** if purge job was deleted. */
  deleted?: Maybe<Scalars['Boolean']['output']>;
  /** the id of purge job. */
  id?: Maybe<Scalars['Int']['output']>;
  /** the job kind of purge job. */
  jobKind?: Maybe<Scalars['String']['output']>;
  /** the job name of purge job. */
  jobName?: Maybe<Scalars['String']['output']>;
  /** the job parameters of purge job. */
  jobParameters?: Maybe<Scalars['String']['output']>;
  /** the status of purge job. */
  jobStatus?: Maybe<Scalars['String']['output']>;
  schedule?: Maybe<ScheduleObj>;
  /** the update time of purge job. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type Execution = {
  __typename?: 'Execution';
  /** The end time of execution */
  endTime?: Maybe<Scalars['String']['output']>;
  extraAttrs?: Maybe<Scalars['JSON']['output']>;
  /** The ID of execution */
  id?: Maybe<Scalars['Int']['output']>;
  metrics?: Maybe<Metrics>;
  /** The start time of execution */
  startTime?: Maybe<Scalars['String']['output']>;
  /** The status of execution */
  status?: Maybe<Scalars['String']['output']>;
  /** The status message of execution */
  statusMessage?: Maybe<Scalars['String']['output']>;
  /** The trigger of execution */
  trigger?: Maybe<Scalars['String']['output']>;
  /** The vendor id of execution */
  vendorId?: Maybe<Scalars['Int']['output']>;
  /** The vendor type of execution */
  vendorType?: Maybe<Scalars['String']['output']>;
};

/** The style of the resource filter */
export type FilterStyle = {
  __typename?: 'FilterStyle';
  /** The filter style */
  style?: Maybe<Scalars['String']['output']>;
  /** The filter type */
  type?: Maybe<Scalars['String']['output']>;
  /** The filter values */
  values?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type GcHistory = {
  __typename?: 'GCHistory';
  /** the creation time of gc job. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** if gc job was deleted. */
  deleted?: Maybe<Scalars['Boolean']['output']>;
  /** the id of gc job. */
  id?: Maybe<Scalars['Int']['output']>;
  /** the job kind of gc job. */
  jobKind?: Maybe<Scalars['String']['output']>;
  /** the job name of gc job. */
  jobName?: Maybe<Scalars['String']['output']>;
  /** the job parameters of gc job. */
  jobParameters?: Maybe<Scalars['String']['output']>;
  /** the status of gc job. */
  jobStatus?: Maybe<Scalars['String']['output']>;
  schedule?: Maybe<ScheduleObj>;
  /** the update time of gc job. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type GeneralInfo = {
  __typename?: 'GeneralInfo';
  /** The auth mode of current Harbor instance. */
  authMode?: Maybe<Scalars['String']['output']>;
  authproxySettings?: Maybe<AuthproxySetting>;
  /** The current time of the server. */
  currentTime?: Maybe<Scalars['String']['output']>;
  /** The external URL of Harbor, with protocol. */
  externalUrl?: Maybe<Scalars['String']['output']>;
  /** The build version of Harbor. */
  harborVersion?: Maybe<Scalars['String']['output']>;
  /** Indicate whether there is a ca root cert file ready for download in the file system. */
  hasCaRoot?: Maybe<Scalars['Boolean']['output']>;
  /** The flag to indicate whether notification mechanism is enabled on Harbor instance. */
  notificationEnable?: Maybe<Scalars['Boolean']['output']>;
  /** Indicate who can create projects, it could be 'adminonly' or 'everyone'. */
  projectCreationRestriction?: Maybe<Scalars['String']['output']>;
  /** The flag to indicate whether Harbor is in readonly mode. */
  readOnly?: Maybe<Scalars['Boolean']['output']>;
  /** The storage provider's name of Harbor registry */
  registryStorageProviderName?: Maybe<Scalars['String']['output']>;
  /** The url of registry against which the docker command should be issued. */
  registryUrl?: Maybe<Scalars['String']['output']>;
  /** Indicate whether the Harbor instance enable user to register himself. */
  selfRegistration?: Maybe<Scalars['Boolean']['output']>;
  /** If the Harbor instance is deployed with nested chartmuseum. */
  withChartmuseum?: Maybe<Scalars['Boolean']['output']>;
  /** If the Harbor instance is deployed with nested notary. */
  withNotary?: Maybe<Scalars['Boolean']['output']>;
};

export type Icon = {
  __typename?: 'Icon';
  /** The base64 encoded content of the icon */
  content?: Maybe<Scalars['String']['output']>;
  /** The content type of the icon */
  contentType?: Maybe<Scalars['String']['output']>;
};

/** ImageListItem describes a single VM image. */
export type ImagesListItem = {
  __typename?: 'ImagesListItem';
  /** The name identifying a single image. */
  name: Scalars['String']['output'];
  /** The list of versions the image is available in. */
  versions: Array<Maybe<Scalars['String']['output']>>;
};

/** ImageListItem describes a single VM image. */
export type ImagesListItemInput = {
  /** The name identifying a single image. */
  name: Scalars['String']['input'];
  /** The list of versions the image is available in. */
  versions: Array<InputMaybe<Scalars['String']['input']>>;
};

export type ImmutableRule = {
  __typename?: 'ImmutableRule';
  action?: Maybe<Scalars['String']['output']>;
  disabled?: Maybe<Scalars['Boolean']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  params?: Maybe<Scalars['JSON']['output']>;
  priority?: Maybe<Scalars['Int']['output']>;
  scopeSelectors?: Maybe<Scalars['JSON']['output']>;
  tagSelectors?: Maybe<Array<Maybe<ImmutableSelector>>>;
  template?: Maybe<Scalars['String']['output']>;
};

export type ImmutableSelector = {
  __typename?: 'ImmutableSelector';
  decoration?: Maybe<Scalars['String']['output']>;
  extras?: Maybe<Scalars['String']['output']>;
  kind?: Maybe<Scalars['String']['output']>;
  pattern?: Maybe<Scalars['String']['output']>;
};

export type Instance = {
  __typename?: 'Instance';
  /** The auth credential data if exists */
  authInfo?: Maybe<Scalars['JSON']['output']>;
  /** The authentication way supported */
  authMode?: Maybe<Scalars['String']['output']>;
  /** Whether the instance is default or not */
  default?: Maybe<Scalars['Boolean']['output']>;
  /** Description of instance */
  description?: Maybe<Scalars['String']['output']>;
  /** Whether the instance is activated or not */
  enabled?: Maybe<Scalars['Boolean']['output']>;
  /** The service endpoint of this instance */
  endpoint?: Maybe<Scalars['String']['output']>;
  /** Unique ID */
  id?: Maybe<Scalars['Int']['output']>;
  /** Whether the instance endpoint is insecure or not */
  insecure?: Maybe<Scalars['Boolean']['output']>;
  /** Instance name */
  name?: Maybe<Scalars['String']['output']>;
  /** The timestamp of instance setting up */
  setupTimestamp?: Maybe<Scalars['BigInt']['output']>;
  /** The health status */
  status?: Maybe<Scalars['String']['output']>;
  /** Based on which driver, identified by ID */
  vendor?: Maybe<Scalars['String']['output']>;
};

/**
 * Instance is the reference to the persistent VM instance to be snapshotted.
 * The instance should not be running, otherwise it won't be possible to
 * steal the volume and extract its content.
 */
export type InstanceRef = {
  __typename?: 'InstanceRef';
  /** The name of the resource to be referenced. */
  name: Scalars['String']['output'];
  /**
   * The namespace containing the resource to be referenced. It should be left
   * empty in case of cluster-wide resources.
   */
  namespace?: Maybe<Scalars['String']['output']>;
};

/**
 * Instance is the reference to the persistent VM instance to be snapshotted.
 * The instance should not be running, otherwise it won't be possible to
 * steal the volume and extract its content.
 */
export type InstanceRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String']['input'];
  /**
   * The namespace containing the resource to be referenced. It should be left
   * empty in case of cluster-wide resources.
   */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type IntegerConfigItem = {
  __typename?: 'IntegerConfigItem';
  /** The configure item can be updated or not */
  editable?: Maybe<Scalars['Boolean']['output']>;
  /** The integer value of current config item */
  value?: Maybe<Scalars['Int']['output']>;
};

/** DeleteOptions may be provided when deleting an API object. */
export type IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** When present, indicates that modifications should not be persisted. An invalid or unrecognized dryRun directive will result in an error response and no further processing of the request. Valid values are: - All: all dry run stages will be processed */
  dryRun?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  /** The duration in seconds before the object should be deleted. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period for the specified type will be used. Defaults to a per object value if not specified. zero means delete immediately. */
  gracePeriodSeconds?: InputMaybe<Scalars['BigInt']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** Deprecated: please use the PropagationPolicy, this field will be deprecated in 1.7. Should the dependent objects be orphaned. If true/false, the "orphan" finalizer will be added to/removed from the object's finalizers list. Either this field or PropagationPolicy may be set, but not both. */
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  /** Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out. */
  preconditions?: InputMaybe<IoK8sApimachineryPkgApisMetaV1PreconditionsInput>;
  /** Whether and how garbage collection will be performed. Either this field or OrphanDependents may be set, but not both. The default policy is decided by the existing finalizer set in the metadata.finalizers and the resource-specific default policy. Acceptable values are: 'Orphan' - orphan the dependents; 'Background' - allow the garbage collector to delete the dependents in the background; 'Foreground' - a cascading policy that deletes all dependents in the foreground. */
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};

/** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
export type IoK8sApimachineryPkgApisMetaV1ListMeta = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ListMeta';
  /** continue may be set if the user set a limit on the number of items returned, and indicates that the server has more data available. The value is opaque and may be used to issue another request to the endpoint that served this list to retrieve the next set of available objects. Continuing a consistent list may not be possible if the server configuration has changed or more than a few minutes have passed. The resourceVersion field returned when using this continue value will be identical to the value in the first response, unless you have received this token from an error message. */
  continue?: Maybe<Scalars['String']['output']>;
  /** remainingItemCount is the number of subsequent items in the list which are not included in this list response. If the list request contained label or field selectors, then the number of remaining items is unknown and the field will be left unset and omitted during serialization. If the list is complete (either because it is not chunking or because this is the last chunk), then there are no more remaining items and this field will be left unset and omitted during serialization. Servers older than v1.15 do not set this field. The intended use of the remainingItemCount is *estimating* the size of a collection. Clients should not rely on the remainingItemCount to be set or to be exact. */
  remainingItemCount?: Maybe<Scalars['BigInt']['output']>;
  /** String that identifies the server's internal version of this object that can be used by clients to determine when objects have changed. Value must be treated as opaque by clients and passed unmodified back to the server. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency */
  resourceVersion?: Maybe<Scalars['String']['output']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: Maybe<Scalars['String']['output']>;
};

/** ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource that the fieldset applies to. */
export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry';
  /** APIVersion defines the version of this resource that this field set applies to. The format is "group/version" just like the top-level APIVersion field. It is necessary to track the version of a field set because it cannot be automatically converted. */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** FieldsType is the discriminator for the different fields format and version. There is currently only one possible value: "FieldsV1" */
  fieldsType?: Maybe<Scalars['String']['output']>;
  /**
   * FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.
   *
   * Each key is either a '.' representing the field itself, and will always map to an empty set, or a string representing a sub-field or item. The string will follow one of these four formats: 'f:<name>', where <name> is the name of a field in a struct, or key in a map 'v:<value>', where <value> is the exact json formatted value of a list item 'i:<index>', where <index> is position of a item in a list 'k:<keys>', where <keys> is a map of  a list item's key fields to their unique values If a key maps to an empty Fields value, the field that key represents is part of the set.
   *
   * The exact format is defined in sigs.k8s.io/structured-merge-diff
   */
  fieldsV1?: Maybe<Scalars['JSON']['output']>;
  /** Manager is an identifier of the workflow managing these fields. */
  manager?: Maybe<Scalars['String']['output']>;
  /** Operation is the type of operation which lead to this ManagedFieldsEntry being created. The only valid values for this field are 'Apply' and 'Update'. */
  operation?: Maybe<Scalars['String']['output']>;
  /** Subresource is the name of the subresource used to update that object, or empty string if the object was updated through the main resource. The value of this field is used to distinguish between managers, even if they share the same name. For example, a status update will be distinct from a regular update using the same manager name. Note that the APIVersion field is not related to the Subresource field and it always corresponds to the version of the main resource. */
  subresource?: Maybe<Scalars['String']['output']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  time?: Maybe<Scalars['String']['output']>;
};

/** ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource that the fieldset applies to. */
export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput = {
  /** APIVersion defines the version of this resource that this field set applies to. The format is "group/version" just like the top-level APIVersion field. It is necessary to track the version of a field set because it cannot be automatically converted. */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** FieldsType is the discriminator for the different fields format and version. There is currently only one possible value: "FieldsV1" */
  fieldsType?: InputMaybe<Scalars['String']['input']>;
  /**
   * FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.
   *
   * Each key is either a '.' representing the field itself, and will always map to an empty set, or a string representing a sub-field or item. The string will follow one of these four formats: 'f:<name>', where <name> is the name of a field in a struct, or key in a map 'v:<value>', where <value> is the exact json formatted value of a list item 'i:<index>', where <index> is position of a item in a list 'k:<keys>', where <keys> is a map of  a list item's key fields to their unique values If a key maps to an empty Fields value, the field that key represents is part of the set.
   *
   * The exact format is defined in sigs.k8s.io/structured-merge-diff
   */
  fieldsV1?: InputMaybe<Scalars['JSON']['input']>;
  /** Manager is an identifier of the workflow managing these fields. */
  manager?: InputMaybe<Scalars['String']['input']>;
  /** Operation is the type of operation which lead to this ManagedFieldsEntry being created. The only valid values for this field are 'Apply' and 'Update'. */
  operation?: InputMaybe<Scalars['String']['input']>;
  /** Subresource is the name of the subresource used to update that object, or empty string if the object was updated through the main resource. The value of this field is used to distinguish between managers, even if they share the same name. For example, a status update will be distinct from a regular update using the same manager name. Note that the APIVersion field is not related to the Subresource field and it always corresponds to the version of the main resource. */
  subresource?: InputMaybe<Scalars['String']['input']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  time?: InputMaybe<Scalars['String']['input']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMeta = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta';
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations */
  annotations?: Maybe<Scalars['JSON']['output']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  creationTimestamp?: Maybe<Scalars['String']['output']>;
  /** Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. */
  deletionGracePeriodSeconds?: Maybe<Scalars['BigInt']['output']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  deletionTimestamp?: Maybe<Scalars['String']['output']>;
  /** Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list. */
  finalizers?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /**
   * GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.
   *
   * If this field is specified and the generated name exists, the server will return a 409.
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: Maybe<Scalars['String']['output']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: Maybe<Scalars['BigInt']['output']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels */
  labels?: Maybe<Scalars['JSON']['output']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry>>>;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names */
  name?: Maybe<Scalars['String']['output']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces
   */
  namespace?: Maybe<Scalars['String']['output']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReference>>>;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: Maybe<Scalars['String']['output']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: Maybe<Scalars['String']['output']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#uids
   */
  uid?: Maybe<Scalars['String']['output']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMetaInput = {
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations */
  annotations?: InputMaybe<Scalars['JSON']['input']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  creationTimestamp?: InputMaybe<Scalars['String']['input']>;
  /** Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. */
  deletionGracePeriodSeconds?: InputMaybe<Scalars['BigInt']['input']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  deletionTimestamp?: InputMaybe<Scalars['String']['input']>;
  /** Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list. */
  finalizers?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  /**
   * GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.
   *
   * If this field is specified and the generated name exists, the server will return a 409.
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: InputMaybe<Scalars['String']['input']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: InputMaybe<Scalars['BigInt']['input']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels */
  labels?: InputMaybe<Scalars['JSON']['input']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: InputMaybe<Array<InputMaybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput>>>;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names */
  name?: InputMaybe<Scalars['String']['input']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces
   */
  namespace?: InputMaybe<Scalars['String']['input']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: InputMaybe<Array<InputMaybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceInput>>>;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: InputMaybe<Scalars['String']['input']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#uids
   */
  uid?: InputMaybe<Scalars['String']['input']>;
};

/** OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field. */
export type IoK8sApimachineryPkgApisMetaV1OwnerReference = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1OwnerReference';
  /** API version of the referent. */
  apiVersion: Scalars['String']['output'];
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. See https://kubernetes.io/docs/concepts/architecture/garbage-collection/#foreground-deletion for how the garbage collector interacts with this field and enforces the foreground deletion. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
  blockOwnerDeletion?: Maybe<Scalars['Boolean']['output']>;
  /** If true, this reference points to the managing controller. */
  controller?: Maybe<Scalars['Boolean']['output']>;
  /** Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind: Scalars['String']['output'];
  /** Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names */
  name: Scalars['String']['output'];
  /** UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#uids */
  uid: Scalars['String']['output'];
};

/** OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field. */
export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceInput = {
  /** API version of the referent. */
  apiVersion: Scalars['String']['input'];
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. See https://kubernetes.io/docs/concepts/architecture/garbage-collection/#foreground-deletion for how the garbage collector interacts with this field and enforces the foreground deletion. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
  blockOwnerDeletion?: InputMaybe<Scalars['Boolean']['input']>;
  /** If true, this reference points to the managing controller. */
  controller?: InputMaybe<Scalars['Boolean']['input']>;
  /** Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind: Scalars['String']['input'];
  /** Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#names */
  name: Scalars['String']['input'];
  /** UID of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#uids */
  uid: Scalars['String']['input'];
};

/** Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out. */
export type IoK8sApimachineryPkgApisMetaV1PreconditionsInput = {
  /** Specifies the target ResourceVersion */
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  /** Specifies the target UID. */
  uid?: InputMaybe<Scalars['String']['input']>;
};

/** Status is a return value for calls that don't return other objects. */
export type IoK8sApimachineryPkgApisMetaV1Status = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1Status';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Suggested HTTP return code for this status, 0 if not set. */
  code?: Maybe<Scalars['Int']['output']>;
  /** StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined. */
  details?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusDetails>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** A human-readable description of the status of this operation. */
  message?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
  /** A machine-readable description of why this operation is in the "Failure" status. If this value is empty there is no information available. A Reason clarifies an HTTP status code but does not override it. */
  reason?: Maybe<Scalars['String']['output']>;
  /** Status of the operation. One of: "Success" or "Failure". More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status */
  status?: Maybe<Scalars['String']['output']>;
};

/** StatusCause provides more information about an api.Status failure, including cases when multiple errors are encountered. */
export type IoK8sApimachineryPkgApisMetaV1StatusCause = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusCause';
  /**
   * The field of the resource that has caused this error, as named by its JSON serialization. May include dot and postfix notation for nested attributes. Arrays are zero-indexed.  Fields may appear more than once in an array of causes due to fields having multiple errors. Optional.
   *
   * Examples:
   *   "name" - the field "name" on the current resource
   *   "items[0].name" - the field "name" on the first array entry in "items"
   */
  field?: Maybe<Scalars['String']['output']>;
  /** A human-readable description of the cause of the error.  This field may be presented as-is to a reader. */
  message?: Maybe<Scalars['String']['output']>;
  /** A machine-readable description of the cause of the error. If this value is empty there is no information available. */
  reason?: Maybe<Scalars['String']['output']>;
};

/** StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined. */
export type IoK8sApimachineryPkgApisMetaV1StatusDetails = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusDetails';
  /** The Causes array includes more details associated with the StatusReason failure. Not all StatusReasons may provide detailed causes. */
  causes?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1StatusCause>>>;
  /** The group attribute of the resource associated with the status StatusReason. */
  group?: Maybe<Scalars['String']['output']>;
  /** The kind attribute of the resource associated with the status StatusReason. On some operations may differ from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** The name attribute of the resource associated with the status StatusReason (when there is a single name which can be described). */
  name?: Maybe<Scalars['String']['output']>;
  /** If specified, the time in seconds before the operation should be retried. Some errors may indicate the client must take an alternate action - for those errors this field may indicate how long to wait before taking the alternate action. */
  retryAfterSeconds?: Maybe<Scalars['Int']['output']>;
  /** UID of the resource. (when there is a single resource which can be described). More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names#uids */
  uid?: Maybe<Scalars['String']['output']>;
};

/** ImageList describes the available VM images in the CrownLabs registry. */
export type ItPolitoCrownlabsV1alpha1ImageList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** ImageListSpec is the specification of the desired state of the ImageList. */
  spec?: Maybe<Spec>;
  /** ImageListStatus reflects the most recently observed status of the ImageList. */
  status?: Maybe<Scalars['JSON']['output']>;
};

/** ImageList describes the available VM images in the CrownLabs registry. */
export type ItPolitoCrownlabsV1alpha1ImageListInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** ImageListSpec is the specification of the desired state of the ImageList. */
  spec?: InputMaybe<SpecInput>;
  /** ImageListStatus reflects the most recently observed status of the ImageList. */
  status?: InputMaybe<Scalars['JSON']['input']>;
};

/** ImageListList is a list of ImageList */
export type ItPolitoCrownlabsV1alpha1ImageListList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of imagelists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha1ImageList>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1ImageListUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  updateType?: Maybe<UpdateType>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1Workspace = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: Maybe<Spec2>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: Maybe<Status2>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1WorkspaceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: InputMaybe<Spec2Input>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: InputMaybe<Status2Input>;
};

/** WorkspaceList is a list of Workspace */
export type ItPolitoCrownlabsV1alpha1WorkspaceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of workspaces. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha1Workspace>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1WorkspaceUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  updateType?: Maybe<UpdateType>;
};

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2Instance = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: Maybe<Spec3>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: Maybe<Status3>;
};

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2InstanceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: InputMaybe<Spec3Input>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: InputMaybe<Status3Input>;
};

/** InstanceList is a list of Instance */
export type ItPolitoCrownlabsV1alpha2InstanceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of instances. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha2Instance>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

/** InstanceSnapshot is the Schema for the instancesnapshots API. */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshot = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshot';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: Maybe<Spec4>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: Maybe<Status4>;
};

/** InstanceSnapshot is the Schema for the instancesnapshots API. */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshotInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: InputMaybe<Spec4Input>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: InputMaybe<Status4Input>;
};

/** InstanceSnapshotList is a list of InstanceSnapshot */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshotList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshotList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of instancesnapshots. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  updateType?: Maybe<UpdateType>;
};

export type ItPolitoCrownlabsV1alpha2InstanceUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  updateType?: Maybe<UpdateType>;
};

/** Template describes the template of a CrownLabs environment to be instantiated. */
export type ItPolitoCrownlabsV1alpha2Template = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Template';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** TemplateSpec is the specification of the desired state of the Template. */
  spec?: Maybe<Spec5>;
  /** TemplateStatus reflects the most recently observed status of the Template. */
  status?: Maybe<Scalars['JSON']['output']>;
};

/** Template describes the template of a CrownLabs environment to be instantiated. */
export type ItPolitoCrownlabsV1alpha2TemplateInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** TemplateSpec is the specification of the desired state of the Template. */
  spec?: InputMaybe<Spec5Input>;
  /** TemplateStatus reflects the most recently observed status of the Template. */
  status?: InputMaybe<Scalars['JSON']['input']>;
};

/** TemplateList is a list of Template */
export type ItPolitoCrownlabsV1alpha2TemplateList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of templates. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha2Template>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2TemplateUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  updateType?: Maybe<UpdateType>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha2Tenant = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Tenant';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: Maybe<Spec6>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: Maybe<Status6>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha2TenantInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: InputMaybe<Scalars['String']['input']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: InputMaybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: InputMaybe<Spec6Input>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: InputMaybe<Status6Input>;
};

/** TenantList is a list of Tenant */
export type ItPolitoCrownlabsV1alpha2TenantList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TenantList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']['output']>;
  /** List of tenants. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items: Array<Maybe<ItPolitoCrownlabsV1alpha2Tenant>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']['output']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2TenantUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TenantUpdate';
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  updateType?: Maybe<UpdateType>;
};

/** the job queue info */
export type JobQueue = {
  __typename?: 'JobQueue';
  /** The count of jobs in the job queue */
  count?: Maybe<Scalars['Int']['output']>;
  /** The type of the job queue */
  jobType?: Maybe<Scalars['String']['output']>;
  /** The latency the job queue (seconds) */
  latency?: Maybe<Scalars['Int']['output']>;
  /** The paused status of the job queue */
  paused?: Maybe<Scalars['Boolean']['output']>;
};

export type Label = {
  __typename?: 'Label';
  /** The color the label */
  color?: Maybe<Scalars['String']['output']>;
  /** The creation time the label */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The description the label */
  description?: Maybe<Scalars['String']['output']>;
  /** The ID of the label */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The name the label */
  name?: Maybe<Scalars['String']['output']>;
  /** The ID of project that the label belongs to */
  projectId?: Maybe<Scalars['BigInt']['output']>;
  /** The scope the label */
  scope?: Maybe<Scalars['String']['output']>;
  /** The update time of the label */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The ldap configure properties */
export type LdapConfInput = {
  /** The base dn of ldap service. */
  ldapBaseDn?: InputMaybe<Scalars['String']['input']>;
  /** The connect timeout of ldap service(second). */
  ldapConnectionTimeout?: InputMaybe<Scalars['BigInt']['input']>;
  /** The serach filter of ldap service. */
  ldapFilter?: InputMaybe<Scalars['String']['input']>;
  /** The serach scope of ldap service. */
  ldapScope?: InputMaybe<Scalars['BigInt']['input']>;
  /** The search dn of ldap service. */
  ldapSearchDn?: InputMaybe<Scalars['String']['input']>;
  /** The search password of ldap service. */
  ldapSearchPassword?: InputMaybe<Scalars['String']['input']>;
  /** The serach uid from ldap service attributes. */
  ldapUid?: InputMaybe<Scalars['String']['input']>;
  /** The url of ldap service. */
  ldapUrl?: InputMaybe<Scalars['String']['input']>;
  /** Verify Ldap server certificate. */
  ldapVerifyCert?: InputMaybe<Scalars['Boolean']['input']>;
};

/** The ldap ping result */
export type LdapPingResult = {
  __typename?: 'LdapPingResult';
  /** The ping operation output message. */
  message?: Maybe<Scalars['String']['output']>;
  /** Test success */
  success?: Maybe<Scalars['Boolean']['output']>;
};

export type LdapUser = {
  __typename?: 'LdapUser';
  /** The user email address from "mail" or "email" attribute. */
  email?: Maybe<Scalars['String']['output']>;
  /** The user realname from "uid" or "cn" attribute. */
  realname?: Maybe<Scalars['String']['output']>;
  /** ldap username. */
  username?: Maybe<Scalars['String']['output']>;
};

export type Metadata = {
  __typename?: 'Metadata';
  /** icon */
  icon?: Maybe<Scalars['String']['output']>;
  /** id */
  id?: Maybe<Scalars['String']['output']>;
  /** maintainers */
  maintainers?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** name */
  name?: Maybe<Scalars['String']['output']>;
  /** source */
  source?: Maybe<Scalars['String']['output']>;
  /** version */
  version?: Maybe<Scalars['String']['output']>;
};

export type Metrics = {
  __typename?: 'Metrics';
  /** The count of error task */
  errorTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of pending task */
  pendingTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of running task */
  runningTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of scheduled task */
  scheduledTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of stopped task */
  stoppedTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of success task */
  successTaskCount?: Maybe<Scalars['Int']['output']>;
  /** The count of task */
  taskCount?: Maybe<Scalars['Int']['output']>;
};

export enum Mode {
  Exam = 'EXAM',
  Exercise = 'EXERCISE',
  Standard = 'STANDARD'
}

export type Mutation = {
  __typename?: 'Mutation';
  /**
   * create an ImageList
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  createCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * create a Workspace
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha1/workspaces
   */
  createCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * create an Instance
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances
   */
  createCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * create an InstanceSnapshot
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots
   */
  createCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * create a Template
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates
   */
  createCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * create a Tenant
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha2/tenants
   */
  createCrownlabsPolitoItV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * delete collection of ImageList
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  deleteCrownlabsPolitoItV1alpha1CollectionImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete collection of Workspace
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/workspaces
   */
  deleteCrownlabsPolitoItV1alpha1CollectionWorkspace?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete an ImageList
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  deleteCrownlabsPolitoItV1alpha1ImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete a Workspace
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}
   */
  deleteCrownlabsPolitoItV1alpha1Workspace?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete collection of Instance
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete collection of InstanceSnapshot
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete collection of Template
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete collection of Tenant
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/tenants
   */
  deleteCrownlabsPolitoItV1alpha2CollectionTenant?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete an Instance
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete an InstanceSnapshot
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete a Template
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * delete a Tenant
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/tenants/{name}
   */
  deleteCrownlabsPolitoItV1alpha2Tenant?: Maybe<IoK8sApimachineryPkgApisMetaV1Status>;
  /**
   * partially update the specified ImageList
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  patchCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * partially update status of the specified ImageList
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}/status
   */
  patchCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * partially update the specified Workspace
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}
   */
  patchCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * partially update status of the specified Workspace
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}/status
   */
  patchCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * partially update the specified Instance
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}
   */
  patchCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * partially update the specified InstanceSnapshot
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}
   */
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * partially update status of the specified InstanceSnapshot
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}/status
   */
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * partially update status of the specified Instance
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}/status
   */
  patchCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * partially update the specified Template
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}
   */
  patchCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * partially update status of the specified Template
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}/status
   */
  patchCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * partially update the specified Tenant
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/tenants/{name}
   */
  patchCrownlabsPolitoItV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * partially update status of the specified Tenant
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha2/tenants/{name}/status
   */
  patchCrownlabsPolitoItV1alpha2TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * Create a robot account
   *
   * Equivalent to POST /robots
   */
  reg_createRobot?: Maybe<RobotCreated>;
  /**
   * Create a robot account
   *
   * Equivalent to POST /projects/{project_name_or_id}/robots
   */
  reg_createRobotV1?: Maybe<RobotCreated>;
  /**
   * Deletes the specified scanner registration.
   *
   *
   * Equivalent to DELETE /scanners/{registration_id}
   */
  reg_deleteScanner?: Maybe<ScannerRegistration>;
  /**
   * Export scan data for selected projects
   *
   * Equivalent to POST /export/cve
   */
  reg_exportScanData?: Maybe<ScanDataExportJob>;
  /**
   * This endpoint ping the available ldap service for test related configuration parameters.
   *
   *
   * Equivalent to POST /ldap/ping
   */
  reg_pingLdap?: Maybe<LdapPingResult>;
  /**
   * Refresh the robot secret
   *
   * Equivalent to PATCH /robots/{robot_id}
   */
  reg_refreshSec?: Maybe<RobotSec>;
  /**
   * replace the specified ImageList
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  replaceCrownlabsPolitoItV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * replace status of the specified ImageList
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * replace the specified Workspace
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}
   */
  replaceCrownlabsPolitoItV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * replace status of the specified Workspace
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * replace the specified Instance
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * replace the specified InstanceSnapshot
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * replace status of the specified InstanceSnapshot
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * replace status of the specified Instance
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * replace the specified Template
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * replace status of the specified Template
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * replace the specified Tenant
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/tenants/{name}
   */
  replaceCrownlabsPolitoItV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * replace status of the specified Tenant
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha2/tenants/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha2TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
};


export type MutationCreateCrownlabsPolitoItV1alpha1ImageListArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationCreateCrownlabsPolitoItV1alpha2TenantArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha1CollectionImageListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha1CollectionWorkspaceArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha1ImageListArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceSnapshotArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionNamespacedTemplateArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionTenantArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationDeleteCrownlabsPolitoItV1alpha2TenantArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  gracePeriodSeconds?: InputMaybe<Scalars['Int']['input']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: InputMaybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
  name: Scalars['String']['input'];
  orphanDependents?: InputMaybe<Scalars['Boolean']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  propagationPolicy?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha1ImageListArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2TenantArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationPatchCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  applicationApplyPatchYamlInput: Scalars['String']['input'];
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  force?: InputMaybe<Scalars['Boolean']['input']>;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReg_CreateRobotArgs = {
  robotCreateInput: RobotCreateInput;
};


export type MutationReg_CreateRobotV1Args = {
  projectNameOrId: Scalars['String']['input'];
  robotCreateV1Input: RobotCreateV1Input;
};


export type MutationReg_DeleteScannerArgs = {
  registrationId: Scalars['String']['input'];
};


export type MutationReg_ExportScanDataArgs = {
  scanDataExportRequestInput: ScanDataExportRequestInput;
};


export type MutationReg_PingLdapArgs = {
  ldapConfInput?: InputMaybe<LdapConfInput>;
};


export type MutationReg_RefreshSecArgs = {
  robotId: Scalars['Int']['input'];
  robotSecInput: RobotSecInput;
};


export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2TenantArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};


export type MutationReplaceCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  dryRun?: InputMaybe<Scalars['String']['input']>;
  fieldManager?: InputMaybe<Scalars['String']['input']>;
  fieldValidation?: InputMaybe<Scalars['String']['input']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
};

/** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
export type Namespace = {
  __typename?: 'Namespace';
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['output'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']['output']>;
};

/** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
export type NamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['input'];
  /** The name of the considered resource. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type OidcUserInfo = {
  __typename?: 'OIDCUserInfo';
  /** The creation time of the OIDC user info record. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** the ID of the OIDC info record */
  id?: Maybe<Scalars['Int']['output']>;
  /** the secret of the OIDC user that can be used for CLI to push/pull artifacts */
  secret?: Maybe<Scalars['String']['output']>;
  /** the concatenation of sub and issuer in the ID token */
  subiss?: Maybe<Scalars['String']['output']>;
  /** The update time of the OIDC user info record. */
  updateTime?: Maybe<Scalars['String']['output']>;
  /** the ID of the user */
  userId?: Maybe<Scalars['Int']['output']>;
};

/** The system health status */
export type OverallHealthStatus = {
  __typename?: 'OverallHealthStatus';
  components?: Maybe<Array<Maybe<ComponentHealthStatus>>>;
  /** The overall health status. It is "healthy" only when all the components' status are "healthy" */
  status?: Maybe<Scalars['String']['output']>;
};

/** The parameters of the policy, the values are dependent on the type of the policy. */
export type Parameter = {
  __typename?: 'Parameter';
  /** The offset in seconds of UTC 0 o'clock, only valid when the policy type is "daily" */
  dailyTime?: Maybe<Scalars['Int']['output']>;
};

export type Permission = {
  __typename?: 'Permission';
  /** The permission action */
  action?: Maybe<Scalars['String']['output']>;
  /** The permission resoruce */
  resource?: Maybe<Scalars['String']['output']>;
};

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
export type PersonalNamespace = {
  __typename?: 'PersonalNamespace';
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['output'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']['output']>;
};

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
export type PersonalNamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['input'];
  /** The name of the considered resource. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export enum Phase {
  CreationLoopBackoff = 'CREATION_LOOP_BACKOFF',
  Failed = 'FAILED',
  Importing = 'IMPORTING',
  Off = 'OFF',
  Ready = 'READY',
  ResourceQuotaExceeded = 'RESOURCE_QUOTA_EXCEEDED',
  Running = 'RUNNING',
  Starting = 'STARTING',
  Stopping = 'STOPPING',
  Empty = '_EMPTY_'
}

export enum Phase2 {
  Completed = 'COMPLETED',
  Failed = 'FAILED',
  Pending = 'PENDING',
  Processing = 'PROCESSING',
  Empty = '_EMPTY_'
}

export type Platform = {
  __typename?: 'Platform';
  /** The architecture that the artifact applys to */
  architecture?: Maybe<Scalars['String']['output']>;
  /** The OS that the artifact applys to */
  os?: Maybe<Scalars['String']['output']>;
  /** The features of the OS that the artifact applys to */
  osFeatures?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** The version of the OS that the artifact applys to */
  osVersion?: Maybe<Scalars['String']['output']>;
  /** The variant of the CPU */
  variant?: Maybe<Scalars['String']['output']>;
};

export type PreheatPolicy = {
  __typename?: 'PreheatPolicy';
  /** The Create Time of preheat policy */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The Description of preheat policy */
  description?: Maybe<Scalars['String']['output']>;
  /** Whether the preheat policy enabled */
  enabled?: Maybe<Scalars['Boolean']['output']>;
  /** The Filters of preheat policy */
  filters?: Maybe<Scalars['String']['output']>;
  /** The ID of preheat policy */
  id?: Maybe<Scalars['Int']['output']>;
  /** The Name of preheat policy */
  name?: Maybe<Scalars['String']['output']>;
  /** The ID of preheat policy project */
  projectId?: Maybe<Scalars['Int']['output']>;
  /** The ID of preheat policy provider */
  providerId?: Maybe<Scalars['Int']['output']>;
  /** The Name of preheat policy provider */
  providerName?: Maybe<Scalars['String']['output']>;
  /** The Trigger of preheat policy */
  trigger?: Maybe<Scalars['String']['output']>;
  /** The Update Time of preheat policy */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type Project2 = {
  __typename?: 'Project2';
  /** The total number of charts under this project. */
  chartCount?: Maybe<Scalars['Int']['output']>;
  /** The creation time of the project. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The role ID with highest permission of the current user who triggered the API (for UI).  This attribute is deprecated and will be removed in future versions. */
  currentUserRoleId?: Maybe<Scalars['Int']['output']>;
  /** The list of role ID of the current user who triggered the API (for UI) */
  currentUserRoleIds?: Maybe<Array<Maybe<Scalars['Int']['output']>>>;
  /** The CVE Allowlist for system or project */
  cveAllowlist?: Maybe<CveAllowlist>;
  /** A deletion mark of the project. */
  deleted?: Maybe<Scalars['Boolean']['output']>;
  metadata?: Maybe<ProjectMetadata>;
  /** The name of the project. */
  name?: Maybe<Scalars['String']['output']>;
  /** The owner ID of the project always means the creator of the project. */
  ownerId?: Maybe<Scalars['Int']['output']>;
  /** The owner name of the project. */
  ownerName?: Maybe<Scalars['String']['output']>;
  /** Project ID */
  projectId?: Maybe<Scalars['Int']['output']>;
  /** The ID of referenced registry when the project is a proxy cache project. */
  registryId?: Maybe<Scalars['BigInt']['output']>;
  /** The number of the repositories under this project. */
  repoCount?: Maybe<Scalars['Int']['output']>;
  /** Correspond to the UI about whether the project's publicity is  updatable (for UI) */
  togglable?: Maybe<Scalars['Boolean']['output']>;
  /** The update time of the project. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type ProjectDeletable = {
  __typename?: 'ProjectDeletable';
  /** Whether the project can be deleted. */
  deletable?: Maybe<Scalars['Boolean']['output']>;
  /** The detail message when the project can not be deleted. */
  message?: Maybe<Scalars['String']['output']>;
};

export type ProjectMemberEntity = {
  __typename?: 'ProjectMemberEntity';
  /** the id of entity, if the member is a user, it is user_id in user table. if the member is a user group, it is the user group's ID in user_group table. */
  entityId?: Maybe<Scalars['Int']['output']>;
  /** the name of the group member. */
  entityName?: Maybe<Scalars['String']['output']>;
  /** the entity's type, u for user entity, g for group entity. */
  entityType?: Maybe<Scalars['String']['output']>;
  /** the project member id */
  id?: Maybe<Scalars['Int']['output']>;
  /** the project id */
  projectId?: Maybe<Scalars['Int']['output']>;
  /** the role id */
  roleId?: Maybe<Scalars['Int']['output']>;
  /** the name of the role */
  roleName?: Maybe<Scalars['String']['output']>;
};

export type ProjectMetadata = {
  __typename?: 'ProjectMetadata';
  /** Whether scan images automatically when pushing. The valid values are "true", "false". */
  autoScan?: Maybe<Scalars['String']['output']>;
  /** Whether content trust is enabled or not. If it is enabled, user can't pull unsigned images from this project. The valid values are "true", "false". */
  enableContentTrust?: Maybe<Scalars['String']['output']>;
  /** Whether cosign content trust is enabled or not. If it is enabled, user can't pull images without cosign signature from this project. The valid values are "true", "false". */
  enableContentTrustCosign?: Maybe<Scalars['String']['output']>;
  /** Whether prevent the vulnerable images from running. The valid values are "true", "false". */
  preventVul?: Maybe<Scalars['String']['output']>;
  /** The public status of the project. The valid values are "true", "false". */
  public?: Maybe<Scalars['String']['output']>;
  /** The ID of the tag retention policy for the project */
  retentionId?: Maybe<Scalars['String']['output']>;
  /** Whether this project reuse the system level CVE allowlist as the allowlist of its own.  The valid values are "true", "false". If it is set to "true" the actual allowlist associate with this project, if any, will be ignored. */
  reuseSysCveAllowlist?: Maybe<Scalars['String']['output']>;
  /** If the vulnerability is high than severity defined here, the images can't be pulled. The valid values are "none", "low", "medium", "high", "critical". */
  severity?: Maybe<Scalars['String']['output']>;
};

export type ProjectRepository = {
  __typename?: 'ProjectRepository';
  /** The count of the artifacts inside the repository */
  artifactCount?: Maybe<Scalars['BigInt']['output']>;
  /** The creation time of the repository */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The description of the repository */
  description?: Maybe<Scalars['String']['output']>;
  /** The ID of the repository */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The name of the repository */
  name?: Maybe<Scalars['String']['output']>;
  /** The ID of the project that the repository belongs to */
  projectId?: Maybe<Scalars['BigInt']['output']>;
  /** The count that the artifact inside the repository pulled */
  pullCount?: Maybe<Scalars['BigInt']['output']>;
  /** The update time of the repository */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type ProjectSummary = {
  __typename?: 'ProjectSummary';
  /** The total number of charts under this project. */
  chartCount?: Maybe<Scalars['Int']['output']>;
  /** The total number of developer members. */
  developerCount?: Maybe<Scalars['Int']['output']>;
  /** The total number of guest members. */
  guestCount?: Maybe<Scalars['Int']['output']>;
  /** The total number of limited guest members. */
  limitedGuestCount?: Maybe<Scalars['Int']['output']>;
  /** The total number of maintainer members. */
  maintainerCount?: Maybe<Scalars['Int']['output']>;
  /** The total number of project admin members. */
  projectAdminCount?: Maybe<Scalars['Int']['output']>;
  quota?: Maybe<ProjectSummaryQuota>;
  registry?: Maybe<Registry>;
  /** The number of the repositories under this project. */
  repoCount?: Maybe<Scalars['Int']['output']>;
};

export type ProjectSummaryQuota = {
  __typename?: 'ProjectSummaryQuota';
  hard?: Maybe<Scalars['JSON']['output']>;
  used?: Maybe<Scalars['JSON']['output']>;
};

export type ProviderUnderProject = {
  __typename?: 'ProviderUnderProject';
  default?: Maybe<Scalars['Boolean']['output']>;
  enabled?: Maybe<Scalars['Boolean']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  provider?: Maybe<Scalars['String']['output']>;
};

export type Query = {
  __typename?: 'Query';
  /**
   * read the specified ImageList
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  itPolitoCrownlabsV1alpha1ImageList?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * list objects of kind ImageList
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  itPolitoCrownlabsV1alpha1ImageListList?: Maybe<ItPolitoCrownlabsV1alpha1ImageListList>;
  /**
   * read the specified Workspace
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}
   */
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * list objects of kind Workspace
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/workspaces
   */
  itPolitoCrownlabsV1alpha1WorkspaceList?: Maybe<ItPolitoCrownlabsV1alpha1WorkspaceList>;
  /**
   * read the specified Instance
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}
   */
  itPolitoCrownlabsV1alpha2Instance?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * list objects of kind Instance
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/instances
   */
  itPolitoCrownlabsV1alpha2InstanceList?: Maybe<ItPolitoCrownlabsV1alpha2InstanceList>;
  /**
   * read the specified InstanceSnapshot
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}
   */
  itPolitoCrownlabsV1alpha2InstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * list objects of kind InstanceSnapshot
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/instancesnapshots
   */
  itPolitoCrownlabsV1alpha2InstanceSnapshotList?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotList>;
  /**
   * read the specified Template
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}
   */
  itPolitoCrownlabsV1alpha2Template?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * list objects of kind Template
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates
   */
  itPolitoCrownlabsV1alpha2TemplateList?: Maybe<ItPolitoCrownlabsV1alpha2TemplateList>;
  /**
   * read the specified Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/tenants/{name}
   */
  itPolitoCrownlabsV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * list objects of kind Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/tenants
   */
  itPolitoCrownlabsV1alpha2TenantList?: Maybe<ItPolitoCrownlabsV1alpha2TenantList>;
  /**
   * list objects of kind Instance
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances
   */
  listCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<ItPolitoCrownlabsV1alpha2InstanceList>;
  /**
   * list objects of kind InstanceSnapshot
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots
   */
  listCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotList>;
  /**
   * list objects of kind Template
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/templates
   */
  listCrownlabsPolitoItV1alpha2TemplateForAllNamespaces?: Maybe<ItPolitoCrownlabsV1alpha2TemplateList>;
  /**
   * read status of the specified ImageList
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}/status
   */
  readCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
  /**
   * read status of the specified Workspace
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}/status
   */
  readCrownlabsPolitoItV1alpha1WorkspaceStatus?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
  /**
   * read status of the specified InstanceSnapshot
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}/status
   */
  readCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatus?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>;
  /**
   * read status of the specified Instance
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}/status
   */
  readCrownlabsPolitoItV1alpha2NamespacedInstanceStatus?: Maybe<ItPolitoCrownlabsV1alpha2Instance>;
  /**
   * read status of the specified Template
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}/status
   */
  readCrownlabsPolitoItV1alpha2NamespacedTemplateStatus?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
  /**
   * read status of the specified Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha2/tenants/{name}/status
   */
  readCrownlabsPolitoItV1alpha2TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
  /**
   * Get the artifact specified by the reference under the project and repository. The reference can be digest or tag.
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}
   */
  reg_artifact?: Maybe<Artifact>;
  /**
   * This endpoint let user see the recent operation logs of the projects which he is member of
   *
   *
   * Equivalent to GET /audit-logs
   */
  reg_auditLogs?: Maybe<Array<Maybe<AuditLog>>>;
  /**
   * Get the system level allowlist of CVE.  This API can be called by all authenticated users.
   *
   * Equivalent to GET /system/CVEAllowlist
   */
  reg_cVEAllowlist?: Maybe<CveAllowlist>;
  /**
   * This endpoint is for retrieving system configurations that only provides for admin user.
   *
   *
   * Equivalent to GET /configurations
   */
  reg_configurationsResponse?: Maybe<ConfigurationsResponse>;
  /**
   * This endpoint let user get purge job status filtered by specific ID.
   *
   * Equivalent to GET /system/purgeaudit/{purge_id}
   */
  reg_execHistory?: Maybe<ExecHistory>;
  /**
   * Get a execution detail by id
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}
   */
  reg_execution?: Maybe<Execution>;
  /**
   * Download the scan data report. Default format is CSV
   *
   * Equivalent to GET /export/cve/download/{execution_id}
   */
  reg_exportCveDownload?: Maybe<Scalars['String']['output']>;
  /**
   * This endpoint is for get schedule of gc job.
   *
   * Equivalent to GET /system/gc/schedule
   */
  reg_gCHistory?: Maybe<GcHistory>;
  /**
   * This API is for retrieving general system info, this can be called by anonymous request.  Some attributes will be omitted in the response when this API is called by anonymous request.
   *
   *
   * Equivalent to GET /systeminfo
   */
  reg_generalInfo?: Maybe<GeneralInfo>;
  /**
   * This endpoint let user get gc status filtered by specific ID.
   *
   * Equivalent to GET /system/gc/{gc_id}
   */
  reg_getGC?: Maybe<GcHistory>;
  /**
   * Get the metrics of the latest scheduled scan all process
   *
   * Equivalent to GET /scans/schedule/metrics
   */
  reg_getLatestScheduledScanAllMetrics?: Maybe<Stats>;
  /**
   * This endpoint is for get schedule of purge job.
   *
   * Equivalent to GET /system/purgeaudit/schedule
   */
  reg_getPurgeSchedule?: Maybe<ExecHistory>;
  /**
   * This endpoint returns specific robot account information by robot ID.
   *
   * Equivalent to GET /robots/{robot_id}
   */
  reg_getRobotByID?: Maybe<Robot>;
  /**
   * Retruns the details of the specified scanner registration.
   *
   *
   * Equivalent to GET /scanners/{registration_id}
   */
  reg_getScanner?: Maybe<ScannerRegistration>;
  /**
   * Get a user's profile.
   *
   * Equivalent to GET /users/{user_id}
   */
  reg_getUser?: Maybe<UserResp>;
  /**
   * Get the artifact icon with the specified digest. As the original icon image is resized and encoded before returning, the parameter "digest" in the path doesn't match the hash of the returned content
   *
   * Equivalent to GET /icons/{digest}
   */
  reg_icon?: Maybe<Icon>;
  /**
   * Get a P2P provider instance
   *
   * Equivalent to GET /p2p/preheat/instances/{preheat_instance_name}
   */
  reg_instance?: Maybe<Instance>;
  /**
   * This endpoint is for retrieving system configurations that only provides for internal api call.
   *
   *
   * Equivalent to GET /internalconfig
   */
  reg_internalConfigurationsResponse?: Maybe<Scalars['JSON']['output']>;
  /**
   * Get workers in current pool
   *
   * Equivalent to GET /jobservice/pools/{pool_id}/workers
   */
  reg_jobservicePoolWorkers?: Maybe<Array<Maybe<Worker>>>;
  /**
   * Get worker pools
   *
   * Equivalent to GET /jobservice/pools
   */
  reg_jobservicePools?: Maybe<Array<Maybe<WorkerPool>>>;
  /**
   * list job queue
   *
   * Equivalent to GET /jobservice/queues
   */
  reg_jobserviceQueues?: Maybe<Array<Maybe<JobQueue>>>;
  /**
   * This endpoint let user get the label by specific ID.
   *
   *
   * Equivalent to GET /labels/{label_id}
   */
  reg_label?: Maybe<Label>;
  /**
   * This endpoint let user list labels by name, scope and project_id
   *
   *
   * Equivalent to GET /labels
   */
  reg_labels?: Maybe<Array<Maybe<Label>>>;
  /**
   * This endpoint searches the available ldap groups based on related configuration parameters. support to search by groupname or groupdn.
   *
   *
   * Equivalent to GET /ldap/groups/search
   */
  reg_ldapGroupsSearch?: Maybe<Array<Maybe<UserGroup>>>;
  /**
   * This endpoint searches the available ldap users based on related configuration parameters. Support searched by input ladp configuration, load configuration from the system and specific filter.
   *
   *
   * Equivalent to GET /ldap/users/search
   */
  reg_ldapUsersSearch?: Maybe<Array<Maybe<LdapUser>>>;
  /**
   * Check the status of Harbor components
   *
   * Equivalent to GET /health
   */
  reg_overallHealthStatus?: Maybe<OverallHealthStatus>;
  /**
   * List P2P provider instances
   *
   * Equivalent to GET /p2p/preheat/instances
   */
  reg_p2pPreheatInstances?: Maybe<Array<Maybe<Instance>>>;
  /**
   * List P2P providers
   *
   * Equivalent to GET /p2p/preheat/providers
   */
  reg_p2pPreheatProviders?: Maybe<Array<Maybe<Metadata>>>;
  /**
   * This API simply replies a pong to indicate the process to handle API is up, disregarding the health status of dependent components.
   *
   * Equivalent to GET /ping
   */
  reg_ping?: Maybe<Scalars['String']['output']>;
  /**
   * Get a preheat policy
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies/{preheat_policy_name}
   */
  reg_preheatPolicy?: Maybe<PreheatPolicy>;
  /**
   * This endpoint returns specific project information by project ID.
   *
   * Equivalent to GET /projects/{project_name_or_id}
   */
  reg_project2?: Maybe<Project2>;
  /**
   * Get the deletable status of the project
   *
   * Equivalent to GET /projects/{project_name_or_id}/_deletable
   */
  reg_projectDeletable?: Maybe<ProjectDeletable>;
  /**
   * This endpoint returns the immutable tag rules of a project
   *
   *
   * Equivalent to GET /projects/{project_name_or_id}/immutabletagrules
   */
  reg_projectImmutabletagrules?: Maybe<Array<Maybe<ImmutableRule>>>;
  /**
   * Get recent logs of the projects
   *
   * Equivalent to GET /projects/{project_name}/logs
   */
  reg_projectLogs?: Maybe<Array<Maybe<AuditLog>>>;
  /**
   * Get the project member information
   *
   * Equivalent to GET /projects/{project_name_or_id}/members/{mid}
   */
  reg_projectMemberEntity?: Maybe<ProjectMemberEntity>;
  /**
   * Get all project member information
   *
   * Equivalent to GET /projects/{project_name_or_id}/members
   */
  reg_projectMembers?: Maybe<Array<Maybe<ProjectMemberEntity>>>;
  /**
   * Get the specific metadata of the specific project
   *
   * Equivalent to GET /projects/{project_name_or_id}/metadatas/{meta_name}
   */
  reg_projectMetadata2?: Maybe<Scalars['JSON']['output']>;
  /**
   * Get the metadata of the specific project
   *
   * Equivalent to GET /projects/{project_name_or_id}/metadatas/
   */
  reg_projectMetadatas?: Maybe<Scalars['JSON']['output']>;
  /**
   * List preheat policies
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies
   */
  reg_projectPreheatPolicies?: Maybe<Array<Maybe<PreheatPolicy>>>;
  /**
   * Get the log text stream of the specified task for the given execution
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks/{task_id}/logs
   */
  reg_projectPreheatPolicyExecutionTaskLogs?: Maybe<Scalars['String']['output']>;
  /**
   * List all the related tasks for the given execution
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions/{execution_id}/tasks
   */
  reg_projectPreheatPolicyExecutionTasks?: Maybe<Array<Maybe<Task>>>;
  /**
   * List executions for the given policy
   *
   * Equivalent to GET /projects/{project_name}/preheat/policies/{preheat_policy_name}/executions
   */
  reg_projectPreheatPolicyExecutions?: Maybe<Array<Maybe<Execution>>>;
  /**
   * Get all providers at project level
   *
   * Equivalent to GET /projects/{project_name}/preheat/providers
   */
  reg_projectPreheatProviders?: Maybe<Array<Maybe<ProviderUnderProject>>>;
  /**
   * List repositories of the specified project
   *
   * Equivalent to GET /projects/{project_name}/repositories
   */
  reg_projectRepositories?: Maybe<Array<Maybe<ProjectRepository>>>;
  /**
   * Get the repository specified by name
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}
   */
  reg_projectRepository?: Maybe<ProjectRepository>;
  /**
   * List artifacts under the specific project and repository. Except the basic properties, the other supported queries in "q" includes "tags=*" to list only tagged artifacts, "tags=nil" to list only untagged artifacts, "tags=~v" to list artifacts whose tag fuzzy matches "v", "tags=v" to list artifact whose tag exactly matches "v", "labels=(id1, id2)" to list artifacts that both labels with id1 and id2 are added to
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts
   */
  reg_projectRepositoryArtifacts?: Maybe<Array<Maybe<Artifact>>>;
  /**
   * List accessories of the specific artifact
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/accessories
   */
  reg_projectRepositoryArtifactsAccessories?: Maybe<Array<Maybe<Accessory>>>;
  /**
   * Get the addition of the artifact specified by the reference under the project and repository.
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/additions/{addition}
   */
  reg_projectRepositoryArtifactsAddition?: Maybe<Scalars['String']['output']>;
  /**
   * Get the vulnerabilities addition of the artifact specified by the reference under the project and repository.
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/additions/vulnerabilities
   */
  reg_projectRepositoryArtifactsAdditionsVulnerabilities?: Maybe<Scalars['String']['output']>;
  /**
   * Get the log of the scan report
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/scan/{report_id}/log
   */
  reg_projectRepositoryArtifactsScanLog?: Maybe<Scalars['String']['output']>;
  /**
   * List tags of the specific artifact
   *
   * Equivalent to GET /projects/{project_name}/repositories/{repository_name}/artifacts/{reference}/tags
   */
  reg_projectRepositoryArtifactsTags?: Maybe<Array<Maybe<Tag>>>;
  /**
   * Get all robot accounts of specified project
   *
   * Equivalent to GET /projects/{project_name_or_id}/robots
   */
  reg_projectRobots?: Maybe<Array<Maybe<Robot>>>;
  /**
   * Retrieve the system configured scanner registrations as candidates of setting project level scanner.
   *
   * Equivalent to GET /projects/{project_name_or_id}/scanner/candidates
   */
  reg_projectScannerCandidates?: Maybe<Array<Maybe<ScannerRegistration>>>;
  /**
   * Get summary of the project.
   *
   * Equivalent to GET /projects/{project_name_or_id}/summary
   */
  reg_projectSummary?: Maybe<ProjectSummary>;
  /**
   * This endpoint returns webhook jobs of a project.
   *
   *
   * Equivalent to GET /projects/{project_name_or_id}/webhook/jobs
   */
  reg_projectWebhookJobs?: Maybe<Array<Maybe<WebhookJob>>>;
  /**
   * This endpoint returns last trigger information of project webhook policy.
   *
   *
   * Equivalent to GET /projects/{project_name_or_id}/webhook/lasttrigger
   */
  reg_projectWebhookLasttrigger?: Maybe<Array<Maybe<WebhookLastTrigger>>>;
  /**
   * This endpoint returns webhook policies of a project.
   *
   *
   * Equivalent to GET /projects/{project_name_or_id}/webhook/policies
   */
  reg_projectWebhookPolicies?: Maybe<Array<Maybe<WebhookPolicy>>>;
  /**
   * This endpoint returns projects created by Harbor.
   *
   * Equivalent to GET /projects
   */
  reg_projects?: Maybe<Array<Maybe<Project2>>>;
  /**
   * Get the specified quota
   *
   * Equivalent to GET /quotas/{id}
   */
  reg_quota?: Maybe<Quota>;
  /**
   * List quotas
   *
   * Equivalent to GET /quotas
   */
  reg_quotas?: Maybe<Array<Maybe<Quota>>>;
  /**
   * List the registries
   *
   * Equivalent to GET /registries
   */
  reg_registries?: Maybe<Array<Maybe<Registry>>>;
  /**
   * Get the specific registry
   *
   * Equivalent to GET /registries/{id}
   */
  reg_registry?: Maybe<Registry>;
  /**
   * Get the registry info
   *
   * Equivalent to GET /registries/{id}/info
   */
  reg_registryInfo?: Maybe<RegistryInfo>;
  /**
   * List all registered registry provider information
   *
   * Equivalent to GET /replication/adapterinfos
   */
  reg_replicationAdapterinfos?: Maybe<Scalars['JSON']['output']>;
  /**
   * List registry adapters
   *
   * Equivalent to GET /replication/adapters
   */
  reg_replicationAdapters?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /**
   * Get the replication execution specified by ID
   *
   * Equivalent to GET /replication/executions/{id}
   */
  reg_replicationExecution?: Maybe<ReplicationExecution>;
  /**
   * Get the log of the specific replication task
   *
   * Equivalent to GET /replication/executions/{id}/tasks/{task_id}/log
   */
  reg_replicationExecutionTaskLog?: Maybe<Scalars['String']['output']>;
  /**
   * List replication tasks for a specific execution
   *
   * Equivalent to GET /replication/executions/{id}/tasks
   */
  reg_replicationExecutionTasks?: Maybe<Array<Maybe<ReplicationTask>>>;
  /**
   * List replication executions
   *
   * Equivalent to GET /replication/executions
   */
  reg_replicationExecutions?: Maybe<Array<Maybe<ReplicationExecution>>>;
  /**
   * List replication policies
   *
   * Equivalent to GET /replication/policies
   */
  reg_replicationPolicies?: Maybe<Array<Maybe<ReplicationPolicy>>>;
  /**
   * Get the specific replication policy
   *
   * Equivalent to GET /replication/policies/{id}
   */
  reg_replicationPolicy?: Maybe<ReplicationPolicy>;
  /**
   * List all authorized repositories
   *
   * Equivalent to GET /repositories
   */
  reg_repositories?: Maybe<Array<Maybe<ProjectRepository>>>;
  /**
   * Get Retention job task log, tags ratain or deletion detail will be shown in a table.
   *
   * Equivalent to GET /retentions/{id}/executions/{eid}/tasks/{tid}
   */
  reg_retentionExecutionTask2?: Maybe<Scalars['String']['output']>;
  /**
   * Get Retention tasks, each repository as a task.
   *
   * Equivalent to GET /retentions/{id}/executions/{eid}/tasks
   */
  reg_retentionExecutionTasks?: Maybe<Array<Maybe<RetentionExecutionTask>>>;
  /**
   * Get Retention executions, execution status may be delayed before job service schedule it up.
   *
   * Equivalent to GET /retentions/{id}/executions
   */
  reg_retentionExecutions?: Maybe<Array<Maybe<RetentionExecution>>>;
  /**
   * Get Retention Metadatas.
   *
   * Equivalent to GET /retentions/metadatas
   */
  reg_retentionMetadata?: Maybe<RetentionMetadata>;
  /**
   * Get Retention Policy.
   *
   * Equivalent to GET /retentions/{id}
   */
  reg_retentionPolicy?: Maybe<RetentionPolicy>;
  /**
   * This endpoint returns specific robot account information by robot ID.
   *
   * Equivalent to GET /projects/{project_name_or_id}/robots/{robot_id}
   */
  reg_robot?: Maybe<Robot>;
  /**
   * List the robot accounts with the specified level and project.
   *
   * Equivalent to GET /robots
   */
  reg_robots?: Maybe<Array<Maybe<Robot>>>;
  /**
   * Get the scan data export execution specified by ID
   *
   * Equivalent to GET /export/cve/execution/{execution_id}
   */
  reg_scanDataExportExecution?: Maybe<ScanDataExportExecution>;
  /**
   * Get a list of specific scan data export execution jobs for a specified user
   *
   * Equivalent to GET /export/cve/executions
   */
  reg_scanDataExportExecutionList?: Maybe<ScanDataExportExecutionList>;
  /**
   * Get the metadata of the specified scanner registration, including the capabilities and customized properties.
   *
   *
   * Equivalent to GET /scanners/{registration_id}/metadata
   */
  reg_scannerAdapterMetadata?: Maybe<ScannerAdapterMetadata>;
  /**
   * Get the scanner registration of the specified project. If no scanner registration is configured for the specified project, the system default scanner registration will be returned.
   *
   * Equivalent to GET /projects/{project_name_or_id}/scanner
   */
  reg_scannerRegistration?: Maybe<ScannerRegistration>;
  /**
   * Returns a list of currently configured scanner registrations.
   *
   *
   * Equivalent to GET /scanners
   */
  reg_scanners?: Maybe<Array<Maybe<ScannerRegistration>>>;
  /**
   * This endpoint is for getting a schedule for the scan all job, which scans all of images in Harbor.
   *
   * Equivalent to GET /system/scanAll/schedule
   */
  reg_schedule?: Maybe<Schedule>;
  /**
   * Get scheduler paused status
   *
   * Equivalent to GET /schedules/{job_type}/paused
   */
  reg_schedulerStatus?: Maybe<SchedulerStatus>;
  /**
   * List schedules
   *
   * Equivalent to GET /schedules
   */
  reg_schedules?: Maybe<Array<Maybe<ScheduleTask>>>;
  /**
   * The Search endpoint returns information about the projects, repositories and helm charts offered at public status or related to the current logged in user. The response includes the project, repository list and charts in a proper display order.
   *
   * Equivalent to GET /search
   */
  reg_search?: Maybe<Search>;
  /**
   * Get the statistic information about the projects and repositories
   *
   * Equivalent to GET /statistics
   */
  reg_statistic?: Maybe<Statistic>;
  /**
   * Get the metrics of the latest scan all process
   *
   * Equivalent to GET /scans/all/metrics
   */
  reg_stats?: Maybe<Stats>;
  /**
   * Get supportted event types and notify types.
   *
   * Equivalent to GET /projects/{project_name_or_id}/webhook/events
   */
  reg_supportedWebhookEventTypes?: Maybe<SupportedWebhookEventTypes>;
  /**
   * This endpoint let user get gc execution history.
   *
   * Equivalent to GET /system/gc
   */
  reg_systemGc?: Maybe<Array<Maybe<GcHistory>>>;
  /**
   * This endpoint let user get gc job logs filtered by specific ID.
   *
   * Equivalent to GET /system/gc/{gc_id}/log
   */
  reg_systemGcLog?: Maybe<Scalars['String']['output']>;
  /**
   * This endpoint is for retrieving system volume info that only provides for admin user.  Note that the response only reflects the storage status of local disk.
   *
   *
   * Equivalent to GET /systeminfo/volumes
   */
  reg_systemInfo?: Maybe<SystemInfo>;
  /**
   * get purge job execution history.
   *
   * Equivalent to GET /system/purgeaudit
   */
  reg_systemPurgeaudit?: Maybe<Array<Maybe<ExecHistory>>>;
  /**
   * This endpoint let user get purge job logs filtered by specific ID.
   *
   * Equivalent to GET /system/purgeaudit/{purge_id}/log
   */
  reg_systemPurgeauditLog?: Maybe<Scalars['String']['output']>;
  /**
   * This endpoint is for downloading a default root certificate.
   *
   *
   * Equivalent to GET /systeminfo/getcert
   */
  reg_systeminfoGetcert?: Maybe<Scalars['String']['output']>;
  /**
   * Get user group information
   *
   * Equivalent to GET /usergroups/{group_id}
   */
  reg_userGroup?: Maybe<UserGroup>;
  /**
   * Get current user info.
   *
   * Equivalent to GET /users/current
   */
  reg_userResp?: Maybe<UserResp>;
  /**
   * Get all user groups information, it is open for system admin
   *
   * Equivalent to GET /usergroups
   */
  reg_usergroups?: Maybe<Array<Maybe<UserGroup>>>;
  /**
   * This endpoint is to search groups by group name.  It's open for all authenticated requests.
   *
   *
   * Equivalent to GET /usergroups/search
   */
  reg_usergroupsSearch?: Maybe<Array<Maybe<UserGroupSearchItem>>>;
  /**
   * List users
   *
   * Equivalent to GET /users
   */
  reg_users?: Maybe<Array<Maybe<UserResp>>>;
  /**
   * Get current user permissions.
   *
   * Equivalent to GET /users/current/permissions
   */
  reg_usersCurrentPermissions?: Maybe<Array<Maybe<Permission>>>;
  /**
   * This endpoint is to search the users by username.  It's open for all authenticated requests.
   *
   *
   * Equivalent to GET /users/search
   */
  reg_usersSearch?: Maybe<Array<Maybe<UserSearchRespItem>>>;
  /**
   * This endpoint returns specified webhook policy of a project.
   *
   *
   * Equivalent to GET /projects/{project_name_or_id}/webhook/policies/{webhook_policy_id}
   */
  reg_webhookPolicy?: Maybe<WebhookPolicy>;
};


export type QueryItPolitoCrownlabsV1alpha1ImageListArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha1ImageListListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha1WorkspaceArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha1WorkspaceListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2InstanceArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2InstanceListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2InstanceSnapshotArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2InstanceSnapshotListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2TemplateArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2TemplateListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2TenantArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryItPolitoCrownlabsV1alpha2TenantListArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryListCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryListCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryListCrownlabsPolitoItV1alpha2TemplateForAllNamespacesArgs = {
  allowWatchBookmarks?: InputMaybe<Scalars['Boolean']['input']>;
  continue?: InputMaybe<Scalars['String']['input']>;
  fieldSelector?: InputMaybe<Scalars['String']['input']>;
  labelSelector?: InputMaybe<Scalars['String']['input']>;
  limit?: InputMaybe<Scalars['Int']['input']>;
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
  resourceVersionMatch?: InputMaybe<Scalars['String']['input']>;
  sendInitialEvents?: InputMaybe<Scalars['Boolean']['input']>;
  timeoutSeconds?: InputMaybe<Scalars['Int']['input']>;
  watch?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String']['input'];
  namespace: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReadCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  name: Scalars['String']['input'];
  pretty?: InputMaybe<Scalars['String']['input']>;
  resourceVersion?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ArtifactArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  reference: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
  withAccessory?: InputMaybe<Scalars['Boolean']['input']>;
  withImmutableStatus?: InputMaybe<Scalars['Boolean']['input']>;
  withLabel?: InputMaybe<Scalars['Boolean']['input']>;
  withScanOverview?: InputMaybe<Scalars['Boolean']['input']>;
  withSignature?: InputMaybe<Scalars['Boolean']['input']>;
  withTag?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryReg_AuditLogsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ExecHistoryArgs = {
  purgeId: Scalars['BigInt']['input'];
};


export type QueryReg_ExecutionArgs = {
  executionId: Scalars['Int']['input'];
  preheatPolicyName: Scalars['String']['input'];
  projectName: Scalars['String']['input'];
};


export type QueryReg_ExportCveDownloadArgs = {
  executionId: Scalars['Int']['input'];
  format?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_GetGcArgs = {
  gcId: Scalars['BigInt']['input'];
};


export type QueryReg_GetRobotByIdArgs = {
  robotId: Scalars['Int']['input'];
};


export type QueryReg_GetScannerArgs = {
  registrationId: Scalars['String']['input'];
};


export type QueryReg_GetUserArgs = {
  userId: Scalars['Int']['input'];
};


export type QueryReg_IconArgs = {
  digest: Scalars['String']['input'];
};


export type QueryReg_InstanceArgs = {
  preheatInstanceName: Scalars['String']['input'];
};


export type QueryReg_JobservicePoolWorkersArgs = {
  poolId: Scalars['String']['input'];
};


export type QueryReg_LabelArgs = {
  labelId: Scalars['BigInt']['input'];
};


export type QueryReg_LabelsArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectId?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  scope?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_LdapGroupsSearchArgs = {
  groupdn?: InputMaybe<Scalars['String']['input']>;
  groupname?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_LdapUsersSearchArgs = {
  username?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_P2pPreheatInstancesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_PreheatPolicyArgs = {
  preheatPolicyName: Scalars['String']['input'];
  projectName: Scalars['String']['input'];
};


export type QueryReg_Project2Args = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectDeletableArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectImmutabletagrulesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectNameOrId: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectLogsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectMemberEntityArgs = {
  mid: Scalars['BigInt']['input'];
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectMembersArgs = {
  entityname?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectMetadata2Args = {
  metaName: Scalars['String']['input'];
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectMetadatasArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectPreheatPoliciesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectPreheatPolicyExecutionTaskLogsArgs = {
  executionId: Scalars['Int']['input'];
  preheatPolicyName: Scalars['String']['input'];
  projectName: Scalars['String']['input'];
  taskId: Scalars['Int']['input'];
};


export type QueryReg_ProjectPreheatPolicyExecutionTasksArgs = {
  executionId: Scalars['Int']['input'];
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  preheatPolicyName: Scalars['String']['input'];
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectPreheatPolicyExecutionsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  preheatPolicyName: Scalars['String']['input'];
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectPreheatProvidersArgs = {
  projectName: Scalars['String']['input'];
};


export type QueryReg_ProjectRepositoriesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectRepositoryArgs = {
  projectName: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
};


export type QueryReg_ProjectRepositoryArtifactsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  repositoryName: Scalars['String']['input'];
  sort?: InputMaybe<Scalars['String']['input']>;
  withAccessory?: InputMaybe<Scalars['Boolean']['input']>;
  withImmutableStatus?: InputMaybe<Scalars['Boolean']['input']>;
  withLabel?: InputMaybe<Scalars['Boolean']['input']>;
  withScanOverview?: InputMaybe<Scalars['Boolean']['input']>;
  withSignature?: InputMaybe<Scalars['Boolean']['input']>;
  withTag?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryReg_ProjectRepositoryArtifactsAccessoriesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  reference: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectRepositoryArtifactsAdditionArgs = {
  addition: Addition;
  projectName: Scalars['String']['input'];
  reference: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
};


export type QueryReg_ProjectRepositoryArtifactsAdditionsVulnerabilitiesArgs = {
  projectName: Scalars['String']['input'];
  reference: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
};


export type QueryReg_ProjectRepositoryArtifactsScanLogArgs = {
  projectName: Scalars['String']['input'];
  reference: Scalars['String']['input'];
  reportId: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
};


export type QueryReg_ProjectRepositoryArtifactsTagsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectName: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  reference: Scalars['String']['input'];
  repositoryName: Scalars['String']['input'];
  sort?: InputMaybe<Scalars['String']['input']>;
  withImmutableStatus?: InputMaybe<Scalars['Boolean']['input']>;
  withSignature?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryReg_ProjectRobotsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectNameOrId: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectScannerCandidatesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectNameOrId: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectSummaryArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectWebhookJobsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  policyId: Scalars['BigInt']['input'];
  projectNameOrId: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
  status?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
};


export type QueryReg_ProjectWebhookLasttriggerArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ProjectWebhookPoliciesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  projectNameOrId: Scalars['String']['input'];
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ProjectsArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  owner?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  public?: InputMaybe<Scalars['Boolean']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
  withDetail?: InputMaybe<Scalars['Boolean']['input']>;
};


export type QueryReg_QuotaArgs = {
  id: Scalars['Int']['input'];
};


export type QueryReg_QuotasArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  reference?: InputMaybe<Scalars['String']['input']>;
  referenceId?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_RegistriesArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_RegistryArgs = {
  id: Scalars['BigInt']['input'];
};


export type QueryReg_RegistryInfoArgs = {
  id: Scalars['BigInt']['input'];
};


export type QueryReg_ReplicationExecutionArgs = {
  id: Scalars['BigInt']['input'];
};


export type QueryReg_ReplicationExecutionTaskLogArgs = {
  id: Scalars['BigInt']['input'];
  taskId: Scalars['BigInt']['input'];
};


export type QueryReg_ReplicationExecutionTasksArgs = {
  id: Scalars['BigInt']['input'];
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  resourceType?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
  status?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ReplicationExecutionsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  policyId?: InputMaybe<Scalars['Int']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
  status?: InputMaybe<Scalars['String']['input']>;
  trigger?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ReplicationPoliciesArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ReplicationPolicyArgs = {
  id: Scalars['BigInt']['input'];
};


export type QueryReg_RepositoriesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_RetentionExecutionTask2Args = {
  eid: Scalars['BigInt']['input'];
  id: Scalars['BigInt']['input'];
  tid: Scalars['BigInt']['input'];
};


export type QueryReg_RetentionExecutionTasksArgs = {
  eid: Scalars['BigInt']['input'];
  id: Scalars['BigInt']['input'];
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
};


export type QueryReg_RetentionExecutionsArgs = {
  id: Scalars['BigInt']['input'];
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
};


export type QueryReg_RetentionPolicyArgs = {
  id: Scalars['BigInt']['input'];
};


export type QueryReg_RobotArgs = {
  projectNameOrId: Scalars['String']['input'];
  robotId: Scalars['Int']['input'];
};


export type QueryReg_RobotsArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_ScanDataExportExecutionArgs = {
  executionId: Scalars['Int']['input'];
};


export type QueryReg_ScannerAdapterMetadataArgs = {
  registrationId: Scalars['String']['input'];
};


export type QueryReg_ScannerRegistrationArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_ScannersArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_SchedulerStatusArgs = {
  jobType: Scalars['String']['input'];
};


export type QueryReg_SchedulesArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
};


export type QueryReg_SearchArgs = {
  q: Scalars['String']['input'];
};


export type QueryReg_SupportedWebhookEventTypesArgs = {
  projectNameOrId: Scalars['String']['input'];
};


export type QueryReg_SystemGcArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_SystemGcLogArgs = {
  gcId: Scalars['BigInt']['input'];
};


export type QueryReg_SystemPurgeauditArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_SystemPurgeauditLogArgs = {
  purgeId: Scalars['BigInt']['input'];
};


export type QueryReg_UserGroupArgs = {
  groupId: Scalars['BigInt']['input'];
};


export type QueryReg_UsergroupsArgs = {
  groupName?: InputMaybe<Scalars['String']['input']>;
  ldapGroupDn?: InputMaybe<Scalars['String']['input']>;
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
};


export type QueryReg_UsergroupsSearchArgs = {
  groupname: Scalars['String']['input'];
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
};


export type QueryReg_UsersArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  q?: InputMaybe<Scalars['String']['input']>;
  sort?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_UsersCurrentPermissionsArgs = {
  relative?: InputMaybe<Scalars['Boolean']['input']>;
  scope?: InputMaybe<Scalars['String']['input']>;
};


export type QueryReg_UsersSearchArgs = {
  page?: InputMaybe<Scalars['BigInt']['input']>;
  pageSize?: InputMaybe<Scalars['BigInt']['input']>;
  username: Scalars['String']['input'];
};


export type QueryReg_WebhookPolicyArgs = {
  projectNameOrId: Scalars['String']['input'];
  webhookPolicyId: Scalars['BigInt']['input'];
};

/** The quota object */
export type Quota = {
  __typename?: 'Quota';
  /** The maximum amount of CPU required by this Workspace. */
  cpu: Scalars['JSON']['output'];
  /** the creation time of the quota */
  creationTime?: Maybe<Scalars['String']['output']>;
  hard?: Maybe<Scalars['JSON']['output']>;
  /** ID of the quota */
  id?: Maybe<Scalars['Int']['output']>;
  /** The maximum number of concurrent instances required by this Workspace. */
  instances: Scalars['Int']['output'];
  /** The maximum amount of RAM memory required by this Workspace. */
  memory: Scalars['JSON']['output'];
  ref?: Maybe<Scalars['JSON']['output']>;
  /** the update time of the quota */
  updateTime?: Maybe<Scalars['String']['output']>;
  used?: Maybe<Scalars['JSON']['output']>;
};

/** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
export type Quota2 = {
  __typename?: 'Quota2';
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['JSON']['output'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int']['output'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['JSON']['output'];
};

/** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
export type Quota2Input = {
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['JSON']['input'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int']['input'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['JSON']['input'];
};

/** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
export type Quota3 = {
  __typename?: 'Quota3';
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['JSON']['output'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int']['output'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['JSON']['output'];
};

/** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
export type Quota3Input = {
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['JSON']['input'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int']['input'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['JSON']['input'];
};

/** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
export type QuotaInput = {
  /** The maximum amount of CPU required by this Workspace. */
  cpu: Scalars['JSON']['input'];
  /** The maximum number of concurrent instances required by this Workspace. */
  instances: Scalars['Int']['input'];
  /** The maximum amount of RAM memory required by this Workspace. */
  memory: Scalars['JSON']['input'];
};

export type Reference = {
  __typename?: 'Reference';
  annotations?: Maybe<Scalars['JSON']['output']>;
  /** The digest of the child artifact */
  childDigest?: Maybe<Scalars['String']['output']>;
  /** The child ID of the reference */
  childId?: Maybe<Scalars['BigInt']['output']>;
  /** The parent ID of the reference */
  parentId?: Maybe<Scalars['BigInt']['output']>;
  platform?: Maybe<Platform>;
  /** The download URLs */
  urls?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type Registry = {
  __typename?: 'Registry';
  /** The create time of the policy. */
  creationTime?: Maybe<Scalars['String']['output']>;
  credential?: Maybe<RegistryCredential>;
  /** Description of the registry. */
  description?: Maybe<Scalars['String']['output']>;
  /** The registry ID. */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** Whether or not the certificate will be verified when Harbor tries to access the server. */
  insecure?: Maybe<Scalars['Boolean']['output']>;
  /** The registry name. */
  name?: Maybe<Scalars['String']['output']>;
  /** Health status of the registry. */
  status?: Maybe<Scalars['String']['output']>;
  /** Type of the registry, e.g. 'harbor'. */
  type?: Maybe<Scalars['String']['output']>;
  /** The update time of the policy. */
  updateTime?: Maybe<Scalars['String']['output']>;
  /** The registry URL string. */
  url?: Maybe<Scalars['String']['output']>;
};

export type RegistryCredential = {
  __typename?: 'RegistryCredential';
  /** Access key, e.g. user name when credential type is 'basic'. */
  accessKey?: Maybe<Scalars['String']['output']>;
  /** Access secret, e.g. password when credential type is 'basic'. */
  accessSecret?: Maybe<Scalars['String']['output']>;
  /** Credential type, such as 'basic', 'oauth'. */
  type?: Maybe<Scalars['String']['output']>;
};

/** The registry info contains the base info and capability declarations of the registry */
export type RegistryInfo = {
  __typename?: 'RegistryInfo';
  /** The description */
  description?: Maybe<Scalars['String']['output']>;
  /** The registry whether support copy by chunk. */
  supportedCopyByChunk?: Maybe<Scalars['Boolean']['output']>;
  /** The filters that the registry supports */
  supportedResourceFilters?: Maybe<Array<Maybe<FilterStyle>>>;
  /** The triggers that the registry supports */
  supportedTriggers?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** The registry type */
  type?: Maybe<Scalars['String']['output']>;
};

/** The replication execution */
export type ReplicationExecution = {
  __typename?: 'ReplicationExecution';
  /** The end time */
  endTime?: Maybe<Scalars['String']['output']>;
  /** The count of failed executions */
  failed?: Maybe<Scalars['Int']['output']>;
  /** The ID of the execution */
  id?: Maybe<Scalars['Int']['output']>;
  /** The count of in_progress executions */
  inProgress?: Maybe<Scalars['Int']['output']>;
  /** The ID if the policy that the execution belongs to */
  policyId?: Maybe<Scalars['Int']['output']>;
  /** The start time */
  startTime?: Maybe<Scalars['String']['output']>;
  /** The status of the execution */
  status?: Maybe<Scalars['String']['output']>;
  /** The status text */
  statusText?: Maybe<Scalars['String']['output']>;
  /** The count of stopped executions */
  stopped?: Maybe<Scalars['Int']['output']>;
  /** The count of succeed executions */
  succeed?: Maybe<Scalars['Int']['output']>;
  /** The total count of all executions */
  total?: Maybe<Scalars['Int']['output']>;
  /** The trigger mode */
  trigger?: Maybe<Scalars['String']['output']>;
};

export type ReplicationFilter = {
  __typename?: 'ReplicationFilter';
  /** matches or excludes the result */
  decoration?: Maybe<Scalars['String']['output']>;
  /** The replication policy filter type. */
  type?: Maybe<Scalars['String']['output']>;
  /** The value of replication policy filter. */
  value?: Maybe<Scalars['JSON']['output']>;
};

export type ReplicationPolicy = {
  __typename?: 'ReplicationPolicy';
  /** Whether to enable copy by chunk. */
  copyByChunk?: Maybe<Scalars['Boolean']['output']>;
  /** The create time of the policy. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** Deprecated, use "replicate_deletion" instead. Whether to replicate the deletion operation. */
  deletion?: Maybe<Scalars['Boolean']['output']>;
  /** The description of the policy. */
  description?: Maybe<Scalars['String']['output']>;
  /** The destination namespace. */
  destNamespace?: Maybe<Scalars['String']['output']>;
  /**
   * Specify how many path components will be replaced by the provided destination namespace.
   * The default value is -1 in which case the legacy mode will be applied.
   */
  destNamespaceReplaceCount?: Maybe<Scalars['Int']['output']>;
  destRegistry?: Maybe<Registry>;
  /** Whether the policy is enabled or not. */
  enabled?: Maybe<Scalars['Boolean']['output']>;
  /** The replication policy filter array. */
  filters?: Maybe<Array<Maybe<ReplicationFilter>>>;
  /** The policy ID. */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The policy name. */
  name?: Maybe<Scalars['String']['output']>;
  /** Whether to override the resources on the destination registry. */
  override?: Maybe<Scalars['Boolean']['output']>;
  /** Whether to replicate the deletion operation. */
  replicateDeletion?: Maybe<Scalars['Boolean']['output']>;
  /** speed limit for each task */
  speed?: Maybe<Scalars['Int']['output']>;
  srcRegistry?: Maybe<Registry>;
  trigger?: Maybe<ReplicationTrigger>;
  /** The update time of the policy. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The replication task */
export type ReplicationTask = {
  __typename?: 'ReplicationTask';
  /** The destination resource that the task operates */
  dstResource?: Maybe<Scalars['String']['output']>;
  /** The end time of the task */
  endTime?: Maybe<Scalars['String']['output']>;
  /** The ID of the execution that the task belongs to */
  executionId?: Maybe<Scalars['Int']['output']>;
  /** The ID of the task */
  id?: Maybe<Scalars['Int']['output']>;
  /** The ID of the underlying job that the task related to */
  jobId?: Maybe<Scalars['String']['output']>;
  /** The operation of the task */
  operation?: Maybe<Scalars['String']['output']>;
  /** The type of the resource that the task operates */
  resourceType?: Maybe<Scalars['String']['output']>;
  /** The source resource that the task operates */
  srcResource?: Maybe<Scalars['String']['output']>;
  /** The start time of the task */
  startTime?: Maybe<Scalars['String']['output']>;
  /** The status of the task */
  status?: Maybe<Scalars['String']['output']>;
};

export type ReplicationTrigger = {
  __typename?: 'ReplicationTrigger';
  triggerSettings?: Maybe<ReplicationTriggerSettings>;
  /** The replication policy trigger type. The valid values are manual, event_based and scheduled. */
  type?: Maybe<Scalars['String']['output']>;
};

export type ReplicationTriggerSettings = {
  __typename?: 'ReplicationTriggerSettings';
  /** The cron string for scheduled trigger */
  cron?: Maybe<Scalars['String']['output']>;
};

/** The amount of computational resources associated with the environment. */
export type Resources = {
  __typename?: 'Resources';
  /** The maximum number of CPU cores made available to the environment (at least 1 core). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu: Scalars['Int']['output'];
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent or container-based environments, while it is silently ignored in the other cases. In case of containers, when this field is not specified, an emptyDir will be attached to the pod but this could result in data loss whenever the pod dies. */
  disk?: Maybe<Scalars['JSON']['output']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory: Scalars['JSON']['output'];
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage: Scalars['Int']['output'];
};

/** The amount of computational resources associated with the environment. */
export type ResourcesInput = {
  /** The maximum number of CPU cores made available to the environment (at least 1 core). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu: Scalars['Int']['input'];
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent or container-based environments, while it is silently ignored in the other cases. In case of containers, when this field is not specified, an emptyDir will be attached to the pod but this could result in data loss whenever the pod dies. */
  disk?: InputMaybe<Scalars['JSON']['input']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory: Scalars['JSON']['input'];
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage: Scalars['Int']['input'];
};

export type RetentionExecution = {
  __typename?: 'RetentionExecution';
  dryRun?: Maybe<Scalars['Boolean']['output']>;
  endTime?: Maybe<Scalars['String']['output']>;
  id?: Maybe<Scalars['BigInt']['output']>;
  policyId?: Maybe<Scalars['BigInt']['output']>;
  startTime?: Maybe<Scalars['String']['output']>;
  status?: Maybe<Scalars['String']['output']>;
  trigger?: Maybe<Scalars['String']['output']>;
};

export type RetentionExecutionTask = {
  __typename?: 'RetentionExecutionTask';
  endTime?: Maybe<Scalars['String']['output']>;
  executionId?: Maybe<Scalars['BigInt']['output']>;
  id?: Maybe<Scalars['BigInt']['output']>;
  jobId?: Maybe<Scalars['String']['output']>;
  repository?: Maybe<Scalars['String']['output']>;
  retained?: Maybe<Scalars['Int']['output']>;
  startTime?: Maybe<Scalars['String']['output']>;
  status?: Maybe<Scalars['String']['output']>;
  statusCode?: Maybe<Scalars['Int']['output']>;
  statusRevision?: Maybe<Scalars['BigInt']['output']>;
  total?: Maybe<Scalars['Int']['output']>;
};

/** the tag retention metadata */
export type RetentionMetadata = {
  __typename?: 'RetentionMetadata';
  /** supported scope selectors */
  scopeSelectors?: Maybe<Array<Maybe<RetentionSelectorMetadata>>>;
  /** supported tag selectors */
  tagSelectors?: Maybe<Array<Maybe<RetentionSelectorMetadata>>>;
  /** templates */
  templates?: Maybe<Array<Maybe<RetentionRuleMetadata>>>;
};

/** retention policy */
export type RetentionPolicy = {
  __typename?: 'RetentionPolicy';
  algorithm?: Maybe<Scalars['String']['output']>;
  id?: Maybe<Scalars['BigInt']['output']>;
  rules?: Maybe<Array<Maybe<RetentionRule>>>;
  scope?: Maybe<RetentionPolicyScope>;
  trigger?: Maybe<RetentionRuleTrigger>;
};

export type RetentionPolicyScope = {
  __typename?: 'RetentionPolicyScope';
  level?: Maybe<Scalars['String']['output']>;
  ref?: Maybe<Scalars['Int']['output']>;
};

export type RetentionRule = {
  __typename?: 'RetentionRule';
  action?: Maybe<Scalars['String']['output']>;
  disabled?: Maybe<Scalars['Boolean']['output']>;
  id?: Maybe<Scalars['Int']['output']>;
  params?: Maybe<Scalars['JSON']['output']>;
  priority?: Maybe<Scalars['Int']['output']>;
  scopeSelectors?: Maybe<Scalars['JSON']['output']>;
  tagSelectors?: Maybe<Array<Maybe<RetentionSelector>>>;
  template?: Maybe<Scalars['String']['output']>;
};

/** the tag retention rule metadata */
export type RetentionRuleMetadata = {
  __typename?: 'RetentionRuleMetadata';
  /** rule action */
  action?: Maybe<Scalars['String']['output']>;
  /** rule display text */
  displayText?: Maybe<Scalars['String']['output']>;
  /** rule params */
  params?: Maybe<Array<Maybe<RetentionRuleParamMetadata>>>;
  /** rule id */
  ruleTemplate?: Maybe<Scalars['String']['output']>;
};

/** rule param */
export type RetentionRuleParamMetadata = {
  __typename?: 'RetentionRuleParamMetadata';
  required?: Maybe<Scalars['Boolean']['output']>;
  type?: Maybe<Scalars['String']['output']>;
  unit?: Maybe<Scalars['String']['output']>;
};

export type RetentionRuleTrigger = {
  __typename?: 'RetentionRuleTrigger';
  kind?: Maybe<Scalars['String']['output']>;
  references?: Maybe<Scalars['JSON']['output']>;
  settings?: Maybe<Scalars['JSON']['output']>;
};

export type RetentionSelector = {
  __typename?: 'RetentionSelector';
  decoration?: Maybe<Scalars['String']['output']>;
  extras?: Maybe<Scalars['String']['output']>;
  kind?: Maybe<Scalars['String']['output']>;
  pattern?: Maybe<Scalars['String']['output']>;
};

/** retention selector */
export type RetentionSelectorMetadata = {
  __typename?: 'RetentionSelectorMetadata';
  decorations?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  displayText?: Maybe<Scalars['String']['output']>;
  kind?: Maybe<Scalars['String']['output']>;
};

export type Robot = {
  __typename?: 'Robot';
  /** The creation time of the robot. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The description of the robot */
  description?: Maybe<Scalars['String']['output']>;
  /** The disable status of the robot */
  disable?: Maybe<Scalars['Boolean']['output']>;
  /** The duration of the robot in days */
  duration?: Maybe<Scalars['BigInt']['output']>;
  /** The editable status of the robot */
  editable?: Maybe<Scalars['Boolean']['output']>;
  /** The expiration data of the robot */
  expiresAt?: Maybe<Scalars['BigInt']['output']>;
  /** The ID of the robot */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The level of the robot, project or system */
  level?: Maybe<Scalars['String']['output']>;
  /** The name of the tag */
  name?: Maybe<Scalars['String']['output']>;
  permissions?: Maybe<Array<Maybe<RobotPermission>>>;
  /** The secret of the robot */
  secret?: Maybe<Scalars['String']['output']>;
  /** The update time of the robot. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The request for robot account creation. */
export type RobotCreateInput = {
  /** The description of the robot */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The disable status of the robot */
  disable?: InputMaybe<Scalars['Boolean']['input']>;
  /** The duration of the robot in days */
  duration?: InputMaybe<Scalars['BigInt']['input']>;
  /** The level of the robot, project or system */
  level?: InputMaybe<Scalars['String']['input']>;
  /** The name of the tag */
  name?: InputMaybe<Scalars['String']['input']>;
  permissions?: InputMaybe<Array<InputMaybe<RobotPermissionInput>>>;
  /** The secret of the robot */
  secret?: InputMaybe<Scalars['String']['input']>;
};

export type RobotCreateV1Input = {
  /** The permission of robot account */
  access?: InputMaybe<Array<InputMaybe<Access2Input>>>;
  /** The description of robot account */
  description?: InputMaybe<Scalars['String']['input']>;
  /** The expiration time on or after which the JWT MUST NOT be accepted for processing. */
  expiresAt?: InputMaybe<Scalars['Int']['input']>;
  /** The name of robot account */
  name?: InputMaybe<Scalars['String']['input']>;
};

/** The response for robot account creation. */
export type RobotCreated = {
  __typename?: 'RobotCreated';
  /** The creation time of the robot. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The expiration data of the robot */
  expiresAt?: Maybe<Scalars['BigInt']['output']>;
  /** The ID of the robot */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The name of the tag */
  name?: Maybe<Scalars['String']['output']>;
  /** The secret of the robot */
  secret?: Maybe<Scalars['String']['output']>;
};

export type RobotPermission = {
  __typename?: 'RobotPermission';
  access?: Maybe<Array<Maybe<Access2>>>;
  /** The kind of the permission */
  kind?: Maybe<Scalars['String']['output']>;
  /** The namespace of the permission */
  namespace?: Maybe<Scalars['String']['output']>;
};

export type RobotPermissionInput = {
  access?: InputMaybe<Array<InputMaybe<Access2Input>>>;
  /** The kind of the permission */
  kind?: InputMaybe<Scalars['String']['input']>;
  /** The namespace of the permission */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

/** The response for refresh/update robot account secret. */
export type RobotSec = {
  __typename?: 'RobotSec';
  /** The secret of the robot */
  secret?: Maybe<Scalars['String']['output']>;
};

/** The response for refresh/update robot account secret. */
export type RobotSecInput = {
  /** The secret of the robot */
  secret?: InputMaybe<Scalars['String']['input']>;
};

export enum Role {
  Candidate = 'CANDIDATE',
  Manager = 'MANAGER',
  User = 'USER'
}

/** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
export type SandboxNamespace = {
  __typename?: 'SandboxNamespace';
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['output'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']['output']>;
};

/** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
export type SandboxNamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean']['input'];
  /** The name of the considered resource. */
  name?: InputMaybe<Scalars['String']['input']>;
};

export type ScanAllPolicy = {
  __typename?: 'ScanAllPolicy';
  /** The parameters of the policy, the values are dependent on the type of the policy. */
  parameter?: Maybe<Parameter>;
  /** The type of scan all policy, currently the valid values are "none" and "daily" */
  type?: Maybe<Scalars['String']['output']>;
};

/** The replication execution */
export type ScanDataExportExecution = {
  __typename?: 'ScanDataExportExecution';
  /** The end time */
  endTime?: Maybe<Scalars['String']['output']>;
  /** Indicates whether the export artifact is present in registry */
  filePresent?: Maybe<Scalars['Boolean']['output']>;
  /** The ID of the execution */
  id?: Maybe<Scalars['Int']['output']>;
  /** The start time */
  startTime?: Maybe<Scalars['String']['output']>;
  /** The status of the execution */
  status?: Maybe<Scalars['String']['output']>;
  /** The status text */
  statusText?: Maybe<Scalars['String']['output']>;
  /** The trigger mode */
  trigger?: Maybe<Scalars['String']['output']>;
  /** The ID if the user triggering the export job */
  userId?: Maybe<Scalars['Int']['output']>;
  /** The name of the user triggering the job */
  userName?: Maybe<Scalars['String']['output']>;
};

/** The list of scan data export executions */
export type ScanDataExportExecutionList = {
  __typename?: 'ScanDataExportExecutionList';
  /** The list of scan data export executions */
  items?: Maybe<Array<Maybe<ScanDataExportExecution>>>;
};

/** The metadata associated with the scan data export job */
export type ScanDataExportJob = {
  __typename?: 'ScanDataExportJob';
  /** The id of the scan data export job */
  id?: Maybe<Scalars['BigInt']['output']>;
};

/** The criteria to select the scan data to export. */
export type ScanDataExportRequestInput = {
  /** CVE-IDs for which to export data. Multiple CVE-IDs can be specified by separating using ',' and enclosed between '{}'. Defaults to all if empty */
  cveIds?: InputMaybe<Scalars['String']['input']>;
  /** Name of the scan data export job */
  jobName?: InputMaybe<Scalars['String']['input']>;
  /** A list of one or more labels for which to export the scan data, defaults to all if empty */
  labels?: InputMaybe<Array<InputMaybe<Scalars['BigInt']['input']>>>;
  /** A list of one or more projects for which to export the scan data, currently only one project is supported due to performance concerns, but define as array for extension in the future. */
  projects?: InputMaybe<Array<InputMaybe<Scalars['BigInt']['input']>>>;
  /** A list of repositories for which to export the scan data, defaults to all if empty */
  repositories?: InputMaybe<Scalars['String']['input']>;
  /** A list of tags enclosed within '{}'. Defaults to all if empty */
  tags?: InputMaybe<Scalars['String']['input']>;
};

export type Scanner = {
  __typename?: 'Scanner';
  /** Name of the scanner */
  name?: Maybe<Scalars['String']['output']>;
  /** Name of the scanner provider */
  vendor?: Maybe<Scalars['String']['output']>;
  /** Version of the scanner adapter */
  version?: Maybe<Scalars['String']['output']>;
};

/** The metadata info of the scanner adapter */
export type ScannerAdapterMetadata = {
  __typename?: 'ScannerAdapterMetadata';
  capabilities?: Maybe<Array<Maybe<ScannerCapability>>>;
  properties?: Maybe<Scalars['JSON']['output']>;
  scanner?: Maybe<Scanner>;
};

export type ScannerCapability = {
  __typename?: 'ScannerCapability';
  consumesMimeTypes?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  producesMimeTypes?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

/**
 * Registration represents a named configuration for invoking a scanner via its adapter.
 *
 */
export type ScannerRegistration = {
  __typename?: 'ScannerRegistration';
  /**
   * An optional value of the HTTP Authorization header sent with each request to the Scanner Adapter API.
   *
   */
  accessCredential?: Maybe<Scalars['String']['output']>;
  /** Optional property to describe the name of the scanner registration */
  adapter?: Maybe<Scalars['String']['output']>;
  /**
   * Specify what authentication approach is adopted for the HTTP communications.
   * Supported types Basic", "Bearer" and api key header "X-ScannerAdapter-API-Key"
   *
   */
  auth?: Maybe<Scalars['String']['output']>;
  /** The creation time of this registration */
  createTime?: Maybe<Scalars['String']['output']>;
  /** An optional description of this registration. */
  description?: Maybe<Scalars['String']['output']>;
  /** Indicate whether the registration is enabled or not */
  disabled?: Maybe<Scalars['Boolean']['output']>;
  /** Indicate the healthy of the registration */
  health?: Maybe<Scalars['String']['output']>;
  /** Indicate if the registration is set as the system default one */
  isDefault?: Maybe<Scalars['Boolean']['output']>;
  /** The name of this registration. */
  name?: Maybe<Scalars['String']['output']>;
  /** Indicate if skip the certificate verification when sending HTTP requests */
  skipCertVerify?: Maybe<Scalars['Boolean']['output']>;
  /** The update time of this registration */
  updateTime?: Maybe<Scalars['String']['output']>;
  /** A base URL of the scanner adapter */
  url?: Maybe<Scalars['String']['output']>;
  /** Indicate whether use internal registry addr for the scanner to pull content or not */
  useInternalAddr?: Maybe<Scalars['Boolean']['output']>;
  /** The unique identifier of this registration. */
  uuid?: Maybe<Scalars['String']['output']>;
  /** Optional property to describe the vendor of the scanner registration */
  vendor?: Maybe<Scalars['String']['output']>;
  /** Optional property to describe the version of the scanner registration */
  version?: Maybe<Scalars['String']['output']>;
};

export type Schedule = {
  __typename?: 'Schedule';
  /** the creation time of the schedule. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The id of the schedule. */
  id?: Maybe<Scalars['Int']['output']>;
  /** The parameters of schedule job */
  parameters?: Maybe<Scalars['JSON']['output']>;
  schedule?: Maybe<ScheduleObj>;
  /** The status of the schedule. */
  status?: Maybe<Scalars['String']['output']>;
  /** the update time of the schedule. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

export type ScheduleObj = {
  __typename?: 'ScheduleObj';
  /** A cron expression, a time-based job scheduler. */
  cron?: Maybe<Scalars['String']['output']>;
  /** The next time to schedule to run the job. */
  nextScheduledTime?: Maybe<Scalars['String']['output']>;
  /**
   * The schedule type. The valid values are 'Hourly', 'Daily', 'Weekly', 'Custom', 'Manual' and 'None'.
   * 'Manual' means to trigger it right away and 'None' means to cancel the schedule.
   *
   */
  type?: Maybe<Type>;
};

/** the schedule task info */
export type ScheduleTask = {
  __typename?: 'ScheduleTask';
  /** the cron of the current schedule task */
  cron?: Maybe<Scalars['String']['output']>;
  /** the id of the Schedule task */
  id?: Maybe<Scalars['Int']['output']>;
  /** the update time of the schedule task */
  updateTime?: Maybe<Scalars['String']['output']>;
  /** the vendor id of the current task */
  vendorId?: Maybe<Scalars['Int']['output']>;
  /** the vendor type of the current schedule task */
  vendorType?: Maybe<Scalars['String']['output']>;
};

/** the scheduler status */
export type SchedulerStatus = {
  __typename?: 'SchedulerStatus';
  /** if the scheduler is paused */
  paused?: Maybe<Scalars['Boolean']['output']>;
};

export type Search = {
  __typename?: 'Search';
  /** Search results of the charts that macthed the filter keywords. */
  chart?: Maybe<Array<Maybe<SearchResult>>>;
  /** Search results of the projects that matched the filter keywords. */
  project?: Maybe<Array<Maybe<Project2>>>;
  /** Search results of the repositories that matched the filter keywords. */
  repository?: Maybe<Array<Maybe<SearchRepository>>>;
};

export type SearchRepository = {
  __typename?: 'SearchRepository';
  /** The count of artifacts in the repository */
  artifactCount?: Maybe<Scalars['Int']['output']>;
  /** The ID of the project that the repository belongs to */
  projectId?: Maybe<Scalars['Int']['output']>;
  /** The name of the project that the repository belongs to */
  projectName?: Maybe<Scalars['String']['output']>;
  /** The flag to indicate the publicity of the project that the repository belongs to (1 is public, 0 is not) */
  projectPublic?: Maybe<Scalars['Boolean']['output']>;
  /** The count how many times the repository is pulled */
  pullCount?: Maybe<Scalars['Int']['output']>;
  /** The name of the repository */
  repositoryName?: Maybe<Scalars['String']['output']>;
};

/** The chart search result item */
export type SearchResult = {
  __typename?: 'SearchResult';
  /** A specified chart entry */
  chart?: Maybe<ChartVersion>;
  /** The chart name with repo name */
  name?: Maybe<Scalars['String']['output']>;
  /** The matched level */
  score?: Maybe<Scalars['Int']['output']>;
};

/** ImageListSpec is the specification of the desired state of the ImageList. */
export type Spec = {
  __typename?: 'Spec';
  /** The list of VM images currently available in CrownLabs. */
  images: Array<Maybe<ImagesListItem>>;
  /** The host name that can be used to access the registry. */
  registryName: Scalars['String']['output'];
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2 = {
  __typename?: 'Spec2';
  /** AutoEnroll capability definition. If omitted, no autoenroll features will be added. */
  autoEnroll?: Maybe<AutoEnroll>;
  /** The human-readable name of the Workspace. */
  prettyName: Scalars['String']['output'];
  /** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
  quota: Quota;
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2Input = {
  /** AutoEnroll capability definition. If omitted, no autoenroll features will be added. */
  autoEnroll?: InputMaybe<AutoEnroll>;
  /** The human-readable name of the Workspace. */
  prettyName: Scalars['String']['input'];
  /** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
  quota: QuotaInput;
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3 = {
  __typename?: 'Spec3';
  /** Optional urls for advanced integration features. */
  customizationUrls?: Maybe<CustomizationUrls>;
  /** Custom name the user can assign and change at any time in order to more easily identify the instance. */
  prettyName?: Maybe<Scalars['String']['output']>;
  /** Whether the current instance is running or not. The meaning of this flag is different depending on whether the instance refers to a persistent environment or not. If the first case, it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. Differently, if the environment is not persistent, it only tears down the exposition objects, making the instance effectively unreachable from outside the cluster, but allowing the subsequent recreation without data loss. */
  running?: Maybe<Scalars['Boolean']['output']>;
  /** The reference to the Template to be instantiated. */
  templateCrownlabsPolitoItTemplateRef: TemplateCrownlabsPolitoItTemplateRef;
  /** The reference to the Tenant which owns the Instance object. */
  tenantCrownlabsPolitoItTenantRef: TenantCrownlabsPolitoItTenantRef;
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3Input = {
  /** Optional urls for advanced integration features. */
  customizationUrls?: InputMaybe<CustomizationUrlsInput>;
  /** Custom name the user can assign and change at any time in order to more easily identify the instance. */
  prettyName?: InputMaybe<Scalars['String']['input']>;
  /** Whether the current instance is running or not. The meaning of this flag is different depending on whether the instance refers to a persistent environment or not. If the first case, it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. Differently, if the environment is not persistent, it only tears down the exposition objects, making the instance effectively unreachable from outside the cluster, but allowing the subsequent recreation without data loss. */
  running?: InputMaybe<Scalars['Boolean']['input']>;
  /** The reference to the Template to be instantiated. */
  templateCrownlabsPolitoItTemplateRef: TemplateCrownlabsPolitoItTemplateRefInput;
  /** The reference to the Tenant which owns the Instance object. */
  tenantCrownlabsPolitoItTenantRef: TenantCrownlabsPolitoItTenantRefInput;
};

/** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
export type Spec4 = {
  __typename?: 'Spec4';
  /**
   * Environment represents the reference to the environment to be snapshotted, in case more are
   * associated with the same Instance. If not specified, the first available environment is considered.
   */
  environmentRef?: Maybe<EnvironmentRef>;
  /** ImageName is the name of the image to pushed in the docker registry. */
  imageName: Scalars['String']['output'];
  /**
   * Instance is the reference to the persistent VM instance to be snapshotted.
   * The instance should not be running, otherwise it won't be possible to
   * steal the volume and extract its content.
   */
  instanceRef: InstanceRef;
};

/** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
export type Spec4Input = {
  /**
   * Environment represents the reference to the environment to be snapshotted, in case more are
   * associated with the same Instance. If not specified, the first available environment is considered.
   */
  environmentRef?: InputMaybe<EnvironmentRefInput>;
  /** ImageName is the name of the image to pushed in the docker registry. */
  imageName: Scalars['String']['input'];
  /**
   * Instance is the reference to the persistent VM instance to be snapshotted.
   * The instance should not be running, otherwise it won't be possible to
   * steal the volume and extract its content.
   */
  instanceRef: InstanceRefInput;
};

/** TemplateSpec is the specification of the desired state of the Template. */
export type Spec5 = {
  __typename?: 'Spec5';
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. If set to "never", the instance will not be automatically terminated. */
  deleteAfter?: Maybe<Scalars['String']['output']>;
  /** A textual description of the Template. */
  description: Scalars['String']['output'];
  /** The list of environments (i.e. VMs or containers) that compose the Template. */
  environmentList: Array<Maybe<EnvironmentListListItem>>;
  /** The human-readable name of the Template. */
  prettyName: Scalars['String']['output'];
  /** The reference to the Workspace this Template belongs to. */
  workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<WorkspaceCrownlabsPolitoItWorkspaceRef>;
};

/** TemplateSpec is the specification of the desired state of the Template. */
export type Spec5Input = {
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. If set to "never", the instance will not be automatically terminated. */
  deleteAfter?: InputMaybe<Scalars['String']['input']>;
  /** A textual description of the Template. */
  description: Scalars['String']['input'];
  /** The list of environments (i.e. VMs or containers) that compose the Template. */
  environmentList: Array<InputMaybe<EnvironmentListListItemInput>>;
  /** The human-readable name of the Template. */
  prettyName: Scalars['String']['input'];
  /** The reference to the Workspace this Template belongs to. */
  workspaceCrownlabsPolitoItWorkspaceRef?: InputMaybe<WorkspaceCrownlabsPolitoItWorkspaceRefInput>;
};

/** TenantSpec is the specification of the desired state of the Tenant. */
export type Spec6 = {
  __typename?: 'Spec6';
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: Maybe<Scalars['Boolean']['output']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email: Scalars['String']['output'];
  /** The first name of the Tenant. */
  firstName: Scalars['String']['output'];
  /** The last login timestamp. */
  lastLogin?: Maybe<Scalars['String']['output']>;
  /** The last name of the Tenant. */
  lastName: Scalars['String']['output'];
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
  quota?: Maybe<Quota2>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: Maybe<Array<Maybe<WorkspacesListItem>>>;
};

/** TenantSpec is the specification of the desired state of the Tenant. */
export type Spec6Input = {
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: InputMaybe<Scalars['Boolean']['input']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email: Scalars['String']['input'];
  /** The first name of the Tenant. */
  firstName: Scalars['String']['input'];
  /** The last login timestamp. */
  lastLogin?: InputMaybe<Scalars['String']['input']>;
  /** The last name of the Tenant. */
  lastName: Scalars['String']['input'];
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: InputMaybe<Array<InputMaybe<Scalars['String']['input']>>>;
  /** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
  quota?: InputMaybe<Quota2Input>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: InputMaybe<Array<InputMaybe<WorkspacesListItemInput>>>;
};

/** ImageListSpec is the specification of the desired state of the ImageList. */
export type SpecInput = {
  /** The list of VM images currently available in CrownLabs. */
  images: Array<InputMaybe<ImagesListItemInput>>;
  /** The host name that can be used to access the registry. */
  registryName: Scalars['String']['input'];
};

export type Statistic = {
  __typename?: 'Statistic';
  /** The count of the private projects */
  privateProjectCount?: Maybe<Scalars['BigInt']['output']>;
  /** The count of the private repositories */
  privateRepoCount?: Maybe<Scalars['BigInt']['output']>;
  /** The count of the public projects */
  publicProjectCount?: Maybe<Scalars['BigInt']['output']>;
  /** The count of the public repositories */
  publicRepoCount?: Maybe<Scalars['BigInt']['output']>;
  /** The count of the total projects, only be seen by the system admin */
  totalProjectCount?: Maybe<Scalars['BigInt']['output']>;
  /** The count of the total repositories, only be seen by the system admin */
  totalRepoCount?: Maybe<Scalars['BigInt']['output']>;
  /** The total storage consumption of blobs, only be seen by the system admin */
  totalStorageConsumption?: Maybe<Scalars['BigInt']['output']>;
};

/** Stats provides the overall progress of the scan all process. */
export type Stats = {
  __typename?: 'Stats';
  /** The number of the finished scan processes triggered by the scan all action */
  completed?: Maybe<Scalars['Int']['output']>;
  /** The metrics data for the each status */
  metrics?: Maybe<Scalars['JSON']['output']>;
  /** A flag indicating job status of scan all. */
  ongoing?: Maybe<Scalars['Boolean']['output']>;
  /** The total number of scan processes triggered by the scan all action */
  total?: Maybe<Scalars['Int']['output']>;
  /** The trigger of the scan all job. */
  trigger?: Maybe<Trigger>;
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type Status2 = {
  __typename?: 'Status2';
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: Maybe<Namespace>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']['output']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: Maybe<Scalars['JSON']['output']>;
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type Status2Input = {
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: InputMaybe<NamespaceInput>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: InputMaybe<Scalars['Boolean']['input']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: InputMaybe<Scalars['JSON']['input']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status3 = {
  __typename?: 'Status3';
  /** Timestamps of the Instance automation phases (check, termination and submission). */
  automation?: Maybe<Automation>;
  /** The amount of time the Instance required to become ready for the first time upon creation. */
  initialReadyTime?: Maybe<Scalars['String']['output']>;
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: Maybe<Scalars['String']['output']>;
  /** The URL where it is possible to access the persistent drive associated with the instance (in case of container-based environments) */
  myDriveUrl?: Maybe<Scalars['String']['output']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: Maybe<Phase>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: Maybe<Scalars['String']['output']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status3Input = {
  /** Timestamps of the Instance automation phases (check, termination and submission). */
  automation?: InputMaybe<AutomationInput>;
  /** The amount of time the Instance required to become ready for the first time upon creation. */
  initialReadyTime?: InputMaybe<Scalars['String']['input']>;
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: InputMaybe<Scalars['String']['input']>;
  /** The URL where it is possible to access the persistent drive associated with the instance (in case of container-based environments) */
  myDriveUrl?: InputMaybe<Scalars['String']['input']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: InputMaybe<Phase>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: InputMaybe<Scalars['String']['input']>;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status4 = {
  __typename?: 'Status4';
  /** Phase represents the current state of the Instance Snapshot. */
  phase: Phase2;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status4Input = {
  /** Phase represents the current state of the Instance Snapshot. */
  phase: Phase2;
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type Status6 = {
  __typename?: 'Status6';
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces: Array<Maybe<Scalars['String']['output']>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
  personalNamespace: PersonalNamespace;
  /** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
  quota?: Maybe<Quota3>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. Will be set to true even when personal workspace is intentionally deleted. */
  ready: Scalars['Boolean']['output'];
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace: SandboxNamespace;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions: Scalars['JSON']['output'];
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type Status6Input = {
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces: Array<InputMaybe<Scalars['String']['input']>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
  personalNamespace: PersonalNamespaceInput;
  /** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
  quota?: InputMaybe<Quota3Input>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. Will be set to true even when personal workspace is intentionally deleted. */
  ready: Scalars['Boolean']['input'];
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace: SandboxNamespaceInput;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions: Scalars['JSON']['input'];
};

export type Storage2 = {
  __typename?: 'Storage2';
  /** Free volume size. */
  free?: Maybe<Scalars['Int']['output']>;
  /** Total volume size. */
  total?: Maybe<Scalars['Int']['output']>;
};

export type StringConfigItem = {
  __typename?: 'StringConfigItem';
  /** The configure item can be updated or not */
  editable?: Maybe<Scalars['Boolean']['output']>;
  /** The string value of current config item */
  value?: Maybe<Scalars['String']['output']>;
};

export type Subscription = {
  __typename?: 'Subscription';
  itPolitoCrownlabsV1alpha1ImageListUpdate?: Maybe<ItPolitoCrownlabsV1alpha1ImageListUpdate>;
  itPolitoCrownlabsV1alpha1WorkspaceUpdate?: Maybe<ItPolitoCrownlabsV1alpha1WorkspaceUpdate>;
  itPolitoCrownlabsV1alpha2InstanceLabelsUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate>;
  itPolitoCrownlabsV1alpha2InstanceUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2TemplateUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TemplateUpdate>;
  itPolitoCrownlabsV1alpha2TenantUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TenantUpdate>;
};


export type SubscriptionItPolitoCrownlabsV1alpha1ImageListUpdateArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace: Scalars['String']['input'];
};


export type SubscriptionItPolitoCrownlabsV1alpha1WorkspaceUpdateArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace: Scalars['String']['input'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceLabelsUpdateArgs = {
  labelSelector?: InputMaybe<Scalars['String']['input']>;
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceSnapshotUpdateArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace: Scalars['String']['input'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceUpdateArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace: Scalars['String']['input'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2TemplateUpdateArgs = {
  name?: InputMaybe<Scalars['String']['input']>;
  namespace: Scalars['String']['input'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2TenantUpdateArgs = {
  name: Scalars['String']['input'];
  namespace?: InputMaybe<Scalars['String']['input']>;
};

/** Supportted webhook event types and notify types. */
export type SupportedWebhookEventTypes = {
  __typename?: 'SupportedWebhookEventTypes';
  eventType?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  notifyType?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
};

export type SystemInfo = {
  __typename?: 'SystemInfo';
  /** The storage of system. */
  storage?: Maybe<Array<Maybe<Storage2>>>;
};

export type Tag = {
  __typename?: 'Tag';
  /** The ID of the artifact that the tag attached to */
  artifactId?: Maybe<Scalars['BigInt']['output']>;
  /** The ID of the tag */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The immutable status of the tag */
  immutable?: Maybe<Scalars['Boolean']['output']>;
  /** The name of the tag */
  name?: Maybe<Scalars['String']['output']>;
  /** The latest pull time of the tag */
  pullTime?: Maybe<Scalars['String']['output']>;
  /** The push time of the tag */
  pushTime?: Maybe<Scalars['String']['output']>;
  /** The ID of the repository that the tag belongs to */
  repositoryId?: Maybe<Scalars['BigInt']['output']>;
  /** The attribute indicates whether the tag is signed or not */
  signed?: Maybe<Scalars['Boolean']['output']>;
};

export type Task = {
  __typename?: 'Task';
  /** The creation time of task */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The end time of task */
  endTime?: Maybe<Scalars['String']['output']>;
  /** The ID of task execution */
  executionId?: Maybe<Scalars['Int']['output']>;
  extraAttrs?: Maybe<Scalars['JSON']['output']>;
  /** The ID of task */
  id?: Maybe<Scalars['Int']['output']>;
  /** The count of task run */
  runCount?: Maybe<Scalars['Int']['output']>;
  /** The start time of task */
  startTime?: Maybe<Scalars['String']['output']>;
  /** The status of task */
  status?: Maybe<Scalars['String']['output']>;
  /** The status message of task */
  statusMessage?: Maybe<Scalars['String']['output']>;
  /** The update time of task */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The reference to the Template to be instantiated. */
export type TemplateCrownlabsPolitoItTemplateRef = {
  __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
  /** The name of the resource to be referenced. */
  name: Scalars['String']['output'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']['output']>;
  templateWrapper?: Maybe<TemplateWrapper>;
};

/** The reference to the Template to be instantiated. */
export type TemplateCrownlabsPolitoItTemplateRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String']['input'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type TemplateWrapper = {
  __typename?: 'TemplateWrapper';
  itPolitoCrownlabsV1alpha2Template?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

/** The reference to the Tenant which owns the Instance object. */
export type TenantCrownlabsPolitoItTenantRef = {
  __typename?: 'TenantCrownlabsPolitoItTenantRef';
  /** The name of the resource to be referenced. */
  name: Scalars['String']['output'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']['output']>;
  tenantV1alpha2Wrapper?: Maybe<TenantV1alpha2Wrapper>;
};

/** The reference to the Tenant which owns the Instance object. */
export type TenantCrownlabsPolitoItTenantRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String']['input'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type TenantV1alpha2Wrapper = {
  __typename?: 'TenantV1alpha2Wrapper';
  itPolitoCrownlabsV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
};

export enum Trigger {
  Event = 'EVENT',
  Manual = 'MANUAL',
  Schedule = 'SCHEDULE'
}

export enum Type {
  Custom = 'CUSTOM',
  Daily = 'DAILY',
  Hourly = 'HOURLY',
  Manual = 'MANUAL',
  None = 'NONE',
  Weekly = 'WEEKLY'
}

export enum UpdateType {
  Added = 'ADDED',
  Deleted = 'DELETED',
  Modified = 'MODIFIED'
}

export type UserGroup = {
  __typename?: 'UserGroup';
  /** The name of the user group */
  groupName?: Maybe<Scalars['String']['output']>;
  /** The group type, 1 for LDAP group, 2 for HTTP group, 3 for OIDC group. */
  groupType?: Maybe<Scalars['Int']['output']>;
  /** The ID of the user group */
  id?: Maybe<Scalars['Int']['output']>;
  /** The DN of the LDAP group if group type is 1 (LDAP group). */
  ldapGroupDn?: Maybe<Scalars['String']['output']>;
};

export type UserGroupSearchItem = {
  __typename?: 'UserGroupSearchItem';
  /** The name of the user group */
  groupName?: Maybe<Scalars['String']['output']>;
  /** The group type, 1 for LDAP group, 2 for HTTP group, 3 for OIDC group. */
  groupType?: Maybe<Scalars['Int']['output']>;
  /** The ID of the user group */
  id?: Maybe<Scalars['Int']['output']>;
};

export type UserResp = {
  __typename?: 'UserResp';
  /** indicate the admin privilege is grant by authenticator (LDAP), is always false unless it is the current login user */
  adminRoleInAuth?: Maybe<Scalars['Boolean']['output']>;
  comment?: Maybe<Scalars['String']['output']>;
  /** The creation time of the user. */
  creationTime?: Maybe<Scalars['String']['output']>;
  email?: Maybe<Scalars['String']['output']>;
  oidcUserMeta?: Maybe<OidcUserInfo>;
  realname?: Maybe<Scalars['String']['output']>;
  sysadminFlag?: Maybe<Scalars['Boolean']['output']>;
  /** The update time of the user. */
  updateTime?: Maybe<Scalars['String']['output']>;
  userId?: Maybe<Scalars['Int']['output']>;
  username?: Maybe<Scalars['String']['output']>;
};

export type UserSearchRespItem = {
  __typename?: 'UserSearchRespItem';
  /** The ID of the user. */
  userId?: Maybe<Scalars['Int']['output']>;
  username?: Maybe<Scalars['String']['output']>;
};

/** The webhook job. */
export type WebhookJob = {
  __typename?: 'WebhookJob';
  /** The webhook job creation time. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The webhook job event type. */
  eventType?: Maybe<Scalars['String']['output']>;
  /** The webhook job ID. */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The webhook job notify detailed data. */
  jobDetail?: Maybe<Scalars['String']['output']>;
  /** The webhook job notify type. */
  notifyType?: Maybe<Scalars['String']['output']>;
  /** The webhook policy ID. */
  policyId?: Maybe<Scalars['BigInt']['output']>;
  /** The webhook job status. */
  status?: Maybe<Scalars['String']['output']>;
  /** The webhook job update time. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The webhook policy and last trigger time group by event type. */
export type WebhookLastTrigger = {
  __typename?: 'WebhookLastTrigger';
  /** The creation time of webhook policy. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** Whether or not the webhook policy enabled. */
  enabled?: Maybe<Scalars['Boolean']['output']>;
  /** The webhook event type. */
  eventType?: Maybe<Scalars['String']['output']>;
  /** The last trigger time of webhook policy. */
  lastTriggerTime?: Maybe<Scalars['String']['output']>;
  /** The webhook policy name. */
  policyName?: Maybe<Scalars['String']['output']>;
};

/** The webhook policy object */
export type WebhookPolicy = {
  __typename?: 'WebhookPolicy';
  /** The create time of the webhook policy. */
  creationTime?: Maybe<Scalars['String']['output']>;
  /** The creator of the webhook policy. */
  creator?: Maybe<Scalars['String']['output']>;
  /** The description of webhook policy. */
  description?: Maybe<Scalars['String']['output']>;
  /** Whether the webhook policy is enabled or not. */
  enabled?: Maybe<Scalars['Boolean']['output']>;
  eventTypes?: Maybe<Array<Maybe<Scalars['String']['output']>>>;
  /** The webhook policy ID. */
  id?: Maybe<Scalars['BigInt']['output']>;
  /** The name of webhook policy. */
  name?: Maybe<Scalars['String']['output']>;
  /** The project ID of webhook policy. */
  projectId?: Maybe<Scalars['Int']['output']>;
  targets?: Maybe<Array<Maybe<WebhookTargetObject>>>;
  /** The update time of the webhook policy. */
  updateTime?: Maybe<Scalars['String']['output']>;
};

/** The webhook policy target object. */
export type WebhookTargetObject = {
  __typename?: 'WebhookTargetObject';
  /** The webhook target address. */
  address?: Maybe<Scalars['String']['output']>;
  /** The webhook auth header. */
  authHeader?: Maybe<Scalars['String']['output']>;
  /** Whether or not to skip cert verify. */
  skipCertVerify?: Maybe<Scalars['Boolean']['output']>;
  /** The webhook target notify type. */
  type?: Maybe<Scalars['String']['output']>;
};

/** worker in the pool */
export type Worker = {
  __typename?: 'Worker';
  /** the checkin of the running job in the worker */
  checkIn?: Maybe<Scalars['String']['output']>;
  /** The checkin time of the worker */
  checkinAt?: Maybe<Scalars['String']['output']>;
  /** the id of the worker */
  id?: Maybe<Scalars['String']['output']>;
  /** the id of the running job in the worker */
  jobId?: Maybe<Scalars['String']['output']>;
  /** the name of the running job in the worker */
  jobName?: Maybe<Scalars['String']['output']>;
  /** the id of the worker pool */
  poolId?: Maybe<Scalars['String']['output']>;
  /** The start time of the worker */
  startAt?: Maybe<Scalars['String']['output']>;
};

/** the worker pool of job service */
export type WorkerPool = {
  __typename?: 'WorkerPool';
  /** The concurrency of the work pool */
  concurrency?: Maybe<Scalars['Int']['output']>;
  /** The heartbeat time of the work pool */
  heartbeatAt?: Maybe<Scalars['String']['output']>;
  /** The host of the work pool */
  host?: Maybe<Scalars['String']['output']>;
  /** the process id of jobservice */
  pid?: Maybe<Scalars['Int']['output']>;
  /** The start time of the work pool */
  startAt?: Maybe<Scalars['String']['output']>;
  /** the id of the worker pool */
  workerPoolId?: Maybe<Scalars['String']['output']>;
};

/** The reference to the Workspace this Template belongs to. */
export type WorkspaceCrownlabsPolitoItWorkspaceRef = {
  __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef';
  /** The name of the resource to be referenced. */
  name: Scalars['String']['output'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']['output']>;
};

/** The reference to the Workspace this Template belongs to. */
export type WorkspaceCrownlabsPolitoItWorkspaceRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String']['input'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: InputMaybe<Scalars['String']['input']>;
};

export type WorkspaceWrapperTenantV1alpha2 = {
  __typename?: 'WorkspaceWrapperTenantV1alpha2';
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItem = {
  __typename?: 'WorkspacesListItem';
  /** The Workspace the Tenant is subscribed to. */
  name: Scalars['String']['output'];
  /** The role of the Tenant in the context of the Workspace. */
  role: Role;
  workspaceWrapperTenantV1alpha2?: Maybe<WorkspaceWrapperTenantV1alpha2>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItemInput = {
  /** The Workspace the Tenant is subscribed to. */
  name: Scalars['String']['input'];
  /** The role of the Tenant in the context of the Workspace. */
  role: Role;
};

export type ApplyInstanceMutationVariables = Exact<{
  instanceId: Scalars['String']['input'];
  tenantNamespace: Scalars['String']['input'];
  patchJson: Scalars['String']['input'];
  manager: Scalars['String']['input'];
}>;


export type ApplyInstanceMutation = { __typename?: 'Mutation', applyInstance?: { __typename?: 'ItPolitoCrownlabsV1alpha2Instance', spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null } | null } | null };

export type ApplyTemplateMutationVariables = Exact<{
  templateId: Scalars['String']['input'];
  workspaceNamespace: Scalars['String']['input'];
  patchJson: Scalars['String']['input'];
  manager: Scalars['String']['input'];
}>;


export type ApplyTemplateMutation = { __typename?: 'Mutation', applyTemplate?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', description: string, name: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, resources: { __typename?: 'Resources', cpu: number, disk?: any | null, memory: any } } | null> } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', id?: string | null } | null } | null };

export type ApplyTenantMutationVariables = Exact<{
  tenantId: Scalars['String']['input'];
  patchJson: Scalars['String']['input'];
  manager: Scalars['String']['input'];
}>;


export type ApplyTenantMutation = { __typename?: 'Mutation', applyTenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null } | null, spec?: { __typename?: 'Spec6', firstName: string, lastName: string, email: string, lastLogin?: string | null, workspaces?: Array<{ __typename?: 'WorkspacesListItem', role: Role, name: string } | null> | null } | null } | null };

export type CreateInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String']['input'];
  templateId: Scalars['String']['input'];
  workspaceNamespace: Scalars['String']['input'];
  tenantId: Scalars['String']['input'];
  generateName?: InputMaybe<Scalars['String']['input']>;
}>;


export type CreateInstanceMutation = { __typename?: 'Mutation', createdInstance?: { __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null, creationTimestamp?: string | null, labels?: any | null } | null, status?: { __typename?: 'Status3', ip?: string | null, phase?: Phase | null, url?: string | null } | null, spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null, templateCrownlabsPolitoItTemplateRef: { __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name: string, namespace?: string | null, templateWrapper?: { __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, environmentType: EnvironmentType } | null> } | null } | null } | null } } | null } | null };

export type CreateTemplateMutationVariables = Exact<{
  workspaceId: Scalars['String']['input'];
  workspaceNamespace: Scalars['String']['input'];
  templateName: Scalars['String']['input'];
  descriptionTemplate: Scalars['String']['input'];
  image: Scalars['String']['input'];
  guiEnabled: Scalars['Boolean']['input'];
  persistent: Scalars['Boolean']['input'];
  mountMyDriveVolume: Scalars['Boolean']['input'];
  resources: ResourcesInput;
  templateId?: InputMaybe<Scalars['String']['input']>;
  environmentType: EnvironmentType;
}>;


export type CreateTemplateMutation = { __typename?: 'Mutation', createdTemplate?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, resources: { __typename?: 'Resources', cpu: number, disk?: any | null, memory: any } } | null> } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null } | null } | null };

export type DeleteInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String']['input'];
  instanceId: Scalars['String']['input'];
}>;


export type DeleteInstanceMutation = { __typename?: 'Mutation', deletedInstance?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: string | null } | null };

export type DeleteLabelSelectorInstancesMutationVariables = Exact<{
  tenantNamespace: Scalars['String']['input'];
  labels?: InputMaybe<Scalars['String']['input']>;
}>;


export type DeleteLabelSelectorInstancesMutation = { __typename?: 'Mutation', deleteLabelSelectorInstances?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: string | null } | null };

export type DeleteTemplateMutationVariables = Exact<{
  workspaceNamespace: Scalars['String']['input'];
  templateId: Scalars['String']['input'];
}>;


export type DeleteTemplateMutation = { __typename?: 'Mutation', deletedTemplate?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: string | null } | null };

export type ImagesQueryVariables = Exact<{ [key: string]: never; }>;


export type ImagesQuery = { __typename?: 'Query', imageList?: { __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList', images: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha1ImageList', spec?: { __typename?: 'Spec', registryName: string, images: Array<{ __typename?: 'ImagesListItem', name: string, versions: Array<string | null> } | null> } | null } | null> } | null };

export type OwnedInstancesQueryVariables = Exact<{
  tenantNamespace: Scalars['String']['input'];
}>;


export type OwnedInstancesQuery = { __typename?: 'Query', instanceList?: { __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList', instances: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null, creationTimestamp?: string | null, labels?: any | null } | null, status?: { __typename?: 'Status3', ip?: string | null, phase?: Phase | null, url?: string | null } | null, spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null, templateCrownlabsPolitoItTemplateRef: { __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name: string, namespace?: string | null, templateWrapper?: { __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, environmentType: EnvironmentType } | null> } | null } | null } | null } } | null } | null> } | null };

export type InstancesLabelSelectorQueryVariables = Exact<{
  labels?: InputMaybe<Scalars['String']['input']>;
}>;


export type InstancesLabelSelectorQuery = { __typename?: 'Query', instanceList?: { __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList', instances: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null, creationTimestamp?: string | null } | null, status?: { __typename?: 'Status3', ip?: string | null, phase?: Phase | null, url?: string | null } | null, spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null, tenantCrownlabsPolitoItTenantRef: { __typename?: 'TenantCrownlabsPolitoItTenantRef', name: string, tenantV1alpha2Wrapper?: { __typename?: 'TenantV1alpha2Wrapper', itPolitoCrownlabsV1alpha2Tenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: { __typename?: 'Spec6', firstName: string, lastName: string } | null } | null } | null }, templateCrownlabsPolitoItTemplateRef: { __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name: string, namespace?: string | null, templateWrapper?: { __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, environmentType: EnvironmentType } | null> } | null } | null } | null } } | null } | null> } | null };

export type WorkspaceTemplatesQueryVariables = Exact<{
  workspaceNamespace: Scalars['String']['input'];
}>;


export type WorkspaceTemplatesQuery = { __typename?: 'Query', templateList?: { __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList', templates: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, resources: { __typename?: 'Resources', cpu: number, disk?: any | null, memory: any } } | null>, workspaceCrownlabsPolitoItWorkspaceRef?: { __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef', name: string } | null } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null } | null } | null> } | null };

export type TenantQueryVariables = Exact<{
  tenantId: Scalars['String']['input'];
}>;


export type TenantQuery = { __typename?: 'Query', tenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: { __typename?: 'Spec6', email: string, firstName: string, lastName: string, lastLogin?: string | null, publicKeys?: Array<string | null> | null, workspaces?: Array<{ __typename?: 'WorkspacesListItem', role: Role, name: string, workspaceWrapperTenantV1alpha2?: { __typename?: 'WorkspaceWrapperTenantV1alpha2', itPolitoCrownlabsV1alpha1Workspace?: { __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', spec?: { __typename?: 'Spec2', prettyName: string } | null, status?: { __typename?: 'Status2', namespace?: { __typename?: 'Namespace', name?: string | null } | null } | null } | null } | null } | null> | null } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null } | null, status?: { __typename?: 'Status6', personalNamespace: { __typename?: 'PersonalNamespace', name?: string | null, created: boolean }, quota?: { __typename?: 'Quota3', cpu: any, instances: number, memory: any } | null } | null } | null };

export type TenantsQueryVariables = Exact<{
  labels?: InputMaybe<Scalars['String']['input']>;
  retrieveWorkspaces?: InputMaybe<Scalars['Boolean']['input']>;
}>;


export type TenantsQuery = { __typename?: 'Query', tenants?: { __typename?: 'ItPolitoCrownlabsV1alpha2TenantList', items: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null } | null, spec?: { __typename?: 'Spec6', firstName: string, lastName: string, email: string, workspaces?: Array<{ __typename?: 'WorkspacesListItem', role: Role, name: string } | null> | null } | null } | null> } | null };

export type WorkspacesQueryVariables = Exact<{
  labels?: InputMaybe<Scalars['String']['input']>;
}>;


export type WorkspacesQuery = { __typename?: 'Query', workspaces?: { __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceList', items: Array<{ __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null } | null, spec?: { __typename?: 'Spec2', prettyName: string, autoEnroll?: AutoEnroll | null } | null } | null> } | null };

export type UpdatedOwnedInstancesSubscriptionVariables = Exact<{
  tenantNamespace: Scalars['String']['input'];
  instanceId?: InputMaybe<Scalars['String']['input']>;
}>;


export type UpdatedOwnedInstancesSubscription = { __typename?: 'Subscription', updateInstance?: { __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate', updateType?: UpdateType | null, instance?: { __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null, creationTimestamp?: string | null, labels?: any | null } | null, status?: { __typename?: 'Status3', ip?: string | null, phase?: Phase | null, url?: string | null } | null, spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null, templateCrownlabsPolitoItTemplateRef: { __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name: string, namespace?: string | null, templateWrapper?: { __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, environmentType: EnvironmentType } | null> } | null } | null } | null } } | null } | null } | null };

export type UpdatedInstancesLabelSelectorSubscriptionVariables = Exact<{
  labels?: InputMaybe<Scalars['String']['input']>;
}>;


export type UpdatedInstancesLabelSelectorSubscription = { __typename?: 'Subscription', updateInstanceLabelSelector?: { __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate', updateType?: UpdateType | null, instance?: { __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null, creationTimestamp?: string | null } | null, status?: { __typename?: 'Status3', ip?: string | null, phase?: Phase | null, url?: string | null } | null, spec?: { __typename?: 'Spec3', running?: boolean | null, prettyName?: string | null, tenantCrownlabsPolitoItTenantRef: { __typename?: 'TenantCrownlabsPolitoItTenantRef', name: string, tenantV1alpha2Wrapper?: { __typename?: 'TenantV1alpha2Wrapper', itPolitoCrownlabsV1alpha2Tenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: { __typename?: 'Spec6', firstName: string, lastName: string } | null } | null } | null }, templateCrownlabsPolitoItTemplateRef: { __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name: string, namespace?: string | null, templateWrapper?: { __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, environmentType: EnvironmentType } | null> } | null } | null } | null } } | null } | null } | null };

export type UpdatedWorkspaceTemplatesSubscriptionVariables = Exact<{
  workspaceNamespace: Scalars['String']['input'];
  templateId?: InputMaybe<Scalars['String']['input']>;
}>;


export type UpdatedWorkspaceTemplatesSubscription = { __typename?: 'Subscription', updatedTemplate?: { __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate', updateType?: UpdateType | null, template?: { __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: { __typename?: 'Spec5', prettyName: string, description: string, environmentList: Array<{ __typename?: 'EnvironmentListListItem', guiEnabled?: boolean | null, persistent?: boolean | null, resources: { __typename?: 'Resources', cpu: number, disk?: any | null, memory: any } } | null>, workspaceCrownlabsPolitoItWorkspaceRef?: { __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef', name: string } | null } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null, namespace?: string | null } | null } | null } | null };

export type UpdatedTenantSubscriptionVariables = Exact<{
  tenantId: Scalars['String']['input'];
}>;


export type UpdatedTenantSubscription = { __typename?: 'Subscription', updatedTenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2TenantUpdate', updateType?: UpdateType | null, tenant?: { __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: { __typename?: 'Spec6', email: string, firstName: string, lastName: string, lastLogin?: string | null, publicKeys?: Array<string | null> | null, workspaces?: Array<{ __typename?: 'WorkspacesListItem', role: Role, name: string, workspaceWrapperTenantV1alpha2?: { __typename?: 'WorkspaceWrapperTenantV1alpha2', itPolitoCrownlabsV1alpha1Workspace?: { __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', spec?: { __typename?: 'Spec2', prettyName: string } | null, status?: { __typename?: 'Status2', namespace?: { __typename?: 'Namespace', name?: string | null } | null } | null } | null } | null } | null> | null } | null, metadata?: { __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: string | null } | null, status?: { __typename?: 'Status6', personalNamespace: { __typename?: 'PersonalNamespace', name?: string | null, created: boolean }, quota?: { __typename?: 'Quota3', cpu: any, instances: number, memory: any } | null } | null } | null } | null };


export const ApplyInstanceDocument = gql`
    mutation applyInstance($instanceId: String!, $tenantNamespace: String!, $patchJson: String!, $manager: String!) {
  applyInstance: patchCrownlabsPolitoItV1alpha2NamespacedInstance(
    name: $instanceId
    namespace: $tenantNamespace
    force: true
    fieldManager: $manager
    applicationApplyPatchYamlInput: $patchJson
  ) {
    spec {
      running
      prettyName
    }
  }
}
    `;
export type ApplyInstanceMutationFn = Apollo.MutationFunction<ApplyInstanceMutation, ApplyInstanceMutationVariables>;

/**
 * __useApplyInstanceMutation__
 *
 * To run a mutation, you first call `useApplyInstanceMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useApplyInstanceMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [applyInstanceMutation, { data, loading, error }] = useApplyInstanceMutation({
 *   variables: {
 *      instanceId: // value for 'instanceId'
 *      tenantNamespace: // value for 'tenantNamespace'
 *      patchJson: // value for 'patchJson'
 *      manager: // value for 'manager'
 *   },
 * });
 */
export function useApplyInstanceMutation(baseOptions?: Apollo.MutationHookOptions<ApplyInstanceMutation, ApplyInstanceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ApplyInstanceMutation, ApplyInstanceMutationVariables>(ApplyInstanceDocument, options);
      }
export type ApplyInstanceMutationHookResult = ReturnType<typeof useApplyInstanceMutation>;
export type ApplyInstanceMutationResult = Apollo.MutationResult<ApplyInstanceMutation>;
export type ApplyInstanceMutationOptions = Apollo.BaseMutationOptions<ApplyInstanceMutation, ApplyInstanceMutationVariables>;
export const ApplyTemplateDocument = gql`
    mutation applyTemplate($templateId: String!, $workspaceNamespace: String!, $patchJson: String!, $manager: String!) {
  applyTemplate: patchCrownlabsPolitoItV1alpha2NamespacedTemplate(
    name: $templateId
    namespace: $workspaceNamespace
    force: true
    fieldManager: $manager
    applicationApplyPatchYamlInput: $patchJson
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
export type ApplyTemplateMutationFn = Apollo.MutationFunction<ApplyTemplateMutation, ApplyTemplateMutationVariables>;

/**
 * __useApplyTemplateMutation__
 *
 * To run a mutation, you first call `useApplyTemplateMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useApplyTemplateMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [applyTemplateMutation, { data, loading, error }] = useApplyTemplateMutation({
 *   variables: {
 *      templateId: // value for 'templateId'
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      patchJson: // value for 'patchJson'
 *      manager: // value for 'manager'
 *   },
 * });
 */
export function useApplyTemplateMutation(baseOptions?: Apollo.MutationHookOptions<ApplyTemplateMutation, ApplyTemplateMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ApplyTemplateMutation, ApplyTemplateMutationVariables>(ApplyTemplateDocument, options);
      }
export type ApplyTemplateMutationHookResult = ReturnType<typeof useApplyTemplateMutation>;
export type ApplyTemplateMutationResult = Apollo.MutationResult<ApplyTemplateMutation>;
export type ApplyTemplateMutationOptions = Apollo.BaseMutationOptions<ApplyTemplateMutation, ApplyTemplateMutationVariables>;
export const ApplyTenantDocument = gql`
    mutation applyTenant($tenantId: String!, $patchJson: String!, $manager: String!) {
  applyTenant: patchCrownlabsPolitoItV1alpha2Tenant(
    name: $tenantId
    force: true
    fieldManager: $manager
    applicationApplyPatchYamlInput: $patchJson
  ) {
    metadata {
      name
    }
    spec {
      firstName
      lastName
      email
      lastLogin
      workspaces {
        role
        name
      }
    }
  }
}
    `;
export type ApplyTenantMutationFn = Apollo.MutationFunction<ApplyTenantMutation, ApplyTenantMutationVariables>;

/**
 * __useApplyTenantMutation__
 *
 * To run a mutation, you first call `useApplyTenantMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useApplyTenantMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [applyTenantMutation, { data, loading, error }] = useApplyTenantMutation({
 *   variables: {
 *      tenantId: // value for 'tenantId'
 *      patchJson: // value for 'patchJson'
 *      manager: // value for 'manager'
 *   },
 * });
 */
export function useApplyTenantMutation(baseOptions?: Apollo.MutationHookOptions<ApplyTenantMutation, ApplyTenantMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<ApplyTenantMutation, ApplyTenantMutationVariables>(ApplyTenantDocument, options);
      }
export type ApplyTenantMutationHookResult = ReturnType<typeof useApplyTenantMutation>;
export type ApplyTenantMutationResult = Apollo.MutationResult<ApplyTenantMutation>;
export type ApplyTenantMutationOptions = Apollo.BaseMutationOptions<ApplyTenantMutation, ApplyTenantMutationVariables>;
export const CreateInstanceDocument = gql`
    mutation createInstance($tenantNamespace: String!, $templateId: String!, $workspaceNamespace: String!, $tenantId: String!, $generateName: String = "instance-") {
  createdInstance: createCrownlabsPolitoItV1alpha2NamespacedInstance(
    namespace: $tenantNamespace
    itPolitoCrownlabsV1alpha2InstanceInput: {kind: "Instance", apiVersion: "crownlabs.polito.it/v1alpha2", metadata: {generateName: $generateName}, spec: {templateCrownlabsPolitoItTemplateRef: {name: $templateId, namespace: $workspaceNamespace}, tenantCrownlabsPolitoItTenantRef: {name: $tenantId, namespace: $tenantNamespace}}}
  ) {
    metadata {
      name
      namespace
      creationTimestamp
      labels
    }
    status {
      ip
      phase
      url
    }
    spec {
      running
      prettyName
      templateCrownlabsPolitoItTemplateRef {
        name
        namespace
        templateWrapper {
          itPolitoCrownlabsV1alpha2Template {
            spec {
              prettyName
              description
              environmentList {
                guiEnabled
                persistent
                environmentType
              }
            }
          }
        }
      }
    }
  }
}
    `;
export type CreateInstanceMutationFn = Apollo.MutationFunction<CreateInstanceMutation, CreateInstanceMutationVariables>;

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
 *      templateId: // value for 'templateId'
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      tenantId: // value for 'tenantId'
 *      generateName: // value for 'generateName'
 *   },
 * });
 */
export function useCreateInstanceMutation(baseOptions?: Apollo.MutationHookOptions<CreateInstanceMutation, CreateInstanceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateInstanceMutation, CreateInstanceMutationVariables>(CreateInstanceDocument, options);
      }
export type CreateInstanceMutationHookResult = ReturnType<typeof useCreateInstanceMutation>;
export type CreateInstanceMutationResult = Apollo.MutationResult<CreateInstanceMutation>;
export type CreateInstanceMutationOptions = Apollo.BaseMutationOptions<CreateInstanceMutation, CreateInstanceMutationVariables>;
export const CreateTemplateDocument = gql`
    mutation createTemplate($workspaceId: String!, $workspaceNamespace: String!, $templateName: String!, $descriptionTemplate: String!, $image: String!, $guiEnabled: Boolean!, $persistent: Boolean!, $mountMyDriveVolume: Boolean!, $resources: ResourcesInput!, $templateId: String = "template-", $environmentType: EnvironmentType!) {
  createdTemplate: createCrownlabsPolitoItV1alpha2NamespacedTemplate(
    namespace: $workspaceNamespace
    itPolitoCrownlabsV1alpha2TemplateInput: {kind: "Template", apiVersion: "crownlabs.polito.it/v1alpha2", spec: {prettyName: $templateName, description: $descriptionTemplate, environmentList: [{name: "default", environmentType: $environmentType, image: $image, guiEnabled: $guiEnabled, persistent: $persistent, resources: $resources, mountMyDriveVolume: $mountMyDriveVolume}], workspaceCrownlabsPolitoItWorkspaceRef: {name: $workspaceId}}, metadata: {generateName: $templateId, namespace: $workspaceNamespace}}
  ) {
    spec {
      prettyName
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
      name
      namespace
    }
  }
}
    `;
export type CreateTemplateMutationFn = Apollo.MutationFunction<CreateTemplateMutation, CreateTemplateMutationVariables>;

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
 *      workspaceId: // value for 'workspaceId'
 *      workspaceNamespace: // value for 'workspaceNamespace'
 *      templateName: // value for 'templateName'
 *      descriptionTemplate: // value for 'descriptionTemplate'
 *      image: // value for 'image'
 *      guiEnabled: // value for 'guiEnabled'
 *      persistent: // value for 'persistent'
 *      mountMyDriveVolume: // value for 'mountMyDriveVolume'
 *      resources: // value for 'resources'
 *      templateId: // value for 'templateId'
 *      environmentType: // value for 'environmentType'
 *   },
 * });
 */
export function useCreateTemplateMutation(baseOptions?: Apollo.MutationHookOptions<CreateTemplateMutation, CreateTemplateMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<CreateTemplateMutation, CreateTemplateMutationVariables>(CreateTemplateDocument, options);
      }
export type CreateTemplateMutationHookResult = ReturnType<typeof useCreateTemplateMutation>;
export type CreateTemplateMutationResult = Apollo.MutationResult<CreateTemplateMutation>;
export type CreateTemplateMutationOptions = Apollo.BaseMutationOptions<CreateTemplateMutation, CreateTemplateMutationVariables>;
export const DeleteInstanceDocument = gql`
    mutation deleteInstance($tenantNamespace: String!, $instanceId: String!) {
  deletedInstance: deleteCrownlabsPolitoItV1alpha2NamespacedInstance(
    namespace: $tenantNamespace
    name: $instanceId
  ) {
    kind
  }
}
    `;
export type DeleteInstanceMutationFn = Apollo.MutationFunction<DeleteInstanceMutation, DeleteInstanceMutationVariables>;

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
 *      instanceId: // value for 'instanceId'
 *   },
 * });
 */
export function useDeleteInstanceMutation(baseOptions?: Apollo.MutationHookOptions<DeleteInstanceMutation, DeleteInstanceMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteInstanceMutation, DeleteInstanceMutationVariables>(DeleteInstanceDocument, options);
      }
export type DeleteInstanceMutationHookResult = ReturnType<typeof useDeleteInstanceMutation>;
export type DeleteInstanceMutationResult = Apollo.MutationResult<DeleteInstanceMutation>;
export type DeleteInstanceMutationOptions = Apollo.BaseMutationOptions<DeleteInstanceMutation, DeleteInstanceMutationVariables>;
export const DeleteLabelSelectorInstancesDocument = gql`
    mutation deleteLabelSelectorInstances($tenantNamespace: String!, $labels: String) {
  deleteLabelSelectorInstances: deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstance(
    namespace: $tenantNamespace
    labelSelector: $labels
  ) {
    kind
  }
}
    `;
export type DeleteLabelSelectorInstancesMutationFn = Apollo.MutationFunction<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables>;

/**
 * __useDeleteLabelSelectorInstancesMutation__
 *
 * To run a mutation, you first call `useDeleteLabelSelectorInstancesMutation` within a React component and pass it any options that fit your needs.
 * When your component renders, `useDeleteLabelSelectorInstancesMutation` returns a tuple that includes:
 * - A mutate function that you can call at any time to execute the mutation
 * - An object with fields that represent the current status of the mutation's execution
 *
 * @param baseOptions options that will be passed into the mutation, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options-2;
 *
 * @example
 * const [deleteLabelSelectorInstancesMutation, { data, loading, error }] = useDeleteLabelSelectorInstancesMutation({
 *   variables: {
 *      tenantNamespace: // value for 'tenantNamespace'
 *      labels: // value for 'labels'
 *   },
 * });
 */
export function useDeleteLabelSelectorInstancesMutation(baseOptions?: Apollo.MutationHookOptions<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables>(DeleteLabelSelectorInstancesDocument, options);
      }
export type DeleteLabelSelectorInstancesMutationHookResult = ReturnType<typeof useDeleteLabelSelectorInstancesMutation>;
export type DeleteLabelSelectorInstancesMutationResult = Apollo.MutationResult<DeleteLabelSelectorInstancesMutation>;
export type DeleteLabelSelectorInstancesMutationOptions = Apollo.BaseMutationOptions<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables>;
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
export type DeleteTemplateMutationFn = Apollo.MutationFunction<DeleteTemplateMutation, DeleteTemplateMutationVariables>;

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
export function useDeleteTemplateMutation(baseOptions?: Apollo.MutationHookOptions<DeleteTemplateMutation, DeleteTemplateMutationVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useMutation<DeleteTemplateMutation, DeleteTemplateMutationVariables>(DeleteTemplateDocument, options);
      }
export type DeleteTemplateMutationHookResult = ReturnType<typeof useDeleteTemplateMutation>;
export type DeleteTemplateMutationResult = Apollo.MutationResult<DeleteTemplateMutation>;
export type DeleteTemplateMutationOptions = Apollo.BaseMutationOptions<DeleteTemplateMutation, DeleteTemplateMutationVariables>;
export const ImagesDocument = gql`
    query images {
  imageList: itPolitoCrownlabsV1alpha1ImageListList {
    images: items {
      spec {
        registryName
        images {
          name
          versions
        }
      }
    }
  }
}
    `;

/**
 * __useImagesQuery__
 *
 * To run a query within a React component, call `useImagesQuery` and pass it any options that fit your needs.
 * When your component renders, `useImagesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useImagesQuery({
 *   variables: {
 *   },
 * });
 */
export function useImagesQuery(baseOptions?: Apollo.QueryHookOptions<ImagesQuery, ImagesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<ImagesQuery, ImagesQueryVariables>(ImagesDocument, options);
      }
export function useImagesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<ImagesQuery, ImagesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<ImagesQuery, ImagesQueryVariables>(ImagesDocument, options);
        }
export function useImagesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<ImagesQuery, ImagesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<ImagesQuery, ImagesQueryVariables>(ImagesDocument, options);
        }
export type ImagesQueryHookResult = ReturnType<typeof useImagesQuery>;
export type ImagesLazyQueryHookResult = ReturnType<typeof useImagesLazyQuery>;
export type ImagesSuspenseQueryHookResult = ReturnType<typeof useImagesSuspenseQuery>;
export type ImagesQueryResult = Apollo.QueryResult<ImagesQuery, ImagesQueryVariables>;
export const OwnedInstancesDocument = gql`
    query ownedInstances($tenantNamespace: String!) {
  instanceList: listCrownlabsPolitoItV1alpha2NamespacedInstance(
    namespace: $tenantNamespace
  ) {
    instances: items {
      metadata {
        name
        namespace
        creationTimestamp
        labels
      }
      status {
        ip
        phase
        url
      }
      spec {
        running
        prettyName
        templateCrownlabsPolitoItTemplateRef {
          name
          namespace
          templateWrapper {
            itPolitoCrownlabsV1alpha2Template {
              spec {
                prettyName
                description
                environmentList {
                  guiEnabled
                  persistent
                  environmentType
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
export function useOwnedInstancesQuery(baseOptions: Apollo.QueryHookOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables> & ({ variables: OwnedInstancesQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(OwnedInstancesDocument, options);
      }
export function useOwnedInstancesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(OwnedInstancesDocument, options);
        }
export function useOwnedInstancesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(OwnedInstancesDocument, options);
        }
export type OwnedInstancesQueryHookResult = ReturnType<typeof useOwnedInstancesQuery>;
export type OwnedInstancesLazyQueryHookResult = ReturnType<typeof useOwnedInstancesLazyQuery>;
export type OwnedInstancesSuspenseQueryHookResult = ReturnType<typeof useOwnedInstancesSuspenseQuery>;
export type OwnedInstancesQueryResult = Apollo.QueryResult<OwnedInstancesQuery, OwnedInstancesQueryVariables>;
export const InstancesLabelSelectorDocument = gql`
    query instancesLabelSelector($labels: String) {
  instanceList: itPolitoCrownlabsV1alpha2InstanceList(labelSelector: $labels) {
    instances: items {
      metadata {
        name
        namespace
        creationTimestamp
      }
      status {
        ip
        phase
        url
      }
      spec {
        running
        prettyName
        tenantCrownlabsPolitoItTenantRef {
          name
          tenantV1alpha2Wrapper {
            itPolitoCrownlabsV1alpha2Tenant {
              spec {
                firstName
                lastName
              }
            }
          }
        }
        templateCrownlabsPolitoItTemplateRef {
          name
          namespace
          templateWrapper {
            itPolitoCrownlabsV1alpha2Template {
              spec {
                prettyName
                description
                environmentList {
                  guiEnabled
                  persistent
                  environmentType
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

/**
 * __useInstancesLabelSelectorQuery__
 *
 * To run a query within a React component, call `useInstancesLabelSelectorQuery` and pass it any options that fit your needs.
 * When your component renders, `useInstancesLabelSelectorQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useInstancesLabelSelectorQuery({
 *   variables: {
 *      labels: // value for 'labels'
 *   },
 * });
 */
export function useInstancesLabelSelectorQuery(baseOptions?: Apollo.QueryHookOptions<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>(InstancesLabelSelectorDocument, options);
      }
export function useInstancesLabelSelectorLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>(InstancesLabelSelectorDocument, options);
        }
export function useInstancesLabelSelectorSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>(InstancesLabelSelectorDocument, options);
        }
export type InstancesLabelSelectorQueryHookResult = ReturnType<typeof useInstancesLabelSelectorQuery>;
export type InstancesLabelSelectorLazyQueryHookResult = ReturnType<typeof useInstancesLabelSelectorLazyQuery>;
export type InstancesLabelSelectorSuspenseQueryHookResult = ReturnType<typeof useInstancesLabelSelectorSuspenseQuery>;
export type InstancesLabelSelectorQueryResult = Apollo.QueryResult<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>;
export const WorkspaceTemplatesDocument = gql`
    query workspaceTemplates($workspaceNamespace: String!) {
  templateList: itPolitoCrownlabsV1alpha2TemplateList(
    namespace: $workspaceNamespace
  ) {
    templates: items {
      spec {
        prettyName
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
        workspaceCrownlabsPolitoItWorkspaceRef {
          name
        }
      }
      metadata {
        name
        namespace
      }
    }
  }
}
    `;

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
export function useWorkspaceTemplatesQuery(baseOptions: Apollo.QueryHookOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables> & ({ variables: WorkspaceTemplatesQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>(WorkspaceTemplatesDocument, options);
      }
export function useWorkspaceTemplatesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>(WorkspaceTemplatesDocument, options);
        }
export function useWorkspaceTemplatesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>(WorkspaceTemplatesDocument, options);
        }
export type WorkspaceTemplatesQueryHookResult = ReturnType<typeof useWorkspaceTemplatesQuery>;
export type WorkspaceTemplatesLazyQueryHookResult = ReturnType<typeof useWorkspaceTemplatesLazyQuery>;
export type WorkspaceTemplatesSuspenseQueryHookResult = ReturnType<typeof useWorkspaceTemplatesSuspenseQuery>;
export type WorkspaceTemplatesQueryResult = Apollo.QueryResult<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>;
export const TenantDocument = gql`
    query tenant($tenantId: String!) {
  tenant: itPolitoCrownlabsV1alpha2Tenant(name: $tenantId) {
    spec {
      email
      firstName
      lastName
      lastLogin
      workspaces {
        role
        name
        workspaceWrapperTenantV1alpha2 {
          itPolitoCrownlabsV1alpha1Workspace {
            spec {
              prettyName
            }
            status {
              namespace {
                name
              }
            }
          }
        }
      }
      publicKeys
    }
    metadata {
      name
    }
    status {
      personalNamespace {
        name
        created
      }
      quota {
        cpu
        instances
        memory
      }
    }
  }
}
    `;

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
export function useTenantQuery(baseOptions: Apollo.QueryHookOptions<TenantQuery, TenantQueryVariables> & ({ variables: TenantQueryVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TenantQuery, TenantQueryVariables>(TenantDocument, options);
      }
export function useTenantLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TenantQuery, TenantQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TenantQuery, TenantQueryVariables>(TenantDocument, options);
        }
export function useTenantSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TenantQuery, TenantQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<TenantQuery, TenantQueryVariables>(TenantDocument, options);
        }
export type TenantQueryHookResult = ReturnType<typeof useTenantQuery>;
export type TenantLazyQueryHookResult = ReturnType<typeof useTenantLazyQuery>;
export type TenantSuspenseQueryHookResult = ReturnType<typeof useTenantSuspenseQuery>;
export type TenantQueryResult = Apollo.QueryResult<TenantQuery, TenantQueryVariables>;
export const TenantsDocument = gql`
    query tenants($labels: String, $retrieveWorkspaces: Boolean = false) {
  tenants: itPolitoCrownlabsV1alpha2TenantList(labelSelector: $labels) {
    items {
      metadata {
        name
      }
      spec {
        firstName
        lastName
        email
        workspaces @include(if: $retrieveWorkspaces) {
          role
          name
        }
      }
    }
  }
}
    `;

/**
 * __useTenantsQuery__
 *
 * To run a query within a React component, call `useTenantsQuery` and pass it any options that fit your needs.
 * When your component renders, `useTenantsQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useTenantsQuery({
 *   variables: {
 *      labels: // value for 'labels'
 *      retrieveWorkspaces: // value for 'retrieveWorkspaces'
 *   },
 * });
 */
export function useTenantsQuery(baseOptions?: Apollo.QueryHookOptions<TenantsQuery, TenantsQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TenantsQuery, TenantsQueryVariables>(TenantsDocument, options);
      }
export function useTenantsLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TenantsQuery, TenantsQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TenantsQuery, TenantsQueryVariables>(TenantsDocument, options);
        }
export function useTenantsSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<TenantsQuery, TenantsQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<TenantsQuery, TenantsQueryVariables>(TenantsDocument, options);
        }
export type TenantsQueryHookResult = ReturnType<typeof useTenantsQuery>;
export type TenantsLazyQueryHookResult = ReturnType<typeof useTenantsLazyQuery>;
export type TenantsSuspenseQueryHookResult = ReturnType<typeof useTenantsSuspenseQuery>;
export type TenantsQueryResult = Apollo.QueryResult<TenantsQuery, TenantsQueryVariables>;
export const WorkspacesDocument = gql`
    query workspaces($labels: String) {
  workspaces: itPolitoCrownlabsV1alpha1WorkspaceList(labelSelector: $labels) {
    items {
      metadata {
        name
      }
      spec {
        prettyName
        autoEnroll
      }
    }
  }
}
    `;

/**
 * __useWorkspacesQuery__
 *
 * To run a query within a React component, call `useWorkspacesQuery` and pass it any options that fit your needs.
 * When your component renders, `useWorkspacesQuery` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the query, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useWorkspacesQuery({
 *   variables: {
 *      labels: // value for 'labels'
 *   },
 * });
 */
export function useWorkspacesQuery(baseOptions?: Apollo.QueryHookOptions<WorkspacesQuery, WorkspacesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WorkspacesQuery, WorkspacesQueryVariables>(WorkspacesDocument, options);
      }
export function useWorkspacesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WorkspacesQuery, WorkspacesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WorkspacesQuery, WorkspacesQueryVariables>(WorkspacesDocument, options);
        }
export function useWorkspacesSuspenseQuery(baseOptions?: Apollo.SkipToken | Apollo.SuspenseQueryHookOptions<WorkspacesQuery, WorkspacesQueryVariables>) {
          const options = baseOptions === Apollo.skipToken ? baseOptions : {...defaultOptions, ...baseOptions}
          return Apollo.useSuspenseQuery<WorkspacesQuery, WorkspacesQueryVariables>(WorkspacesDocument, options);
        }
export type WorkspacesQueryHookResult = ReturnType<typeof useWorkspacesQuery>;
export type WorkspacesLazyQueryHookResult = ReturnType<typeof useWorkspacesLazyQuery>;
export type WorkspacesSuspenseQueryHookResult = ReturnType<typeof useWorkspacesSuspenseQuery>;
export type WorkspacesQueryResult = Apollo.QueryResult<WorkspacesQuery, WorkspacesQueryVariables>;
export const UpdatedOwnedInstancesDocument = gql`
    subscription updatedOwnedInstances($tenantNamespace: String!, $instanceId: String) {
  updateInstance: itPolitoCrownlabsV1alpha2InstanceUpdate(
    namespace: $tenantNamespace
    name: $instanceId
  ) {
    updateType
    instance: payload {
      metadata {
        name
        namespace
        creationTimestamp
        labels
      }
      status {
        ip
        phase
        url
      }
      spec {
        running
        prettyName
        templateCrownlabsPolitoItTemplateRef {
          name
          namespace
          templateWrapper {
            itPolitoCrownlabsV1alpha2Template {
              spec {
                prettyName
                description
                environmentList {
                  guiEnabled
                  persistent
                  environmentType
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
 *      instanceId: // value for 'instanceId'
 *   },
 * });
 */
export function useUpdatedOwnedInstancesSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedOwnedInstancesSubscription, UpdatedOwnedInstancesSubscriptionVariables> & ({ variables: UpdatedOwnedInstancesSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<UpdatedOwnedInstancesSubscription, UpdatedOwnedInstancesSubscriptionVariables>(UpdatedOwnedInstancesDocument, options);
      }
export type UpdatedOwnedInstancesSubscriptionHookResult = ReturnType<typeof useUpdatedOwnedInstancesSubscription>;
export type UpdatedOwnedInstancesSubscriptionResult = Apollo.SubscriptionResult<UpdatedOwnedInstancesSubscription>;
export const UpdatedInstancesLabelSelectorDocument = gql`
    subscription updatedInstancesLabelSelector($labels: String) {
  updateInstanceLabelSelector: itPolitoCrownlabsV1alpha2InstanceLabelsUpdate(
    labelSelector: $labels
  ) {
    updateType
    instance: payload {
      metadata {
        name
        namespace
        creationTimestamp
      }
      status {
        ip
        phase
        url
      }
      spec {
        running
        prettyName
        tenantCrownlabsPolitoItTenantRef {
          name
          tenantV1alpha2Wrapper {
            itPolitoCrownlabsV1alpha2Tenant {
              spec {
                firstName
                lastName
              }
            }
          }
        }
        templateCrownlabsPolitoItTemplateRef {
          name
          namespace
          templateWrapper {
            itPolitoCrownlabsV1alpha2Template {
              spec {
                prettyName
                description
                environmentList {
                  guiEnabled
                  persistent
                  environmentType
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

/**
 * __useUpdatedInstancesLabelSelectorSubscription__
 *
 * To run a query within a React component, call `useUpdatedInstancesLabelSelectorSubscription` and pass it any options that fit your needs.
 * When your component renders, `useUpdatedInstancesLabelSelectorSubscription` returns an object from Apollo Client that contains loading, error, and data properties
 * you can use to render your UI.
 *
 * @param baseOptions options that will be passed into the subscription, supported options are listed on: https://www.apollographql.com/docs/react/api/react-hooks/#options;
 *
 * @example
 * const { data, loading, error } = useUpdatedInstancesLabelSelectorSubscription({
 *   variables: {
 *      labels: // value for 'labels'
 *   },
 * });
 */
export function useUpdatedInstancesLabelSelectorSubscription(baseOptions?: Apollo.SubscriptionHookOptions<UpdatedInstancesLabelSelectorSubscription, UpdatedInstancesLabelSelectorSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<UpdatedInstancesLabelSelectorSubscription, UpdatedInstancesLabelSelectorSubscriptionVariables>(UpdatedInstancesLabelSelectorDocument, options);
      }
export type UpdatedInstancesLabelSelectorSubscriptionHookResult = ReturnType<typeof useUpdatedInstancesLabelSelectorSubscription>;
export type UpdatedInstancesLabelSelectorSubscriptionResult = Apollo.SubscriptionResult<UpdatedInstancesLabelSelectorSubscription>;
export const UpdatedWorkspaceTemplatesDocument = gql`
    subscription updatedWorkspaceTemplates($workspaceNamespace: String!, $templateId: String) {
  updatedTemplate: itPolitoCrownlabsV1alpha2TemplateUpdate(
    namespace: $workspaceNamespace
    name: $templateId
  ) {
    updateType
    template: payload {
      spec {
        prettyName
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
        workspaceCrownlabsPolitoItWorkspaceRef {
          name
        }
      }
      metadata {
        name
        namespace
      }
    }
  }
}
    `;

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
 *      templateId: // value for 'templateId'
 *   },
 * });
 */
export function useUpdatedWorkspaceTemplatesSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedWorkspaceTemplatesSubscription, UpdatedWorkspaceTemplatesSubscriptionVariables> & ({ variables: UpdatedWorkspaceTemplatesSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<UpdatedWorkspaceTemplatesSubscription, UpdatedWorkspaceTemplatesSubscriptionVariables>(UpdatedWorkspaceTemplatesDocument, options);
      }
export type UpdatedWorkspaceTemplatesSubscriptionHookResult = ReturnType<typeof useUpdatedWorkspaceTemplatesSubscription>;
export type UpdatedWorkspaceTemplatesSubscriptionResult = Apollo.SubscriptionResult<UpdatedWorkspaceTemplatesSubscription>;
export const UpdatedTenantDocument = gql`
    subscription updatedTenant($tenantId: String!) {
  updatedTenant: itPolitoCrownlabsV1alpha2TenantUpdate(name: $tenantId) {
    updateType
    tenant: payload {
      spec {
        email
        firstName
        lastName
        lastLogin
        workspaces {
          role
          name
          workspaceWrapperTenantV1alpha2 {
            itPolitoCrownlabsV1alpha1Workspace {
              spec {
                prettyName
              }
              status {
                namespace {
                  name
                }
              }
            }
          }
        }
        publicKeys
      }
      metadata {
        name
      }
      status {
        personalNamespace {
          name
          created
        }
        quota {
          cpu
          instances
          memory
        }
      }
    }
  }
}
    `;

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
export function useUpdatedTenantSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables> & ({ variables: UpdatedTenantSubscriptionVariables; skip?: boolean; } | { skip: boolean; }) ) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables>(UpdatedTenantDocument, options);
      }
export type UpdatedTenantSubscriptionHookResult = ReturnType<typeof useUpdatedTenantSubscription>;
export type UpdatedTenantSubscriptionResult = Apollo.SubscriptionResult<UpdatedTenantSubscription>;