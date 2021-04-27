import ModalExit, { IModalExitProps } from './ModalExit';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/ModalExit',
  component: ModalExit,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IModalExitProps> = {
  showmodal: true,
};

const Template: Story<IModalExitProps> = args => <ModalExit {...args} />;

export const Default = Template.bind({});
Default.args = defaultArgs;
