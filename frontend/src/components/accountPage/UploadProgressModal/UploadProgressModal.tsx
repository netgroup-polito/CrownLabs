import { Button, Modal } from 'antd';
import {
  StopOutlined,
  CaretRightOutlined,
  CaretLeftOutlined,
} from '@ant-design/icons';
import { type FC, useEffect, useState } from 'react';
import { type UserAccountPage, type WorkspaceEntry } from '../../../utils';

import UploadProgressContent from './UploadProgressContent';
import { Role } from '../../../generated-types';
import UploadProgressErrorsModal from './UploadProgressErrorsModal';
import {
  type EnrichedError,
  type SupportedError,
} from '../../../errorHandling/utils';
export interface IUploadProgressModalInterface {
  onClose: () => void;
  confirmUpload: (
    users: UserAccountPage[],
    workspaces: WorkspaceEntry[],
  ) => Promise<boolean>;
  setAbortUploading: (value: boolean) => void;
  setUploadingErrors: (errors: EnrichedError[]) => void;
  genericErrorCatcher: (err: SupportedError) => void;
  show: boolean;
  workspaceName: string;
  uploadedNumber: number;
  abortUploading: boolean;
  uploadingErrors: EnrichedError[];
  uploadedUserNumber: number;
}

// eslint-disable-next-line react-refresh/only-export-components
export enum StepStatus {
  finish = 'finish',
  error = 'error',
  process = 'process',
  wait = 'wait',
}

const UploadProgressModal: FC<IUploadProgressModalInterface> = props => {
  const { setAbortUploading } = props;
  const [usersCSV, setUsersCSV] = useState<UserAccountPage[]>([]);
  const [uploadingStatusResult, setUploadingStatusResult] = useState(false);
  const [editing, setEditing] = useState<boolean>(false);
  const [stepCurrent, setStepCurrent] = useState<number>(0);
  const [stepStatus, setStepStatus] = useState<StepStatus>(StepStatus.process);

  const handleOk = async () => {
    if (stepCurrent === 1) {
      setStepStatus(StepStatus.wait);
      setStepCurrent(2);

      const result = await props.confirmUpload(usersCSV, [
        { name: props.workspaceName, role: Role.User },
      ]);
      setUploadingStatusResult(result);
      setStepCurrent(3);
    } else {
      setStepCurrent(stepCurrent + 1);
    }
  };

  useEffect(() => {
    if (props.abortUploading) setStepCurrent(3);
  }, [props.abortUploading]);

  useEffect(() => {
    if (stepCurrent === 3)
      setStepStatus(
        !props.abortUploading &&
          !props.uploadingErrors.length &&
          uploadingStatusResult
          ? StepStatus.finish
          : StepStatus.error,
      );
  }, [
    props.abortUploading,
    props.uploadingErrors.length,
    stepCurrent,
    uploadingStatusResult,
  ]);
  useEffect(() => {
    const doAlert = (ev: Event) => {
      ev.preventDefault();
      setAbortUploading(true);
      return "WARNING. If you close this window the import process will be interrupted, resulting in an incomplete import. If you proceed, you'll need to run the procedure from the beginning!";
    };
    window.addEventListener('beforeunload', doAlert);
    return () => window.removeEventListener('beforeunload', doAlert);
  });
  useEffect(() => {
    if (props.show) {
      setUsersCSV([]);
      setStepCurrent(0);
      setStepStatus(StepStatus.process);
      setAbortUploading(false);
      setUploadingStatusResult(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.show]);

  const handleClose = () => {
    props.setUploadingErrors([]);
    props.onClose();
  };

  return (
    <Modal
      open={props.show}
      width="800px"
      closable={false}
      footer={[
        stepCurrent < 1 && (
          <Button
            key={'cancel'}
            danger
            onClick={handleClose}
            icon={<StopOutlined />}
            children="Cancel"
          />
        ),
        stepCurrent === 2 && (
          <Button
            key={'abort'}
            danger
            onClick={() => setAbortUploading(true)}
            icon={<StopOutlined />}
            children="Abort"
          />
        ),
        (stepCurrent === 1 || stepCurrent === 3) && (
          <Button
            key={'previous'}
            icon={<CaretLeftOutlined />}
            onClick={() => setStepCurrent(stepCurrent - 1)}
            disabled={stepCurrent === 3 || editing}
            children="Previous"
          />
        ),
        stepCurrent < 3 && (
          <Button
            key={'next'}
            icon={<CaretRightOutlined />}
            onClick={handleOk}
            disabled={usersCSV.length === 0 || stepCurrent === 2 || editing}
            children="Next"
          />
        ),
        stepCurrent === 3 && <Button onClick={props.onClose}>Close</Button>,
      ]}
      destroyOnHidden={true}
    >
      <UploadProgressContent
        setStepCurrent={setStepCurrent}
        stepCurrent={stepCurrent}
        stepStatus={stepStatus}
        workspaceName={props.workspaceName}
        setEditing={setEditing}
        setUsersCSV={setUsersCSV}
        usersCSV={usersCSV}
        uploadedNumber={props.uploadedNumber}
        uploadedUserNumber={props.uploadedUserNumber}
        abortUploading={props.abortUploading}
        uploadingErrors={props.uploadingErrors}
        genericErrorCatcher={props.genericErrorCatcher}
      />

      {stepCurrent === 3 && (
        <UploadProgressErrorsModal
          errors={props.uploadingErrors}
          uploadedUserNumber={props.uploadedUserNumber}
        />
      )}
    </Modal>
  );
};

export default UploadProgressModal;
