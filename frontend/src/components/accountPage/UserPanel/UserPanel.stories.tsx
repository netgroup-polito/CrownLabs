import UserPanel, { IUserPanelProps } from './UserPanel';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/UserPanel',
  component: UserPanel,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IUserPanelProps> = {
  firstName: 'John',
  lastName: 'Doe',
  username: 's123456',
  email: 'john.doe@studenti.polito.it',
};

const Template: Story<IUserPanelProps> = args => <UserPanel {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
