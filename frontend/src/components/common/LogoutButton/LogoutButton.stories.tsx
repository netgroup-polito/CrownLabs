import LogoutButton, { ILogoutButtonProps } from './LogoutButton';
import { Story, Meta } from '@storybook/react';

export default {
  title: 'Components/common/LogoutButton',
  component: LogoutButton,
  argTypes: {
    largeMode: { table: { disable: true } },
    logoutHandler: { table: { disable: true } },
  },
  decorators: [
    (Story: Story) => (
      <div className="flex h-screen items-center justify-center">
        <Story />
      </div>
    ),
  ],
} as Meta;

const Template: Story<ILogoutButtonProps> = args => <LogoutButton {...args} />;

export const Default = Template.bind({});

Default.args = {
  logoutHandler: () => null,
  iconStyle: { fontSize: '200px' },
};

export const LargeMode = Template.bind({});

LargeMode.args = {
  logoutHandler: () => null,
  largeMode: true,
};
