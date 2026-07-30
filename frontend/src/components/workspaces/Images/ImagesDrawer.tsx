import { DeleteOutlined } from '@ant-design/icons';
import {
  Badge,
  Button,
  Drawer,
  Empty,
  Table,
  Tooltip,
} from 'antd';
import { useState, type FC } from 'react';
import { ModalAlert } from '../../common/ModalAlert';

type MockSnapshotImage = {
  id: string;
  name: string;
  size: string;
  createdAt: string;
};

const initialMockImages: MockSnapshotImage[] = [
  {
    id: 'snapshot-image-1',
    name: 'snapshot-of-ubuntu-dev',
    size: '12.4 GiB',
    createdAt: '29 Jul 2026, 14:35',
  },
  {
    id: 'snapshot-image-2',
    name: 'snapshot-of-cloud-vm',
    size: '8.7 GiB',
    createdAt: '30 Jul 2026, 09:10',
  },
];

const ImagesDrawer: FC = () => {
  const [open, setOpen] = useState(false);
  const [images, setImages] =
    useState<MockSnapshotImage[]>(initialMockImages);
  const [selectedImage, setSelectedImage] = useState<MockSnapshotImage>();
  const [showDeleteModalConfirm, setShowDeleteModalConfirm] = useState(false);

  const columns = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
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
      render: (_: unknown, image: MockSnapshotImage) => (
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
                  type="primary"
                  onClick={() => setShowDeleteModalConfirm(false)}
                >
                  Close
                </Button>,
                <Button
                  key="delete"
                  shape="round"
                  className="ml-2 w-24"
                  color="red"
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
