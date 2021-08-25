import SidebarInfo, { ISidebarInfoProps } from './SidebarInfo';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';
import { DialogOpenDecorator } from '../../../decorators/DialogOpenDecorator';

export default {
  title: 'Components/common/SidebarInfo',
  component: SidebarInfo,
  argTypes: {
    show: { table: { disable: true } },
    setShow: { table: { disable: true } },
  },
  decorators: [DialogOpenDecorator],
} as Meta;

const defaultArgs: someKeysOf<ISidebarInfoProps> = {
  position: 'left',
};

const Template: Story<ISidebarInfoProps> = args => <SidebarInfo {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
