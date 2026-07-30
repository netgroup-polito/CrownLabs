### Mermaid Legend:

```mermaid
%% ==== OVERVIEW GRAPH ====
flowchart TB

LogicalDomain
Controller
Workspace
CRD
Tenant
Environment

%% General Class Definitions - Border Styles (BS)
classDef GenericDashedBS fill:none,rx:12,ry:12,stroke-dasharray:8
classDef ControllerBS stroke:#E45756,rx:12,ry:12,fill:none
classDef WorkspaceBS stroke:#54724B,rx:12,ry:12,fill:none
classDef CustomResourceBS stroke:#FF0000,rx:12,ry:12,fill:none
classDef TenantBS stroke:#2222FF,rx:12,ry:12,fill:none
classDef EnvironmentBS stroke:#BB55BB,rx:12,ry:12,fill:none

%% Class Applications
class LogicalDomain GenericDashedBS
class Controller ControllerBS
class Workspace WorkspaceBS
class CRD CustomResourceBS
class Tenant TenantBS
class Environment EnvironmentBS
```

# CrownLabs Big Picture
The following will be placed in **home**

```mermaid
flowchart LR

Argo["Argo/Helm¹"]
Releases["Deployment Releases¹"]
subgraph ReleasesNS["Deployment Releases Namespace"]
    Controllers@{shape: docs, label: "Controller Set¹"}
    Frontend["Frontend-app⁴"]
    GraphQL["GraphQL Relay⁵"]

    Frontend .-> GraphQL
end
TenantCR["`Tenant CR
    [cluster-wide]
    _tenant: xyz-efg_`"]
Controllers --> TenantCR
subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_ ²`"]
    InstanceCR["`Instance CR
        _instance: bar_`"]
    InstanceEnv@{shape: docs, label: "bar Environments"}

    InstanceCR --> InstanceEnv
end
subgraph K8S["K8S Provided Infratructure"]
    APIServer["API Server"]
    GWAPI["Load Balancer + Gateway/Ingress¹"]
end
RBAC["RBACs"]
WorkspaceCR["`Workspace CR
    [cluster-wide]
    _workspace: abc_`"]
Controllers --> WorkspaceCR
subgraph WorkspaceNS["`Workspace Namespace: _workspace-abc_ ³`"]
    TemplateCR["`Template CR
        template: _foo_`"]
    TemplateEnv@{shape: docs, label: "foo Environments"}

    TemplateCR --> TemplateEnv
end

Argo .-> Releases
Releases -. hosts .-> Controllers
Frontend -. login .-> TenantNS
Frontend -. connects to .-> GWAPI
GraphQL .-> APIServer
GWAPI .-> InstanceEnv
WorkspaceCR --> TemplateCR
InstanceCR -. references to .-> TemplateCR
TenantCR --> InstanceCR
APIServer --> RBAC

classDef GenericDashedBS fill:none,rx:12,ry:12,stroke-dasharray:8
classDef ControllerDashedBS stroke:#E45756,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef ControllerBS stroke:#E45756,rx:12,ry:12,fill:none
classDef WorkspaceDashedBS stroke:#54724B,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef CustomResourceBS stroke:#FF0000,rx:12,ry:12,fill:none
classDef TenantDashedBS stroke:#2222FF,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef InstanceBS stroke:#BB55BB,rx:12,ry:12,fill:none

class K8S GenericDashedBS
class ReleasesNS ControllerDashedBS
class Controllers ControllerBS
class WorkspaceCR,TenantCR,InstanceCR,TemplateCR CustomResourceBS
class WorkspaceNS WorkspaceDashedBS
class TenantNS TenantDashedBS
class InstanceEnv,TemplateEnv InstanceBS
```
¹More about this [here](BIGPICTURE.md#crownlabs-deployment).  
²More about this [here](BIGPICTURE.md#tenant-business-logic).  
³More about this [here](BIGPICTURE.md#workspace-business-logic).  
⁴More about this  
⁵More about this  


## CrownLabs Deployment Releases
The following will be placed in **home/deploy/crownlabs**

```mermaid
flowchart LR

ArgoHelm["Helm / ArgoCD"] 
ProdDeploy["`_crownlabs-production_
    Production`"]
PreProdDeploy["`_crownlabs-pre-production_
    Pre-production`"]
StagingDeploy["`_crownlabs-staging-123_
    Staging (PR-associated)`"]

ArgoHelm --> ProdDeploy
ArgoHelm --> PreProdDeploy
ArgoHelm --> StagingDeploy

subgraph Controllers["`Controllers (_operator-selector = production_)`"]
    TenantController["Tenant Controller"]
    InstanceController["Instance Controller"]
    InstanceAutomationController["Instance Automation Controller"]
    InstanceSnapshotController["Instance Snapshot Controller"]
    WorkspaceController["Workspace Controller"]
    BastionController["SSH Bastion"]
    WebSSHController["WebSSH"]
    FrontendController["CrownLabs Frontend"]
    QLKubeController["QLKube"]
    ShVolController["Shared Volume Controller"]
end

ProdDeploy -. hosts .-> WorkspaceController
ProdDeploy -. hosts .-> TenantController
ProdDeploy -. hosts .-> InstanceController
ProdDeploy -. hosts .-> InstanceAutomationController
ProdDeploy -. hosts .-> InstanceSnapshotController
ProdDeploy -. hosts .-> BastionController
ProdDeploy -. hosts .-> WebSSHController
ProdDeploy -. hosts .-> FrontendController
ProdDeploy -. hosts .-> QLKubeController
ProdDeploy -. hosts .-> ShVolController

classDef ControllerBS stroke:#E45756,rx:12,ry:12,fill:none

class Controllers ControllerBS
```

## Workspace Business Logic
The following will be placed in **home/operators/operator/workspace**

```mermaid
flowchart LR

WorkspaceController["Workspace Controller"]
WorkspaceCR["`Workspace CR
_workspace: abc_`"]

WorkspaceController -- reconciles --> WorkspaceCR

subgraph WorkspaceNS["`Workspace Namespace: _workspace-abc_`"]
    subgraph TemplateFoo["`Template: _foo_`"]
        EnvContainer["`Environment 1
            _(container-based)_`"]
        EnvVM["`Environment 2
            _(vm-based)_`"]
    end

    TemplateOther["`Another Template
    _(does nothing if not referenced in an instance)_`"]
end

WorkspaceCR --> WorkspaceNS

classDef ControllerBS stroke:#E45756,rx:12,ry:12,fill:none
classDef WorkspaceBS stroke:#54724B,rx:12,ry:12,fill:none
classDef WorkspaceDashedBS stroke:#54724B,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef CustomResourceBS stroke:#FF0000,rx:12,ry:12,fill:none
classDef EnvironmentBS stroke:#BB55BB,rx:12,ry:12,fill:none

class WorkspaceNS WorkspaceDashedBS
class TemplateFoo,TemplateOther WorkspaceBS
class WorkspaceController ControllerBS
class WorkspaceCR CustomResourceBS
class EnvContainer,EnvVM EnvironmentBS
```

## Tenant Business Logic
The following will be placed in **home/operators/operator/tenant**

```mermaid
flowchart LR

TenantController["Tenant Controller"]
TenantCR["`Tenant CR
_tenant: xyz.efg_`"]

TenantController -- reconciles --> TenantCR
BastionController -. watches .-> TenantCR

subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_ (has label for _operator-selector=production_)`"]
    subgraph Env1ContainerInst["`Instanced Environment 1 _(container-based)_`"]
        ExpositionEnv1["Ingress / HTTPRoute"]
        ServiceEnv1["Service"]
        DeploymentEnv1["Deployment"]

        DeploymentEnv1 -. exposes .-> ServiceEnv1
        ServiceEnv1 -. routes .-> ExpositionEnv1
    end

    subgraph Env2VMInst["`Instanced Environment 2 _(VM-based)_`"]
        ExpositionEnv2["Ingress / HTTPRoute"]
        ServiceEnv2["Service"]
        VirtualMachine["VirtualMachine"]

        VirtualMachine -. exposes .-> ServiceEnv2
        ServiceEnv2 -. routes .-> ExpositionEnv2
    end

    VirtLauncherPod["KubeVirt VirtLauncher Pod"]
    InstanceCR["`Instance CR
        _instance: bar_`"]
    ShVol["Shared Volume"]

    VirtualMachine -- becomes --> VirtLauncherPod
    InstanceCR -- becomes --> Env1ContainerInst
    InstanceCR -- becomes --> Env2VMInst

    Env1ContainerInst -. attaches .-> ShVol
    Env2VMInst -. attaches .-> ShVol
end

subgraph TemplateFoo["`Template: _foo_`"]
    EnvContainer["`Environment 1
        _(container-based)_`"]
    EnvVM["`Environment 2
        _(vm-based)_`"]
end 

EnvContainer -. configures .-> DeploymentEnv1
EnvVM -. configures .-> VirtualMachine
InstanceCR -. references .-> TemplateFoo

SharedVolumeController -- reconciles --> ShVol

TenantCR -- manages --> TenantNS
InstanceController -- reconciles --> InstanceCR
InstanceAutomationController -- reconciles --> InstanceCR
InstanceSnapshotController -- reconciles --> InstanceCR 

GWAPI["Load Balancer + Gateway/Ingress"]
ExpositionEnv2 -. routes .-> GWAPI
ExpositionEnv1 -. routes .-> GWAPI

classDef GenericDashedBS fill:none,rx:12,ry:12,stroke-dasharray:8
classDef ControllerBS stroke:#E45756,rx:12,ry:12,fill:none
classDef WorkspaceBS stroke:#54724B,rx:12,ry:12,fill:none
classDef CustomResourceBS stroke:#FF0000,rx:12,ry:12,fill:none
classDef TenantBS stroke:#2222FF,rx:12,ry:12,fill:none
classDef EnvironmentBS stroke:#BB55BB,rx:12,ry:12,fill:none

classDef TenantBorderStyle fill:none,stroke:#5472FB,rx:12,ry:12
class TenantNS,Env1ContainerInst,Env2VMInst TenantBorderStyle
classDef ControllerBorderStyle stroke:#E45756
class InstanceController,InstanceAutomationController,InstanceSnapshotController,SharedVolumeController,TenantController,BastionController ControllerBorderStyle
classDef CRBorderStyle stroke:#FF0000
class TenantCR,InstanceCR CRBorderStyle
classDef WorkspaceBorderStyle fill:none,stroke:#54724B,stroke-width:2px,rx:12,ry:12
class TemplateFoo WorkspaceBorderStyle
```
