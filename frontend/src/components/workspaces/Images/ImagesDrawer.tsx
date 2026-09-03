import { DeleteOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { Badge, Button, Drawer, Empty, Table, Tooltip } from 'antd';
import { useState, type FC } from 'react';
import { ModalAlert } from '../../common/ModalAlert';

type MockImage = {
  id: string;
  name: string;
  description: string;
  size: string;
  createdAt: string;
};

const initialMockImages: MockImage[] = [
  {
    id: 'image-1',
    name: 'image-of-ubuntu-dev',
    description:
      'Ubuntu 24.04 development image with cloud-init reset, Docker, Git, Python, and the CrownLabs base toolchain preinstalled.',
    size: '12.4 GiB',
    createdAt: '29 Jul 2026, 14:35',
  },
  {
    id: 'image-2',
    name: 'image-of-cloud-vm',
    description:
      'Cloud VM image prepared for networking labs, with common diagnostic tools and the initial system configuration already completed.',
    size: '8.7 GiB',
    createdAt: '30 Jul 2026, 09:10',
  },
];

const ImagesDrawer: FC = () => {
  const [open, setOpen] = useState(false);
  const [images, setImages] = useState<MockImage[]>(initialMockImages);
  const [selectedImage, setSelectedImage] = useState<MockImage>();
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, image: MockImage) => (
        <span className="flex items-center gap-2">
          {name}
          <Tooltip
            title={image.description}
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
      render: (_: unknown, image: MockImage) => (
        <Tooltip title="Delete image">
          <DeleteOutlined
            className="cursor-pointer"
            aria-label={`Delete ${image.name}`}
            onClick={() => {
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
        {images.length ? (
          <>
            <Table
              columns={columns}
              dataSource={images}
              rowKey="id"
              pagination={false}
            />
            <ModalAlert
              headTitle="Confirm image deletion"
              message={
                <>
                  Do you really want to delete <b>{selectedImage?.name}</b>?
                  <br />
                  The image will no longer be available.
                </>
              }
              description="This is a frontend prototype. The deletion only affects the mock data."
              type="warning"
              buttons={[
                <Button
                  key="close"
                  shape="round"
                  className="mr-2 w-24"
                  type="default"
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
                  onClick={() => {
                    if (selectedImage) {
                      setImages(current =>
                        current.filter(
                          currentImage => currentImage.id !== selectedImage.id,
                        ),
                      );
                    }
                    setShowDeleteModalConfirm(false);
                    setSelectedImage(undefined);
                  }}
                >
                  Delete
                </Button>,
              ]}
              show={showDeleteModalConfirm}
              setShow={setShowDeleteModalConfirm}
            />
          </>
        ) : (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="No images found"
          />
        )}
      </Drawer>
    </div>
  );
};

export default ImagesDrawer;
