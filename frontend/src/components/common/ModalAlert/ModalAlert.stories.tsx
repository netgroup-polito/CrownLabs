import ModalAlert, { IModalAlertProps } from './ModalAlert';
import { Story, Meta } from '@storybook/react';
import Button from 'antd-button-color';
import { Space, Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

export default {
  title: 'Components/common/ModalAlert',
  component: ModalAlert,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const Template: Story<IModalAlertProps> = args => <ModalAlert {...args} />;

export const Loading = Template.bind({});
Loading.args = {
  headTitle: 'Loading modal ...',
  showModal: true,
  alertType: 'warning',
  alertMessage: 'Crownlabs is creating your vm ...',
  alertDescription:
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
  setShowModal: x => null,
};

export const Ready = Template.bind({});
Ready.args = {
  headTitle: 'Ready modal',
  showModal: true,
  alertType: 'success',
  alertMessage: 'your VM is ready',
  alertDescription:
    'Please remember to turn off your VM in the active section when you don’t need it anymore.',
  buttons: [
    <Button key={0} type="success" shape="round" size={'middle'}>
      Go to active
    </Button>,
  ],
  setShowModal: x => null,
};

export const Exit = Template.bind({});
Exit.args = {
  headTitle: 'Wait before going out',
  showModal: true,
  alertType: 'error',
  alertMessage: 'You have some VM still running',
  alertDescription: 'Please turn off your VM if you don’t need it anymore',
  buttons: [
    <Button ghost type="primary" shape="round" size={'middle'}>
      Exit
    </Button>,
    <Button type="danger" className="ml-5" shape="round" size={'middle'}>
      Go to active
    </Button>,
  ],
  setShowModal: x => null,
};
