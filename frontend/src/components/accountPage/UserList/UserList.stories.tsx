import UserList, { IUserListProps } from './UserList';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/AccountPage/UserList',
  component: UserList,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IUserListProps> = {};

const Template: Story<IUserListProps> = args => <UserList {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
