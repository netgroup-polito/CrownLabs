import { gql } from '@apollo/client';
import * as React from 'react';
import * as Apollo from '@apollo/client';
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
  /** The `JSON` scalar type represents JSON values as specified by [ECMA-404](http://www.ecma-international.org/publications/files/ECMA-ST/ECMA-404.pdf). */
  JSON: any;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItem = {
  __typename?: 'EnvironmentListListItem';
  /** The type of environment to be instantiated, among VirtualMachine and Container. */
  environmentType?: Maybe<EnvironmentType>;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: Maybe<Scalars['Boolean']>;
  /** The VM or container to be started when instantiating the environment. */
  image?: Maybe<Scalars['String']>;
  /** The name identifying the specific environment. */
  name?: Maybe<Scalars['String']>;
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: Maybe<Scalars['Boolean']>;
  /** The amount of computational resources associated with the environment. */
  resources?: Maybe<Resources>;
};

/** Environment defines the characteristics of an environment composing the Template. */
export type EnvironmentListListItemInput = {
  /** The type of environment to be instantiated, among VirtualMachine and Container. */
  environmentType: EnvironmentType;
  /** Whether the environment is characterized by a graphical desktop or not. */
  guiEnabled?: Maybe<Scalars['Boolean']>;
  /** The VM or container to be started when instantiating the environment. */
  image: Scalars['String'];
  /** The name identifying the specific environment. */
  name: Scalars['String'];
  /** Whether the environment should be persistent (i.e. preserved when the corresponding instance is terminated) or not. */
  persistent?: Maybe<Scalars['Boolean']>;
  /** The amount of computational resources associated with the environment. */
  resources: ResourcesInput;
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
}

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
export type IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input = {
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
  /**
   * selfLink is a URL representing this object. Populated by the system. Read-only.
   *
   * DEPRECATED Kubernetes will stop propagating this field in 1.20 release and the field is planned to be removed in 1.21 release.
   */
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
  /** Time is a wrapper around time.Time which supports correct marshaling to YAML and JSON.  Wrappers are provided for many of the factory methods that the time package offers. */
  time?: Maybe<Scalars['String']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMetaV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1ObjectMetaV2';
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations */
  annotations?: Maybe<Scalars['JSON']>;
  /** The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request. */
  clusterName?: Maybe<Scalars['String']>;
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
   * If this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header).
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: Maybe<Scalars['String']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: Maybe<Scalars['Float']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels */
  labels?: Maybe<Scalars['JSON']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntry>>
  >;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name?: Maybe<Scalars['String']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces
   */
  namespace?: Maybe<Scalars['String']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2>>
  >;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: Maybe<Scalars['String']>;
  /**
   * SelfLink is a URL representing this object. Populated by the system. Read-only.
   *
   * DEPRECATED Kubernetes will stop propagating this field in 1.20 release and the field is planned to be removed in 1.21 release.
   */
  selfLink?: Maybe<Scalars['String']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids
   */
  uid?: Maybe<Scalars['String']>;
};

/** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
export type IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input = {
  /** Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata. They are not queryable and should be preserved when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations */
  annotations?: Maybe<Scalars['JSON']>;
  /** The name of the cluster which the object belongs to. This is used to distinguish resources with same name and namespace in different clusters. This field is not set anywhere right now and apiserver is going to ignore it if set in create or update request. */
  clusterName?: Maybe<Scalars['String']>;
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
   * If this field is specified and the generated name exists, the server will NOT return a 409 - instead, it will either return 201 Created or 500 with Reason ServerTimeout indicating a unique name could not be found in the time allotted, and the client should retry (optionally after the time indicated in the Retry-After header).
   *
   * Applied only if Name is not specified. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#idempotency
   */
  generateName?: Maybe<Scalars['String']>;
  /** A sequence number representing a specific generation of the desired state. Populated by the system. Read-only. */
  generation?: Maybe<Scalars['Float']>;
  /** Map of string keys and values that can be used to organize and categorize (scope and select) objects. May match selectors of replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels */
  labels?: Maybe<Scalars['JSON']>;
  /** ManagedFields maps workflow-id and version to the set of fields that are managed by that workflow. This is mostly for internal housekeeping, and users typically shouldn't need to set or understand this field. A workflow can be the user's name, a controller's name, or the name of a specific apply path like "ci-cd". The set of fields is always in the version that the workflow used when modifying the object. */
  managedFields?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1ManagedFieldsEntryInput>>
  >;
  /** Name must be unique within a namespace. Is required when creating resources, although some resources may allow a client to request the generation of an appropriate name automatically. Name is primarily intended for creation idempotence and configuration definition. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names */
  name?: Maybe<Scalars['String']>;
  /**
   * Namespace defines the space within which each name must be unique. An empty namespace is equivalent to the "default" namespace, but "default" is the canonical representation. Not all objects are required to be scoped to a namespace - the value of this field for those objects will be empty.
   *
   * Must be a DNS_LABEL. Cannot be updated. More info: http://kubernetes.io/docs/user-guide/namespaces
   */
  namespace?: Maybe<Scalars['String']>;
  /** List of objects depended by this object. If ALL objects in the list have been deleted, this object will be garbage collected. If this object is managed by a controller, then an entry in this list will point to this controller, with the controller field set to true. There cannot be more than one managing controller. */
  ownerReferences?: Maybe<
    Array<Maybe<IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2Input>>
  >;
  /**
   * An opaque value that represents the internal version of this object that can be used by clients to determine when objects have changed. May be used for optimistic concurrency, change detection, and the watch operation on a resource or set of resources. Clients must treat these values as opaque and passed unmodified back to the server. They may only be valid for a particular resource or set of resources.
   *
   * Populated by the system. Read-only. Value must be treated as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency
   */
  resourceVersion?: Maybe<Scalars['String']>;
  /**
   * SelfLink is a URL representing this object. Populated by the system. Read-only.
   *
   * DEPRECATED Kubernetes will stop propagating this field in 1.20 release and the field is planned to be removed in 1.21 release.
   */
  selfLink?: Maybe<Scalars['String']>;
  /**
   * UID is the unique in time and space value for this object. It is typically generated by the server on successful creation of a resource and is not allowed to change on PUT operations.
   *
   * Populated by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids
   */
  uid?: Maybe<Scalars['String']>;
};

/** OwnerReference contains enough information to let you identify an owning object. An owning object must be in the same namespace as the dependent, or be cluster-scoped, so there is no namespace field. */
export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2';
  /** API version of the referent. */
  apiVersion?: Maybe<Scalars['String']>;
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
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
export type IoK8sApimachineryPkgApisMetaV1OwnerReferenceV2Input = {
  /** API version of the referent. */
  apiVersion: Scalars['String'];
  /** If true, AND if the owner has the "foregroundDeletion" finalizer, then the owner cannot be deleted from the key-value store until this reference is removed. Defaults to false. To set this field, a user needs "delete" permission of the owner, otherwise 422 (Unprocessable Entity) will be returned. */
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
export type IoK8sApimachineryPkgApisMetaV1StatusDetailsV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusDetailsV2';
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

/** Status is a return value for calls that don't return other objects. */
export type IoK8sApimachineryPkgApisMetaV1StatusV2 = {
  __typename?: 'IoK8sApimachineryPkgApisMetaV1StatusV2';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Suggested HTTP return code for this status, 0 if not set. */
  code?: Maybe<Scalars['Int']>;
  /** StatusDetails is a set of additional properties that MAY be set by the server to provide additional information about a response. The Reason field of a Status object defines what attributes will be set. Clients must ignore fields that do not match the defined type of each attribute, and should assume that any attribute may be empty, invalid, or under defined. */
  details?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusDetailsV2>;
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

/** ImageListList is a list of ImageList */
export type ItPolitoCrownlabsV1alpha1ImageListList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1ImageListList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of imagelists. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha1Tenant = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: Maybe<Spec>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: Maybe<Status>;
};

/** Tenant describes a user of CrownLabs. */
export type ItPolitoCrownlabsV1alpha1TenantInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  /** TenantSpec is the specification of the desired state of the Tenant. */
  spec?: Maybe<SpecInput>;
  /** TenantStatus reflects the most recently observed status of the Tenant. */
  status?: Maybe<StatusInput>;
};

