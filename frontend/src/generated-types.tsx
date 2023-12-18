import { gql } from '@apollo/client';
import * as Apollo from '@apollo/client';
import * as React from 'react';
import * as ApolloReactComponents from '@apollo/client/react/components';
export type Maybe<T> = T | null;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type Omit<T, K extends keyof T> = Pick<T, Exclude<keyof T, K>>;
const defaultOptions =  {}
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: string;
  String: string;
  Boolean: boolean;
  Int: number;
  Float: number;
  /** The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf). */
  JSON: any;
};

export enum AutoEnroll {
  None = '_',
  WithApproval = 'withApproval',
  Immediate = 'immediate'
}

/** Timestamps of the Instance automation phases (check, termination and submission). */
export type Automation = {
  __typename?: 'Automation';
  /** The last time the Instance desired status was checked. */
  lastCheckTime?: Maybe<Scalars['String']>;
  /** The time the Instance content submission has been completed. */
  submissionTime?: Maybe<Scalars['String']>;
  /** The (possibly expected) termination time of the Instance. */
  terminationTime?: Maybe<Scalars['String']>;
};

/** Timestamps of the Instance automation phases (check, termination and submission). */
export type AutomationInput = {
  /** The last time the Instance desired status was checked. */
  lastCheckTime?: Maybe<Scalars['String']>;
  /** The time the Instance content submission has been completed. */
  submissionTime?: Maybe<Scalars['String']>;
  /** The (possibly expected) termination time of the Instance. */
  terminationTime?: Maybe<Scalars['String']>;
};

/** Options to customize container startup */
export type ContainerStartupOptions = {
  __typename?: 'ContainerStartupOptions';
  /** Path on which storage (EmptyDir/Storage) will be mounted and into which, if given in SourceArchiveURL, will be extracted the archive */
  contentPath?: Maybe<Scalars['String']>;
  /** Whether forcing the container working directory to be the same as the contentPath (or default mydrive path if not specified) */
  enforceWorkdir?: Maybe<Scalars['Boolean']>;
  /** URL from which GET the archive to be extracted into ContentPath */
  sourceArchiveURL?: Maybe<Scalars['String']>;
  /** Arguments to be passed to the application container on startup */
  startupArgs?: Maybe<Array<Maybe<Scalars['String']>>>;
};

/** Options to customize container startup */
export type ContainerStartupOptionsInput = {
  /** Path on which storage (EmptyDir/Storage) will be mounted and into which, if given in SourceArchiveURL, will be extracted the archive */
  contentPath?: Maybe<Scalars['String']>;
  /** Whether forcing the container working directory to be the same as the contentPath (or default mydrive path if not specified) */
  enforceWorkdir?: Maybe<Scalars['Boolean']>;
  /** URL from which GET the archive to be extracted into ContentPath */
  sourceArchiveURL?: Maybe<Scalars['String']>;
  /** Arguments to be passed to the application container on startup */
  startupArgs?: Maybe<Array<Maybe<Scalars['String']>>>;
};

/** Optional urls for advanced integration features. */
export type CustomizationUrls = {
  __typename?: 'CustomizationUrls';
  /** URL to which POST an archive with the contents found (at instance termination) in Template.ContainerStartupOptions.ContentPath. */
  contentDestination?: Maybe<Scalars['String']>;
  /** URL from which GET the archive to be extracted into Template.ContainerStartupOptions.ContentPath. This field, if set, OVERRIDES Template.ContainerStartupOptions.SourceArchiveURL. */
  contentOrigin?: Maybe<Scalars['String']>;
  /** URL which is periodically checked (with a GET request) to determine automatic instance shutdown. Should return any 2xx status code if the instance has to keep running, any 4xx otherwise. In case of 2xx response, it should output a JSON with a `deadline` field containing a ISO_8601 compliant date/time string of the expected instance termination time. See instautoctrl.StatusCheckResponse for exact definition. */
  statusCheck?: Maybe<Scalars['String']>;
};

