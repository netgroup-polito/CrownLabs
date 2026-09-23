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

User@{shape: trap-b, label: "User Browser⁴"}

subgraph Cluster["Kubernetes Cluster"]
    %% Cells
    Argo["Argo/Helm¹"]
    Release["CL Argo Application¹"]

    %% subgraphs
    subgraph ReleaseNS["CL Release Namespace¹"]
        Controllers@{shape: docs, label: "Controller Set¹"}
        Frontend["Frontend-app⁴"]:::FrontendStyle
        GraphQL["GraphQL Relay⁵"]
        Bastion["SSH Bastion"]
    end
    WorkspaceCR["`Workspace CR
        [cluster-wide]
        _workspace: abc_`"]:::CustomResourceStyle
    Controllers -- manages --> WorkspaceCR
    subgraph WorkspaceNS["`Workspace Namespace:
    _workspace-abc_ ³`"]
        direction LR
        TemplateCR["`Template CR
            template: _foo_`"]:::CustomResourceStyle
    end
    TenantCR["`Tenant CR
        [cluster-wide]
        _tenant: xyz-efg_`"]:::CustomResourceStyle
    subgraph TenantNS["`Tenant Namespace: _tenant-xyz-efg_ ²`"]
        InstanceCR["`Instance CR
            _instance: bar_`"]:::CustomResourceStyle
        InstanceEnv@{shape: docs, label: "bar Environments"}

    end
    subgraph K8S["K8S Provided Infratructure"]
        APIServer["API Server"]
        GWAPI["Gateway Balancer"]
    end
end

Controllers -- manages --> TenantCR
Argo .-> Release
Release -. becomes .-> ReleaseNS
Frontend -. exposed through .-> GWAPI
Frontend -. delivered to .-> User
GraphQL .-> APIServer
GWAPI ---> InstanceEnv
WorkspaceCR -- reconciles to --> WorkspaceNS
TenantCR -- reconciles to --> TenantNS
User --> GWAPI
User -. interacts with .-> GraphQL
User -. interacts with .-> InstanceEnv
InstanceCR -. references to ..-> TemplateCR
InstanceCR --> InstanceEnv
GraphQL -. exposed through .-> GWAPI
User --> Bastion
Bastion  --> InstanceEnv

classDef ControllerStyle stroke:#FF8000,rx:12,ry:12,fill:none
classDef WorkspaceStyle stroke:#0A6522,rx:12,ry:12,fill:none
classDef CustomResourceStyle stroke:#FF0000,rx:12,ry:12,fill:none
classDef TenantStyle stroke:#2222FF,rx:12,ry:12,fill:none
classDef EnvironmentStyle stroke:#A865B5,rx:12,ry:12,fill:none
classDef FrontendStyle stroke:#FDDCD7,rx:12,ry:12,fill:none
classDef ClusterStyle stroke:#80FF00,rx:12,ry:12,fill:none

classDef GenericDomainStyle fill:none,rx:12,ry:12,stroke-dasharray:8
classDef WorkspaceNSStyle stroke:#0A6522,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef TenantNSStyle stroke:#2222FF,rx:12,ry:12,fill:none,stroke-dasharray:8
classDef ReleaseNSStyle stroke:#FF8000,rx:12,ry:12,fill:none,stroke-dasharray:8

class K8S GenericDomainStyle
class ReleaseNS ReleaseNSStyle
class Controllers ControllerStyle
class WorkspaceNS WorkspaceNSStyle
class TemplateEnv,InstanceEnv EnvironmentStyle
class TenantNS TenantNSStyle
class Cluster ClusterStyle
```
¹More about Deployment+Gateway [here](BIGPICTURE.md#crownlabs-deployment).  
²More about Tenant [here](BIGPICTURE.md#tenant-business-logic).  
³More about Workspace [here](BIGPICTURE.md#workspace-business-logic).  
⁴More about Frontend [here](BIGPICTURE.md#frontend-logic)  
⁵More about QLKube  


