import UserListForm, { IUserListFormProps } from './UserListForm';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/AccountPage/UserListForm',
  component: UserListForm,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IUserListFormProps> = {};

const Template: Story<IUserListFormProps> = args => <UserListForm {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