/** Optional urls for advanced integration features. */
export type CustomizationUrlsInput = {
  /** URL to which POST an archive with the contents found (at instance termination) in Template.ContainerStartupOptions.ContentPath. */
  contentDestination?: Maybe<Scalars['String']>;
  /** URL from which GET the archive to be extracted into Template.ContainerStartupOptions.ContentPath. This field, if set, OVERRIDES Template.ContainerStartupOptions.SourceArchiveURL. */
  contentOrigin?: Maybe<Scalars['String']>;
  /** URL which is periodically checked (with a GET request) to determine automatic instance shutdown. Should return any 2xx status code if the instance has to keep running, any 4xx otherwise. In case of 2xx response, it should output a JSON with a `deadline` field containing a ISO_8601 compliant date/time string of the expected instance termination time. See instautoctrl.StatusCheckResponse for exact definition. */
  statusCheck?: Maybe<Scalars['String']>;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItem = {
  __typename?: 'EnvironmentListListItem';
  /** Options to customize container startup */
  containerStartupOptions?: Maybe<ContainerStartupOptions>;
  /** For VNC based containers, hide the noVNC control bar when true */
  disableControls?: Maybe<Scalars['Boolean']>;
  /** The type of environment to be instantiated, among VirtualMachine, Container, CloudVM and Standalone. */
  environmentType?: Maybe<EnvironmentType>;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: Maybe<Scalars['Boolean']>;
  /** The VM or container to be started when instantiating the environment. */
  image?: Maybe<Scalars['String']>;
  /** The mode associated with the environment (Standard, Exam, Exercise) */
  mode?: Maybe<Mode>;
  /** Whether the instance has to have the user's MyDrive volume */
  mountMyDriveVolume?: Maybe<Scalars['Boolean']>;
  /** The name identifying the specific environment. */
  name?: Maybe<Scalars['String']>;
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: Maybe<Scalars['Boolean']>;
  /** The amount of computational resources associated with the environment. */
  resources?: Maybe<Resources>;
  /** Whether the environment needs the URL Rewrite or not. */
  rewriteURL?: Maybe<Scalars['Boolean']>;
  /** Name of the storage class to be used for the persistent volume (when needed) */
  storageClassName?: Maybe<Scalars['String']>;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItemInput = {
  /** Options to customize container startup */
  containerStartupOptions?: Maybe<ContainerStartupOptionsInput>;
  /** For VNC based containers, hide the noVNC control bar when true */
  disableControls?: Maybe<Scalars['Boolean']>;
  /** The type of environment to be instantiated, among VirtualMachine, Container, CloudVM and Standalone. */
  environmentType: EnvironmentType;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: Maybe<Scalars['Boolean']>;
  /** The VM or container to be started when instantiating the environment. */
  image: Scalars['String'];
  /** The mode associated with the environment (Standard, Exam, Exercise) */
  mode?: Maybe<Mode>;
  /** Whether the instance has to have the user's MyDrive volume */
  mountMyDriveVolume: Scalars['Boolean'];
  /** The name identifying the specific environment. */
  name: Scalars['String'];
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: Maybe<Scalars['Boolean']>;
  /** The amount of computational resources associated with the environment. */
  resources: ResourcesInput;
  /** Whether the environment needs the URL Rewrite or not. */
  rewriteURL?: Maybe<Scalars['Boolean']>;
  /** Name of the storage class to be used for the persistent volume (when needed) */
  storageClassName?: Maybe<Scalars['String']>;
};

/** Environment represents the reference to the environment to be snapshotted, in case more are associated with the same Instance. If not specified, the first available environment is considered. */
export type EnvironmentRef = {
  __typename?: 'EnvironmentRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

/** Environment represents the reference to the environment to be snapshotted, in case more are associated with the same Instance. If not specified, the first available environment is considered. */
export type EnvironmentRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

export enum EnvironmentType {
  VirtualMachine = 'VirtualMachine',
  Container = 'Container',
  CloudVm = 'CloudVM',
  Standalone = 'Standalone'
}

/** ImageListItem describes a single VM image. */
export type ImagesListItem = {
  __typename?: 'ImagesListItem';
  /** The name identifying a single image. */
  name?: Maybe<Scalars['String']>;
  /** The list of versions the image is available in. */
  versions?: Maybe<Array<Maybe<Scalars['String']>>>;
};

/** ImageListItem describes a single VM image. */
export type ImagesListItemInput = {
  /** The name identifying a single image. */
  name: Scalars['String'];
  /** The list of versions the image is available in. */
  versions: Array<Maybe<Scalars['String']>>;
};

/** Instance is the reference to the persistent VM instance to be snapshotted. The instance should not be running, otherwise it won't be possible to steal the volume and extract its content. */
export type InstanceRef = {
  __typename?: 'InstanceRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

/** Instance is the reference to the persistent VM instance to be snapshotted. The instance should not be running, otherwise it won't be possible to steal the volume and extract its content. */
export type InstanceRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

/** DeleteOptions may be provided when deleting an API object. */
export type IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** When present, indicates that modifications should not be persisted. An invalid or unrecognized dryRun directive will result in an error response and no further processing of the request. Valid values are: - All: all dry run stages will be processed */
  dryRun?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The duration in seconds before the object should be deleted. Value must be non-negative integer. The value zero indicates delete immediately. If this value is nil, the default grace period for the specified type will be used. Defaults to a per object value if not specified. zero means delete immediately. */
  gracePeriodSeconds?: Maybe<Scalars['Float']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** Deprecated: please use the PropagationPolicy, this field will be deprecated in 1.7. Should the dependent objects be orphaned. If true/false, the "orphan" finalizer will be added to/removed from the object's finalizers list. Either this field or PropagationPolicy may be set, but not both. */
  orphanDependents?: Maybe<Scalars['Boolean']>;
  /** Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out. */
  preconditions?: Maybe<IoK8sApimachineryPkgApisMetaV1PreconditionsInput>;
  /** Whether and how garbage collection will be performed. Either this field or OrphanDependents may be set, but not both. The default policy is decided by the existing finalizer set in the metadata.finalizers and the resource-specific default policy. Acceptable values are: 'Orphan' - orphan the dependents; 'Background' - allow the garbage collector to delete the dependents in the background; 'Foreground' - a cascading policy that deletes all dependents in the foreground. */
  propagationPolicy?: Maybe<Scalars['String']>;
};

/** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
export type IoK8sApimachineryPkgApisMetaV1ListMeta = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ListMeta';
  /** continue may be set if the user set a limit on the number of items returned, and indicates that the server has more data available. The value is opaque and may be used to issue another request to the endpoint that served this list to retrieve the next set of available objects. Continuing a consistent list may not be possible if the server configuration has changed or more than a few minutes have passed. The resourceVersion field returned when using this continue value will be identical to the value in the first response, unless you have received this token from an error message. */
  continue?: Maybe<Scalars['String']>;
  /** remainingItemCount is the number of subsequent items in the list which are not included in this list response. If the list request contained label or field selectors, then the number of remaining items is unknown and the field will be left unset and omitted during serialization. If the list is complete (either because it is not chunking or because this is the last chunk), then there are no more remaining items and this field will be left unset and omitted during serialization. Servers older than v1.15 do not set this field. The intended use of the remainingItemCount is *estimating* the size of a collection. Clients should not rely on the remainingItemCount to be set or to be exact. */
  remainingItemCount?: Maybe<Scalars['Float']>;
  /** String that identifies the server's internal version of this object that can be used by clients to determine when objects have changed. Value must be treated as opaque by clients and passed unmodified back to the server. Populated by the system. Read-only. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency */
  resourceVersion?: Maybe<Scalars['String']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: Maybe<Scalars['String']>;
};

/** ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource that the fieldset applies to. */
export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry';
  /** APIVersion defines the version of this resource that this field set applies to. The format is "group/version" just like the top-level APIVersion field. It is necessary to track the version of a field set because it cannot be automatically converted. */
  apiVersion?: Maybe<Scalars['String']>;
  /** FieldsType is the discriminator for the different fields format and version. There is currently only one possible value: "FieldsV1" */
  fieldsType?: Maybe<Scalars['String']>;
  /**
   * FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.
   *
   * Each key is either a '.' representing the field itself, and will always map to an empty set, or a string representing a sub-field or item. The string will follow one of these four formats: 'f:<name>', where <name> is the name of a field in a struct, or key in a map 'v:<value>', where <value> is the exact json formatted value of a list item 'i:<index>', where <index> is position of a item in a list 'k:<keys>', where <keys> is a map of  a list item's key fields to their unique values If a key maps to an empty Fields value, the field that key represents is part of the set.
   *
   * The exact format is defined in sigs.k8s.io/structured-merge-diff
   */
  fieldsV1?: Maybe<Scalars['String']>;
  /** Manager is an identifier of the workflow managing these fields. */
  manager?: Maybe<Scalars['String']>;
  /** Operation is the type of operation which lead to this ManagedFieldsEntry being created. The only valid values for this field are 'Apply' and 'Update'. */
  operation?: Maybe<Scalars['String']>;
  /** Subresource is the name of the subresource used to update that object, or empty string if the object was updated through the main resource. The value of this field is used to distinguish between managers, even if they share the same name. For example, a status update will be distinct from a regular update using the same manager name. Note that the APIVersion field is not related to the Subresource field and it always corresponds to the version of the main resource. */
  subresource?: Maybe<Scalars['String']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  time?: Maybe<Scalars['String']>;
};

/** ManagedFieldsEntry is a workflow-id, a FieldSet and the group version of the resource that the fieldset applies to. */
export type IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput = {
  /** APIVersion defines the version of this resource that this field set applies to. The format is "group/version" just like the top-level APIVersion field. It is necessary to track the version of a field set because it cannot be automatically converted. */
  apiVersion?: Maybe<Scalars['String']>;
  /** FieldsType is the discriminator for the different fields format and version. There is currently only one possible value: "FieldsV1" */
  fieldsType?: Maybe<Scalars['String']>;
  /**
   * FieldsV1 stores a set of fields in a data structure like a Trie, in JSON format.
   *
   * Each key is either a '.' representing the field itself, and will always map to an empty set, or a string representing a sub-field or item. The string will follow one of these four formats: 'f:<name>', where <name> is the name of a field in a struct, or key in a map 'v:<value>', where <value> is the exact json formatted value of a list item 'i:<index>', where <index> is position of a item in a list 'k:<keys>', where <keys> is a map of  a list item's key fields to their unique values If a key maps to an empty Fields value, the field that key represents is part of the set.
   *
   * The exact format is defined in sigs.k8s.io/structured-merge-diff
   */
  fieldsV1?: Maybe<Scalars['String']>;
  /** Manager is an identifier of the workflow managing these fields. */
  manager?: Maybe<Scalars['String']>;
  /** Operation is the type of operation which lead to this ManagedFieldsEntry being created. The only valid values for this field are 'Apply' and 'Update'. */
  operation?: Maybe<Scalars['String']>;
  /** Subresource is the name of the subresource used to update that object, or empty string if the object was updated through the main resource. The value of this field is used to distinguish between managers, even if they share the same name. For example, a status update will be distinct from a regular update using the same manager name. Note that the APIVersion field is not related to the Subresource field and it always corresponds to the version of the main resource. */
  subresource?: Maybe<Scalars['String']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  time?: Maybe<Scalars['String']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMeta = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta';
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations */
  annotations?: Maybe<Scalars['JSON']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  creationTimestamp?: Maybe<Scalars['String']>;
  /** Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. */
  deletionGracePeriodSeconds?: Maybe<Scalars['Float']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  deletionTimestamp?: Maybe<Scalars['String']>;
  /** Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list. */
  finalizers?: Maybe<Array<Maybe<Scalars['String']>>>;
  /**
   * GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.
   *
   * If this field is specified and the generated name exists, the server will return a 409.
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: Maybe<Scalars['String']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: Maybe<Scalars['Float']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels */
  labels?: Maybe<Scalars['JSON']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry>>>;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name?: Maybe<Scalars['String']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces
   */
  namespace?: Maybe<Scalars['String']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReference>>>;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: Maybe<Scalars['String']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: Maybe<Scalars['String']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids
   */
  uid?: Maybe<Scalars['String']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMetaInput = {
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations */
  annotations?: Maybe<Scalars['JSON']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  creationTimestamp?: Maybe<Scalars['String']>;
  /** Number of seconds allowed for this object to gracefully terminate before it will be removed from the system. Only set when deletionTimestamp is also set. May only be shortened. Read-only. */
  deletionGracePeriodSeconds?: Maybe<Scalars['Float']>;
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  deletionTimestamp?: Maybe<Scalars['String']>;
  /** Must be empty before the object is deleted from the registry. Each entry is an identifier for the responsible component that will remove the entry from the list. If the deletionTimestamp of the object is non-nil, entries in this list can only be removed. Finalizers may be processed and removed in any order.  Order is NOT enforced because it introduces significant risk of stuck finalizers. finalizers is a shared field, any actor with permission can reorder it. If the finalizer list is processed in order, then this can lead to a situation in which the component responsible for the first finalizer in the list is waiting for a signal (field value, external system, or other) produced by a component responsible for a finalizer later in the list, resulting in a deadlock. Without enforced ordering finalizers are free to order amongst themselves and are not vulnerable to ordering changes in the list. */
  finalizers?: Maybe<Array<Maybe<Scalars['String']>>>;
  /**
   * GenerateName is an optional prefix, used by the server, to generate a unique name ONLY IF the Name field has not been provided. If this field is used, the name returned to the client will be different than the name passed. This value will also be combined with a unique suffix. The provided value has the same validation rules as the Name field, and may be truncated by the length of the suffix required to make the value unique on the server.
   *
   * If this field is specified and the generated name exists, the server will return a 409.
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: Maybe<Scalars['String']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: Maybe<Scalars['Float']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels */
  labels?: Maybe<Scalars['JSON']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput>>>;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name?: Maybe<Scalars['String']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces
   */
  namespace?: Maybe<Scalars['String']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceInput>>>;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: Maybe<Scalars['String']>;
  /** Deprecated: selfLink is a legacy read-only field that is no longer populated by the system. */
  selfLink?: Maybe<Scalars['String']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids
   */
  uid?: Maybe<Scalars['String']>;
};

/** OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field. */
export type IoK8sApimachineryPkgApisMetaV1OwnerReference = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1OwnerReference';
  /** API version of the referent. */
  apiVersion?: Maybe<Scalars['String']>;
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. See https://kubernetes.io/docs/concepts/architecture/garbage-collection/#foreground-deletion for how the garbage collector interacts with this field and enforces the foreground deletion. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
  blockOwnerDeletion?: Maybe<Scalars['Boolean']>;
  /** If true, this reference points to the managing controller. */
  controller?: Maybe<Scalars['Boolean']>;
  /** Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name?: Maybe<Scalars['String']>;
  /** UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids */
  uid?: Maybe<Scalars['String']>;
};

/** OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field. */
export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceInput = {
  /** API version of the referent. */
  apiVersion: Scalars['String'];
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. See https://kubernetes.io/docs/concepts/architecture/garbage-collection/#foreground-deletion for how the garbage collector interacts with this field and enforces the foreground deletion. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
  blockOwnerDeletion?: Maybe<Scalars['Boolean']>;
  /** If true, this reference points to the managing controller. */
  controller?: Maybe<Scalars['Boolean']>;
  /** Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind: Scalars['String'];
  /** Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name: Scalars['String'];
  /** UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids */
  uid: Scalars['String'];
};

/** Preconditions must be fulfilled before an operation (update, delete, etc.) is carried out. */
export type IoK8sApimachineryPkgApisMetaV1PreconditionsInput = {
  /** Specifies the target ResourceVersion */
  resourceVersion?: Maybe<Scalars['String']>;
  /** Specifies the target UID. */
  uid?: Maybe<Scalars['String']>;
};

/** Status is a return value for calls that don't return other objects. */
export type IoK8sApimachineryPkgApisMetaV1Status = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1Status';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Suggested HTTP return code for this status, 0 if not set. */
  code?: Maybe<Scalars['Int']>;
  /** StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined. */
  details?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusDetails>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** A human-readable description of the status of this operation. */
  message?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
  /** A machine-readable description of why this operation is in the "Failure" status. If this value is empty there is no information available. A Reason clarifies an HTTP status code but does not override it. */
  reason?: Maybe<Scalars['String']>;
  /** Status of the operation. One of: "Success" or "Failure". More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status */
  status?: Maybe<Scalars['String']>;
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
  field?: Maybe<Scalars['String']>;
  /** A human-readable description of the cause of the error.  This field may be presented as-is to a reader. */
  message?: Maybe<Scalars['String']>;
  /** A machine-readable description of the cause of the error. If this value is empty there is no information available. */
  reason?: Maybe<Scalars['String']>;
};

/** StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined. */
export type IoK8sApimachineryPkgApisMetaV1StatusDetails = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusDetails';
  /** The Causes array includes more details associated with the StatusReason failure. Not all StatusReasons may provide detailed causes. */
  causes?: Maybe<Array<Maybe<IoK8sApimachineryPkgApisMetaV1StatusCause>>>;
  /** The group attribute of the resource associated with the status StatusReason. */
  group?: Maybe<Scalars['String']>;
  /** The kind attribute of the resource associated with the status StatusReason. On some operations may differ from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** The name attribute of the resource associated with the status StatusReason (when there is a single name which can be described). */
  name?: Maybe<Scalars['String']>;
  /** If specified, the time in seconds before the operation should be retried. Some errors may indicate the client must take an alternate action - for those errors this field may indicate how long to wait before taking the alternate action. */
  retryAfterSeconds?: Maybe<Scalars['Int']>;
  /** UID of the resource. (when there is a single resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids */
  uid?: Maybe<Scalars['String']>;
};

/** ImageList describes the available VM images in the CrownLabs registry. */
export type ItPolitoCrownlabsV1alpha1ImageList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** ImageListSpec is the specification of the desired state of the ImageList. */
  spec?: Maybe<Spec>;
  /** ImageListStatus reflects the most recently observed status of the ImageList. */
  status?: Maybe<Scalars['String']>;
};

/** ImageList describes the available VM images in the CrownLabs registry. */
export type ItPolitoCrownlabsV1alpha1ImageListInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** ImageListSpec is the specification of the desired state of the ImageList. */
  spec?: Maybe<SpecInput>;
  /** ImageListStatus reflects the most recently observed status of the ImageList. */
  status?: Maybe<Scalars['String']>;
};

/** ImageListList is a list of ImageList */
export type ItPolitoCrownlabsV1alpha1ImageListList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of imagelists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1ImageList>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1ImageListUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha1ImageList>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1Workspace = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: Maybe<Spec2>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: Maybe<Status>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1WorkspaceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: Maybe<Spec2Input>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: Maybe<StatusInput>;
};

/** WorkspaceList is a list of Workspace */
export type ItPolitoCrownlabsV1alpha1WorkspaceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of workspaces. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1Workspace>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha1WorkspaceUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2Instance = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: Maybe<Spec3>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: Maybe<Status2>;
};

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2InstanceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: Maybe<Spec3Input>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: Maybe<Status2Input>;
};

/** InstanceList is a list of Instance */
export type ItPolitoCrownlabsV1alpha2InstanceList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of instances. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2Instance>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

/** InstanceSnapshot is the Schema for the instancesnapshots API. */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshot = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshot';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: Maybe<Spec4>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: Maybe<Status3>;
};

/** InstanceSnapshot is the Schema for the instancesnapshots API. */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshotInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: Maybe<Spec4Input>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: Maybe<Status3Input>;
};

/** InstanceSnapshotList is a list of InstanceSnapshot */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshotList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2InstanceSnapshotList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of instancesnapshots. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshot>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
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

/** Template describes the template of a CrownLabs environment to be instantiated. */
export type ItPolitoCrownlabsV1alpha2Template = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Template';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** TemplateSpec is the specification of the desired state of the Template. */
  spec?: Maybe<Spec5>;
  /** TemplateStatus reflects the most recently observed status of the Template. */
  status?: Maybe<Scalars['String']>;
};

/** Template describes the template of a CrownLabs environment to be instantiated. */
export type ItPolitoCrownlabsV1alpha2TemplateInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** TemplateSpec is the specification of the desired state of the Template. */
  spec?: Maybe<Spec5Input>;
  /** TemplateStatus reflects the most recently observed status of the Template. */
  status?: Maybe<Scalars['String']>;
};

/** TemplateList is a list of Template */
export type ItPolitoCrownlabsV1alpha2TemplateList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of templates. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2Template>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2TemplateUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha2Tenant = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Tenant';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMeta>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: Maybe<Spec6>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: Maybe<Status4>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha2TenantInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaInput>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: Maybe<Spec6Input>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: Maybe<Status4Input>;
};

/** TenantList is a list of Tenant */
export type ItPolitoCrownlabsV1alpha2TenantList = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TenantList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of tenants. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha2Tenant>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

export type ItPolitoCrownlabsV1alpha2TenantUpdate = {
  __typename?: 'ItPolitoCrownlabsV1alpha2TenantUpdate';
  updateType?: Maybe<UpdateType>;
  payload?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
};


export enum Mode {
  Standard = 'Standard',
  Exam = 'Exam',
  Exercise = 'Exercise'
}

/** The start of any mutation */
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


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha1ImageListArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};


/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2TenantArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
};


/** The start of any mutation */
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


/** The start of any mutation */
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


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
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


