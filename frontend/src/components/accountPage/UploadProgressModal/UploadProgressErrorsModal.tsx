import { FC, useState } from 'react';
import { Alert, Modal, Button } from 'antd';
export interface IUploadProgressErrorsModal {
  errors: any[];
  uploadedUserNumber: number;
}

const tryExtractError = (e: any): string => {
  try {
    return e.graphQLErrors[0].extensions.k8s.reason;
  } catch (_) {
    return 'Unknown error (see details)';
  }
};

const UploadProgressErrorsModal: FC<IUploadProgressErrorsModal> = props => {
  const [showModal, setShowModal] = useState(false);
  const failedEntities = props.errors.filter(e => e.entity);
  return (
    <div className="mt-2">
      {props.errors.length > 0 && (
        <Alert
          message={`${props.errors.length} errors and ${props.uploadedUserNumber} successes.`}
          showIcon
          description={
            <>
              Some errors occured while uploading users from csv.
              {failedEntities && (
                <p className="mt-2">
                  The following users could not be synchronized:
                  <ul>
                    {failedEntities.map(e => (
                      <li>
                        {e.entity}: {tryExtractError(e)}
                      </li>
                    ))}
                  </ul>
                </p>
              )}
            </>
          }
          type="error"
          action={
            <Button
              size="small"
              danger
              onClick={() => setShowModal(true)}
              children="Details"
            />
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
              message={e.entity && <b>User: {e.entity}</b>}
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
