# OPA Policies

## Available policies

This section details and documents the different policies available in CrownLabs:
* **Verify Tenant Patch**: verifies that a tenant creation or patch is allowed
* **Verify Instance-Template Reference**: this policy verifies that an instance refers to an existing template in the correct namespace when it is created or updated.

### Verify Tenant Patch

This policy verifies that a tenant creation or patch is allowed. In particular :

- for creation
  - is allowed only if the creator is manager in all workspaces for which he creates the new tenant resource.
  - is never allowed for a user.
  - the creator must already have his own tenant in the cluster.
  - is not allowed to create a tenant that already exists

- for update
  - a user can modify any tenant workspaces for which he is manager
  - a user can always update the publicKeys on his own tenant
  - is not allowed to modify any other field

- for a cluster-admin or an operator all operations are allowed

## How to deploy

The creation of the Gatekeeper resources and their deployment is automated through an Helm Chart.

This allows the creators of the policies to focus on writing the appropriate constraints in the `rego` language (together with the corresponding tests), without dealing with the Gatekeeper manifests.
Additionally, it also prevents the duplication of the policies themselves in two files (i.e. the `rego` policy, which is used to run the tests, and the Gatekeeper ConstraintTemplate), avoiding possible inconsistencies between the two.

In order to add a new policy to the Helm chart, it is sufficient to an add an entry to the `policies` array of the [values.yaml](values.yaml) file. Specifically, each entry contains the following fields:

* **name**: the name assigned to the policy resources (in camel-case);
* **file**: the path of the `rego` file implementing the policy;
* **dryRun**: whether the policy is enforced or violations are only logged;
* **resources**: the list of resources the policy applies to;
* **sync**: the list of resources that need to be accessed from the policy (optional).

## OPA POLICIES

Policies are rules that can be applied when any of your cluster object is _created_, _updated_, _patched_ and _deleted_.

Policies are always executed _before_ the action takes place and their goal is to allow or prevent the requested actions.

The language used to express OPA policies is [REGO](https://www.openpolicyagent.org/docs/latest/policy-language/).

### HOW TO USE GATEKEEPER

To use OPA gatekeeper you need to:

- Install Gatekeeper in your cluster, deploying the CRDs and the associated operators/components. See [this README](../infrastructure/policy-enforcement/README.md) for further information about the installation process.

- Create your own constraint and constraintTemplate files and, if needed, the config-sync file. See [the examples folder](../infrastructure/policy-enforcement/examples) for an example of the three resources.

Now you can test your policies on your cluster objects.



### ConstraintTemplate

In the ConstraintTemplate you should define your own constraint CRD schema and associate it to your rego policies, such as in this example: [**ConstraintTemplate.yaml**](../infrastructure/policy-enforcement/examples/constraintTemplate.yaml).

Under the _crd_ tag you have to define your constraint CRD schema:
- _kind_: must be the same as the constraintTemplate name in camelcase (mandatory)
- _properties_: must declare your fixed parameters (optional)

Under the _targets_ tag you have to define your rego policy.
The example REGO policy executes the following operations:
- the `provided` variable retrieves all the label keys of the input object
- the `required` variable retrieves all the label values of the constraint parameters
- the `missing` variable calculates the result of subtraction between required and provided, therefore it will contain all the labels required but not provided
- if `missing` is empty then `count(missing)>0` will be false, the execution will be ended and so there will be no violation
- if `missing` is not empty then the condition will be true, the execution will continue and so there will be a violation
- the `msg` variable, containg the `sprintf` value, will be the returned value of the violation together with the `details` object

As you could see, you can retrieve your input object using `input` variable.
In OPA, input is a reserved, global variable whose value is the Kubernetes AdmissionReview object that the API server hands to any admission control webhook.


### Constraint

In Constraint you should declare the resources on which your ConstraintTemplate operates and define possible optional parameters (e.g., reject a modification if a value is greater than a given threshold), for example: [**constraint.yaml**](../infrastructure/policy-enforcement/examples/constraint.yaml).

The _kind_ tag must be the one specified in the _kind_ tag of the related constraintTemplate.

The _match_ tag must list the cluster resources on which you want to enforce the policy.

The _paramters_ tag must list your fixed parameters, if any.


### Config-Sync

Some constraints cannot be written without having access to more cluster resources than just the object under test. For example, it is impossible to know if an instance is valid unless a rule has access to all the templates. To make such rules possible, we enable syncing of data into OPA. [1] 

The OPA rules do not require cluster resources by default, however if you need them you must first cache your objects before using them in your contraint violations.
To do so Gatekeeper provides a Config CRD where you can define which resources need to be cached.

**Note:** The Config resource has to be named `config` for it to be reconciled by Gatekeeper. Gatekeeper will ignore the resource if you do not name it `config`.

Kubernetes data can be replicated into OPA via the sync config resource. Currently resources defined in `syncOnly` will be synced into OPA. Updating syncOnly should dynamically update what objects are synced.
Once data is synced into OPA, rules can access the cached data under the `data.inventory` document, for example: [**config-sync.yaml**](../infrastructure/policy-enforcement/examples/config-sync.yaml).

Under _syncOnly_ tag yuo should list the resources to be cached, for each object define its _group_, _version_ and _kind_.


### Retrieve cached resources 

 To retrieve the cached resources in the rego rules you can do as follow [1]:

- For cluster-scoped objects: `data.inventory.cluster[<groupVersion>][<kind>][<name>]`

- Example referencing the Gatekeeper namespace: `data.inventory.cluster["v1"].Namespace["gatekeeper"]`

- For namespace-scoped objects: `data.inventory.namespace[<namespace>][groupVersion][<kind>][<name>]`

- Example referencing the Gatekeeper pod: `data.inventory.namespace["gatekeeper"]["v1"]["Pod"]["gatekeeper-controller-manager-d4c98b788-j7d92"]`

### Retrieve input object

The `input` document contains the following fields [2]:

-   `input.review.kind`  specifies the type of the object (e.g.,  `Pod`,  `Service`, etc.)
-   `input.review.operation`  specifies the type of the operation, i.e.,  `CREATE`,  `UPDATE`,  `DELETE`,  `CONNECT`.
-   `input.review.userInfo`  specifies the identity of the caller.
-   `input.review.object`  contains the entire Kubernetes object.
-   `input.review.oldObject`  specifies the previous version of the Kubernetes object on  `UPDATE`  and  `DELETE`.

### How to write rego policies in gatekeeper
You should use `violation` functions to write your rules.

    violation[{"msg":msg,"details":{}}]{
    }
    
 `msg` and `details`are the values returned if a violation occurs, otherwise nothing is returned.
The rule must have at least one boolean condition to take a decision whether the violation occurs or not.
If you need to specify more boolean conditions, proceed as follows:

 - For `and` conditions (all your conditions must be true in order for the violation to take place):
    write all of them in one violation function
  
 - For `or` conditions (at least one condition must be true in order for the violation to take place):
   write one violation function for each condition
  
In general OPA policies in gatekeeper work as black list mode: so you specify what can't be done.

 You can learn [REGO](https://www.openpolicyagent.org/docs/latest/#rego) language.
 
 You can exploit [PLAYGROUND](https://play.openpolicyagent.org/) to try your rules.

### Note
[1] This information is taken from the official [documentation](https://open-policy-agent.github.io/gatekeeper/website/docs/sync).

[2] This information is taken from the official [documentation](https://www.openpolicyagent.org/docs/latest/kubernetes-introduction/#how-does-it-work-with-plain-opa-and-kube-mgmt).
