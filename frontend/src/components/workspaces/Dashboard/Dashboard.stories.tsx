import Dashboard, { IDashboardProps } from './Dashboard';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/workspaces/Dashboard',
  component: Dashboard,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IDashboardProps> = {
  workspaces: [
    {
      role: 'manager',
      workspaceId: 'Workspace 1',
      workspaceNamespace: 'w1',
    },
    {
      role: 'user',
      workspaceId: 'Workspace 2',
      workspaceNamespace: 'w2',
    },
    {
      role: 'user',
      workspaceId: 'Workspace 3',
      workspaceNamespace: 'w3',
    },
    {
      role: 'manager',
      workspaceId: 'Workspace 4',
      workspaceNamespace: 'w4',
    },
    {
      role: 'user',
      workspaceId: 'Workspace 5',
      workspaceNamespace: 'w5',
    },
  ],
};

const Template: Story<IDashboardProps> = args => <Dashboard {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
