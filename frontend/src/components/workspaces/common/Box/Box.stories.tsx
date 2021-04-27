import Box, { IBoxProps } from './Box';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../../utils';

import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';

export default {
  title: 'Components/Box',
  component: Box,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IBoxProps> = {
  headLeft: (
    <Button
      type="primary"
      shape="circle"
      size="large"
      icon={<UserSwitchOutlined />}
    />
  ),
  headRight: (
    <Button
      type="lightdark"
      shape="circle"
      size="large"
      icon={<PlusOutlined />}
    />
  ),
  headTitle: 'Box Professor View',
  footer: (
    <Button type="success" shape="round" size={'large'} disabled={false}>
      Button Footer
    </Button>
  ),
};

const Template: Story<IBoxProps> = args => <Box {...args} />;

export const ProfessorView = Template.bind({});

ProfessorView.args = defaultArgs;

export const StudentView = Template.bind({});

StudentView.args = {
  ...defaultArgs,
  headTitle: 'Box Student View',
  headLeft: null,
  headRight: null,
};
