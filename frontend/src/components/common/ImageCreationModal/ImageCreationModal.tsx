import {
  Alert,
  Breadcrumb,
  Card,
  Checkbox,
  Input,
  Modal,
  Radio,
  Select,
  Space,
  theme,
  Typography,
} from 'antd';
import { useEffect, useMemo, useState, type FC } from 'react';
import {
  CROWNLABS_IMAGE_CREATION_HELP,
  VITE_APP_CROWNLABS_IMAGE_CREATION_WARNING,
} from '../../../env';

const { Text } = Typography;
const MAX_DESCRIPTION_CHARACTERS = 200;
const MAX_IMAGE_NAME_CHARACTERS = 63;
const RFC_1035_LABEL_PATTERN = /^[a-z](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

export type ImageDestination =
  | 'this-workspace'
  | 'another-workspace'
  | 'public-registry';

export type ImageDestinationSelection = {
  imageName: string;
  description: string;
  destination: ImageDestination;
  workspace?: string;
};

type WorkspaceOption = {
  value: string;
  label: string;
};

export interface ImageCreationModalProps {
  open: boolean;
  defaultImageName: string;
  breadcrumbItems: string[];
  currentWorkspaceName: string;
  isPersonalWorkspace: boolean;
  personalWorkspaceAvailable: boolean;
  canUseWorkspaceDestinations: boolean;
  canPublishToPublicRegistry: boolean;
  canCreateImage: boolean;
  creating?: boolean;
  loadingWorkspaces?: boolean;
  otherWorkspaceOptions: WorkspaceOption[];
  initialSelection?: ImageDestinationSelection;
  onCancel: () => void;
  onCreate: (selection: ImageDestinationSelection) => void;
}

const destinationOptions: Array<{
  value: ImageDestination;
  label: string;
  description: string;
}> = [
  {
    value: 'this-workspace',
    label: 'This Workspace',
    description: 'Keep the image in the workspace you are currently using.',
  },
  {
    value: 'another-workspace',
    label: 'Another Workspace',
    description: 'Make the image available in a different workspace.',
  },
  {
    value: 'public-registry',
    label: 'Public Registry',
    description: 'Publish the image in the shared public registry.',
  },
];

export const ImageCreationModal: FC<ImageCreationModalProps> = ({
  open,
  defaultImageName,
  breadcrumbItems,
  currentWorkspaceName,
  isPersonalWorkspace,
  personalWorkspaceAvailable,
  canUseWorkspaceDestinations,
  canPublishToPublicRegistry,
  canCreateImage,
  creating = false,
  loadingWorkspaces = false,
  otherWorkspaceOptions,
  initialSelection,
  onCancel,
  onCreate,
}) => {
  const { token } = theme.useToken();
  const [imageName, setImageName] = useState(
    initialSelection?.imageName ?? defaultImageName,
  );
  const [description, setDescription] = useState(
    initialSelection?.description ?? '',
  );
  const [destination, setDestination] = useState<ImageDestination>(
    initialSelection?.destination ??
      (canUseWorkspaceDestinations ? 'this-workspace' : 'public-registry'),
  );
  const [workspace, setWorkspace] = useState<string | undefined>(
    initialSelection?.workspace,
  );
  const [warningAcknowledged, setWarningAcknowledged] = useState(false);

  const selectableWorkspaces = useMemo(() => {
    const options = otherWorkspaceOptions.filter(
      option => option.value !== currentWorkspaceName,
    );

    if (isPersonalWorkspace || !personalWorkspaceAvailable) return options;

    return [
      { value: 'personal', label: 'Personal Workspace' },
      ...options.filter(option => option.value !== 'personal'),
    ];
  }, [
    currentWorkspaceName,
    isPersonalWorkspace,
    otherWorkspaceOptions,
    personalWorkspaceAvailable,
  ]);

  const visibleDestinationOptions = useMemo(
    () =>
      destinationOptions.filter(option => {
        if (option.value === 'public-registry') {
          return canPublishToPublicRegistry;
        }

        return canUseWorkspaceDestinations;
      }),
    [canPublishToPublicRegistry, canUseWorkspaceDestinations],
  );

  useEffect(() => {
    if (!open) return;
    setImageName(initialSelection?.imageName ?? defaultImageName);
    setDescription(initialSelection?.description ?? '');
    const initialDestination = initialSelection?.destination;
    setDestination(
      initialDestination &&
        visibleDestinationOptions.some(
          option => option.value === initialDestination,
        )
        ? initialDestination
        : (visibleDestinationOptions[0]?.value ?? 'public-registry'),
    );
    setWorkspace(initialSelection?.workspace);
    setWarningAcknowledged(false);
  }, [defaultImageName, initialSelection, open, visibleDestinationOptions]);

  const selectDestination = (selected: ImageDestination) => {
    setDestination(selected);
    if (selected !== 'another-workspace') setWorkspace(undefined);
  };

  const imageNameIsEmpty = imageName.trim().length === 0;
  const imageNameIsValid = RFC_1035_LABEL_PATTERN.test(imageName);
  const imageNameError = imageNameIsEmpty
    ? 'Image name is required.'
    : !imageNameIsValid
      ? 'Use 1–63 lowercase letters, numbers or hyphens. Start with a letter and end with a letter or number.'
      : undefined;
  const destinationIsAvailable = visibleDestinationOptions.some(
    option => option.value === destination,
  );
  const createDisabled =
    !imageNameIsValid ||
    !destinationIsAvailable ||
    !canCreateImage ||
    !warningAcknowledged ||
    (destination === 'another-workspace' && workspace === undefined);

  return (
    <Modal
      title={
        <Space size="middle" wrap>
          <span>Create new image</span>
          <Breadcrumb
            separator="/"
            items={breadcrumbItems.map(item => ({ title: item }))}
          />
        </Space>
      }
      open={open}
      okText="Create"
      onCancel={onCancel}
      onOk={() =>
        onCreate({
          imageName: imageName.trim(),
          description: description.trim(),
          destination,
          workspace,
        })
      }
      okButtonProps={{ disabled: createDisabled, loading: creating }}
      cancelButtonProps={{ disabled: creating }}
      centered
      width={760}
      destroyOnHidden
    >
      <Space direction="vertical" size="middle" className="w-full py-2">
        <div>
          <Input
            value={imageName}
            status={imageNameError ? 'error' : undefined}
            aria-label="Image name"
            placeholder="Enter an image name"
            maxLength={MAX_IMAGE_NAME_CHARACTERS}
            showCount
            onChange={event => setImageName(event.target.value)}
          />
          {imageNameError && (
            <Text type="danger" className="text-xs">
              {imageNameError}
            </Text>
          )}
          <Input.TextArea
            className="mt-2"
            value={description}
            aria-label="Image description"
            placeholder="Add a short image description"
            maxLength={MAX_DESCRIPTION_CHARACTERS}
            showCount
            autoSize={{ minRows: 2, maxRows: 4 }}
            onKeyDown={event => {
              if (event.key === 'Enter') event.preventDefault();
            }}
            onChange={event => setDescription(event.target.value)}
          />
        </div>

        <Text type="secondary">
          Select where the new image will be available:
        </Text>

        <Radio.Group
          value={destination}
          onChange={event => selectDestination(event.target.value)}
          className="w-full"
        >
          <div
            className="grid gap-3 items-stretch"
            style={{
              gridTemplateColumns: `repeat(${visibleDestinationOptions.length}, minmax(0, 1fr))`,
            }}
          >
            {visibleDestinationOptions.map(option => (
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
                      placeholder="Search or select a workspace"
                      value={workspace}
                      options={selectableWorkspaces}
                      loading={loadingWorkspaces}
                      showSearch
                      optionFilterProp="label"
                      onClick={event => event.stopPropagation()}
                      onChange={setWorkspace}
                    />
                  )}
              </Card>
            ))}
          </div>
        </Radio.Group>

        <Alert
          type="warning"
          showIcon
          message={
            <span style={{ fontSize: token.fontSizeLG }}>
              {VITE_APP_CROWNLABS_IMAGE_CREATION_WARNING.replace(/[.\s]+$/, '')}{' '}
              (
              <a
                href={CROWNLABS_IMAGE_CREATION_HELP}
                style={{
                  color: '#fff',
                  textDecoration: 'underline',
                  fontStyle: 'italic',
                }}
                target="_blank"
                rel="noopener noreferrer"
              >
                help!
              </a>
              )
            </span>
          }
          description={
            <Checkbox
              checked={warningAcknowledged}
              onChange={event => setWarningAcknowledged(event.target.checked)}
            >
              I have read and understood this information.
            </Checkbox>
          }
        />

        <Text type="secondary">
          Once the image has been created successfully, you can use it as the
          base for new templates.
        </Text>
      </Space>
    </Modal>
  );
};