/** The start of any mutation */
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


/** The start of any mutation */
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


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha2CollectionTenantArgs = {
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


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha2TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsInput?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsInput>;
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  force?: Maybe<Scalars['Boolean']>;
  applicationApplyPatchYamlInput: Scalars['String'];
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: ItPolitoCrownlabsV1alpha1ImageListInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
};


/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  fieldValidation?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TenantInput: ItPolitoCrownlabsV1alpha2TenantInput;
};

/** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
export type Namespace = {
  __typename?: 'Namespace';
  /** Whether the creation succeeded or not. */
  created?: Maybe<Scalars['Boolean']>;
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
export type NamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
export type PersonalNamespace = {
  __typename?: 'PersonalNamespace';
  /** Whether the creation succeeded or not. */
  created?: Maybe<Scalars['Boolean']>;
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
export type PersonalNamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

export enum Phase {
  Importing = 'Importing',
  Starting = 'Starting',
  ResourceQuotaExceeded = 'ResourceQuotaExceeded',
  Running = 'Running',
  Ready = 'Ready',
  Stopping = 'Stopping',
  Off = 'Off',
  Failed = 'Failed',
  CreationLoopBackoff = 'CreationLoopBackoff'
}

/** The start of any query */
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
};


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
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


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
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


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha2InstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
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


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha2InstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
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


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha2TemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
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


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha2TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryItPolitoCrownlabsV1alpha2TenantListArgs = {
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


/** The start of any query */
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


/** The start of any query */
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


/** The start of any query */
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


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};


/** The start of any query */
export type QueryReadCrownlabsPolitoItV1alpha2TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

/** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
export type Quota = {
  __typename?: 'Quota';
  /** The maximum amount of CPU required by this Workspace. */
  cpu?: Maybe<Scalars['String']>;
  /** The maximum number of concurrent instances required by this Workspace. */
  instances?: Maybe<Scalars['Int']>;
  /** The maximum amount of RAM memory required by this Workspace. */
  memory?: Maybe<Scalars['String']>;
};

/** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
export type Quota2 = {
  __typename?: 'Quota2';
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu?: Maybe<Scalars['String']>;
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances?: Maybe<Scalars['Int']>;
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory?: Maybe<Scalars['String']>;
};

/** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
export type Quota2Input = {
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['String'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['String'];
};

/** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
export type Quota3 = {
  __typename?: 'Quota3';
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu?: Maybe<Scalars['String']>;
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances?: Maybe<Scalars['Int']>;
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory?: Maybe<Scalars['String']>;
};

/** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
export type Quota3Input = {
  /** The maximum amount of CPU which can be used by this Tenant. */
  cpu: Scalars['String'];
  /** The maximum number of concurrent instances which can be created by this Tenant. */
  instances: Scalars['Int'];
  /** The maximum amount of RAM memory which can be used by this Tenant. */
  memory: Scalars['String'];
};

/** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
export type QuotaInput = {
  /** The maximum amount of CPU required by this Workspace. */
  cpu: Scalars['String'];
  /** The maximum number of concurrent instances required by this Workspace. */
  instances: Scalars['Int'];
  /** The maximum amount of RAM memory required by this Workspace. */
  memory: Scalars['String'];
};

/** The amount of computational resources associated with the environment. */
export type Resources = {
  __typename?: 'Resources';
  /** The maximum number of CPU cores made available to the environment (at least 1 core). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu?: Maybe<Scalars['Int']>;
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent or container-based environments, while it is silently ignored in the other cases. In case of containers, when this field is not specified, an emptyDir will be attached to the pod but this could result in data loss whenever the pod dies. */
  disk?: Maybe<Scalars['String']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory?: Maybe<Scalars['String']>;
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage?: Maybe<Scalars['Int']>;
};

/** The amount of computational resources associated with the environment. */
export type ResourcesInput = {
  /** The maximum number of CPU cores made available to the environment (at least 1 core). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu: Scalars['Int'];
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent or container-based environments, while it is silently ignored in the other cases. In case of containers, when this field is not specified, an emptyDir will be attached to the pod but this could result in data loss whenever the pod dies. */
  disk?: Maybe<Scalars['String']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory: Scalars['String'];
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage: Scalars['Int'];
};

export enum Role {
  Manager = 'manager',
  User = 'user',
  Candidate = 'candidate'
}

/** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
export type SandboxNamespace = {
  __typename?: 'SandboxNamespace';
  /** Whether the creation succeeded or not. */
  created?: Maybe<Scalars['Boolean']>;
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
export type SandboxNamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** ImageListSpec is the specification of the desired state of the ImageList. */
export type Spec = {
  __typename?: 'Spec';
  /** The list of VM images currently available in CrownLabs. */
  images?: Maybe<Array<Maybe<ImagesListItem>>>;
  /** The host name that can be used to access the registry. */
  registryName?: Maybe<Scalars['String']>;
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2 = {
  __typename?: 'Spec2';
  /** AutoEnroll capability definition. If omitted, no autoenroll features will be added. */
  autoEnroll?: Maybe<AutoEnroll>;
  /** The human-readable name of the Workspace. */
  prettyName?: Maybe<Scalars['String']>;
  /** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
  quota?: Maybe<Quota>;
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2Input = {
  /** AutoEnroll capability definition. If omitted, no autoenroll features will be added. */
  autoEnroll?: Maybe<AutoEnroll>;
  /** The human-readable name of the Workspace. */
  prettyName: Scalars['String'];
  /** The amount of resources associated with this workspace, and inherited by enrolled tenants. */
  quota: QuotaInput;
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3 = {
  __typename?: 'Spec3';
  /** Optional urls for advanced integration features. */
  customizationUrls?: Maybe<CustomizationUrls>;
  /** Custom name the user can assign and change at any time in order to more easily identify the instance. */
  prettyName?: Maybe<Scalars['String']>;
  /** Whether the current instance is running or not. The meaning of this flag is different depending on whether the instance refers to a persistent environment or not. If the first case, it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. Differently, if the environment is not persistent, it only tears down the exposition objects, making the instance effectively unreachable from outside the cluster, but allowing the subsequent recreation without data loss. */
  running?: Maybe<Scalars['Boolean']>;
  /** The reference to the Template to be instantiated. */
  templateCrownlabsPolitoItTemplateRef?: Maybe<TemplateCrownlabsPolitoItTemplateRef>;
  /** The reference to the Tenant which owns the Instance object. */
  tenantCrownlabsPolitoItTenantRef?: Maybe<TenantCrownlabsPolitoItTenantRef>;
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3Input = {
  /** Optional urls for advanced integration features. */
  customizationUrls?: Maybe<CustomizationUrlsInput>;
  /** Custom name the user can assign and change at any time in order to more easily identify the instance. */
  prettyName?: Maybe<Scalars['String']>;
  /** Whether the current instance is running or not. The meaning of this flag is different depending on whether the instance refers to a persistent environment or not. If the first case, it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. Differently, if the environment is not persistent, it only tears down the exposition objects, making the instance effectively unreachable from outside the cluster, but allowing the subsequent recreation without data loss. */
  running?: Maybe<Scalars['Boolean']>;
  /** The reference to the Template to be instantiated. */
  templateCrownlabsPolitoItTemplateRef: TemplateCrownlabsPolitoItTemplateRefInput;
  /** The reference to the Tenant which owns the Instance object. */
  tenantCrownlabsPolitoItTenantRef: TenantCrownlabsPolitoItTenantRefInput;
};

/** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
export type Spec4 = {
  __typename?: 'Spec4';
  /** Environment represents the reference to the environment to be snapshotted, in case more are associated with the same Instance. If not specified, the first available environment is considered. */
  environmentRef?: Maybe<EnvironmentRef>;
  /** ImageName is the name of the image to pushed in the docker registry. */
  imageName?: Maybe<Scalars['String']>;
  /** Instance is the reference to the persistent VM instance to be snapshotted. The instance should not be running, otherwise it won't be possible to steal the volume and extract its content. */
  instanceRef?: Maybe<InstanceRef>;
};

/** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
export type Spec4Input = {
  /** Environment represents the reference to the environment to be snapshotted, in case more are associated with the same Instance. If not specified, the first available environment is considered. */
  environmentRef?: Maybe<EnvironmentRefInput>;
  /** ImageName is the name of the image to pushed in the docker registry. */
  imageName: Scalars['String'];
  /** Instance is the reference to the persistent VM instance to be snapshotted. The instance should not be running, otherwise it won't be possible to steal the volume and extract its content. */
  instanceRef: InstanceRefInput;
};

/** TemplateSpec is the specification of the desired state of the Template. */
export type Spec5 = {
  __typename?: 'Spec5';
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. If set to "never", the instance will not be automatically terminated. */
  deleteAfter?: Maybe<Scalars['String']>;
  /** A textual description of the Template. */
  description?: Maybe<Scalars['String']>;
  /** The list of environments (i.e. VMs or containers) that compose the Template. */
  environmentList?: Maybe<Array<Maybe<EnvironmentListListItem>>>;
  /** The human-readable name of the Template. */
  prettyName?: Maybe<Scalars['String']>;
  /** The reference to the Workspace this Template belongs to. */
  workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<WorkspaceCrownlabsPolitoItWorkspaceRef>;
};

/** TemplateSpec is the specification of the desired state of the Template. */
export type Spec5Input = {
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. If set to "never", the instance will not be automatically terminated. */
  deleteAfter?: Maybe<Scalars['String']>;
  /** A textual description of the Template. */
  description: Scalars['String'];
  /** The list of environments (i.e. VMs or containers) that compose the Template. */
  environmentList: Array<Maybe<EnvironmentListListItemInput>>;
  /** The human-readable name of the Template. */
  prettyName: Scalars['String'];
  /** The reference to the Workspace this Template belongs to. */
  workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<WorkspaceCrownlabsPolitoItWorkspaceRefInput>;
};

/** TenantSpec is the specification of the desired state of the Tenant. */
export type Spec6 = {
  __typename?: 'Spec6';
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: Maybe<Scalars['Boolean']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email?: Maybe<Scalars['String']>;
  /** The first name of the Tenant. */
  firstName?: Maybe<Scalars['String']>;
  /** The last login timestamp. */
  lastLogin?: Maybe<Scalars['String']>;
  /** The last name of the Tenant. */
  lastName?: Maybe<Scalars['String']>;
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
  quota?: Maybe<Quota2>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: Maybe<Array<Maybe<WorkspacesListItem>>>;
};

/** TenantSpec is the specification of the desired state of the Tenant. */
export type Spec6Input = {
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: Maybe<Scalars['Boolean']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email: Scalars['String'];
  /** The first name of the Tenant. */
  firstName: Scalars['String'];
  /** The last login timestamp. */
  lastLogin?: Maybe<Scalars['String']>;
  /** The last name of the Tenant. */
  lastName: Scalars['String'];
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The amount of resources associated with this Tenant, if defined it overrides the one computed from the workspaces the tenant is enrolled in. */
  quota?: Maybe<Quota2Input>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: Maybe<Array<Maybe<WorkspacesListItemInput>>>;
};

/** ImageListSpec is the specification of the desired state of the ImageList. */
export type SpecInput = {
  /** The list of VM images currently available in CrownLabs. */
  images: Array<Maybe<ImagesListItemInput>>;
  /** The host name that can be used to access the registry. */
  registryName: Scalars['String'];
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type Status = {
  __typename?: 'Status';
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: Maybe<Namespace>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: Maybe<Scalars['JSON']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status2 = {
  __typename?: 'Status2';
  /** Timestamps of the Instance automation phases (check, termination and submission). */
  automation?: Maybe<Automation>;
  /** The amount of time the Instance required to become ready for the first time upon creation. */
  initialReadyTime?: Maybe<Scalars['String']>;
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: Maybe<Scalars['String']>;
  /** The URL where it is possible to access the persistent drive associated with the instance (in case of container-based environments) */
  myDriveUrl?: Maybe<Scalars['String']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: Maybe<Phase>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: Maybe<Scalars['String']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status2Input = {
  /** Timestamps of the Instance automation phases (check, termination and submission). */
  automation?: Maybe<AutomationInput>;
  /** The amount of time the Instance required to become ready for the first time upon creation. */
  initialReadyTime?: Maybe<Scalars['String']>;
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: Maybe<Scalars['String']>;
  /** The URL where it is possible to access the persistent drive associated with the instance (in case of container-based environments) */
  myDriveUrl?: Maybe<Scalars['String']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: Maybe<Phase>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: Maybe<Scalars['String']>;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status3 = {
  __typename?: 'Status3';
  /** Phase represents the current state of the Instance Snapshot. */
  phase?: Maybe<Scalars['String']>;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status3Input = {
  /** Phase represents the current state of the Instance Snapshot. */
  phase: Scalars['String'];
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type Status4 = {
  __typename?: 'Status4';
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
  personalNamespace?: Maybe<PersonalNamespace>;
  /** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
  quota?: Maybe<Quota3>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. Will be set to true even when personal workspace is intentionally deleted. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace?: Maybe<SandboxNamespace>;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions?: Maybe<Scalars['JSON']>;
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type Status4Input = {
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces: Array<Maybe<Scalars['String']>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quota, network policies, ...) created by the tenant-operator. */
  personalNamespace: PersonalNamespaceInput;
  /** The amount of resources associated with this Tenant, either inherited from the Workspaces in which he/she is enrolled, or manually overridden. */
  quota?: Maybe<Quota3Input>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. Will be set to true even when personal workspace is intentionally deleted. */
  ready: Scalars['Boolean'];
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace: SandboxNamespaceInput;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions: Scalars['JSON'];
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type StatusInput = {
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: Maybe<NamespaceInput>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: Maybe<Scalars['JSON']>;
};

export type Subscription = {
  __typename?: 'Subscription';
  itPolitoCrownlabsV1alpha2InstanceUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2InstanceLabelsUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceSnapshotUpdate>;
  itPolitoCrownlabsV1alpha2TemplateUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TemplateUpdate>;
  itPolitoCrownlabsV1alpha2TenantUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TenantUpdate>;
  itPolitoCrownlabsV1alpha1WorkspaceUpdate?: Maybe<ItPolitoCrownlabsV1alpha1WorkspaceUpdate>;
  itPolitoCrownlabsV1alpha1ImageListUpdate?: Maybe<ItPolitoCrownlabsV1alpha1ImageListUpdate>;
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceLabelsUpdateArgs = {
  labelSelector?: Maybe<Scalars['String']>;
};


export type SubscriptionItPolitoCrownlabsV1alpha2InstanceSnapshotUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2TemplateUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};


export type SubscriptionItPolitoCrownlabsV1alpha2TenantUpdateArgs = {
  name: Scalars['String'];
  namespace?: Maybe<Scalars['String']>;
};


export type SubscriptionItPolitoCrownlabsV1alpha1WorkspaceUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};


export type SubscriptionItPolitoCrownlabsV1alpha1ImageListUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

/** The reference to the Template to be instantiated. */
export type TemplateCrownlabsPolitoItTemplateRef = {
  __typename?: 'TemplateCrownlabsPolitoItTemplateRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
  templateWrapper?: Maybe<TemplateWrapper>;
};

/** The reference to the Template to be instantiated. */
export type TemplateCrownlabsPolitoItTemplateRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

export type TemplateWrapper = {
  __typename?: 'TemplateWrapper';
  itPolitoCrownlabsV1alpha2Template?: Maybe<ItPolitoCrownlabsV1alpha2Template>;
};

/** The reference to the Tenant which owns the Instance object. */
export type TenantCrownlabsPolitoItTenantRef = {
  __typename?: 'TenantCrownlabsPolitoItTenantRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
  tenantV1alpha2Wrapper?: Maybe<TenantV1alpha2Wrapper>;
};

/** The reference to the Tenant which owns the Instance object. */
export type TenantCrownlabsPolitoItTenantRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

export type TenantV1alpha2Wrapper = {
  __typename?: 'TenantV1alpha2Wrapper';
  itPolitoCrownlabsV1alpha2Tenant?: Maybe<ItPolitoCrownlabsV1alpha2Tenant>;
};

export enum UpdateType {
  Added = 'ADDED',
  Modified = 'MODIFIED',
  Deleted = 'DELETED'
}

/** The reference to the Workspace this Template belongs to. */
export type WorkspaceCrownlabsPolitoItWorkspaceRef = {
  __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

/** The reference to the Workspace this Template belongs to. */
export type WorkspaceCrownlabsPolitoItWorkspaceRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

export type WorkspaceWrapperTenantV1alpha2 = {
  __typename?: 'WorkspaceWrapperTenantV1alpha2';
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItem = {
  __typename?: 'WorkspacesListItem';
  /** The Workspace the Tenant is subscribed to. */
  name?: Maybe<Scalars['String']>;
  /** The role of the Tenant in the context of the Workspace. */
  role?: Maybe<Role>;
  workspaceWrapperTenantV1alpha2?: Maybe<WorkspaceWrapperTenantV1alpha2>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItemInput = {
  /** The Workspace the Tenant is subscribed to. */
  name: Scalars['String'];
  /** The role of the Tenant in the context of the Workspace. */
  role: Role;
};

export type ApplyInstanceMutationVariables = Exact<{
  instanceId: Scalars['String'];
  tenantNamespace: Scalars['String'];
  patchJson: Scalars['String'];
  manager: Scalars['String'];
}>;


export type ApplyInstanceMutation = { __typename?: 'Mutation', applyInstance?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string> }> }> };

export type ApplyTemplateMutationVariables = Exact<{
  templateId: Scalars['String'];
  workspaceNamespace: Scalars['String'];
  patchJson: Scalars['String'];
  manager: Scalars['String'];
}>;


export type ApplyTemplateMutation = { __typename?: 'Mutation', applyTemplate?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', description?: Maybe<string>, name?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, resources?: Maybe<{ __typename?: 'Resources', cpu?: Maybe<number>, disk?: Maybe<string>, memory?: Maybe<string> }> }>>> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', id?: Maybe<string> }> }> };

export type ApplyTenantMutationVariables = Exact<{
  tenantId: Scalars['String'];
  patchJson: Scalars['String'];
  manager: Scalars['String'];
}>;


export type ApplyTenantMutation = { __typename?: 'Mutation', applyTenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec6', firstName?: Maybe<string>, lastName?: Maybe<string>, email?: Maybe<string>, lastLogin?: Maybe<string>, workspaces?: Maybe<Array<Maybe<{ __typename?: 'WorkspacesListItem', role?: Maybe<Role>, name?: Maybe<string> }>>> }> }> };

export type CreateInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String'];
  templateId: Scalars['String'];
  workspaceNamespace: Scalars['String'];
  tenantId: Scalars['String'];
  generateName?: Maybe<Scalars['String']>;
}>;


export type CreateInstanceMutation = { __typename?: 'Mutation', createdInstance?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string>, creationTimestamp?: Maybe<string>, labels?: Maybe<any> }>, status?: Maybe<{ __typename?: 'Status2', ip?: Maybe<string>, phase?: Maybe<Phase>, url?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string>, templateCrownlabsPolitoItTemplateRef?: Maybe<{ __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name?: Maybe<string>, namespace?: Maybe<string>, templateWrapper?: Maybe<{ __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, environmentType?: Maybe<EnvironmentType> }>>> }> }> }> }> }> }> };

export type CreateTemplateMutationVariables = Exact<{
  workspaceId: Scalars['String'];
  workspaceNamespace: Scalars['String'];
  templateName: Scalars['String'];
  descriptionTemplate: Scalars['String'];
  image: Scalars['String'];
  guiEnabled: Scalars['Boolean'];
  persistent: Scalars['Boolean'];
  mountMyDriveVolume: Scalars['Boolean'];
  resources: ResourcesInput;
  templateId?: Maybe<Scalars['String']>;
  environmentType: EnvironmentType;
}>;


export type CreateTemplateMutation = { __typename?: 'Mutation', createdTemplate?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, resources?: Maybe<{ __typename?: 'Resources', cpu?: Maybe<number>, disk?: Maybe<string>, memory?: Maybe<string> }> }>>> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string> }> }> };

