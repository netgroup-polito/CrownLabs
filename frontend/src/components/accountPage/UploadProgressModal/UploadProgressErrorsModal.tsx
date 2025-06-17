import { type FC, useState } from 'react';
import { Alert, Modal, Button } from 'antd';
import type { ApolloError } from '@apollo/client';
import type { EnrichedError } from '../../../errorHandling/utils';
export interface IUploadProgressErrorsModal {
  errors: EnrichedError[];
  uploadedUserNumber: number;
}

const unknownError = 'Unknown error (see details)';

const tryExtractError = (e: EnrichedError): string => {
  try {
    if (typeof e === 'object' && e !== null && 'graphQLErrors' in e) {
      const err = e as ApolloError;
      const k8s = err.graphQLErrors?.[0]?.extensions?.k8s;
      return typeof k8s === 'object' && k8s !== null && 'reason' in k8s
        ? ((k8s as { reason?: string }).reason ?? unknownError)
        : unknownError;
    }
  } catch (_) {
    /* empty */
  }
  return unknownError;
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
        open={showModal}
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
