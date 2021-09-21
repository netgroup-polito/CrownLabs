import { FC, useState } from 'react';
import { Alert, Modal, Button } from 'antd';
export interface IUploadProgressErrorsModal {
  errors: any[];
  uploadedUserNumber: number;
}
const UploadProgressErrorsModal: FC<IUploadProgressErrorsModal> = props => {
  const [showModal, setShowModal] = useState(false);
  return (
    <div className="mt-2">
      {props.errors.length > 0 && (
        <Alert
          message={`${props.errors.length} errors and ${props.uploadedUserNumber} success.`}
          showIcon
          description="Some errors occured while uploading users from csv."
          type="error"
          action={
            <Button size="small" danger onClick={() => setShowModal(true)}>
              Detail
            </Button>
          }
        />
      )}
      <Modal
        visible={showModal}
        closable={true}
        onCancel={() => setShowModal(false)}
      >
        <div className="overflow-auto mt-5 pt-0 pr-2 h-96">
          {props.errors.map(e => (
            <Alert
              className="mt-1"
              message=""
              description={e.message}
              type="error"
            />
          ))}
        </div>
      </Modal>
    </div>
  );
};
export default UploadProgressErrorsModal;
