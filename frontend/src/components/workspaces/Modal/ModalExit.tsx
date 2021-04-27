import { FC } from 'react';
import { Modal, Alert } from 'antd';
import Button from 'antd-button-color';
import '../../../index.less'; //To delete, usefull only to storybook

export interface IModalExitProps {
  showmodal: boolean;
  setshowmodal: (status: boolean) => void;
}

const ModalExit: FC<IModalExitProps> = ({ ...props }) => {
  const { showmodal, setshowmodal } = props;

  const handleCancel = () => {
    setshowmodal(false);
  };

  const footerDiv = (
    <div className="flex justify-center mt-6">
      <Button
        ghost
        type="primary"
        shape="round"
        size={'middle'}
        onClick={() => setshowmodal(false)}
      >
        Exit
      </Button>
      <Button type="danger" shape="round" size={'middle'}>
        Go to active
      </Button>
    </div>
  );

  const alert = (
    <Alert
      message="You have some VM still running"
      description="Please turn off your VM if you don’t need it anymore"
      type="error"
      showIcon
    />
  );

  return (
    <Modal
      centered
      footer={null}
      title="Wait before leave out"
      visible={showmodal}
      onCancel={handleCancel}
    >
      {alert}
      {footerDiv}
    </Modal>
  );
};

export default ModalExit;