export type DeleteInstanceMutationVariables = Exact<{
  tenantNamespace: Scalars['String'];
  instanceId: Scalars['String'];
}>;


export type DeleteInstanceMutation = { __typename?: 'Mutation', deletedInstance?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: Maybe<string> }> };

export type DeleteLabelSelectorInstancesMutationVariables = Exact<{
  tenantNamespace: Scalars['String'];
  labels?: Maybe<Scalars['String']>;
}>;


export type DeleteLabelSelectorInstancesMutation = { __typename?: 'Mutation', deleteLabelSelectorInstances?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: Maybe<string> }> };

export type DeleteTemplateMutationVariables = Exact<{
  workspaceNamespace: Scalars['String'];
  templateId: Scalars['String'];
}>;


export type DeleteTemplateMutation = { __typename?: 'Mutation', deletedTemplate?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1Status', kind?: Maybe<string> }> };

export type ImagesQueryVariables = Exact<{ [key: string]: never; }>;


export type ImagesQuery = { __typename?: 'Query', imageList?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList', images?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1ImageList', spec?: Maybe<{ __typename?: 'Spec', registryName?: Maybe<string>, images?: Maybe<Array<Maybe<{ __typename?: 'ImagesListItem', name?: Maybe<string>, versions?: Maybe<Array<Maybe<string>>> }>>> }> }>>> }> };

export type OwnedInstancesQueryVariables = Exact<{
  tenantNamespace: Scalars['String'];
}>;


export type OwnedInstancesQuery = { __typename?: 'Query', instanceList?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList', instances?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string>, creationTimestamp?: Maybe<string>, labels?: Maybe<any> }>, status?: Maybe<{ __typename?: 'Status2', ip?: Maybe<string>, phase?: Maybe<Phase>, url?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string>, templateCrownlabsPolitoItTemplateRef?: Maybe<{ __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name?: Maybe<string>, namespace?: Maybe<string>, templateWrapper?: Maybe<{ __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, environmentType?: Maybe<EnvironmentType> }>>> }> }> }> }> }> }>>> }> };

export type InstancesLabelSelectorQueryVariables = Exact<{
  labels?: Maybe<Scalars['String']>;
}>;


export type InstancesLabelSelectorQuery = { __typename?: 'Query', instanceList?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2InstanceList', instances?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string>, creationTimestamp?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status2', ip?: Maybe<string>, phase?: Maybe<Phase>, url?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string>, tenantCrownlabsPolitoItTenantRef?: Maybe<{ __typename?: 'TenantCrownlabsPolitoItTenantRef', name?: Maybe<string>, tenantV1alpha2Wrapper?: Maybe<{ __typename?: 'TenantV1alpha2Wrapper', itPolitoCrownlabsV1alpha2Tenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: Maybe<{ __typename?: 'Spec6', firstName?: Maybe<string>, lastName?: Maybe<string> }> }> }> }>, templateCrownlabsPolitoItTemplateRef?: Maybe<{ __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name?: Maybe<string>, namespace?: Maybe<string>, templateWrapper?: Maybe<{ __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, environmentType?: Maybe<EnvironmentType> }>>> }> }> }> }> }> }>>> }> };

export type WorkspaceTemplatesQueryVariables = Exact<{
  workspaceNamespace: Scalars['String'];
}>;


export type WorkspaceTemplatesQuery = { __typename?: 'Query', templateList?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2TemplateList', templates?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, resources?: Maybe<{ __typename?: 'Resources', cpu?: Maybe<number>, disk?: Maybe<string>, memory?: Maybe<string> }> }>>>, workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<{ __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef', name?: Maybe<string> }> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string> }> }>>> }> };

export type TenantQueryVariables = Exact<{
  tenantId: Scalars['String'];
}>;


export type TenantQuery = { __typename?: 'Query', tenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: Maybe<{ __typename?: 'Spec6', email?: Maybe<string>, firstName?: Maybe<string>, lastName?: Maybe<string>, lastLogin?: Maybe<string>, publicKeys?: Maybe<Array<Maybe<string>>>, workspaces?: Maybe<Array<Maybe<{ __typename?: 'WorkspacesListItem', role?: Maybe<Role>, name?: Maybe<string>, workspaceWrapperTenantV1alpha2?: Maybe<{ __typename?: 'WorkspaceWrapperTenantV1alpha2', itPolitoCrownlabsV1alpha1Workspace?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', spec?: Maybe<{ __typename?: 'Spec2', prettyName?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status', namespace?: Maybe<{ __typename?: 'Namespace', name?: Maybe<string> }> }> }> }> }>>> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status4', personalNamespace?: Maybe<{ __typename?: 'PersonalNamespace', name?: Maybe<string>, created?: Maybe<boolean> }>, quota?: Maybe<{ __typename?: 'Quota3', cpu?: Maybe<string>, instances?: Maybe<number>, memory?: Maybe<string> }> }> }> };

export type TenantsQueryVariables = Exact<{
  labels?: Maybe<Scalars['String']>;
  retrieveWorkspaces?: Maybe<Scalars['Boolean']>;
}>;


export type TenantsQuery = { __typename?: 'Query', tenants?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2TenantList', items?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec6', firstName?: Maybe<string>, lastName?: Maybe<string>, email?: Maybe<string>, workspaces?: Maybe<Array<Maybe<{ __typename?: 'WorkspacesListItem', role?: Maybe<Role>, name?: Maybe<string> }>>> }> }>>> }> };

