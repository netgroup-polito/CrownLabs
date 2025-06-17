import { InfoCircleOutlined } from '@ant-design/icons';
import { Button, Form, Input, InputNumber, Modal, Tooltip } from 'antd';
import type { FC } from 'react';
import { useEffect } from 'react';
import { findKeyByValue } from '../../../../utils';

// eslint-disable-next-line react-refresh/only-export-components
export enum Actions {
  Create = 'Create a new Shared Volume',
  Update = 'Update an existing Shared Volume',
}

export interface ISharedVolumesFormProps {
  workspaceNamespace: string;
  workspaceName: string;
  initialName?: string;
  initialSize?: number;
  open: boolean;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  action: Actions;
  mutation: (p: {
    wsName: string;
    wsNs: string;
    prettyName: string;
    size: string;
  }) => Promise<unknown>;
  loading: boolean;
  reload: () => void;
}

const SharedVolumeForm: FC<ISharedVolumesFormProps> = ({ ...props }) => {
  const {
    workspaceNamespace,
    workspaceName,
    initialName,
    initialSize,
    open,
    setOpen,
    action,
    mutation,
    reload,
  } = props;
  let loading = props.loading;
  const [form] = Form.useForm<{ name: string; size: number }>();

  useEffect(() => {
    if (open) {
      form.resetFields();

      if (initialName || initialSize) {
        form.setFieldsValue({
          name: initialName || '',
          size: initialSize || 1,
        });
      }
    }
  }, [form, open, initialName, initialSize]);

  return (
    <Modal
      title={action}
      style={{
        top: '25%',
        right: '15%',
        position: 'fixed',
      }}
      centered={false}
      open={open}
      onOk={() => setOpen(false)}
      onCancel={() => setOpen(false)}
      footer={[
        <Button key="cancel" onClick={() => setOpen(false)}>
          Cancel
        </Button>,
        <Button
          key="ok"
          type="primary"
          loading={loading}
          onClick={async () => {
            const values = form.getFieldsValue();
            const { name, size } = values;
            loading = true;
            await mutation({
              wsName: workspaceName,
              wsNs: workspaceNamespace,
              prettyName: name,
              size: String(size + 'G'),
            });
            reload();
            setOpen(false);
          }}
        >
          {findKeyByValue(Actions, action)}
        </Button>,
      ]}
    >
      <Form form={form} layout="vertical" autoComplete="off">
        <Form.Item
          name="name"
          label="Name"
          rules={[
            { required: true, message: 'Missing Name' },
            {
              validator: (_, value) =>
                value && value.trim().length > 2
                  ? Promise.resolve()
                  : Promise.reject(
                      new Error('Name must be at least 3 characters long'),
                    ),
            },
          ]}
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="size"
          label={
            <>
              Size{' '}
              <Tooltip title="Max size is 20GB, for larger Volumes please reach out to Crownlabs manager.">
                <InfoCircleOutlined style={{ marginLeft: '8px' }} />
              </Tooltip>
            </>
          }
        >
          <InputNumber
            min={initialSize}
            step={0.5}
            max={20}
            style={{ width: '120px' }}
            addonAfter="GB"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default SharedVolumeForm;
