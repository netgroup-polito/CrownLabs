import Example, { IExampleProps } from './Example';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../utils';

export default {
  title: 'Components/Example',
  component: Example,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IExampleProps> = {
  text: 'Example',
  disabled: false,
  specialCSS: false,
};

const Template: Story<IExampleProps> = args => <Example {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;

export const CustomCSS = Template.bind({});

CustomCSS.args = { ...defaultArgs, specialCSS: true };

export const Disabled = Template.bind({});

Disabled.args = { ...defaultArgs, disabled: true };

export const Large = Template.bind({});

Large.args = { ...defaultArgs, size: 'large' };
