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
        'Center_Only_small',
        'Center_Only_middle',
        'Center_Only_large',
        'Left_Center_Right_small',
        'Left_Center_Right_middle',
        'Left_Center_Right_large',
      ],
      mapping: {
        Center_Only_small: {
          size: 'small' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Simple Header small</b>
              </p>
            </div>
          ),
        },
        Center_Only_middle: {
          size: 'middle' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Simple Header middle</b>
              </p>
            </div>
          ),
        },
        Center_Only_large: {
          size: 'large' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Simple Header large</b>
              </p>
            </div>
          ),
        },
        Left_Center_Right_small: {
          size: 'small' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Custom Header small</b>
              </p>
            </div>
          ),
          left: (
            <div className="h-full flex justify-center items-center pl-10">
              <Button
                type="primary"
                shape="circle"
                size="large"
                icon={<UserSwitchOutlined />}
              />
            </div>
          ),
          right: (
            <div className="h-full flex justify-center items-center pr-10">
              <Button
                type="lightdark"
                shape="circle"
                size="large"
                icon={<PlusOutlined />}
              />
            </div>
          ),
        },
        Left_Center_Right_middle: {
          size: 'middle' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Custom Header middle</b>
              </p>
            </div>
          ),
          left: (
            <div className="h-full flex justify-center items-center pl-10">
              <Button
                type="primary"
                shape="circle"
                size="large"
                icon={<UserSwitchOutlined />}
              />
            </div>
          ),
          right: (
            <div className="h-full flex justify-center items-center pr-10">
              <Button
                type="lightdark"
                shape="circle"
                size="large"
                icon={<PlusOutlined />}
              />
            </div>
          ),
        },
        Left_Center_Right_large: {
          size: 'large' as BoxHeaderSize,
          center: (
            <div className="h-full flex justify-center items-center px-5">
              <p className="md:text-4xl text-2xl text-center mb-0">
                <b>Custom Header large</b>
              </p>
            </div>
          ),
          left: (
            <div className="h-full flex justify-center items-center pl-10">
              <Button
                type="primary"
                shape="circle"
                size="large"
                icon={<UserSwitchOutlined />}
              />
            </div>
          ),
          right: (
            <div className="h-full flex justify-center items-center pr-10">
              <Button
                type="lightdark"
                shape="circle"
                size="large"
                icon={<PlusOutlined />}
              />
            </div>
          ),
        },
      },
    },
    footer: {
      options: ['Yes', 'None'],
      mapping: {
        Yes: (
          <div className="w-full py-10 flex justify-center">
            <Button
              type="success"
              shape="round"
              size={'large'}
              disabled={false}
            >
              Button Footer Example
            </Button>
          </div>
        ),
        None: undefined,
      },
    },
    headerMinHeight: { table: { disable: true } },
    footerMinHeight: { table: { disable: true } },
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
      <div className="h-full flex justify-center items-center pl-10">
        <p className="md:text-4xl text-2xl text-center mb-0">
          <b>User View</b>
        </p>
      </div>
    ),
    size: 'large' as BoxHeaderSize,
  },
  footer: (
    <div className="w-full py-10 flex justify-center">
      <Button type="success" shape="round" size={'large'} disabled={false}>
        Button Footer Example
      </Button>
    </div>
  ),
};

export const Manager = Template.bind({});

Manager.args = {
  ...defaultArgs,
  header: {
    center: (
      <div className="h-full flex justify-center items-center px-5">
        <p className="md:text-4xl text-2xl text-center mb-0">
          <b>Manager View</b>
        </p>
      </div>
    ),
    left: (
      <div className="h-full flex justify-center items-center pl-10">
        <Button
          type="primary"
          shape="circle"
          size="large"
          icon={<UserSwitchOutlined />}
        />
      </div>
    ),
    right: (
      <div className="h-full flex justify-center items-center pr-10">
        <Button
          type="lightdark"
          shape="circle"
          size="large"
          icon={<PlusOutlined />}
        />
      </div>
    ),
    size: 'large' as BoxHeaderSize,
  },
  footer: (
    <div className="w-full py-10 flex justify-center">
      <Button type="success" shape="round" size={'large'} disabled={false}>
        Button Footer Example
      </Button>
    </div>
  ),
};
