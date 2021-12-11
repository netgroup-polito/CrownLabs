import React, { FC, ReactNode } from 'react';
import { Modal, Alert, AlertProps } from 'antd';

export interface IModalAlertProps extends AlertProps {
  headTitle: ReactNode;
  show: boolean;
  buttons: Array<React.ReactNode>;
  setShow: (status: boolean) => void;
}

const ModalAlert: FC<IModalAlertProps> = ({ ...props }) => {
  const { headTitle, show, type, message, description, buttons, setShow } =
    props;

  return (
    <Modal
      footer={false}
      centered
      title={headTitle}
      visible={show}
      onCancel={() => setShow(false)}
    >
      <Alert message={message} description={description} type={type} showIcon />

      <div className="flex justify-center mt-6">{buttons}</div>
    </Modal>
  );
};

export default ModalAlert;
