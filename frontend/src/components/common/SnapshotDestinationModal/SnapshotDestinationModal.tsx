import {
  CheckOutlined,
  EditOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Input,
  Modal,
  Radio,
  Select,
  Space,
  theme,
  Tooltip,
  Typography,
} from 'antd';
import { useEffect, useState, type FC } from 'react';

const { Text } = Typography;

export type SnapshotDestination =
  | 'personal-workspace'
  | 'another-workspace'
  | 'public-registry';

export type SnapshotDestinationSelection = {
  snapshotName: string;
  destination: SnapshotDestination;
  workspace?: string;
};

export interface SnapshotDestinationModalProps {
  open: boolean;
  defaultSnapshotName: string;
  initialSelection?: SnapshotDestinationSelection;
  onCancel: () => void;
  onConfirm: (selection: SnapshotDestinationSelection) => void;
}

const destinationOptions: Array<{
  value: SnapshotDestination;
  label: string;
  description: string;
}> = [
  {
    value: 'personal-workspace',
    label: 'Personal Workspace',
    description: 'Keep the snapshot in your personal CrownLabs workspace.',
  },
  {
    value: 'another-workspace',
    label: 'Another Workspace',
    description: 'Share the snapshot with a different workspace.',
  },
  {
    value: 'public-registry',
    label: 'Public Registry',
    description: 'Publish the snapshot in the shared public registry.',
  },
];

const workspaceOptions = ['workspaceX', 'workspaceY', 'workspaceZ'].map(
  workspace => ({
    value: workspace,
    label: workspace,
  }),
);

export const SnapshotDestinationModal: FC<
  SnapshotDestinationModalProps
> = ({
  open,
  defaultSnapshotName,
  initialSelection,
  onCancel,
  onConfirm,
}) => {
  const { token } = theme.useToken();
  const [snapshotName, setSnapshotName] = useState(
    initialSelection?.snapshotName ?? defaultSnapshotName,
  );
  const [editingName, setEditingName] = useState(false);
  const [destination, setDestination] = useState<SnapshotDestination>(
    initialSelection?.destination ?? 'personal-workspace',
  );
  const [workspace, setWorkspace] = useState<string | undefined>(
    initialSelection?.workspace,
  );

  useEffect(() => {
    if (!open) return;
    setSnapshotName(initialSelection?.snapshotName ?? defaultSnapshotName);
    setEditingName(false);
    setDestination(initialSelection?.destination ?? 'personal-workspace');
    setWorkspace(initialSelection?.workspace);
  }, [defaultSnapshotName, initialSelection, open]);

  const selectDestination = (selected: SnapshotDestination) => {
    setDestination(selected);
    if (selected !== 'another-workspace') setWorkspace(undefined);
  };

  const confirmDisabled =
    snapshotName.trim().length === 0 ||
    (destination === 'another-workspace' && workspace === undefined);

  return (
    <Modal
      title="Snapshot destination"
      open={open}
      okText="Confirm"
      onCancel={onCancel}
      onOk={() =>
        onConfirm({
          snapshotName: snapshotName.trim(),
          destination,
          workspace,
        })
      }
      okButtonProps={{ disabled: confirmDisabled }}
      centered
      width={760}
      destroyOnHidden
    >
      <Space direction="vertical" size="middle" className="w-full py-2">
        <Space.Compact className="w-full">
          <Input
            value={snapshotName}
            readOnly={!editingName}
            status={snapshotName.trim().length === 0 ? 'error' : undefined}
            aria-label="Snapshot name"
            onChange={event => setSnapshotName(event.target.value)}
            onPressEnter={() => setEditingName(false)}
          />
          <Button
            aria-label={editingName ? 'Save snapshot name' : 'Edit snapshot name'}
            icon={editingName ? <CheckOutlined /> : <EditOutlined />}
            onClick={() => setEditingName(current => !current)}
          >
            {editingName ? 'Save' : 'Edit'}
          </Button>
        </Space.Compact>

        <Space size={6}>
          <Text type="secondary">
            Select where the snapshot should be stored.
          </Text>
          <Tooltip title="Before creating a snapshot, reset cloud-init and complete any required manual configuration inside the VM.">
            <InfoCircleOutlined
              className="cursor-help"
              style={{ color: token.colorTextSecondary }}
            />
          </Tooltip>
        </Space>

        <Radio.Group
          value={destination}
          onChange={event => selectDestination(event.target.value)}
          className="w-full"
        >
          <div className="grid grid-cols-3 gap-3 items-stretch">
            {destinationOptions.map(option => (
              <Card
                key={option.value}
                size="small"
                hoverable
                onClick={() => selectDestination(option.value)}
                className="w-full h-full cursor-pointer"
                style={{
                  borderColor:
                    destination === option.value
                      ? token.colorPrimary
                      : undefined,
                }}
                styles={{ body: { padding: 14 } }}
              >
                <Radio value={option.value}>
                  <Text strong>{option.label}</Text>
                </Radio>

                <div className="ml-6 mt-1">
                  <Text type="secondary">{option.description}</Text>
                </div>

                {option.value === 'another-workspace' &&
                  destination === 'another-workspace' && (
                    <Select
                      className="mt-3 w-full"
                      placeholder="Select a workspace"
                      value={workspace}
                      options={workspaceOptions}
                      onClick={event => event.stopPropagation()}
                      onChange={setWorkspace}
                    />
                  )}
              </Card>
            ))}
          </div>
        </Radio.Group>

        <Text type="secondary">
          Once the image has been created successfully, you can use it as the
          base for new templates.
        </Text>
      </Space>
    </Modal>
  );
};
