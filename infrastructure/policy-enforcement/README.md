
CUSTOM POLICY ENFORCEMENT
=====================================
## GOALS 
Kubernetes allows you to write policies on your cluster objects by means of admission controller webhooks which are executed everytime a cluster component is created or modified.
[OPA](https://www.openpolicyagent.org/) (Open Policy Agent) is an open source, general-purpose policy engine to write unifed policies across different applications. It exploits [REGO](https://www.openpolicyagent.org/docs/latest/policy-language/) language to write policies.
[Gatekeeper](https://github.com/open-policy-agent/gatekeeper) allows you to integrate easily OPA on kubernetes by adding:
 - native kubernetes CRDs for instantiating the policy library (Constraints)
 - native kubernetes CRDs for extending the policy library (ConstraintTemplates)


## HOW TO INSTALL
### Prerequisites 

- minimun kubernetes version: 1.14
- make sure you have cluster admin permissions
#### Deploying a Release using Prebuilt Image
Run the following command:

    kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.2/deploy/gatekeeper.yaml
For more information visit the [gatekeeper official repository](https://github.com/open-policy-agent/gatekeeper)

## HOW TO USE GATEKEEPER
### ConstraintTemplates
In the ConstraintTemplates you should define your own constraint CRD and associate it to your rego policies, for example : [ **constrainttemplate.yaml**](./examples/constrainttemplate.yaml)


  
 ### Constraint 
In Constraint you should declare on which resources your ConstraintTemplate must be enforced and define your parameters values, for example: **[constraint.yaml](./examples/constraint.yaml)**

     

### Config-Sync
If you need to get resources from the cluster to write your policy you need first to cache those and then you will be able to retrieve them in the ConstraintTemplate rules,
for example: **[config-sync.yaml](./examples/config-sync.yaml)**

### Retrieve cached resources
To retrieve the cached resources in the rego rules you can do as follow:

-   For cluster-scoped objects:  `data.inventory.cluster[<groupVersion>][<kind>][<name>]`
    -   Example referencing the Gatekeeper namespace:  `data.inventory.cluster["v1"].Namespace["gatekeeper"]`
-   For namespace-scoped objects:  `data.inventory.namespace[<namespace>][groupVersion][<kind>][<name>]`
    -   Example referencing the Gatekeeper pod:  `data.inventory.namespace["gatekeeper"]["v1"]["Pod"]["gatekeeper-controller-manager-d4c98b788-j7d92"]`