export type WorkspacesQueryVariables = Exact<{
  labels?: Maybe<Scalars['String']>;
}>;


export type WorkspacesQuery = { __typename?: 'Query', workspaces?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1WorkspaceList', items?: Maybe<Array<Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec2', prettyName?: Maybe<string>, autoEnroll?: Maybe<AutoEnroll> }> }>>> }> };

export type UpdatedOwnedInstancesSubscriptionVariables = Exact<{
  tenantNamespace: Scalars['String'];
  instanceId?: Maybe<Scalars['String']>;
}>;


export type UpdatedOwnedInstancesSubscription = { __typename?: 'Subscription', updateInstance?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate', updateType?: Maybe<UpdateType>, instance?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string>, creationTimestamp?: Maybe<string>, labels?: Maybe<any> }>, status?: Maybe<{ __typename?: 'Status2', ip?: Maybe<string>, phase?: Maybe<Phase>, url?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string>, templateCrownlabsPolitoItTemplateRef?: Maybe<{ __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name?: Maybe<string>, namespace?: Maybe<string>, templateWrapper?: Maybe<{ __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, environmentType?: Maybe<EnvironmentType> }>>> }> }> }> }> }> }> }> };

export type UpdatedInstancesLabelSelectorSubscriptionVariables = Exact<{
  labels?: Maybe<Scalars['String']>;
}>;


export type UpdatedInstancesLabelSelectorSubscription = { __typename?: 'Subscription', updateInstanceLabelSelector?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2InstanceUpdate', updateType?: Maybe<UpdateType>, instance?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Instance', metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string>, creationTimestamp?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status2', ip?: Maybe<string>, phase?: Maybe<Phase>, url?: Maybe<string> }>, spec?: Maybe<{ __typename?: 'Spec3', running?: Maybe<boolean>, prettyName?: Maybe<string>, tenantCrownlabsPolitoItTenantRef?: Maybe<{ __typename?: 'TenantCrownlabsPolitoItTenantRef', name?: Maybe<string>, tenantV1alpha2Wrapper?: Maybe<{ __typename?: 'TenantV1alpha2Wrapper', itPolitoCrownlabsV1alpha2Tenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: Maybe<{ __typename?: 'Spec6', firstName?: Maybe<string>, lastName?: Maybe<string> }> }> }> }>, templateCrownlabsPolitoItTemplateRef?: Maybe<{ __typename?: 'TemplateCrownlabsPolitoItTemplateRef', name?: Maybe<string>, namespace?: Maybe<string>, templateWrapper?: Maybe<{ __typename?: 'TemplateWrapper', itPolitoCrownlabsV1alpha2Template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, environmentType?: Maybe<EnvironmentType> }>>> }> }> }> }> }> }> }> };

export type UpdatedWorkspaceTemplatesSubscriptionVariables = Exact<{
  workspaceNamespace: Scalars['String'];
  templateId?: Maybe<Scalars['String']>;
}>;


export type UpdatedWorkspaceTemplatesSubscription = { __typename?: 'Subscription', updatedTemplate?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2TemplateUpdate', updateType?: Maybe<UpdateType>, template?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Template', spec?: Maybe<{ __typename?: 'Spec5', prettyName?: Maybe<string>, description?: Maybe<string>, environmentList?: Maybe<Array<Maybe<{ __typename?: 'EnvironmentListListItem', guiEnabled?: Maybe<boolean>, persistent?: Maybe<boolean>, resources?: Maybe<{ __typename?: 'Resources', cpu?: Maybe<number>, disk?: Maybe<string>, memory?: Maybe<string> }> }>>>, workspaceCrownlabsPolitoItWorkspaceRef?: Maybe<{ __typename?: 'WorkspaceCrownlabsPolitoItWorkspaceRef', name?: Maybe<string> }> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string>, namespace?: Maybe<string> }> }> }> };

export type UpdatedTenantSubscriptionVariables = Exact<{
  tenantId: Scalars['String'];
}>;


