import type { FC } from 'react';
import { useState, useContext, useEffect } from 'react';
import { Modal, Form, Input, Select, Button } from 'antd';
import { useCreateWorkspaceMutation, useApplyWorkspaceMutation, AutoEnroll } from '../../../generated-types';
import type { ApolloError } from '@apollo/client';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { convertToGiB } from '../../../utils';
import QuotaFields from '../../shared/QuotaFields';

export interface WorkspaceEditData {
  name: string;
  prettyName: string;
  autoEnroll?: string | null;
  cpu: string;
  memory: string;
  instances: number;
}

export interface IModalCreateWorkspaceProps {
  show: boolean;
  setShow: (status: boolean) => void;
  onSuccess?: () => void;
  editWorkspace?: WorkspaceEditData | null;
  existingWorkspaceNames?: string[];
}

interface WorkspaceFormValues {
  name: string;
  prettyName: string;
  autoEnroll?: string;
  cpu: number;
  memory: number;
  instances: number;
}

const ModalCreateWorkspace: FC<IModalCreateWorkspaceProps> = ({
  show,
  setShow,
  onSuccess,
  editWorkspace,
  existingWorkspaceNames = [],
}) => {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [form] = Form.useForm<WorkspaceFormValues>();
  const [loading, setLoading] = useState(false);
  const isEditMode = !!editWorkspace;

  const [createWorkspace] = useCreateWorkspaceMutation({
    onError: apolloErrorCatcher,
  });

  const [applyWorkspace] = useApplyWorkspaceMutation({
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (show && editWorkspace) {
      form.setFieldsValue({
        name: editWorkspace.name,
        prettyName: editWorkspace.prettyName,
        autoEnroll: editWorkspace.autoEnroll || undefined,
        cpu: parseFloat(editWorkspace.cpu),
        memory: convertToGiB(editWorkspace.memory),
        instances: editWorkspace.instances,
      });
    } else if (show && !editWorkspace) {
      form.resetFields();
    }
  }, [show, editWorkspace, form]);

  // Convert GraphQL enum value to Kubernetes expected value
  const normalizeAutoEnroll = (value: string | undefined | null): AutoEnroll | null => {
    if (!value || value === AutoEnroll.Empty) return null;
    return value as AutoEnroll;
  };

  const handleSubmit = async (values: WorkspaceFormValues) => {
    setLoading(true);
    try {
      if (isEditMode) {
        // Edit mode: use apply mutation with JSON patch
        const autoEnrollValue = normalizeAutoEnroll(values.autoEnroll);
        const patchJson = JSON.stringify([
          { op: 'replace', path: '/spec/prettyName', value: values.prettyName },
          { op: 'replace', path: '/spec/autoEnroll', value: autoEnrollValue },
          { op: 'replace', path: '/spec/quota/cpu', value: String(values.cpu) },
          { op: 'replace', path: '/spec/quota/memory', value: `${values.memory}Gi` },
          { op: 'replace', path: '/spec/quota/instances', value: values.instances },
        ]);
        await applyWorkspace({
          variables: {
            name: values.name,
            patchJson: patchJson,
            manager: 'frontend-workspace-edit',
          },
        });
      } else {
        // Create mode
        await createWorkspace({
          variables: {
            name: values.name,
            prettyName: values.prettyName,
            autoEnroll: normalizeAutoEnroll(values.autoEnroll),
            cpu: String(values.cpu),
            memory: `${values.memory}Gi`, // Kubernetes Quantity format
            instances: values.instances,
          },
        });
      }
      form.resetFields();
      setShow(false);
      onSuccess?.();
    } catch (error) {
      apolloErrorCatcher(error as ApolloError);
    } finally {
      setLoading(false);
    }
  };

  const handleCancel = () => {
    form.resetFields();
    setShow(false);
  };

  return (
    <Modal
      title={isEditMode ? 'Edit Workspace' : 'Create New Workspace'}
      open={show}
      onCancel={handleCancel}
      width={600}
      footer={[
        <Button key="cancel" onClick={handleCancel}>
          Cancel
        </Button>,
        <Button
          key="submit"
          type="primary"
          loading={loading}
          onClick={() => form.submit()}
        >
          {isEditMode ? 'Save' : 'Create'}
        </Button>,
      ]}
    >
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        initialValues={{
          autoEnroll: AutoEnroll.Empty,
          cpu: 4,
          memory: 8,
          instances: 5,
        }}
      >
        <Form.Item
          label="Name"
          name="name"
          rules={[
            { required: true, message: 'Please input workspace name!' },
            {
              pattern: /^[a-z0-9]([-a-z0-9]*[a-z0-9])?$/,
              message:
                'Name must be lowercase alphanumeric with hyphens (e.g., my-workspace)',
            },
            {
              validator: async (_, value) => {
                if (!isEditMode && value && existingWorkspaceNames.includes(value)) {
                  return Promise.reject(new Error(`Workspace "${value}" already exists. Please choose a different name.`));
                }
                return Promise.resolve();
              },
            },
          ]}
          tooltip="Unique identifier for the workspace (lowercase, alphanumeric, hyphens only)"
        >
          <Input placeholder="my-workspace" disabled={isEditMode} />
        </Form.Item>

        <Form.Item
          label="Pretty Name"
          name="prettyName"
          rules={[{ required: true, message: 'Please input pretty name!' }]}
          tooltip="Human-readable name for the workspace"
        >
          <Input placeholder="My Workspace" />
        </Form.Item>

        <Form.Item
          label="Allow self-enrollment"
          name="autoEnroll"
          tooltip="Enabling self-enrollment makes this workspace visible by every CrownLabs user. The option defines if the self-enrollment is immediate or happens after managers' approval"
        >
          <Select
            options={[
              { label: 'No', value: AutoEnroll.Empty },
              { label: 'Yes', value: AutoEnroll.Immediate },
              { label: 'Require approval', value: AutoEnroll.WithApproval },
            ]}
          />
        </Form.Item>

        <QuotaFields
          rules={{
            cpu: [{ required: true, message: 'Please input CPU quota!' }],
            memory: [{ required: true, message: 'Please input memory quota!' }],
            instances: [{ required: true, message: 'Please input max instances!' }],
          }}
          limits={{
            cpu: { min: 1, max: 128 },
            memory: { min: 1, max: 512 },
            instances: { min: 1, max: 100 },
          }}
          tooltips={{
            cpu: 'Maximum number of CPU cores',
            memory: 'Maximum memory in gibibytes',
            instances: 'Maximum number of concurrent instances',
          }}
        />
      </Form>
    </Modal>
  );
};

export default ModalCreateWorkspace;