/** TenantList is a list of Tenant */
export type ItPolitoCrownlabsV1alpha1TenantList = {
  __typename?: 'ItPolitoCrownlabsV1alpha1TenantList';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** List of tenants. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md */
  items?: Maybe<Array<Maybe<ItPolitoCrownlabsV1alpha1Tenant>>>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ListMeta describes metadata that synthetic resources must have, including lists and various status objects. A resource may have only one of {ObjectMeta, ListMeta}. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ListMeta>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1Workspace = {
  __typename?: 'ItPolitoCrownlabsV1alpha1Workspace';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: Maybe<Spec2>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: Maybe<Status2>;
};

/** Workspace describes a workspace in CrownLabs. */
export type ItPolitoCrownlabsV1alpha1WorkspaceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  /** WorkspaceSpec is the specification of the desired state of the Workspace. */
  spec?: Maybe<Spec2Input>;
  /** WorkspaceStatus reflects the most recently observed status of the Workspace. */
  status?: Maybe<Status2Input>;
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

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2Instance = {
  __typename?: 'ItPolitoCrownlabsV1alpha2Instance';
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: Maybe<Spec3>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: Maybe<Status3>;
};

/** Instance describes the instance of a CrownLabs environment Template. */
export type ItPolitoCrownlabsV1alpha2InstanceInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  /** InstanceSpec is the specification of the desired state of the Instance. */
  spec?: Maybe<Spec3Input>;
  /** InstanceStatus reflects the most recently observed status of the Instance. */
  status?: Maybe<Status3Input>;
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
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: Maybe<Spec4>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: Maybe<Status4>;
};

