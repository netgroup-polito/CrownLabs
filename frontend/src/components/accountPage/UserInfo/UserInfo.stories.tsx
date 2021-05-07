import UserInfo, { IUserInfoProps } from './UserInfo';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/AccountPage/UserInfo',
  component: UserInfo,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IUserInfoProps> = {
  firstName: 'John',
  lastName: 'Doe',
  username: 's123456',
  email: 'john.doe@studenti.polito.it',
};

const Template: Story<IUserInfoProps> = args => <UserInfo {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
