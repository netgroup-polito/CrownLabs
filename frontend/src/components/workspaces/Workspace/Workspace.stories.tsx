import Workspace, { IWorkspaceProps } from './Workspace';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/Workspace',
  component: Workspace,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceProps> = {
  workspace: {
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
};

const Template: Story<IWorkspaceProps> = args => <Workspace {...args} />;

export const LANDC = Template.bind({});

LANDC.args = defaultArgs;

export const TSR = Template.bind({});

TSR.args = {
  ...defaultArgs,
  workspace: {
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
};

export const EMPTY = Template.bind({});

EMPTY.args = {
  ...defaultArgs,
  workspace: {
    id: 8,
    title: 'Software Networking',
    templates: [],
  },
};