/** InstanceSnapshot is the Schema for the instancesnapshots API. */
export type ItPolitoCrownlabsV1alpha2InstanceSnapshotInput = {
  /** APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources */
  apiVersion?: Maybe<Scalars['String']>;
  /** Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds */
  kind?: Maybe<Scalars['String']>;
  /** ObjectMeta is metadata that all persisted resources must have, which includes all objects users must create. */
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
  /** InstanceSnapshotSpec defines the desired state of InstanceSnapshot. */
  spec?: Maybe<Spec4Input>;
  /** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
  status?: Maybe<Status4Input>;
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
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2>;
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
  metadata?: Maybe<IoK8sApimachineryPkgApisMetaV1ObjectMetaV2Input>;
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

/** The start of any mutation */
export type Mutation = {
  __typename?: 'Mutation';
  /**
   * create an ImageList
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  createCrownlabsPolitoItV1alpha1ImageList?: Maybe<Scalars['String']>;
  /**
   * create a Tenant
   *
   * Equivalent to POST /apis/crownlabs.polito.it/v1alpha1/tenants
   */
  createCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
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
   * delete collection of ImageList
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  deleteCrownlabsPolitoItV1alpha1CollectionImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete collection of Tenant
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/tenants
   */
  deleteCrownlabsPolitoItV1alpha1CollectionTenant?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete collection of Workspace
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/workspaces
   */
  deleteCrownlabsPolitoItV1alpha1CollectionWorkspace?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete an ImageList
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  deleteCrownlabsPolitoItV1alpha1ImageList?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete a Tenant
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/tenants/{name}
   */
  deleteCrownlabsPolitoItV1alpha1Tenant?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete a Workspace
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha1/workspaces/{name}
   */
  deleteCrownlabsPolitoItV1alpha1Workspace?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete collection of Instance
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete collection of InstanceSnapshot
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete collection of Template
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates
   */
  deleteCrownlabsPolitoItV1alpha2CollectionNamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete an Instance
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instances/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedInstance?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete an InstanceSnapshot
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/instancesnapshots/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshot?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * delete a Template
   *
   * Equivalent to DELETE /apis/crownlabs.polito.it/v1alpha2/namespaces/{namespace}/templates/{name}
   */
  deleteCrownlabsPolitoItV1alpha2NamespacedTemplate?: Maybe<IoK8sApimachineryPkgApisMetaV1StatusV2>;
  /**
   * partially update the specified ImageList
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  patchCrownlabsPolitoItV1alpha1ImageList?: Maybe<Scalars['String']>;
  /**
   * partially update status of the specified ImageList
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}/status
   */
  patchCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<Scalars['String']>;
  /**
   * partially update the specified Tenant
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/tenants/{name}
   */
  patchCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  /**
   * partially update status of the specified Tenant
   *
   * Equivalent to PATCH /apis/crownlabs.polito.it/v1alpha1/tenants/{name}/status
   */
  patchCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
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
   * replace the specified ImageList
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  replaceCrownlabsPolitoItV1alpha1ImageList?: Maybe<Scalars['String']>;
  /**
   * replace status of the specified ImageList
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<Scalars['String']>;
  /**
   * replace the specified Tenant
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/tenants/{name}
   */
  replaceCrownlabsPolitoItV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  /**
   * replace status of the specified Tenant
   *
   * Equivalent to PUT /apis/crownlabs.polito.it/v1alpha1/tenants/{name}/status
   */
  replaceCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
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
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha1ImageListArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha1TenantArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

/** The start of any mutation */
export type MutationCreateCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
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
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
};

/** The start of any mutation */
export type MutationDeleteCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  gracePeriodSeconds?: Maybe<Scalars['Int']>;
  orphanDependents?: Maybe<Scalars['Boolean']>;
  propagationPolicy?: Maybe<Scalars['String']>;
  ioK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input?: Maybe<IoK8sApimachineryPkgApisMetaV1DeleteOptionsV2Input>;
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

/** The start of any mutation */
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

/** The start of any mutation */
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

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationPatchCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  applicationJsonPatchJsonInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1ImageListStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1ImageListInput: Scalars['String'];
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1TenantStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1TenantInput: ItPolitoCrownlabsV1alpha1TenantInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha1WorkspaceStatusArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha1WorkspaceInput: ItPolitoCrownlabsV1alpha1WorkspaceInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceSnapshotStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceSnapshotInput: ItPolitoCrownlabsV1alpha2InstanceSnapshotInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedInstanceStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2InstanceInput: ItPolitoCrownlabsV1alpha2InstanceInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
};

/** The start of any mutation */
export type MutationReplaceCrownlabsPolitoItV1alpha2NamespacedTemplateStatusArgs = {
  name: Scalars['String'];
  namespace: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  dryRun?: Maybe<Scalars['String']>;
  fieldManager?: Maybe<Scalars['String']>;
  itPolitoCrownlabsV1alpha2TemplateInput: ItPolitoCrownlabsV1alpha2TemplateInput;
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

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quotas, network policies, ...) created by the tenant-operator. */
export type PersonalNamespace = {
  __typename?: 'PersonalNamespace';
  /** Whether the creation succeeded or not. */
  created?: Maybe<Scalars['Boolean']>;
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quotas, network policies, ...) created by the tenant-operator. */
export type PersonalNamespaceInput = {
  /** Whether the creation succeeded or not. */
  created: Scalars['Boolean'];
  /** The name of the considered resource. */
  name?: Maybe<Scalars['String']>;
};

