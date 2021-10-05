import WorkspaceContainer, {
  IWorkspaceContainerProps,
} from './WorkspaceContainer';
import { Story, Meta } from '@storybook/react';
import { someKeysOf, WorkspaceRole } from '../../../utils';

export default {
  title: 'Components/workspaces/WorkspaceContainer',
  component: WorkspaceContainer,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceContainerProps> = {
  workspace: {
    id: 0,
    title: 'Reti Locali e Data Center',
    role: WorkspaceRole.manager,
    workspaceNamespace: 'workspaceNamespace',
    workspaceName: 'workspaceNamespace',
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
    role: WorkspaceRole.user,
    workspaceNamespace: 'workspaceNamespace',
    workspaceName: 'workspaceNamespace',
  },
};
