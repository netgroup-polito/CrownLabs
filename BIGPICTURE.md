### Mermaid Legend:

```mermaid
%% ==== OVERVIEW GRAPH ====
flowchart LR


Domain["Logical Domain"]:::DomainStyle
Entity["Single Entity"]:::EntitySyle

Domain -- physical link --> Entity
Domain -. logical link .-> Entity

classDef DomainStyle rx:12,ry:12,fill:none,stroke-dasharray:8
classDef EntitySyle rx:12,ry:12,fill:none
```

# CrownLabs Big Picture
The following will be placed in **home**

```mermaid
flowchart LR

%% Cells
Argo["Argo/Helm¹"]
Releases["Deployment Releases¹"]
User@{shape: trap-b, label: "User Browser⁴"}

%% subgraphs
subgraph ReleasesNS["Deployment Releases Namespace"]
    Controllers@{shape: docs, label: "Controller Set¹"}
    Frontend["Frontend-app⁴"]:::FrontendStyle
    GraphQL["GraphQL Relay⁵"]
end
WorkspaceCR["`Workspace CR
    [cluster-wide]
    _workspace: abc_`"]:::CustomResourceStyle
Controllers -- manages --> WorkspaceCR
subgraph WorkspaceNS["`Workspace Namespace: _workspace-abc_ ³`"]
    direction LR
    TemplateCR["`Template CR
        template: _foo_`"]:::CustomResourceStyle
end
TenantCR["`Tenant CR
    [cluster-wide]
    _tenant: xyz-efg_`"]:::CustomResourceStyle
Controllers -- manages --> TenantCR
subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_ ²`"]
    InstanceCR["`Instance CR
        _instance: bar_`"]:::CustomResourceStyle
    InstanceEnv@{shape: docs, label: "bar Environments"}

    InstanceCR --> InstanceEnv
end
subgraph K8S["K8S Provided Infratructure"]
    APIServer["API Server"]
    GWAPI["Gateway (Balancer)¹"]
end

Argo .-> Releases
Releases -. deploys .-> ReleasesNS
Frontend -. exposed through .-> GWAPI
Frontend -. delivered to .-> User
GraphQL -. exposed through .-> GWAPI
GraphQL .-> APIServer
GWAPI ---> InstanceEnv
WorkspaceCR -- reconciles to --> WorkspaceNS
InstanceCR -. references to .-> TemplateCR
TenantCR -- reconciles to --> TenantNS
User --> GWAPI
User -. interacts with .-> GraphQL

classDef ControllerStyle stroke:#FF8000,rx:12,ry:12,fill:none
classDef WorkspaceStyle stroke:#0A6522,rx:12,ry:12,fill:none
classDef CustomResourceStyle stroke:#FF0000,rx:12,ry:12,fill:none
classDef TenantStyle stroke:#2222FF,rx:12,ry:12,fill:none
classDef EnvironmentStyle stroke:#A865B5,rx:12,ry:12,fill:none
classDef FrontendStyle stroke:#FDDCD7,rx:12,ry:12,fill:none

classDef GenericDomainStyle fill:none,rx:12,ry:12,stroke-dasharray:8
classDef WorkspaceNSStyle stroke:#0A6522,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef TenantNSStyle stroke:#2222FF,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef ReleasesNSStyle stroke:#FF8000,rx:12,ry:12,fill:none,stroke-dasharray:8

class K8S GenericDomainStyle
class ReleasesNS ReleasesNSStyle
class Controllers ControllerStyle
class WorkspaceNS WorkspaceNSStyle
class TemplateEnv,InstanceEnv EnvironmentStyle
class TenantNS TenantNSStyle
```
¹More about Deployment+Gateway [here](BIGPICTURE.md#crownlabs-deployment).  
²More about Tenant [here](BIGPICTURE.md#tenant-business-logic).  
³More about Workspace [here](BIGPICTURE.md#workspace-business-logic).  
⁴More about Frontend [here](BIGPICTURE.md#frontend-logic)  
⁵More about QLKube  


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
subgraph Controllers["`Controller Set (_operator-selector = prod_)`"]
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

ArgoHelm --> ProdDeploy
ArgoHelm --> PreProdDeploy
ArgoHelm --> StagingDeploy
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

classDef ReleasesNSStyle stroke:#FF8000,rx:12,ry:12,fill:none,stroke-dasharray:8

class Controllers ReleasesNSStyle
```

```mermaid
flowchart LR

Gateway["`Gateway
    (Envoy Gateway API)`"]
Proxy["Envoy Proxy"]:::GenericDomainStyle
GatewayClass["GatewayClass"]:::GenericDomainStyle
GatewayController["Gateway Controller"]
APIServer["API Server"]
subgraph Routes["HTTPRoutes"]
    subgraph PublicRoutes["`Public HTTPRoutes (has label _crownlabs.polito.it/public-route: true_)`"]
    WebSSHRoute["`WebSSH HTTPRoute 
        _/webssh_`"]
    FrontendRoute["`Frontend-app HTTPRoute
        _/_`"]
    QLKubeRoute["`QLKube HTTPRoute
        _/graph_`"]
    end
    CallbackRoute["`callback HTTPRoute 
        _/app/instauth/callback_`"]
    OtherRoute["`another HTTPRoute
        (if Tenant has label 
        _crownlabs.polito.it/gw-access = crownlabs-main-production_)`"]:::TenantStyle
end
subgraph Backends["Backend Services"]
    WebSSHService["WebSSH Service"]
    FrontendService["Frontend-app Service"]
    QLKubeService["QLKube Service"]
    OtherService["another Service"]:::TenantStyle
    DummyService["dummy Service"]
    VM@{shape: docs, label: "VMs"}
end
subgraph SecPols["Security Policies"]
    SecPolG["Security Policy"]
    SecPolDummy["`Security Policy
        (dummy, only if route has label 
        _crownlabs.polito.it/public-route: true_)`"]
end
Traffic["Request Traffic"]:::FrontendStyle

Gateway -- manages --> Proxy
Gateway -. accepts .-> Routes
Gateway -. authenticates .-> SecPols
Traffic .-> Proxy .-> VM

WebSSHRoute --> WebSSHService
FrontendRoute --> FrontendService
QLKubeRoute --> QLKubeService
OtherRoute --> OtherService
CallbackRoute --> DummyService

WebSSHService --> VM
FrontendService --> VM
QLKubeService --> VM
OtherService --> VM

GatewayController -. watches .-> APIServer
GatewayController -. creates/updates .-> Gateway
GatewayController -. creates/updates .-> Routes

APIServer --> Gateway
GatewayClass -. used by .-> Gateway

classDef GenericDomainStyle fill:none,rx:12,ry:12,stroke-dasharray:8
classDef TenantStyle stroke:#2222FF,rx:12,ry:12,fill:none
classDef FrontendStyle stroke:#FDDCD7,rx:12,ry:12,fill:none

class Routes,PublicRoutes,SecPols,Backends GenericDomainStyle
class VM TenantStyle
```

## Frontend Logic
The following will be placed in **home/frontend-app**

```mermaid
flowchart LR

Home["`Home Page
    _crownlabs.polito.it_`"]:::FrontendStyle
Login["Login Page"]:::FrontendStyle
Frontend["`Front-end
    _crownlabs.polito.it/app_`"]:::FrontendStyle
Remote["Remote Desktop"]
OIDC["`OIDC Server
    (Authentication)`"]
Graph["`GraphQL
    _crownlabs.polito.it/graph_`"]:::FrontendStyle
GWAPI["Gateway (Balancer)"]

Home --> Login
Login --> Frontend
Login --> OIDC
Frontend --> Graph
Frontend -. connects .-> Remote
Frontend -. connects .-> GWAPI
Remote --> GWAPI

classDef FrontendStyle stroke:#FDDCD7,rx:12,ry:12,fill:none
```

## Workspace Business Logic
The following will be placed in **home/operators/operator/workspace**

```mermaid
flowchart LR

WorkspaceController["Workspace Controller"]:::ControllerStyle
WorkspaceCR["`Workspace CR
    _workspace: abc_`"]:::CustomResourceStyle
subgraph WorkspaceNS["`Workspace Namespace: _workspace-abc_`"]
    subgraph TemplateFoo["`Template: _foo_`"]
        EnvContainer["`Environment 1
            _(container-based)_`"]:::EnvironmentStyle
        EnvVM["`Environment 2
            _(vm-based)_`"]:::EnvironmentStyle
    end

    TemplateOther["`Another Template
        _(does nothing if not referenced in an instance)_`"]
end

WorkspaceController -- reconciles --> WorkspaceCR
WorkspaceCR --> WorkspaceNS

classDef ControllerStyle stroke:#FF8000,rx:12,ry:12,fill:none
classDef TemplateStyle stroke:#0A6522,rx:12,ry:12,fill:none
classDef CustomResourceStyle stroke:#FF0000,rx:12,ry:12,fill:none
classDef EnvironmentStyle stroke:#A865B5,rx:12,ry:12,fill:none
classDef WorkspaceNSStyle stroke:#0A6522,rx:12,ry:12,fill:none,stroke-dasharray:8

class WorkspaceNS WorkspaceNSStyle
class TemplateFoo,TemplateOther TemplateStyle
```

## Tenant Business Logic
The following will be placed in **home/operators/operator/tenant**

```mermaid
flowchart LR

TenantController["Tenant Controller"]:::ControllerStyle
BastionController["Bastion Controller"]:::ControllerStyle
SharedVolumeController["Shared Volume Controller"]:::ControllerStyle

TenantCR["`Tenant CR
_tenant: xyz.efg_`"]:::CustomResourceStyle
subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_ (has label for _operator-selector=production_)`"]
    InstanceCR["`Instance CR
        _instance: bar_`"]:::CustomResourceStyle
    ShVol["Shared Volume"]
    ContainerEnv["Container-based Environment"]:::EnvironmentStyle
    VMEnv["VM-based Environment"]:::EnvironmentStyle
end

subgraph TemplateFoo["`Template: _foo_`"]
    EnvContainer["`Environment 1
        _(container-based)_`"]:::EnvironmentStyle
    EnvVM["`Environment 2
        _(vm-based)_`"]:::EnvironmentStyle
end

TenantController -- reconciles --> TenantCR
BastionController -. watches .-> TenantCR
TenantCR -- manages --> TenantNS
SharedVolumeController -- reconciles --> ShVol
InstanceCR -. references .-> TemplateFoo
EnvContainer -. generates .-> ContainerEnv
EnvVM -. generates .-> VMEnv
ContainerEnv -. attaches .-> ShVol
VMEnv -. attaches .-> ShVol

classDef ControllerStyle stroke:#FF8000,rx:12,ry:12,fill:none
classDef CustomResourceStyle stroke:#FF0000,rx:12,ry:12,fill:none
classDef EnvironmentStyle stroke:#A865B5,rx:12,ry:12,fill:none
classDef WorkspaceNSStyle stroke:#0A6522,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef TenantNSStyle stroke:#2222FF,rx:12,ry:12,fill:none,stroke-dasharray:8

class TenantNS TenantNSStyle
class TemplateFoo WorkspaceNSStyle
```

## Instances Deep Dive
The following will be placed in **home/operators/.../instctl**

```mermaid
flowchart LR

InstanceController["Instance Controller"]:::ControllerStyle
InstanceAutomationController["Instance Automation Controller"]:::ControllerStyle
InstanceSnapshotController["Instance Snapshot Controller"]:::ControllerStyle

subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_`"]
    InstanceCR["`Instance CR
        _instance: bar_`"]:::CustomResourceStyle
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

    VirtualMachine -- becomes --> VirtLauncherPod
    InstanceCR -- becomes --> Env1ContainerInst
    InstanceCR -- becomes --> Env2VMInst
end
GWAPI["Load Balancer + Gateway/Ingress"]

InstanceController -- reconciles --> InstanceCR
InstanceAutomationController -- reconciles --> InstanceCR
InstanceSnapshotController -- reconciles --> InstanceCR
ExpositionEnv2 -. routes .-> GWAPI
ExpositionEnv1 -. routes .-> GWAPI

classDef ControllerStyle stroke:#FF8000,rx:12,ry:12,fill:none
classDef CustomResourceStyle stroke:#FF0000,rx:12,ry:12,fill:none
classDef EnvironmentStyle stroke:#A865B5,rx:12,ry:12,fill:none
classDef WorkspaceNSStyle stroke:#0A6522,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef TenantNSStyle stroke:#2222FF,rx:12,ry:12,fill:none,stroke-dasharray:8

class TenantNS TenantNSStyle
class Env1ContainerInst,Env2VMInst EnvironmentStyle
```