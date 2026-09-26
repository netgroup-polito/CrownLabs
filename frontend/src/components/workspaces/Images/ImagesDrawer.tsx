import {
  CloseCircleOutlined,
  DeleteOutlined,
  InfoCircleOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import { Badge, Button, Drawer, Empty, Table, Tooltip } from 'antd';
import { useContext, useEffect, useMemo, useState, type FC } from 'react';
import {
  Phase4,
  type UpdatedWorkspaceImagesSubscription,
  useDeleteWorkspaceImageMutation,
  useWorkspaceImagesQuery,
} from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { ErrorTypes } from '../../../errorHandling/utils';
import { ModalAlert } from '../../common/ModalAlert';
import { updateWorkspaceImages } from './workspaceImagesUpdates';
import { updatedWorkspaceImages } from '../../../graphql-components/subscription';

type WorkspaceImage = {
  id: string;
  resourceName: string;
  name: string;
  author: string;
  description: string;
  size: string;
  createdAt: string;
  phase?: Phase4;
};

export interface ImagesDrawerProps {
  workspaceNamespace: string;
}

const formatCreationDate = (value?: string | null) => {
  if (!value) return '—';

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '—';

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date);
};

const ImagesDrawer: FC<ImagesDrawerProps> = ({ workspaceNamespace }) => {
  const { apolloErrorCatcher, makeErrorCatcher } = useContext(ErrorContext);
  const [open, setOpen] = useState(false);
  const [selectedImage, setSelectedImage] = useState<WorkspaceImage>();
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);

  const {
    data,
    loading: loadingImages,
    error: imagesError,
    subscribeToMore,
  } = useWorkspaceImagesQuery({
    variables: { workspaceNamespace },
    skip: !workspaceNamespace,
    fetchPolicy: 'network-only',
    onError: apolloErrorCatcher,
  });

  useEffect(() => {
    if (!workspaceNamespace || loadingImages || imagesError) return;

    // Keep the subscription active while mounted, even with the drawer closed,
    // so the badge and mutations from other views stay synchronized too.
    const unsubscribe = subscribeToMore<UpdatedWorkspaceImagesSubscription>({
      document: updatedWorkspaceImages,
      variables: { workspaceNamespace },
      onError: makeErrorCatcher(ErrorTypes.GenericError),
      updateQuery: (previous, { subscriptionData }) =>
        updateWorkspaceImages(
          previous,
          subscriptionData.data?.updatedImage,
          workspaceNamespace,
        ),
    });

    return () => unsubscribe();
  }, [
    makeErrorCatcher,
    imagesError,
    loadingImages,
    subscribeToMore,
    workspaceNamespace,
  ]);

  const [deleteWorkspaceImage, { loading: deletingImage }] =
    useDeleteWorkspaceImageMutation({
      onError: apolloErrorCatcher,
    });

  const images = useMemo<WorkspaceImage[]>(
    () =>
      (data?.imageList?.images ?? [])
        .filter(image => image?.metadata?.name)
        .map(image => ({
          id: image?.metadata?.uid ?? image?.metadata?.name ?? '',
          resourceName: image?.metadata?.name ?? '',
          name: image?.spec?.imageName ?? image?.metadata?.name ?? '',
          author: image?.spec?.tenantRef?.name ?? 'Unknown',
          description:
            image?.spec?.description?.trim() || 'No description provided.',
          size: image?.status?.artifact?.volumeSize
            ? String(image.status.artifact.volumeSize)
            : '—',
          createdAt: formatCreationDate(image?.metadata?.creationTimestamp),
          phase: image?.status?.phase,
        }))
        .sort((a, b) => a.name.localeCompare(b.name)),
    [data?.imageList?.images],
  );

  const confirmDelete = async () => {
    if (!selectedImage) return;

    await deleteWorkspaceImage({
      variables: {
        workspaceNamespace,
        imageName: selectedImage.resourceName,
      },
    });
    setShowDeleteModalConfirm(false);
    setSelectedImage(undefined);
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, image: WorkspaceImage) => (
        <span className="flex items-center gap-2">
          {name}
          <Tooltip
            title={
              <div>
                <div>{image.description}</div>
                <div className="mt-1">
                  <b>Author:</b> {image.author}
                </div>
              </div>
            }
            placement="top"
            trigger="hover"
            zIndex={2000}
            getPopupContainer={() => document.body}
          >
            <span
              className="inline-flex cursor-help"
              aria-label={`Description for ${image.name}`}
            >
              <InfoCircleOutlined />
            </span>
          </Tooltip>
          {image.phase === Phase4.Completed ? (
            <span className="success-color-fg font-medium">ready</span>
          ) : image.phase === Phase4.Failed ? (
            <Tooltip title="Image creation failed">
              <CloseCircleOutlined
                className="danger-color-fg"
                aria-label="Image creation failed"
              />
            </Tooltip>
          ) : (
            <Tooltip title="Image creation in progress">
              <LoadingOutlined
                className="warning-color-fg"
                aria-label="Image creation in progress"
                spin
              />
            </Tooltip>
          )}
        </span>
      ),
    },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
    },
    {
      title: 'Created',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: 'Action',
      key: 'action',
      width: 90,
      render: (_: unknown, image: WorkspaceImage) => (
        <Tooltip title="Delete image">
          <DeleteOutlined
            className="cursor-pointer"
            aria-label={`Delete ${image.name}`}
            onClick={() => {
              if (deletingImage) return;
              setSelectedImage(image);
              setShowDeleteModalConfirm(true);
            }}
          />
        </Tooltip>
      ),
    },
  ];

  return (
    <div
      className="flex justify-center items-center"
      style={{ marginTop: '1em' }}
    >
      <Badge count={images.length}>
        <Button
          className="xs:block"
          type="primary"
          shape="round"
          size="middle"
          onClick={() => setOpen(true)}
        >
          Images
        </Button>
      </Badge>

      <Drawer
        title="Images"
        placement="bottom"
        height={300}
        getContainer={false}
        open={open}
        onClose={() => setOpen(false)}
        rootStyle={{
          position: 'absolute',
          zIndex: 1000,
        }}
      >
        {images.length || loadingImages ? (
          <Table
            columns={columns}
            dataSource={images}
            rowKey="id"
            pagination={false}
            loading={loadingImages}
          />
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="No images found"
          />
        )}

        <ModalAlert
          headTitle="Confirm image deletion"
          message={
            <>
              Do you really want to delete <b>{selectedImage?.name}</b>?
              <br />
              The image and its associated data will no longer be available.
            </>
          }
          description="This action cannot be undone."
          type="warning"
          buttons={[
            <Button
              key="close"
              shape="round"
              className="mr-2 w-24"
              type="default"
              disabled={deletingImage}
              onClick={() => setShowDeleteModalConfirm(false)}
            >
              Close
            </Button>,
            <Button
              key="delete"
              shape="round"
              className="ml-2 w-24"
              type="primary"
              danger
              loading={deletingImage}
              onClick={() => void confirmDelete().catch(() => undefined)}
            >
              Delete
            </Button>,
          ]}
          show={showDeleteModalConfirm}
          setShow={setShowDeleteModalConfirm}
        />
      </Drawer>
    </div>
  );
};

export default ImagesDrawer;
