import Dashboard, { IDashboardProps } from './Dashboard';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/workspaces/Dashboard',
  component: Dashboard,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IDashboardProps> = {};

const Template: Story<IDashboardProps> = args => <Dashboard {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
