import type { ApolloError } from '@apollo/client';
import { Button, Divider, Modal } from 'antd';
import type { FC } from 'react';
import type { Instance, InstanceJobTemplate } from '../../../../utils';
import type { CrownJob } from './crownJobUtils';
import RunJobForm from './RunJobForm';
import RunJobList from './RunJobList';

export interface IRunJobModalProps {
  instance: Instance;
  job: InstanceJobTemplate;
  open: boolean;
  onClose: () => void;
  jobs: CrownJob[];
  usedNames: ReadonlySet<string>;
  refreshError?: ApolloError;
  refresh: () => Promise<unknown>;
}

const RunJobModal: FC<IRunJobModalProps> = ({
  instance,
  job,
  open,
  onClose,
  jobs,
  usedNames,
  refreshError,
  refresh,
}) => (
  <Modal
    title={`Run job: ${instance.prettyName || instance.name}`}
    width={720}
    open={open}
    onCancel={onClose}
    footer={<Button onClick={onClose}>Close</Button>}
    centered
    destroyOnHidden
  >
    <RunJobForm
      instance={instance}
      job={job}
      usedNames={usedNames}
      onCreated={refresh}
    />
    <Divider className="my-4" />
    <RunJobList
      tenantNamespace={instance.tenantNamespace}
      jobs={jobs}
      refreshError={refreshError}
      onChanged={refresh}
    />
  </Modal>
);

export default RunJobModal;
