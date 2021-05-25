import { IWorkspace, ITemplate } from './NestedTables/NestedTables';

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
