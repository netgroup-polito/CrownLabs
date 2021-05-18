import WorkspaceContainer, {
  IWorkspaceContainerProps,
} from './WorkspaceContainer';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/workspaces/WorkspaceContainer',
  component: WorkspaceContainer,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceContainerProps> = {
  workspace: {
    id: 0,
    title: 'Reti Locali e Data Center',
    role: 'manager',
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

const Template: Story<IWorkspaceContainerProps> = args => (
  <WorkspaceContainer {...args} />
);

export const Full = Template.bind({});

Full.args = defaultArgs;

export const Empty = Template.bind({});

Empty.args = {
  ...defaultArgs,
  workspace: {
    id: 8,
    title: 'Software Networking',
    role: 'user',
    templates: [],
  },
};
