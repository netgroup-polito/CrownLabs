import ThemeSwitcher, { IThemeSwitcherProps } from './ThemeSwitcher';
import { Story, Meta } from '@storybook/react';
import ThemeContextProvider from '../../../contexts/ThemeContext';

export default {
  title: 'Components/Misc/ThemeSwitcher',
  component: ThemeSwitcher,
  argTypes: {
    onClick: { action: 'clicked' },
  },
  decorators: [
    (Story: Story) => {
      return (
        <ThemeContextProvider>
          <Story />
        </ThemeContextProvider>
      );
    },
  ],
} as Meta;

const Template: Story<IThemeSwitcherProps> = args => (
  <ThemeSwitcher {...args} />
);

export const Default = Template.bind({});
