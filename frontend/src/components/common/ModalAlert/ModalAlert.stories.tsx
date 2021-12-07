import ModalAlert, { IModalAlertProps } from './ModalAlert';
import { Story, Meta } from '@storybook/react';
import Button from 'antd-button-color';
import { Space, Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import { DialogOpenDecorator } from '../../../decorators/DialogOpenDecorator';

export default {
  title: 'Components/common/ModalAlert',
  component: ModalAlert,
  argTypes: {
    show: { table: { disable: true } },
    setShow: { table: { disable: true } },
  },
  decorators: [DialogOpenDecorator],
} as Meta;

const Template: Story<IModalAlertProps> = args => <ModalAlert {...args} />;

export const Loading = Template.bind({});
Loading.args = {
  headTitle: 'Loading modal',
  type: 'warning',
  message: 'CrownLabs is creating your vm ...',
  description:
    'Please remember to turn off your VM in the active section when you don’t need it anymore.',
  buttons: [
    <Button key={0} type="primary" shape="round" size={'middle'}>
      <Space size="middle">
        Loading
        <Spin
          indicator={
            <LoadingOutlined style={{ fontSize: 20, color: 'white' }} spin />
          }
        />
      </Space>
    </Button>,
  ],
};

export const Ready = Template.bind({});
Ready.args = {
  headTitle: 'Ready modal',
  type: 'success',
  message: 'your VM is ready',
  description:
    'Please remember to turn off your VM in the active section when you don’t need it anymore.',
  buttons: [
    <Button key={0} type="success" shape="round" size={'middle'}>
      Go to active
    </Button>,
  ],
};

export const Exit = Template.bind({});
Exit.args = {
  headTitle: 'Wait before going out',
  type: 'error',
  message: 'You have some VM still running',
  description: 'Please turn off your VM if you don’t need it anymore',
  buttons: [
    <Button ghost type="primary" shape="round" size={'middle'}>
      Exit
    </Button>,
    <Button type="danger" className="ml-5" shape="round" size={'middle'}>
      Go to active
    </Button>,
  ],
};
