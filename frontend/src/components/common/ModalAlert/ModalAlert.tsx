import React, { FC } from 'react';
import { Modal, Alert } from 'antd';

export interface IModalAlertProps {
  headTitle: string;
  showModal: boolean;
  alertMessage: string;
  alertDescription: string;
  alertType: 'info' | 'success' | 'error' | 'warning';
  buttons: Array<React.ReactNode>;
  setShowModal: (status: boolean) => void;
}

const ModalAlert: FC<IModalAlertProps> = ({ ...props }) => {
  const {
    headTitle,
    showModal,
    alertType,
    alertMessage,
    alertDescription,
    buttons,
    setShowModal,
  } = props;

  return (
    <Modal
      footer={false}
      centered
      title={headTitle}
      visible={showModal}
      onCancel={() => setShowModal(false)}
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
