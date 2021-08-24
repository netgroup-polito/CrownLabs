import Navbar, { INavbarProps } from './Navbar';
import { Story, Meta } from '@storybook/react';
import { NavbarDecorator } from '../../../Decorators';

export default {
  title: 'Components/common/Navbar',
  component: Navbar,
  decorators: [NavbarDecorator],
  argTypes: {
    logoutHandler: { table: { disable: true } },
  },
} as Meta;

const Template: Story<INavbarProps> = args => <Navbar {...args} />;

export const Default = Template.bind({});

Default.args = {
  logoutHandler: () => null,
  routes: [
    { path: '/', name: 'Dashboard' },
    { path: '/active', name: 'Active' },
    { path: 'https://nextcloud.com/', name: 'Drive' },
    { path: '/account', name: 'Account' },
  ],
};

export const Extra = Template.bind({});

Extra.args = {
  logoutHandler: () => null,
  routes: [
    { path: '/', name: 'Dashboard' },
    { path: '/active', name: 'Active' },
    { path: 'https://nextcloud.com/', name: 'Drive' },
    { path: '/account', name: 'Account' },
    { path: 'https://grafana.com', name: 'Grafana' },
    {
      path: 'https://ticketing.crownlabs.polito.it/',
      name: 'Ticketing',
    },
  ],
};