export type UpdatedTenantSubscription = { __typename?: 'Subscription', updatedTenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2TenantUpdate', updateType?: Maybe<UpdateType>, tenant?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha2Tenant', spec?: Maybe<{ __typename?: 'Spec6', email?: Maybe<string>, firstName?: Maybe<string>, lastName?: Maybe<string>, lastLogin?: Maybe<string>, publicKeys?: Maybe<Array<Maybe<string>>>, workspaces?: Maybe<Array<Maybe<{ __typename?: 'WorkspacesListItem', role?: Maybe<Role>, name?: Maybe<string>, workspaceWrapperTenantV1alpha2?: Maybe<{ __typename?: 'WorkspaceWrapperTenantV1alpha2', itPolitoCrownlabsV1alpha1Workspace?: Maybe<{ __typename?: 'ItPolitoCrownlabsV1alpha1Workspace', spec?: Maybe<{ __typename?: 'Spec2', prettyName?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status', namespace?: Maybe<{ __typename?: 'Namespace', name?: Maybe<string> }> }> }> }> }>>> }>, metadata?: Maybe<{ __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMeta', name?: Maybe<string> }>, status?: Maybe<{ __typename?: 'Status4', personalNamespace?: Maybe<{ __typename?: 'PersonalNamespace', name?: Maybe<string>, created?: Maybe<boolean> }>, quota?: Maybe<{ __typename?: 'Quota3', cpu?: Maybe<string>, instances?: Maybe<number>, memory?: Maybe<string> }> }> }> }> };


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
export type ApplyInstanceComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<ApplyInstanceMutation, ApplyInstanceMutationVariables>, 'mutation'>;

    export const ApplyInstanceComponent = (props: ApplyInstanceComponentProps) => (
      <ApolloReactComponents.Mutation<ApplyInstanceMutation, ApplyInstanceMutationVariables> mutation={ApplyInstanceDocument} {...props} />
    );
    

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
export type ApplyTemplateComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<ApplyTemplateMutation, ApplyTemplateMutationVariables>, 'mutation'>;

    export const ApplyTemplateComponent = (props: ApplyTemplateComponentProps) => (
      <ApolloReactComponents.Mutation<ApplyTemplateMutation, ApplyTemplateMutationVariables> mutation={ApplyTemplateDocument} {...props} />
    );
    

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
export type ApplyTenantComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<ApplyTenantMutation, ApplyTenantMutationVariables>, 'mutation'>;

    export const ApplyTenantComponent = (props: ApplyTenantComponentProps) => (
      <ApolloReactComponents.Mutation<ApplyTenantMutation, ApplyTenantMutationVariables> mutation={ApplyTenantDocument} {...props} />
    );
    

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
export type CreateInstanceComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<CreateInstanceMutation, CreateInstanceMutationVariables>, 'mutation'>;

    export const CreateInstanceComponent = (props: CreateInstanceComponentProps) => (
      <ApolloReactComponents.Mutation<CreateInstanceMutation, CreateInstanceMutationVariables> mutation={CreateInstanceDocument} {...props} />
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
export type CreateTemplateComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<CreateTemplateMutation, CreateTemplateMutationVariables>, 'mutation'>;

    export const CreateTemplateComponent = (props: CreateTemplateComponentProps) => (
      <ApolloReactComponents.Mutation<CreateTemplateMutation, CreateTemplateMutationVariables> mutation={CreateTemplateDocument} {...props} />
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
export type DeleteInstanceComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<DeleteInstanceMutation, DeleteInstanceMutationVariables>, 'mutation'>;

    export const DeleteInstanceComponent = (props: DeleteInstanceComponentProps) => (
      <ApolloReactComponents.Mutation<DeleteInstanceMutation, DeleteInstanceMutationVariables> mutation={DeleteInstanceDocument} {...props} />
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
export type DeleteLabelSelectorInstancesComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables>, 'mutation'>;

    export const DeleteLabelSelectorInstancesComponent = (props: DeleteLabelSelectorInstancesComponentProps) => (
      <ApolloReactComponents.Mutation<DeleteLabelSelectorInstancesMutation, DeleteLabelSelectorInstancesMutationVariables> mutation={DeleteLabelSelectorInstancesDocument} {...props} />
    );
    

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
export type DeleteTemplateComponentProps = Omit<ApolloReactComponents.MutationComponentOptions<DeleteTemplateMutation, DeleteTemplateMutationVariables>, 'mutation'>;

    export const DeleteTemplateComponent = (props: DeleteTemplateComponentProps) => (
      <ApolloReactComponents.Mutation<DeleteTemplateMutation, DeleteTemplateMutationVariables> mutation={DeleteTemplateDocument} {...props} />
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
export type ImagesComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<ImagesQuery, ImagesQueryVariables>, 'query'>;

    export const ImagesComponent = (props: ImagesComponentProps) => (
      <ApolloReactComponents.Query<ImagesQuery, ImagesQueryVariables> query={ImagesDocument} {...props} />
    );
    

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
export type ImagesQueryHookResult = ReturnType<typeof useImagesQuery>;
export type ImagesLazyQueryHookResult = ReturnType<typeof useImagesLazyQuery>;
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
export type OwnedInstancesComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables>, 'query'> & ({ variables: OwnedInstancesQueryVariables; skip?: boolean; } | { skip: boolean; });

    export const OwnedInstancesComponent = (props: OwnedInstancesComponentProps) => (
      <ApolloReactComponents.Query<OwnedInstancesQuery, OwnedInstancesQueryVariables> query={OwnedInstancesDocument} {...props} />
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
export function useOwnedInstancesQuery(baseOptions: Apollo.QueryHookOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(OwnedInstancesDocument, options);
      }
export function useOwnedInstancesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<OwnedInstancesQuery, OwnedInstancesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<OwnedInstancesQuery, OwnedInstancesQueryVariables>(OwnedInstancesDocument, options);
        }
export type OwnedInstancesQueryHookResult = ReturnType<typeof useOwnedInstancesQuery>;
export type OwnedInstancesLazyQueryHookResult = ReturnType<typeof useOwnedInstancesLazyQuery>;
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
export type InstancesLabelSelectorComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables>, 'query'>;

    export const InstancesLabelSelectorComponent = (props: InstancesLabelSelectorComponentProps) => (
      <ApolloReactComponents.Query<InstancesLabelSelectorQuery, InstancesLabelSelectorQueryVariables> query={InstancesLabelSelectorDocument} {...props} />
    );
    

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
export type InstancesLabelSelectorQueryHookResult = ReturnType<typeof useInstancesLabelSelectorQuery>;
export type InstancesLabelSelectorLazyQueryHookResult = ReturnType<typeof useInstancesLabelSelectorLazyQuery>;
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
export type WorkspaceTemplatesComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>, 'query'> & ({ variables: WorkspaceTemplatesQueryVariables; skip?: boolean; } | { skip: boolean; });

    export const WorkspaceTemplatesComponent = (props: WorkspaceTemplatesComponentProps) => (
      <ApolloReactComponents.Query<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables> query={WorkspaceTemplatesDocument} {...props} />
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
export function useWorkspaceTemplatesQuery(baseOptions: Apollo.QueryHookOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>(WorkspaceTemplatesDocument, options);
      }
export function useWorkspaceTemplatesLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<WorkspaceTemplatesQuery, WorkspaceTemplatesQueryVariables>(WorkspaceTemplatesDocument, options);
        }
export type WorkspaceTemplatesQueryHookResult = ReturnType<typeof useWorkspaceTemplatesQuery>;
export type WorkspaceTemplatesLazyQueryHookResult = ReturnType<typeof useWorkspaceTemplatesLazyQuery>;
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
export type TenantComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<TenantQuery, TenantQueryVariables>, 'query'> & ({ variables: TenantQueryVariables; skip?: boolean; } | { skip: boolean; });

    export const TenantComponent = (props: TenantComponentProps) => (
      <ApolloReactComponents.Query<TenantQuery, TenantQueryVariables> query={TenantDocument} {...props} />
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
export function useTenantQuery(baseOptions: Apollo.QueryHookOptions<TenantQuery, TenantQueryVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useQuery<TenantQuery, TenantQueryVariables>(TenantDocument, options);
      }
export function useTenantLazyQuery(baseOptions?: Apollo.LazyQueryHookOptions<TenantQuery, TenantQueryVariables>) {
          const options = {...defaultOptions, ...baseOptions}
          return Apollo.useLazyQuery<TenantQuery, TenantQueryVariables>(TenantDocument, options);
        }
export type TenantQueryHookResult = ReturnType<typeof useTenantQuery>;
export type TenantLazyQueryHookResult = ReturnType<typeof useTenantLazyQuery>;
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
export type TenantsComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<TenantsQuery, TenantsQueryVariables>, 'query'>;

    export const TenantsComponent = (props: TenantsComponentProps) => (
      <ApolloReactComponents.Query<TenantsQuery, TenantsQueryVariables> query={TenantsDocument} {...props} />
    );
    

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
export type TenantsQueryHookResult = ReturnType<typeof useTenantsQuery>;
export type TenantsLazyQueryHookResult = ReturnType<typeof useTenantsLazyQuery>;
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
export type WorkspacesComponentProps = Omit<ApolloReactComponents.QueryComponentOptions<WorkspacesQuery, WorkspacesQueryVariables>, 'query'>;

    export const WorkspacesComponent = (props: WorkspacesComponentProps) => (
      <ApolloReactComponents.Query<WorkspacesQuery, WorkspacesQueryVariables> query={WorkspacesDocument} {...props} />
    );
    

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
export type WorkspacesQueryHookResult = ReturnType<typeof useWorkspacesQuery>;
export type WorkspacesLazyQueryHookResult = ReturnType<typeof useWorkspacesLazyQuery>;
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
export type UpdatedOwnedInstancesComponentProps = Omit<ApolloReactComponents.SubscriptionComponentOptions<UpdatedOwnedInstancesSubscription, UpdatedOwnedInstancesSubscriptionVariables>, 'subscription'>;

    export const UpdatedOwnedInstancesComponent = (props: UpdatedOwnedInstancesComponentProps) => (
      <ApolloReactComponents.Subscription<UpdatedOwnedInstancesSubscription, UpdatedOwnedInstancesSubscriptionVariables> subscription={UpdatedOwnedInstancesDocument} {...props} />
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
 *      instanceId: // value for 'instanceId'
 *   },
 * });
 */
export function useUpdatedOwnedInstancesSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedOwnedInstancesSubscription, UpdatedOwnedInstancesSubscriptionVariables>) {
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
export type UpdatedInstancesLabelSelectorComponentProps = Omit<ApolloReactComponents.SubscriptionComponentOptions<UpdatedInstancesLabelSelectorSubscription, UpdatedInstancesLabelSelectorSubscriptionVariables>, 'subscription'>;

    export const UpdatedInstancesLabelSelectorComponent = (props: UpdatedInstancesLabelSelectorComponentProps) => (
      <ApolloReactComponents.Subscription<UpdatedInstancesLabelSelectorSubscription, UpdatedInstancesLabelSelectorSubscriptionVariables> subscription={UpdatedInstancesLabelSelectorDocument} {...props} />
    );
    

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
export type UpdatedWorkspaceTemplatesComponentProps = Omit<ApolloReactComponents.SubscriptionComponentOptions<UpdatedWorkspaceTemplatesSubscription, UpdatedWorkspaceTemplatesSubscriptionVariables>, 'subscription'>;

    export const UpdatedWorkspaceTemplatesComponent = (props: UpdatedWorkspaceTemplatesComponentProps) => (
      <ApolloReactComponents.Subscription<UpdatedWorkspaceTemplatesSubscription, UpdatedWorkspaceTemplatesSubscriptionVariables> subscription={UpdatedWorkspaceTemplatesDocument} {...props} />
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
 *      templateId: // value for 'templateId'
 *   },
 * });
 */
export function useUpdatedWorkspaceTemplatesSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedWorkspaceTemplatesSubscription, UpdatedWorkspaceTemplatesSubscriptionVariables>) {
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
export type UpdatedTenantComponentProps = Omit<ApolloReactComponents.SubscriptionComponentOptions<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables>, 'subscription'>;

    export const UpdatedTenantComponent = (props: UpdatedTenantComponentProps) => (
      <ApolloReactComponents.Subscription<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables> subscription={UpdatedTenantDocument} {...props} />
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
export function useUpdatedTenantSubscription(baseOptions: Apollo.SubscriptionHookOptions<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables>) {
        const options = {...defaultOptions, ...baseOptions}
        return Apollo.useSubscription<UpdatedTenantSubscription, UpdatedTenantSubscriptionVariables>(UpdatedTenantDocument, options);
      }
export type UpdatedTenantSubscriptionHookResult = ReturnType<typeof useUpdatedTenantSubscription>;
export type UpdatedTenantSubscriptionResult = Apollo.SubscriptionResult<UpdatedTenantSubscription>;
