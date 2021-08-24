import SidebarInfo, { ISidebarInfoProps } from './SidebarInfo';
import { Story, Meta, StoryContext } from '@storybook/react';
import { someKeysOf } from '../../../utils';
import { Layout, Skeleton } from 'antd';
import Button from 'antd-button-color';
import { useState } from 'react';

export default {
  title: 'Components/common/SidebarInfo',
  component: SidebarInfo,
  argTypes: {
    visible: { table: { disable: true } },
    onClose: { table: { disable: true } },
  },
  decorators: [
    (Story: Story, context: StoryContext) => {
      const [show, setShow] = useState(false);
      context.args.visible = show;
      context.args.onClose = () => setShow(false);
      return (
        <Layout>
          <Layout.Content>
            <Story />
            <Skeleton />
            <div className=" flex justify-center mt-12">
              <Button
                onClick={() => setShow(true)}
                type="primary"
                shape="round"
                size="large"
              >
                Open Sidebar
              </Button>
            </div>
          </Layout.Content>
        </Layout>
      );
    },
  ],
} as Meta;

const defaultArgs: someKeysOf<ISidebarInfoProps> = {
  position: 'left',
};

const Template: Story<ISidebarInfoProps> = args => <SidebarInfo {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
