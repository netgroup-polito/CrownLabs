import type { FC } from 'react';
import { useState, useEffect } from 'react';
import type { UserAccountPage } from '../../../utils';
import {
  Row,
  Button,
  Upload,
  Typography,
  Steps,
  Col,
  Progress,
  type UploadFile,
} from 'antd';
import { UploadOutlined, LoadingOutlined } from '@ant-design/icons';
import EditableTable from './EditableTable';
import { StepStatus } from './UploadProgressModal';
import Papa from 'papaparse';
import { Role } from '../../../generated-types';
import type {
  EnrichedError,
  SupportedError,
} from '../../../errorHandling/utils';
import type { UploadChangeParam } from 'antd/lib/upload';
const { Text } = Typography;
const { Step } = Steps;
export interface IUploadProgressContent {
  workspaceName: string;
  setEditing: (value: boolean) => void;
  stepCurrent: number;
  stepStatus: StepStatus;
  setStepCurrent: (value: number) => void;
  usersCSV: UserAccountPage[];
  setUsersCSV: (users: UserAccountPage[]) => void;
  uploadedNumber: number;
  abortUploading: boolean;
  uploadingErrors: EnrichedError[];
  uploadedUserNumber: number;
  genericErrorCatcher: (err: SupportedError) => void;
}

enum StatusProgressBar {
  active = 'active',
  normal = 'normal',
  success = 'success',
  exception = 'exception',
}

enum CSVFields {
  name = 'NOME',
  surname = 'COGNOME',
  surname_inserted = 'COGNOME - (*) Inserito dal docente',
  userid = 'MATRICOLA',
  email = 'EMAIL',
}

const UploadProgressContent: FC<IUploadProgressContent> = props => {
  const [fileError, setFileError] = useState<string>('');
  const [statusProgressBar, setStatusProgessBar] = useState<StatusProgressBar>(
    StatusProgressBar.active,
  );
  const { usersCSV, setUsersCSV } = props;

  const capitalizeName = (name: string) => {
    return name
      .toLowerCase()
      .replace(/\b(\w)/g, s => s.toUpperCase())
      .trim();
  };

  const handleUserId = (userId: string) => {
    const regex = new RegExp(/^[0-9]{6}$/);
    return regex.test(userId)
      ? 's' + userId.trim()
      : userId.toLowerCase().trim();
  };

  useEffect(() => {
    if (props.stepStatus === StepStatus.error || props.abortUploading)
      setStatusProgessBar(StatusProgressBar.exception);
    else if (props.uploadedUserNumber === usersCSV.length)
      setStatusProgessBar(StatusProgressBar.success);
    else setStatusProgessBar(StatusProgressBar.active);
  }, [
    props.stepStatus,
    props.uploadedUserNumber,
    usersCSV.length,
    props.abortUploading,
  ]);

  const handleUserCSV = (user: Record<string, string>) => {
    return {
      key: user[CSVFields.userid].trim(),
      name: capitalizeName(user[CSVFields.name]) ?? '',
      surname:
        capitalizeName(
          user[CSVFields.surname_inserted]?.replace(/\(\*+\)/, '') ?? '',
        ) ??
        capitalizeName(user[CSVFields.surname]) ??
        '',
      userid: handleUserId(user[CSVFields.userid]) ?? '',
      email: user[CSVFields.email].toLowerCase().trim() ?? '',
      currentRole: Role.User,
      workspaces: [],
    };
  };
  const onCsvUploaded = (fileInfo: UploadChangeParam<UploadFile>) => {
    if (fileInfo.file.status === 'removed') {
      setUsersCSV([]);
      return;
    }

    console.log(fileInfo.file);
    if (!fileInfo.file) {
      setFileError('No file selected or file could not be read.');
      return;
    }
    Papa.parse<Record<string, string>>(fileInfo.file as unknown as File, {
      header: true,
      skipEmptyLines: true,
      complete: (result, _) => {
        for (const line of result.data) {
          if (
            !line[CSVFields.name] ||
            !(line[CSVFields.surname_inserted] || line[CSVFields.surname]) ||
            !line[CSVFields.userid] ||
            !line[CSVFields.email]
          ) {
            setFileError(
              'Invalid file format, must contain <MATRICOLA, NOME, COGNOME (o COGNOME - (*) Inserito dal docente), EMAIL>',
            );
            return;
          }
        }
        const users = result.data.map((user, _index) => handleUserCSV(user));
        setUsersCSV(users);
        props.setStepCurrent(1);
        setFileError('');
      },
    });
  };

  return (
    <>
      <Row className="flex justify-center mb-4">
        <Steps
          direction="horizontal"
          initial={0}
          current={props.stepCurrent}
          status={props.stepStatus}
        >
          <Step title="Upload " description="Upload your CSV file" />
          <Step title="Edit and review" description="Fix possible errors" />

          <Step
            title="Synchronization"
            description="Sync changes"
            icon={props.stepCurrent === 2 && <LoadingOutlined />}
          />

          {props.stepStatus === StepStatus.error ? (
            <Step
              title={props.abortUploading ? 'Aborted' : 'Error'}
              description={
                props.abortUploading ? 'Aborted uploading' : 'Upload failure'
              }
            />
          ) : (
            <Step title="Completed" description="Operation results" />
          )}
        </Steps>
      </Row>

      <Row className="flex justify-center mt-4">
        <Col>
          {props.stepCurrent === 0 && (
            <Upload
              className="flex justify-center"
              name="file"
              accept=".csv"
              onChange={onCsvUploaded}
              beforeUpload={() => false}
              fileList={[]}
              maxCount={1}
              disabled={props.stepCurrent > 0}
            >
              <Button
                type="primary"
                disabled={props.stepCurrent > 0}
                className="m-6"
                icon={<UploadOutlined />}
              >
                Upload CSV
              </Button>
            </Upload>
          )}
          {fileError && (
            <Text className="m-2" type="danger">
              {fileError}
            </Text>
          )}
        </Col>
      </Row>
      <Row>
        <Col flex={4}>
          {props.stepCurrent === 1 && (
            <EditableTable
              data={usersCSV}
              updateUserCSV={(users: UserAccountPage[]) => {
                setUsersCSV(users);
              }}
              setEditing={(value: boolean) => props.setEditing(value)}
              genericErrorCatcher={props.genericErrorCatcher}
            />
          )}
          {props.stepCurrent > 1 && (
            <div className="flex justify-center my-8">
              <Progress
                type="circle"
                size={200}
                status={statusProgressBar}
                percent={Math.floor(
                  (props.uploadedNumber * 100) / usersCSV.length,
                )}
              />
            </div>
          )}
        </Col>
      </Row>
    </>
  );
};

export default UploadProgressContent;
