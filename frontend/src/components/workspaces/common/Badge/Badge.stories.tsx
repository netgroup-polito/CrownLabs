import Badge, { IBadgeProps } from './Badge';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/Badge',
  component: Badge,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IBadgeProps> = {
  value: 5,
};

const Template: Story<IBadgeProps> = args => <Badge {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
