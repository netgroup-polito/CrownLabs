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
    },
    {
      role: 'user',
      workspaceId: 'Workspace 2',
    },
    {
      role: 'user',
      workspaceId: 'Workspace 3',
    },
    {
      role: 'manager',
      workspaceId: 'Workspace 4',
    },
    {
      role: 'user',
      workspaceId: 'Workspace 5',
    },
  ],
};

const Template: Story<IDashboardProps> = args => <Dashboard {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
