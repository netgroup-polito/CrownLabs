import Box, { IBoxProps } from './Box';
import { Story, Meta } from '@storybook/react';
import { BoxHeaderSize, someKeysOf } from '../../../utils';

import { UserSwitchOutlined, PlusOutlined } from '@ant-design/icons';
import Button from 'antd-button-color';

export default {
  title: 'Components/common/Box',
  component: Box,
  argTypes: {
    header: {
      options: [
        'Small_Simple',
        'Middle_Simple',
        'Large_Simple',
        'Small_Custom',
        'Middle_Custom',
        'Large_Custom',
      ],
      mapping: {
        Small_Simple: {
          size: 'small' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Simple Header small</b>
            </p>
          ),
        },
        Middle_Simple: {
          size: 'middle' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Simple Header middle</b>
            </p>
          ),
        },
        Large_Simple: {
          size: 'large' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Simple Header large</b>
            </p>
          ),
        },
        Small_Custom: {
          size: 'small' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Custom Header small</b>
            </p>
          ),
          left: (
            <Button
              type="primary"
              shape="circle"
              size="large"
              icon={<UserSwitchOutlined />}
            />
          ),
          right: (
            <Button
              type="lightdark"
              shape="circle"
              size="large"
              icon={<PlusOutlined />}
            />
          ),
        },
        Middle_Custom: {
          size: 'middle' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Custom Header middle</b>
            </p>
          ),
          left: (
            <Button
              type="primary"
              shape="circle"
              size="large"
              icon={<UserSwitchOutlined />}
            />
          ),
          right: (
            <Button
              type="lightdark"
              shape="circle"
              size="large"
              icon={<PlusOutlined />}
            />
          ),
        },
        Large_Custom: {
          size: 'large' as BoxHeaderSize,
          center: (
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Custom Header large</b>
            </p>
          ),
          left: (
            <Button
              type="primary"
              shape="circle"
              size="large"
              icon={<UserSwitchOutlined />}
            />
          ),
          right: (
            <Button
              type="lightdark"
              shape="circle"
              size="large"
              icon={<PlusOutlined />}
            />
          ),
        },
      },
    },
    footer: {
      options: ['Yes', 'None'],
      mapping: {
        Yes: (
          <Button type="success" shape="round" size={'large'} disabled={false}>
            Button Footer Example
          </Button>
        ),
        None: undefined,
      },
    },
  },
} as Meta;

const defaultArgs: someKeysOf<IBoxProps> = {};

const Template: Story<IBoxProps> = args => (
  <Box {...args}>
    <div className="w-full flex-grow flex flex-wrap content-center justify-center py-5 2xl:py-52">
      <p className="text-xl text-center px-5 xs:px-24 block">
        <span className="text-3xl">Example Content</span> <br />
        This is an example <b>{`<Box>`}</b> Component created for all containers
        of new frontend.
        <br />
        You can optionally insert a <b>Header</b> (3 areas: Left, Center and
        Right), a <b>Content</b> like this and a <b>Footer</b>
      </p>
    </div>
  </Box>
);

export const Default = Template.bind({});

Default.args = defaultArgs;

export const User = Template.bind({});

User.args = {
  ...defaultArgs,
  header: {
    center: (
      <p className="md:text-4xl text-2xl text-center mb-0">
        <b>User View</b>
      </p>
    ),
    size: 'large' as BoxHeaderSize,
  },
  footer: (
    <Button type="success" shape="round" size={'large'} disabled={false}>
      Button Footer
    </Button>
  ),
};

export const Manager = Template.bind({});

Manager.args = {
  ...defaultArgs,
  header: {
    center: (
      <p className="md:text-4xl text-2xl text-center mb-0">
        <b>Manager View</b>
      </p>
    ),
    left: (
      <Button
        type="primary"
        shape="circle"
        size="large"
        icon={<UserSwitchOutlined />}
      />
    ),
    right: (
      <Button
        type="lightdark"
        shape="circle"
        size="large"
        icon={<PlusOutlined />}
      />
    ),
    size: 'large' as BoxHeaderSize,
  },
  footer: (
    <Button type="success" shape="round" size={'large'} disabled={false}>
      Button Footer
    </Button>
  ),
};
