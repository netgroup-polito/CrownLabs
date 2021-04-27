import UserInfo, { IUserInfoProps } from './UserInfo';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/UserInfo',
  component: UserInfo,
  argTypes: { onClick: { action: 'clicked' } },
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
