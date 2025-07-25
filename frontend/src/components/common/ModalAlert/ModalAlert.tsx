import type { Dispatch, FC, ReactNode, SetStateAction } from 'react';
import React from 'react';
import type { AlertProps } from 'antd';
import { Modal, Alert, Checkbox } from 'antd';

export interface IModalAlertProps extends AlertProps {
  headTitle: ReactNode;
  show: boolean;
  buttons: Array<React.ReactNode>;
  setShow: (status: boolean) => void;
  checkbox?: {
    confirmCheckbox: boolean;
    setConfirmCheckbox: Dispatch<SetStateAction<boolean>>;
    checkboxLabel: string;
  };
}

const ModalAlert: FC<IModalAlertProps> = ({ ...props }) => {
  const {
    headTitle,
    show,
    type,
    message,
    description,
    buttons,
    setShow,
    checkbox,
  } = props;

  const { confirmCheckbox, setConfirmCheckbox, checkboxLabel } = checkbox || {};

  return (
    <Modal
      footer={false}
      centered
      title={headTitle}
      open={show}
      onCancel={() => setShow(false)}
    >
      <Alert message={message} description={description} type={type} showIcon />
      {checkbox && (
        <div className="flex justify-center mt-3">
          <Checkbox
            checked={confirmCheckbox}
            onChange={() => setConfirmCheckbox!(old => !old)}
          >
            {checkboxLabel}
          </Checkbox>
        </div>
      )}
      <div className="flex justify-center mt-6">{buttons}</div>
    </Modal>
  );
};

export default ModalAlert;
