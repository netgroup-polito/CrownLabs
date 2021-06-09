import SSHKeysForm, { ISSHKeysFormProps } from './SSHKeysForm';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/AccountPage/SSHKeysForm',
  component: SSHKeysForm,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<ISSHKeysFormProps> = {};

const Template: Story<ISSHKeysFormProps> = args => <SSHKeysForm {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
