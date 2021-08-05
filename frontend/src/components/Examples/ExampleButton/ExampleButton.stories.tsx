import ExampleButton, { IExampleButtonProps } from './ExampleButton';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/Examples/ExampleButton',
  component: ExampleButton,
} as Meta;

const defaultArgs: someKeysOf<IExampleButtonProps> = {
  text: 'Example Button',
  disabled: false,
  specialCSS: false,
};

const Template: Story<IExampleButtonProps> = args => (
  <ExampleButton {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;

export const CustomCSS = Template.bind({});

CustomCSS.args = { ...defaultArgs, specialCSS: true };

export const Disabled = Template.bind({});

Disabled.args = { ...defaultArgs, disabled: true };

export const Large = Template.bind({});

Large.args = { ...defaultArgs, size: 'large' };
