import React, { FC } from 'react';
import { Modal, Alert } from 'antd';

export interface IModalAlertProps {
  headTitle: string;
  show: boolean;
  alertMessage: string;
  alertDescription: string;
  alertType: 'info' | 'success' | 'error' | 'warning';
  buttons: Array<React.ReactNode>;
  setShow: (status: boolean) => void;
}

const ModalAlert: FC<IModalAlertProps> = ({ ...props }) => {
  const {
    headTitle,
    show,
    alertType,
    alertMessage,
    alertDescription,
    buttons,
    setShow,
  } = props;

  return (
    <Modal
      footer={false}
      centered
      title={headTitle}
      visible={show}
      onCancel={() => setShow(false)}
    >
      <Alert
        message={alertMessage}
        description={alertDescription}
        type={alertType}
        showIcon
      />

      <div className="flex justify-center mt-6">{buttons}</div>
    </Modal>
  );
};

export default ModalAlert;
