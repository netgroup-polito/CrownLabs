import Dashboard, { IDashboardProps } from './Dashboard';
import { Story, Meta } from '@storybook/react';
import { someKeysOf, WorkspaceRole } from '../../../utils';

export default {
  title: 'Components/workspaces/Dashboard',
  component: Dashboard,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IDashboardProps> = {
  workspaces: [
    {
      role: WorkspaceRole.manager,
      workspaceId: 'Workspace 1',
      workspaceNamespace: 'w1',
      workspaceName: 'w1',
    },
    {
      role: WorkspaceRole.user,
      workspaceId: 'Workspace 2',
      workspaceNamespace: 'w2',
      workspaceName: 'w2',
    },
    {
      role: WorkspaceRole.user,
      workspaceId: 'Workspace 3',
      workspaceNamespace: 'w3',
      workspaceName: 'w3',
    },
    {
      role: WorkspaceRole.user,
      workspaceId: 'Workspace 4',
      workspaceNamespace: 'w4',
      workspaceName: 'w4',
    },
    {
      role: WorkspaceRole.user,
      workspaceId: 'Workspace 5',
      workspaceNamespace: 'w5',
      workspaceName: 'w5',
    },
  ],
};

const Template: Story<IDashboardProps> = args => <Dashboard {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