/** The start of any query */
export type Query = {
  __typename?: 'Query';
  /**
   * read the specified ImageList
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/imagelists/{name}
   */
  itPolitoCrownlabsV1alpha1ImageList?: Maybe<Scalars['String']>;
  /**
   * list objects of kind ImageList
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/imagelists
   */
  itPolitoCrownlabsV1alpha1ImageListList?: Maybe<ItPolitoCrownlabsV1alpha1ImageListList>;
  /**
   * read the specified Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/tenants/{name}
   */
  itPolitoCrownlabsV1alpha1Tenant?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
  /**
   * list objects of kind Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/tenants
   */
  itPolitoCrownlabsV1alpha1TenantList?: Maybe<ItPolitoCrownlabsV1alpha1TenantList>;
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
  readCrownlabsPolitoItV1alpha1ImageListStatus?: Maybe<Scalars['String']>;
  /**
   * read status of the specified Tenant
   *
   * Equivalent to GET /apis/crownlabs.polito.it/v1alpha1/tenants/{name}/status
   */
  readCrownlabsPolitoItV1alpha1TenantStatus?: Maybe<ItPolitoCrownlabsV1alpha1Tenant>;
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
export type QueryItPolitoCrownlabsV1alpha1TenantArgs = {
  name: Scalars['String'];
  pretty?: Maybe<Scalars['String']>;
  resourceVersion?: Maybe<Scalars['String']>;
};

/** The start of any query */
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
export type QueryReadCrownlabsPolitoItV1alpha1TenantStatusArgs = {
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

/** The amount of computational resources associated with the environment. */
export type Resources = {
  __typename?: 'Resources';
  /** The maximum number of CPU cores made available to the environment (ranging between 1 and 8 cores). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu?: Maybe<Scalars['Int']>;
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent environments, while it is silently ignored in the other cases. */
  disk?: Maybe<Scalars['String']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory?: Maybe<Scalars['String']>;
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage?: Maybe<Scalars['Int']>;
};

/** The amount of computational resources associated with the environment. */
export type ResourcesInput = {
  /** The maximum number of CPU cores made available to the environment (ranging between 1 and 8 cores). This maps to the 'limits' specified for the actual pod representing the environment. */
  cpu: Scalars['Int'];
  /** The size of the persistent disk allocated for the given environment. This field is meaningful only in case of persistent environments, while it is silently ignored in the other cases. */
  disk?: Maybe<Scalars['String']>;
  /** The amount of RAM memory assigned to the given environment. Requests and limits do correspond to avoid OOMKill issues. */
  memory: Scalars['String'];
  /** The percentage of reserved CPU cores, ranging between 1 and 100, with respect to the 'CPU' value. Essentially, this corresponds to the 'requests' specified for the actual pod representing the environment. */
  reservedCPUPercentage: Scalars['Int'];
};

export enum Role {
  Manager = 'manager',
  User = 'user',
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

/** TenantSpec is the specification of the desired state of the Tenant. */
export type Spec = {
  __typename?: 'Spec';
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: Maybe<Scalars['Boolean']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email?: Maybe<Scalars['String']>;
  /** The first name of the Tenant. */
  firstName?: Maybe<Scalars['String']>;
  /** The last name of the Tenant. */
  lastName?: Maybe<Scalars['String']>;
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: Maybe<Array<Maybe<WorkspacesListItem>>>;
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2 = {
  __typename?: 'Spec2';
  /** The human-readable name of the Workspace. */
  prettyName?: Maybe<Scalars['String']>;
};

/** WorkspaceSpec is the specification of the desired state of the Workspace. */
export type Spec2Input = {
  /** The human-readable name of the Workspace. */
  prettyName: Scalars['String'];
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3 = {
  __typename?: 'Spec3';
  /** Whether the current instance is running or not. This field is meaningful only in case the Instance refers to persistent environments, and it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. The flag, on the other hand, is silently ignored in case of non-persistent environments, as the state cannot be preserved among reboots. */
  running?: Maybe<Scalars['Boolean']>;
  /** The reference to the Template to be instantiated. */
  templateCrownlabsPolitoItTemplateRef?: Maybe<TemplateCrownlabsPolitoItTemplateRef>;
  /** The reference to the Tenant which owns the Instance object. */
  tenantCrownlabsPolitoItTenantRef?: Maybe<TenantCrownlabsPolitoItTenantRef>;
};

/** InstanceSpec is the specification of the desired state of the Instance. */
export type Spec3Input = {
  /** Whether the current instance is running or not. This field is meaningful only in case the Instance refers to persistent environments, and it allows to stop the environment (e.g. the underlying VM) without deleting the associated disk. Setting the flag to true will restart the environment, attaching it to the same disk used previously. The flag, on the other hand, is silently ignored in case of non-persistent environments, as the state cannot be preserved among reboots. */
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
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. */
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
  /** The maximum lifetime of an Instance referencing the current Template. Once this period is expired, the Instance may be automatically deleted or stopped to save resources. */
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
export type SpecInput = {
  /** Whether a sandbox namespace should be created to allow the Tenant play with Kubernetes. */
  createSandbox?: Maybe<Scalars['Boolean']>;
  /** The email associated with the Tenant, which will be used to log-in into the system. */
  email: Scalars['String'];
  /** The first name of the Tenant. */
  firstName: Scalars['String'];
  /** The last name of the Tenant. */
  lastName: Scalars['String'];
  /** The list of the SSH public keys associated with the Tenant. These will be used to enable to access the remote environments through the SSH protocol. */
  publicKeys?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The list of the Workspaces the Tenant is subscribed to, along with his/her role in each of them. */
  workspaces?: Maybe<Array<Maybe<WorkspacesListItemInput>>>;
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type Status = {
  __typename?: 'Status';
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces?: Maybe<Array<Maybe<Scalars['String']>>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quotas, network policies, ...) created by the tenant-operator. */
  personalNamespace?: Maybe<PersonalNamespace>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace?: Maybe<SandboxNamespace>;
  /** The list of the subscriptions to external services (e.g. Keycloak, Nextcloud, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions?: Maybe<Scalars['JSON']>;
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type Status2 = {
  __typename?: 'Status2';
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: Maybe<Namespace>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, Nextcloud, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: Maybe<Scalars['JSON']>;
};

/** WorkspaceStatus reflects the most recently observed status of the Workspace. */
export type Status2Input = {
  /** The namespace containing all CrownLabs related objects of the Workspace. This is the namespace that groups multiple related templates, together with all the accessory resources (e.g. RBACs) created by the tenant operator. */
  namespace?: Maybe<NamespaceInput>;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready?: Maybe<Scalars['Boolean']>;
  /** The list of the subscriptions to external services (e.g. Keycloak, Nextcloud, ...), indicating for each one whether it succeeded or an error occurred. */
  subscription?: Maybe<Scalars['JSON']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status3 = {
  __typename?: 'Status3';
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: Maybe<Scalars['String']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: Maybe<Scalars['String']>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: Maybe<Scalars['String']>;
};

/** InstanceStatus reflects the most recently observed status of the Instance. */
export type Status3Input = {
  /** The internal IP address associated with the remote environment, which can be used to access it through the SSH protocol (leveraging the SSH bastion in case it is not contacted from another CrownLabs Instance). */
  ip?: Maybe<Scalars['String']>;
  /** The current status Instance, with reference to the associated environment (e.g. VM). This conveys which resource is being created, as well as whether the associated VM is being scheduled, is running or ready to accept incoming connections. */
  phase?: Maybe<Scalars['String']>;
  /** The URL where it is possible to access the remote desktop of the instance (in case of graphical environments) */
  url?: Maybe<Scalars['String']>;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status4 = {
  __typename?: 'Status4';
  /** Phase represents the current state of the Instance Snapshot. */
  phase?: Maybe<Scalars['String']>;
};

/** InstanceSnapshotStatus defines the observed state of InstanceSnapshot. */
export type Status4Input = {
  /** Phase represents the current state of the Instance Snapshot. */
  phase: Scalars['String'];
};

/** TenantStatus reflects the most recently observed status of the Tenant. */
export type StatusInput = {
  /** The list of Workspaces that are throwing errors during subscription. This mainly happens if .spec.Workspaces contains references to Workspaces which do not exist. */
  failingWorkspaces: Array<Maybe<Scalars['String']>>;
  /** The namespace containing all CrownLabs related objects of the Tenant. This is the namespace that groups his/her own Instances, together with all the accessory resources (e.g. RBACs, resource quotas, network policies, ...) created by the tenant-operator. */
  personalNamespace: PersonalNamespaceInput;
  /** Whether all subscriptions and resource creations succeeded or an error occurred. In case of errors, the other status fields provide additional information about which problem occurred. */
  ready: Scalars['Boolean'];
  /** The namespace that can be freely used by the Tenant to play with Kubernetes. This namespace is created only if the .spec.CreateSandbox flag is true. */
  sandboxNamespace: SandboxNamespaceInput;
  /** The list of the subscriptions to external services (e.g. Keycloak, Nextcloud, ...), indicating for each one whether it succeeded or an error occurred. */
  subscriptions: Scalars['JSON'];
};

export type Subscription = {
  __typename?: 'Subscription';
  itPolitoCrownlabsV1alpha2InstanceUpdate?: Maybe<ItPolitoCrownlabsV1alpha2InstanceUpdate>;
  itPolitoCrownlabsV1alpha2TemplateUpdate?: Maybe<ItPolitoCrownlabsV1alpha2TemplateUpdate>;
};

export type SubscriptionItPolitoCrownlabsV1alpha2InstanceUpdateArgs = {
  name?: Maybe<Scalars['String']>;
  namespace: Scalars['String'];
};

export type SubscriptionItPolitoCrownlabsV1alpha2TemplateUpdateArgs = {
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
  tenantWrapper?: Maybe<TenantWrapper>;
};

/** The reference to the Tenant which owns the Instance object. */
export type TenantCrownlabsPolitoItTenantRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
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

/** The reference to the Workspace resource the Tenant is subscribed to. */
export type WorkspaceRef = {
  __typename?: 'WorkspaceRef';
  /** The name of the resource to be referenced. */
  name?: Maybe<Scalars['String']>;
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
  workspaceWrapper?: Maybe<WorkspaceWrapper>;
};

/** The reference to the Workspace resource the Tenant is subscribed to. */
export type WorkspaceRefInput = {
  /** The name of the resource to be referenced. */
  name: Scalars['String'];
  /** The namespace containing the resource to be referenced. It should be left empty in case of cluster-wide resources. */
  namespace?: Maybe<Scalars['String']>;
};

export type WorkspaceWrapper = {
  __typename?: 'WorkspaceWrapper';
  itPolitoCrownlabsV1alpha1Workspace?: Maybe<ItPolitoCrownlabsV1alpha1Workspace>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItem = {
  __typename?: 'WorkspacesListItem';
  /** The number of the group the Tenant belongs to. Empty means no group. */
  groupNumber?: Maybe<Scalars['Int']>;
  /** The role of the Tenant in the context of the Workspace. */
  role?: Maybe<Role>;
  /** The reference to the Workspace resource the Tenant is subscribed to. */
  workspaceRef?: Maybe<WorkspaceRef>;
};

/** TenantWorkspaceEntry contains the information regarding one of the Workspaces the Tenant is subscribed to, including his/her role. */
export type WorkspacesListItemInput = {
  /** The number of the group the Tenant belongs to. Empty means no group. */
  groupNumber?: Maybe<Scalars['Int']>;
  /** The role of the Tenant in the context of the Workspace. */
  role: Role;
  /** The reference to the Workspace resource the Tenant is subscribed to. */
  workspaceRef: WorkspaceRefInput;
};

export type TenantQueryVariables = Exact<{
  tenantId: Scalars['String'];
}>;

export type TenantQuery = {
  __typename?: 'Query';
  tenant?: Maybe<{
    __typename?: 'ItPolitoCrownlabsV1alpha1Tenant';
    spec?: Maybe<{
      __typename?: 'Spec';
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
                    __typename?: 'Spec2';
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
