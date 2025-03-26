import { InfoCircleOutlined } from '@ant-design/icons';
import { Button, Form, Input, InputNumber, Modal, Tooltip } from 'antd';
import { FC, useEffect } from 'react';
import { getShVolPatchJson } from '../../../../graphql-components/utils';
import { findKeyByValue } from '../../../../utils';

export enum Actions {
  Create = 'Create a new Shared Volume',
  Update = 'Update an existing Shared Volume',
}

export interface ISharedVolumesFormProps {
  workspaceNamespace: string;
  workspaceName?: string;
  initialName?: string;
  initialSize?: number;
  open: boolean;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  action: Actions;
  mutation: Function;
  loading: boolean;
  reload: Function;
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
            if (action === Actions.Update) {
              await mutation({
                variables: {
                  workspaceNamespace: workspaceNamespace,
                  name: workspaceName,
                  patchJson: getShVolPatchJson({
                    prettyName: name,
                    size: String(size + 'G'),
                  }),
                  manager: 'frontend-shvol-editor',
                },
              });
            } else {
              await mutation({
                variables: {
                  workspaceNamespace: workspaceNamespace,
                  prettyName: name,
                  size: String(size + 'G'),
                },
              });
            }
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
                      new Error('Name must be at least 3 characters long')
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
