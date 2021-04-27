import ModalCreateInstance, {
  IModalCreateInstanceProps,
} from './ModalCreateInstance';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/ModalCreateInstance',
  component: ModalCreateInstance,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IModalCreateInstanceProps> = {
  showmodal: true,
  loadingVm: true,
  headTitle: 'Loading modal',
};

const Template: Story<IModalCreateInstanceProps> = args => (
  <ModalCreateInstance {...args} />
);

export const Loading = Template.bind({});
Loading.args = defaultArgs;

export const Ready = Template.bind({});
Ready.args = { ...defaultArgs, loadingVm: false, headTitle: 'Ready modal' };
