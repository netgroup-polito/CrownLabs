const data = [
  {
    id: 0,
    title: 'Reti Locali e Data Center',
    templates: [
      {
        id: '0_1',
        name: 'Ubuntu VM',
        gui: true,
        instances: [
          { id: 1, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 2, name: 'Ubuntu VM', ip: '192.168.0.1', status: false },
        ],
      },
      { id: '0_2', name: 'Ubuntu VM', gui: false, instances: [] },
      {
        id: '0_3',
        name: 'Windows VM',
        gui: true,
        instances: [
          { id: 1, name: 'Windows VM', ip: '192.168.0.1', status: true },
        ],
      },
      { id: '0_4', name: 'Console (Linux)', gui: false, instances: [] },
      {
        id: '0_5',
        name: 'Ubuntu VM',
        gui: true,
        instances: [
          { id: 1, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 2, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 3, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
        ],
      },
      {
        id: '0_6',
        name: 'Ubuntu VM',
        gui: true,
        instances: [
          { id: 1, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 2, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 3, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
        ],
      },
    ],
  },
  {
    id: 1,
    title: 'Tecnologie e Servizi di Rete',
    templates: [
      {
        id: '1_1',
        name: 'Ubuntu VM',
        gui: true,
        instances: [
          { id: 1, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 2, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
          { id: 3, name: 'Ubuntu VM', ip: '192.168.0.1', status: true },
        ],
      },
      { id: '1_2', name: 'Ubuntu VM', gui: false, instances: [] },
      { id: '1_3', name: 'Windows VM', gui: true, instances: [] },
    ],
  },
  {
    id: 2,
    title: 'Applicazioni Web I',
    templates: [
      { id: '2_1', name: 'Ubuntu VM', gui: true, instances: [] },
      { id: '2_2', name: 'Windows VM', gui: true, instances: [] },
      { id: '2_3', name: 'Console (Linux)', gui: false, instances: [] },
    ],
  },
  {
    id: 3,
    title: 'Cloud Computing',
    templates: [
      { id: '3_1', name: 'Console (Linux)', gui: false, instances: [] },
    ],
  },
  {
    id: 4,
    title: 'Programmazione di Sistema',
    templates: [],
  },
  {
    id: 5,
    title: 'Information System Security',
    templates: [],
  },
  {
    id: 6,
    title: 'Ingegneria del Software',
    templates: [],
  },
  {
    id: 7,
    title: 'Data Science',
    templates: [],
  },
  {
    id: 8,
    title: 'Software Networking',
    templates: [],
  },
];

export default data;
