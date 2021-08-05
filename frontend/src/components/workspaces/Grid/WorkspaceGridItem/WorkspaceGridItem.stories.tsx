import WorkspaceGridItem, {
  IWorkspaceGridItemProps,
} from './WorkspaceGridItem';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/workspaces/Grid/WorkspaceGridItem',
  component: WorkspaceGridItem,
  argTypes: {
    onClick: { table: { disable: true } },
  },
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceGridItemProps> = {
  title: 'Reti locali e data center',
  isActive: true,
  id: 0,
};

const Template: Story<IWorkspaceGridItemProps> = args => (
  <WorkspaceGridItem {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
