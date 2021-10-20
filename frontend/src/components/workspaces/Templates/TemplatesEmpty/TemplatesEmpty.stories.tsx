import TemplatesEmpty, { ITemplatesEmptyProps } from './TemplatesEmpty';
import { Story, Meta } from '@storybook/react';
import { someKeysOf, WorkspaceRole } from '../../../../utils';

export default {
  title: 'Components/workspaces/Templates/TemplatesEmpty',
  component: TemplatesEmpty,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<ITemplatesEmptyProps> = {
  role: WorkspaceRole.manager,
};

const Template: Story<ITemplatesEmptyProps> = args => (
  <TemplatesEmpty {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
