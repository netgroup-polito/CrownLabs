import { Badge, Button, Drawer, Empty, Space, Table, Tooltip } from 'antd';
import { approximate, convertToGB, SharedVolume } from '../../../utils';
import { FC, useContext, useState } from 'react';
import {
  useApplySharedVolumeMutation,
  useCreateSharedVolumeMutation,
  useDeleteSharedVolumeMutation,
  useWorkspaceSharedVolumesQuery,
} from '../../../generated-types';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { makeGuiSharedVolume } from '../../../utilsLogic';
import { FetchPolicy } from '@apollo/client';
import { EditOutlined, DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import RowShVolStatus from './RowShVolStatus/RowShVolStatus';
import SharedVolumeForm, {
  Actions,
} from './SharedVolumeForms/SharedVolumeForm';

export interface ISharedVolumesDrawerProps {
  workspaceNamespace: string;
}

const fetchPolicy_networkOnly: FetchPolicy = 'network-only';
const hoverColor = 'rgb(129, 181, 255)';

const SharedVolumeDrawer: FC<ISharedVolumesDrawerProps> = ({ ...props }) => {
  const { workspaceNamespace } = props;

  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [open, setOpen] = useState(false);
  const [isCreateOpen, setCreateOpen] = useState(false);
  const [isEditOpen, setEditOpen] = useState(false);
  const [editShVolWorkspaceName, setEditShVolWorkspaceName] =
    useState<string>('');
  const [editName, setEditName] = useState<string>('');
  const [editSize, setEditSize] = useState<number>(1);
  const [dataShVols, setDataShVols] = useState<SharedVolume[]>([]);

  const {
    loading: loadingSharedVolumes,
    error: errorSharedVolumes,
    refetch: refetchSharedVolumes,
  } = useWorkspaceSharedVolumesQuery({
    variables: { workspaceNamespace },
    onError: apolloErrorCatcher,
    onCompleted: data =>
      setDataShVols(
        data.sharedvolumeList?.sharedvolumes
          ?.map(sv => makeGuiSharedVolume(sv))
          .sort((a, b) =>
            (a.prettyName ?? '').localeCompare(b.prettyName ?? '')
          ) ?? []
      ),
    fetchPolicy: fetchPolicy_networkOnly,
  });

  const [createShVolMutation, { loading: loadingCreateShVolMutation }] =
    useCreateSharedVolumeMutation({
      onError: apolloErrorCatcher,
    });

  const [applyShVolMutation, { loading: loadingApplyShVolMutation }] =
    useApplySharedVolumeMutation({
      onError: apolloErrorCatcher,
    });

  const [deleteShVolMutation, { loading: loadingDeleteShVolMutation }] =
    useDeleteSharedVolumeMutation({
      onError: apolloErrorCatcher,
    });

  const [intervalReload, setIntervalReload] = useState<NodeJS.Timeout>();
  const reloadSharedVolumes = async () => {
    if (intervalReload) clearInterval(intervalReload);

    let intv = setInterval(async () => {
      let res = await refetchSharedVolumes();
      setDataShVols(
        res.data?.sharedvolumeList?.sharedvolumes
          ?.map(sv => makeGuiSharedVolume(sv))
          .sort((a, b) =>
            (a.prettyName ?? '').localeCompare(b.prettyName ?? '')
          ) ?? []
      );
    }, 1500);
    setIntervalReload(intv);

    setTimeout(() => clearInterval(intv), 15000);
  };

  const columns = [
    {
      title: 'Name',
      dataIndex: 'prettyName',
      key: 'prettyName',
      render: (prettyName: string, shvol: SharedVolume) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <RowShVolStatus status={shvol.status} />
          <span>{prettyName}</span>
        </div>
      ),
    },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
      render: (size: string) => <span>{size}B</span>,
    },
    {
      title: 'Action',
      dataIndex: 'id',
      key: 'action',
      // eslint-disable-next-line react/no-multi-comp
      render: (_: any, shvol: SharedVolume) => (
        <span style={{ display: 'flex', gap: '8px' }}>
          <EditOutlined
            style={{ cursor: 'pointer' }}
            onMouseEnter={e => (e.currentTarget.style.color = hoverColor)}
            onMouseLeave={e => (e.currentTarget.style.color = '')}
            onClick={() => {
              setEditShVolWorkspaceName(shvol.name);
              setEditName(shvol.prettyName);
              setEditSize(approximate(convertToGB(shvol.size), 2) || 0.01);
              setEditOpen(true);
            }}
          />
          <Tooltip title="Be mindful you can't delete a Shared Volume that is mounted on a Template. Unmount it before deletion.">
            <DeleteOutlined
              style={{ cursor: 'pointer' }}
              onMouseEnter={e => (e.currentTarget.style.color = hoverColor)}
              onMouseLeave={e => (e.currentTarget.style.color = '')}
              onClick={async () => {
                if (!loadingDeleteShVolMutation) {
                  await deleteShVolMutation({
                    variables: {
                      workspaceNamespace: shvol.namespace,
                      name: shvol.name,
                    },
                  });
                  reloadSharedVolumes();
                }
              }}
            />
          </Tooltip>
        </span>
      ),
    },
  ];

  return (
    <div>
      {!loadingSharedVolumes && !errorSharedVolumes && dataShVols ? (
        <div
          className="flex justify-center items-center"
          style={{
            marginTop: '1em',
          }}
        >
          <Badge count={dataShVols.length}>
            <Button
              className="xs:block"
              type="primary"
              shape="round"
              size={'middle'}
              onClick={() => setOpen(true)}
            >
              Shared Volumes
            </Button>
          </Badge>

          {/*FIXME: Someone makes appear a scroll bar when Drawer is opening. */}
          <Drawer
            title="Shared Volumes"
            placement="bottom"
            height={open ? 300 : 0}
            getContainer={false}
            destroyOnClose={true}
            open={open}
            closable={true}
            onClose={() => setOpen(false)}
            style={{
              position: 'absolute',
              opacity: open ? 1 : 0,
              transition: 'all 0.3',
              overflow: open ? 'auto' : 'hidden',
            }}
            extra={
              <Space>
                <Button
                  type="primary"
                  shape="circle"
                  icon={<PlusOutlined />}
                  size="middle"
                  onClick={() => setCreateOpen(true)}
                />
                <SharedVolumeForm
                  key="create-form"
                  open={isCreateOpen}
                  setOpen={setCreateOpen}
                  workspaceNamespace={workspaceNamespace}
                  action={Actions.Create}
                  mutation={createShVolMutation}
                  loading={loadingCreateShVolMutation}
                  reload={reloadSharedVolumes}
                  size={0.5}
                />
              </Space>
            }
          >
            {dataShVols.length ? (
              <>
                <Table columns={columns} dataSource={dataShVols} />
                <SharedVolumeForm
                  key={`edit-form-${editName}`}
                  open={isEditOpen}
                  setOpen={setEditOpen}
                  workspaceNamespace={workspaceNamespace}
                  workspaceName={editShVolWorkspaceName}
                  action={Actions.Update}
                  name={editName}
                  size={editSize}
                  mutation={applyShVolMutation}
                  loading={loadingApplyShVolMutation}
                  reload={reloadSharedVolumes}
                />
              </>
            ) : (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="No Shared Volumes found"
              />
            )}
          </Drawer>
        </div>
      ) : null}
    </div>
  );
};

export default SharedVolumeDrawer;
