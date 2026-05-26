import { useContext, useMemo, useState } from 'react';
import { Modal, Table, Input, Spin, Col, Tooltip, Button, Badge, message } from 'antd';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useWorkspacesQuery, useAllTemplatesQuery, useDeleteWorkspaceMutation, AutoEnroll } from '../../../generated-types';
import Box from '../../common/Box';
import { DeleteOutlined, EditOutlined, ExclamationCircleOutlined, PlusOutlined, UserSwitchOutlined } from '@ant-design/icons';
import ModalCreateWorkspace, { type WorkspaceEditData } from '../ModalCreateWorkspace';
import UserListLogic from '../../accountPage/UserListLogic/UserListLogic';
import { WorkspaceRole, type Workspace } from '../../../utils';

interface WorkspaceData {
  name: string;
  prettyName: string;
  deleting: boolean;
  autoEnroll: string;
  cpu: string;
  memory: string;
  instances: number;
  templateCount: number;
  key: string;
}

export default function WorkspaceListPage() {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const [modal, modalContextHolder] = Modal.useModal();
  const [messageApi, messageContextHolder] = message.useMessage();
  const [deleteWorkspace] = useDeleteWorkspaceMutation({
    onError: apolloErrorCatcher,
  });

  const [searchText, setSearchText] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editWorkspace, setEditWorkspace] = useState<WorkspaceEditData | null>(null);
  const [usersModalWorkspace, setUsersModalWorkspace] = useState<WorkspaceData | null>(null);

  const { data, loading, error, refetch } = useWorkspacesQuery({
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
  });

  const { data: templatesData, loading: templatesLoading } = useAllTemplatesQuery({
    onError: apolloErrorCatcher,
  });

  const templateCountByWorkspace = useMemo(() => {
    const countMap = new Map<string, number>();
    if (templatesData?.allTemplates?.items) {
      templatesData.allTemplates.items.forEach((template) => {
        const namespace = template?.metadata?.namespace;
        if (namespace) {
          countMap.set(namespace, (countMap.get(namespace) || 0) + 1);
        }
      });
    }
    return countMap;
  }, [templatesData]);

  const workspaces = useMemo<WorkspaceData[]>(() => {
    if (!data?.workspaces?.items) return [];
    return data.workspaces.items.map(ws => {
      const metadata = ws?.metadata as
        | { name?: string | null; deletionTimestamp?: string | null }
        | undefined;
      const workspaceName = ws?.metadata?.name || '';
      const workspaceNamespace = `workspace-${workspaceName}`;
      return {
        name: workspaceName,
        prettyName: ws?.spec?.prettyName || '',
        deleting: Boolean(metadata?.deletionTimestamp),
        autoEnroll: ws?.spec?.autoEnroll || AutoEnroll.Empty,
        cpu: ws?.spec?.quota?.cpu || '0',
        memory: ws?.spec?.quota?.memory || '0Gi',
        instances: ws?.spec?.quota?.instances || 0,
        templateCount: templateCountByWorkspace.get(workspaceNamespace) || 0,
        key: workspaceName,
      };
    });
  }, [data, templateCountByWorkspace]);

  const filteredWorkspaces = useMemo(
    () =>
      workspaces.filter(workspace =>
        [workspace.name, workspace.prettyName]
          .join(' ')
          .toLowerCase()
          .includes(searchText),
      ),
    [workspaces, searchText],
  );

  const handleSearch = (value: string) => {
    setSearchText(value.toLowerCase());
  };

  const handleEditWorkspace = (workspace: WorkspaceData) => {
    setEditWorkspace({
      name: workspace.name,
      prettyName: workspace.prettyName,
      autoEnroll: workspace.autoEnroll,
      cpu: workspace.cpu,
      memory: workspace.memory,
      instances: workspace.instances,
    });
    setShowCreateModal(true);
  };

  const handleCloseModal = () => {
    setShowCreateModal(false);
    setEditWorkspace(null);
  };

  const handleDeleteWorkspace = (workspace: WorkspaceData) => {
    modal.confirm({
      title: 'Delete Workspace',
      icon: <ExclamationCircleOutlined />,
      content: `Are you sure you want to delete workspace "${workspace.prettyName}"? This action cannot be undone.`,
      okText: 'Delete',
      okType: 'danger',
      cancelText: 'Cancel',
      onOk: async () => {
        await deleteWorkspace({ variables: { name: workspace.name } });
        messageApi.success(`Workspace "${workspace.prettyName}" deleted successfully`);
        refetch();
      },
    });
  };

  return (
    <Col span={24} lg={22} xxl={20} className="h-full min-h-0">
      {modalContextHolder}
      {messageContextHolder}
      <Box
        header={{
          size: 'large',
          right: (
            <div className="h-full flex-none flex justify-center items-center w-20">
              <Tooltip title="Create new workspace">
                <Button
                  type="primary"
                  shape="circle"
                  size="large"
                  icon={<PlusOutlined />}
                  onClick={() => setShowCreateModal(true)}
                />
              </Tooltip>
            </div>
          ),
          center: (
            <div className="h-full flex flex-col justify-center items-center gap-4">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Manage workspaces</b>
              </p>

              <Input.Search
                placeholder="Search workspaces"
                style={{ width: 300 }}
                onSearch={handleSearch}
                enterButton
                allowClear={true}
              />
            </div>
          ),
        }}
      >
        <Spin spinning={loading || templatesLoading || error != null}>
          <Table
            pagination={false}
            dataSource={filteredWorkspaces}
            size="small"
            scroll={{ x: 'max-content' }}
          >
            <Table.Column
              title="Name"
              dataIndex="name"
              sorter={(a: WorkspaceData, b: WorkspaceData) =>
                a.name.localeCompare(b.name)
              }
              key="name"
              width={200}
            />
            <Table.Column
              responsive={['md', 'lg']}
              title="Pretty Name"
              dataIndex="prettyName"
              sorter={(a: WorkspaceData, b: WorkspaceData) =>
                a.prettyName.localeCompare(b.prettyName)
              }
              key="prettyName"
              width={200}
            />
            <Table.Column
              responsive={['sm', 'md', 'lg']}
              title="Auto Enroll"
              dataIndex="autoEnroll"
              key="autoEnroll"
              width={150}
              render={(autoEnroll: string) => {
                const labels: Record<string, string> = {
                  [AutoEnroll.Empty]: 'No',
                  [AutoEnroll.Immediate]: 'Yes',
                  [AutoEnroll.WithApproval]: 'Require approval',
                };
                return labels[autoEnroll] ?? autoEnroll;
              }}
            />
            <Table.Column
              title="Templates"
              dataIndex="templateCount"
              key="templateCount"
              width={100}
              sorter={(a: WorkspaceData, b: WorkspaceData) =>
                a.templateCount - b.templateCount
              }
            />
            <Table.Column
              title="Actions"
              key="actions"
              width={100}
              render={(workspace: WorkspaceData) => (
                workspace.deleting ? (
                  <Badge status="processing" text="Deleting..." />
                ) : (
                  <div className="flex gap-2">
                    <Tooltip title="Edit workspace">
                      <EditOutlined onClick={() => handleEditWorkspace(workspace)} />
                    </Tooltip>
                    <Tooltip title="Manage users">
                      <UserSwitchOutlined onClick={() => setUsersModalWorkspace(workspace)} />
                    </Tooltip>
                    <Tooltip
                      title={
                        workspace.templateCount > 0
                          ? `Cannot delete: workspace has ${workspace.templateCount} template${workspace.templateCount > 1 ? 's' : ''}`
                          : 'Delete workspace'
                      }
                    >
                      <DeleteOutlined
                        className={workspace.templateCount > 0 ? 'cursor-not-allowed opacity-30' : 'text-red-500'}
                        onClick={() => workspace.templateCount === 0 && handleDeleteWorkspace(workspace)}
                      />
                    </Tooltip>
                  </div>
                )
              )}
            />
          </Table>
        </Spin>
      </Box>
      <ModalCreateWorkspace
        show={showCreateModal}
        setShow={handleCloseModal}
        onSuccess={() => refetch()}
        editWorkspace={editWorkspace}
        existingWorkspaceNames={workspaces.map(ws => ws.name)}
      />
      <Modal
        destroyOnHidden={true}
        title={`Users in ${usersModalWorkspace?.prettyName ?? ''}`}
        width={800}
        open={usersModalWorkspace !== null}
        footer={null}
        onCancel={() => setUsersModalWorkspace(null)}
      >
        {usersModalWorkspace && (
          <UserListLogic
            workspace={{
              name: usersModalWorkspace.name,
              namespace: `workspace-${usersModalWorkspace.name}`,
              prettyName: usersModalWorkspace.prettyName,
              role: WorkspaceRole.manager,
            } satisfies Workspace}
          />
        )}
      </Modal>
    </Col>
  );
}
