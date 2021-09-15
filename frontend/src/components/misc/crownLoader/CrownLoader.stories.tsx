import CrownLoader, { ICrownLoaderProps } from './CrownLoader';
import { Story, Meta } from '@storybook/react';

export default {
  title: 'Components/misc/CrownLoader',
  component: CrownLoader,
  argTypes: {
    color: {
      control: { type: 'color' },
    },
    size: {
      control: { type: 'text' },
    },
    duration: {
      control: { type: 'number' },
    },
  },
} as Meta;

const customArgs: ICrownLoaderProps = {
  color: '#1c7afd',
  duration: 3,
  size: '300',
};

const Template: Story<ICrownLoaderProps> = args => <CrownLoader {...args} />;

export const Default = Template.bind({});

export const Customized = Template.bind({});

Customized.args = customArgs;
