import Navbar, { INavbarProps } from './Navbar';
import { Story, Meta } from '@storybook/react';
import { MemoryRouter, Route } from 'react-router-dom';
import ThemeContextProvider from '../../../contexts/ThemeContext';
import { Layout, Skeleton } from 'antd';
import { LinkPosition } from '../../../utils';

export default {
  title: 'Components/common/Navbar',
  component: Navbar,
  decorators: [
    (Story: Story) => (
      <MemoryRouter>
        <ThemeContextProvider>
          <Route path="/">
            <Layout>
              <Story />
              <Layout.Content>
                <div className="m-4">
                  <Skeleton />
                  <Skeleton />
                  <Skeleton />
                </div>
              </Layout.Content>
            </Layout>
          </Route>
        </ThemeContextProvider>
      </MemoryRouter>
    ),
  ],
  argTypes: {
    logoutHandler: { table: { disable: true } },
  },
} as Meta;

const Template: Story<INavbarProps> = args => <Navbar {...args} />;

export const Default = Template.bind({});

Default.args = {
  logoutHandler: () => null,
  routes: [
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/', name: 'Dashboard' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/active', name: 'Active' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: 'https://nextcloud.com/', name: 'Drive' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/account', name: 'Account' },
    },
  ],
};

export const Extra = Template.bind({});

Extra.args = {
  logoutHandler: () => null,
  routes: [
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/', name: 'Dashboard' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/active', name: 'Active' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: 'https://nextcloud.com/', name: 'Drive' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: '/account', name: 'Account' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: { path: 'https://grafana.com', name: 'Grafana' },
    },
    {
      linkPosition: LinkPosition.NavbarButton,
      route: {
        path: 'https://ticketing.crownlabs.polito.it/',
        name: 'Ticketing',
      },
    },
  ],
};
