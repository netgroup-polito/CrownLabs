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
Keycloak["`Keycloak
    (OIDC Authentication)`"]
Graph["`GraphQL
    _crownlabs.polito.it/graph_`"]:::FrontendStyle
GWAPI["Gateway (Balancer)"]

Home --> Login
Login -. OIDC sign-in request .-> Keycloak
Keycloak -. redirects back with authorization code .-> Frontend
Frontend -. exchanges code for tokens .-> Keycloak
Frontend -. GraphQL requests with access token .-> Graph
Graph -. routes through .-> GWAPI
Frontend -. connects .-> Remote
Frontend -. connects .-> GWAPI
Remote --> GWAPI

classDef FrontendStyle stroke:#FDDCD7,rx:12,ry:12,fill:none
```

## Workspace Business Logic
The following will be placed in **home/operators/operator/workspace**

```mermaid
flowchart LR

Keyclock
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
WorkspaceController -- groups provisioning --> Keyclock

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