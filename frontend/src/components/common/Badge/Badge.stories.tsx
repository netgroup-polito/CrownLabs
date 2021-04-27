import Badge, { IBadgeProps } from './Badge';
import { Story, Meta } from '@storybook/react';
import { BadgeSize, someKeysOf } from '../../../utils';

export default {
  title: 'Components/common/Badge',
  component: Badge,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IBadgeProps> = {
  value: 5,
  size: 'middle' as BadgeSize,
};

const Template: Story<IBadgeProps> = args => <Badge {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
