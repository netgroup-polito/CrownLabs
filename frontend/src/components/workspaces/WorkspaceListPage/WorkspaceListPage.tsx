import { useContext, useMemo, useState } from 'react';
import { Table, Input, Spin, Col, Tooltip, Button } from 'antd';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useWorkspacesQuery, useAllTemplatesQuery, AutoEnroll } from '../../../generated-types';
import Box from '../../common/Box';
import { EditOutlined, PlusOutlined } from '@ant-design/icons';
import ModalCreateWorkspace, { type WorkspaceEditData } from '../ModalCreateWorkspace';

interface WorkspaceData {
  name: string;
  prettyName: string;
  autoEnroll: string;
  cpu: string;
  memory: string;
  instances: number;
  templateCount: number;
  key: string;
}

export default function WorkspaceListPage() {
  const { apolloErrorCatcher } = useContext(ErrorContext);

  const [searchText, setSearchText] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editWorkspace, setEditWorkspace] = useState<WorkspaceEditData | null>(null);
  const [editWorkspaceTemplateCount, setEditWorkspaceTemplateCount] = useState<number>(0);

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
      const workspaceName = ws?.metadata?.name || '';
      const workspaceNamespace = `workspace-${workspaceName}`;
      return {
        name: workspaceName,
        prettyName: ws?.spec?.prettyName || '',
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
    setEditWorkspaceTemplateCount(workspace.templateCount);
    setShowCreateModal(true);
  };

  const handleCloseModal = () => {
    setShowCreateModal(false);
    setEditWorkspace(null);
    setEditWorkspaceTemplateCount(0);
  };

  return (
    <Col span={24} lg={22} xxl={20} className="h-full">
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
            pagination={{ defaultPageSize: 10 }}
            dataSource={filteredWorkspaces}
            size="small"
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
              render={(autoEnroll: string) => 
                autoEnroll === '_EMPTY_' ? '-' : autoEnroll
              }
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
              width={60}
              render={(workspace: WorkspaceData) => (
                <Tooltip title="Edit workspace">
                  <EditOutlined
                    className="mr-2"
                    onClick={() => handleEditWorkspace(workspace)}
                  />
                </Tooltip>
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
        templateCount={editWorkspaceTemplateCount}
      />
    </Col>
  );
}
