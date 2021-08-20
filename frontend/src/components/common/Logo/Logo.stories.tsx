import Logo, { ILogoProps } from './Logo';
import { Story, Meta } from '@storybook/react';
import { LogoDecorator } from '../../../Decorators';

export default {
  title: 'Components/common/Logo',
  component: Logo,
  decorators: [LogoDecorator],
  argTypes: {
    widthPx: {
      control: { type: 'range', min: 0, max: 800, step: 1 },
    },
  },
} as Meta;

const Template: Story<ILogoProps> = args => <Logo {...args} />;

export const Default = Template.bind({});
Default.argTypes = {
  color: { table: { disable: true } },
};
Default.args = {
  widthPx: 300,
};

export const NoWidth = Template.bind({});
NoWidth.argTypes = {
  color: { table: { disable: true } },
};

export const FixedColor = Template.bind({});
FixedColor.args = {
  widthPx: 300,
  color: 'rgba(255, 0, 0, 1)',
};
