
CUSTOM POLICY ENFORCEMENT
=====================================
## GOALS 
Kubernetes allows you to write policies on your cluster objects by means of admission controller webhooks which are executed everytime a cluster component is created or modified.
[OPA](https://www.openpolicyagent.org/) (Open Policy Agent) is an open source, general-purpose policy engine to write unifed policies across different applications. It exploits [REGO](https://www.openpolicyagent.org/docs/latest/policy-language/) language to write policies.
[Gatekeeper](https://github.com/open-policy-agent/gatekeeper) allows you to integrate easily OPA on kubernetes by adding:
 - native kubernetes CRDs for instantiating the policy library (Constraint)
 - native kubernetes CRDs for extending the policy library (ConstraintTemplate)


## HOW TO INSTALL
### Prerequisites 

- minimun kubernetes version: 1.14
- make sure you have cluster admin permissions
#### Deploying a Release using Prebuilt Image
Run the following command:

    kubectl apply -f https://raw.githubusercontent.com/open-policy-agent/gatekeeper/release-3.3/deploy/gatekeeper.yaml
For more information visit the [gatekeeper official repository](https://github.com/open-policy-agent/gatekeeper).

## HOW TO USE GATEKEEPER

The following pages contain some more [information](../../policies/README.md) and [examples](./examples) about how to use gatekeeper and OPA policy.
The examples are taken from the original [repository](https://open-policy-agent.github.io/gatekeeper/website/docs/howto/)