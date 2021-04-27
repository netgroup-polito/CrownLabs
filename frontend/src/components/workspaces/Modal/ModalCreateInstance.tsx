import { FC } from 'react';
import { Modal, Alert, Space } from 'antd';
import Button from 'antd-button-color';
import { Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import '../../../index.less'; //To delete, usefull only to storybook

export interface IModalCreateInstanceProps {
  headTitle: string;
  showmodal: boolean;
  setshowmodal: (status: boolean) => void;
  loadingVm: boolean;
}

const ModalCreateInstance: FC<IModalCreateInstanceProps> = ({ ...props }) => {
  const { headTitle, showmodal, setshowmodal, loadingVm } = props;

  const handleCancel = () => {
    setshowmodal(false);
  };

  const antIcon = (
    <LoadingOutlined style={{ fontSize: 20, color: 'white' }} spin />
  );

  const footerDivWaiting = (
    <div className="flex justify-center mt-6">
      <Button type="primary" shape="round" size={'middle'}>
        <Space size="middle">
          Loading
          <Spin indicator={antIcon} />
        </Space>
      </Button>
    </div>
  );

  const footerDivReady = (
    <div className="flex justify-center mt-6">
      <Button type="success" shape="round" size={'middle'}>
        Go to active
      </Button>
    </div>
  );

  const alertWarning = (
    <Alert
      message="Crownlabs is creating your vm ..."
      description="Please remember to turn off your VM in the active section when you don’t need it anymore."
      type="warning"
      showIcon
    />
  );

  const alertReady = (
    <Alert
      message="Your VM is ready"
      description="Please remember to turn off your VM in the active section when you don’t need it anymore."
      type="success"
      showIcon
    />
  );

  return (
    <Modal
      footer={false}
      centered
      title={headTitle}
      visible={showmodal}
      onCancel={handleCancel}
    >
      {loadingVm ? alertWarning : alertReady}
      {loadingVm ? footerDivWaiting : footerDivReady}
    </Modal>
  );
};

export default ModalCreateInstance;
