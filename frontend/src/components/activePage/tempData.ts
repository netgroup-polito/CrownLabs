import { IWorkspace, ITemplate } from './NestedTables/NestedTables';
import { IInstance } from './Instances/InstancesTable/InstancesTable';

export const templates: Array<ITemplate> = [
  {
    key: 'LANDC-labs',
    id: 'LANDC-labs',
    name: 'Laboratorio di Reti Locali e Data Center',
    workspace: '01SQOOV',
    nActiveInstances: 3,
  },
  {
    key: 'TSR-labs',
    id: 'TSR-labs',
    name: 'Laboratorio di Tecnologie e Servizi di Rete',
    workspace: '02KPNOV',
    nActiveInstances: 2,
  },
  {
    key: 'LANDC-project',
    id: 'LANDC-project',
    name: 'Progetto Design Rete Aziendale',
    workspace: '01SQOOV',
    nActiveInstances: 1,
  },
];

export const workspaces: Array<IWorkspace> = [
  {
    key: '01SQOOV',
    id: '01SQOOV',
    name: 'Reti Locali e Data Center',
    templates: [templates[0], templates[2]],
  },
  {
    key: '02KPNOV',
    id: '02KPNOV',
    name: 'Tecnologie e Servizi di Rete',
    templates: [templates[1]],
  },
];

export const instances: Array<IInstance> = [
  {
    id: 'inst1',
    templateId: 'LANDC-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'ready',
    ip: '192,168.1.1',
    cliOnly: false,
  },
  {
    id: 'inst2',
    templateId: 'LANDC-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'ready',
    ip: '192,168.1.1',
    cliOnly: true,
  },
  {
    id: 'inst3',
    templateId: 'LANDC-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'ready',
    ip: '192,168.1.1',
    cliOnly: false,
  },
  {
    id: 'inst4',
    templateId: 'LANDC-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'creating',
    ip: '192,168.1.1',
    cliOnly: true,
  },
  {
    id: 'inst5',
    templateId: 'LANDC-project',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'failed',
    ip: '192,168.1.1',
    cliOnly: true,
  },
  {
    id: 'inst6',
    templateId: 'TSR-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'stopping',
    ip: '192,168.1.1',
    cliOnly: false,
  },
  {
    id: 'inst7',
    templateId: 'TSR-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'off',
    ip: '192,168.1.1',
    cliOnly: false,
  },
  {
    id: 'inst8',
    templateId: 'TSR-labs',
    tenantId: 's123456',
    tenantDisplayName: 'Name Surname',
    displayName: 'VM Name',
    phase: 'off',
    ip: '192,168.1.1',
    cliOnly: false,
  },
];
