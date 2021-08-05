import ThemeSwitcher, { IThemeSwitcherProps } from './ThemeSwitcher';
import { Story, Meta } from '@storybook/react';
import { CenterDecorator } from '../../../decorators/CenterDecorator';

export default {
  title: 'Components/Misc/ThemeSwitcher',
  component: ThemeSwitcher,
  argTypes: {
    onClick: { action: 'clicked' },
  },
  decorators: [CenterDecorator],
} as Meta;

const Template: Story<IThemeSwitcherProps> = args => (
  <ThemeSwitcher {...args} />
);

export const Default = Template.bind({});
