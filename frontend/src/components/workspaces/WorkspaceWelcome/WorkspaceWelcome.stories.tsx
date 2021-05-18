import WorkspaceWelcome, { IWorkspaceWelcomeProps } from './WorkspaceWelcome';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/workspaces/Welcome',
  component: WorkspaceWelcome,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceWelcomeProps> = {};

const Template: Story<IWorkspaceWelcomeProps> = args => <WorkspaceWelcome />;

export const Default = Template.bind({});

Default.args = defaultArgs;
