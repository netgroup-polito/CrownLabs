import React, { useContext, useMemo, useState } from 'react';
import {
  Table,
  Input,
  Spin,
  Col,
  Tooltip,
  Space,
  DatePicker,
  Popconfirm,
  message,
  Button,
  Select,
  InputNumber,
} from 'antd';
import dayjs from 'dayjs';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import {
  useTenantsQuery,
  useDeleteTenantMutation,
} from '../../../generated-types';
import { makeTenantsList } from '../../../utilsLogic';
import { multiStringIncludes, type Tenant } from '../../../utils';
import Box from '../../common/Box';
import { EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

export default function TenantListPage() {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const navigate = useNavigate();

  const [searchText, setSearchText] = useState('');
  const [registrationDateRange, setRegistrationDateRange] = useState<
    [dayjs.Dayjs | null, dayjs.Dayjs | null] | null
  >(null);
  const [lastLoginDateRange, setLastLoginDateRange] = useState<
    [dayjs.Dayjs | null, dayjs.Dayjs | null] | null
  >(null);
  const [labelKeyFilter, setLabelKeyFilter] = useState('');
  const [labelValueFilter, setLabelValueFilter] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [selectedOperators, setSelectedOperators] = useState<string[]>([]);
  const [minWorkspaces, setMinWorkspaces] = useState<number | null>(null);
  const [maxWorkspaces, setMaxWorkspaces] = useState<number | null>(null);
  const [personalWorkspaceActive, setPersonalWorkspaceActive] = useState<string | null>(null);

  const { data, loading, error, refetch } = useTenantsQuery({
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
    variables: {
      retrieveWorkspaces: true,
    },
  });

  const [deleteTenantMutation, { loading: deleteLoading }] =
    useDeleteTenantMutation({
      onError: apolloErrorCatcher,
    });

  const handleDeleteTenant = async (name: string) => {
    try {
      await deleteTenantMutation({ variables: { name } });
      message.success('Tenant deleted successfully');
      refetch();
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteMultipleTenants = async () => {
    try {
      await Promise.all(
        selectedRowKeys.map(name =>
          deleteTenantMutation({ variables: { name: name as string } }),
        ),
      );
      message.success(`${selectedRowKeys.length} tenants deleted successfully`);
      setSelectedRowKeys([]);
      refetch();
    } catch (e) {
      console.error(e);
    }
  };

  const tenants = useMemo(() => makeTenantsList(data), [data]);

  const operatorSelectorValues = useMemo(() => {
    const values = new Set<string>();
    tenants.forEach(tenant => {
      const val = tenant.labels?.['crownlabsPolitoItOperatorSelector'];
      if (val) values.add(val);
    });
    return Array.from(values).sort();
  }, [tenants]);

  const filteredTenants = useMemo(
    () =>
      tenants.filter(tenant => {
        const searchMatches = multiStringIncludes(
          searchText,
          tenant.name,
          tenant.surname,
          tenant.userid,
        );

        let matchesReg = true;
        if (
          registrationDateRange &&
          registrationDateRange[0] &&
          registrationDateRange[1]
        ) {
          if (!tenant.creationDate) matchesReg = false;
          else {
            const date = dayjs(tenant.creationDate);
            matchesReg =
              date.isAfter(registrationDateRange[0].startOf('day')) &&
              date.isBefore(registrationDateRange[1].endOf('day'));
          }
        }

        let matchesLogin = true;
        if (
          lastLoginDateRange &&
          lastLoginDateRange[0] &&
          lastLoginDateRange[1]
        ) {
          if (!tenant.lastLogin) matchesLogin = false;
          else {
            const date = dayjs(tenant.lastLogin);
            matchesLogin =
              date.isAfter(lastLoginDateRange[0].startOf('day')) &&
              date.isBefore(lastLoginDateRange[1].endOf('day'));
          }
        }

        let matchesLabel = true;
        if (labelKeyFilter || labelValueFilter) {
          if (!tenant.labels) {
            matchesLabel = false;
          } else {
            if (labelKeyFilter && labelValueFilter) {
              matchesLabel =
                tenant.labels[labelKeyFilter] !== undefined &&
                tenant.labels[labelKeyFilter]
                  .toLowerCase()
                  .includes(labelValueFilter.toLowerCase());
            } else if (labelKeyFilter) {
              matchesLabel = tenant.labels[labelKeyFilter] !== undefined;
            } else if (labelValueFilter) {
              matchesLabel = Object.values(tenant.labels).some(v =>
                v.toLowerCase().includes(labelValueFilter.toLowerCase()),
              );
            }
          }
        }

        let matchesOperator = true;
        if (selectedOperators.length > 0) {
          const val = tenant.labels?.['crownlabsPolitoItOperatorSelector'];
          if (!val || !selectedOperators.includes(val)) {
            matchesOperator = false;
          }
        }

        let matchesWorkspaces = true;
        const wsCount = tenant.workspaces?.length || 0;
        if (minWorkspaces !== null && wsCount < minWorkspaces) matchesWorkspaces = false;
        if (maxWorkspaces !== null && wsCount > maxWorkspaces) matchesWorkspaces = false;

        let matchesPW = true;
        if (personalWorkspaceActive === 'yes' && !tenant.personalWorkspace) matchesPW = false;
        if (personalWorkspaceActive === 'no' && tenant.personalWorkspace) matchesPW = false;

        return (
          searchMatches &&
          matchesReg &&
          matchesLogin &&
          matchesLabel &&
          matchesOperator &&
          matchesWorkspaces &&
          matchesPW
        );
      }),
    [
      tenants,
      searchText,
      registrationDateRange,
      lastLoginDateRange,
      labelKeyFilter,
      labelValueFilter,
      selectedOperators,
      minWorkspaces,
      maxWorkspaces,
      personalWorkspaceActive,
    ],
  );

  const handleSearch = (value: string) => {
    setSearchText(value.toLowerCase());
  };

  return (
    <Col span={24} lg={22} xxl={20} className="h-full">
      <Box
        header={{
          className: 'py-4 md:py-6 h-auto',
          center: (
            <div className="flex flex-col justify-center items-center gap-4 w-full px-2 sm:px-4">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Manage users</b>
              </p>

              <div className="flex flex-col w-full items-center gap-4">
                <Input.Search
                  placeholder="Search users"
                  className="w-full max-w-xs"
                  onSearch={handleSearch}
                  enterButton
                  allowClear={true}
                />

                <div className="flex flex-wrap justify-center gap-3 w-full">
                  {selectedRowKeys.length > 0 && (
                    <Popconfirm
                      title={`You are going to delete ${selectedRowKeys.length} users. Are you sure?`}
                      onConfirm={handleDeleteMultipleTenants}
                      okText="Yes"
                      cancelText="No"
                    >
                      <Button
                        type="primary"
                        danger
                        icon={<DeleteOutlined />}
                        loading={deleteLoading}
                      >
                        Delete Selected ({selectedRowKeys.length})
                      </Button>
                    </Popconfirm>
                  )}
                  <DatePicker.RangePicker
                    className="w-full sm:w-auto"
                    placeholder={['Reg. Start', 'Reg. End']}
                    onChange={dates => setRegistrationDateRange(dates)}
                    allowClear
                  />
                  <DatePicker.RangePicker
                    className="w-full sm:w-auto"
                    placeholder={['Login Start', 'Login End']}
                    onChange={dates => setLastLoginDateRange(dates)}
                    allowClear
                  />
                  <Space.Compact className="w-full sm:w-auto flex">
                    <Input
                      placeholder="Label Key"
                      onChange={e => setLabelKeyFilter(e.target.value)}
                      allowClear
                      style={{ width: '40%' }}
                    />
                    <Input
                      placeholder="Label Value"
                      onChange={e => setLabelValueFilter(e.target.value)}
                      allowClear
                      style={{ width: '60%' }}
                    />
                  </Space.Compact>
                  <Select
                    mode="multiple"
                    allowClear
                    placeholder="Filter by Operator"
                    className="w-full sm:w-auto"
                    style={{ minWidth: 200 }}
                    value={selectedOperators}
                    onChange={setSelectedOperators}
                    options={operatorSelectorValues.map(v => ({
                      label: v,
                      value: v,
                    }))}
                  />
                  <Space.Compact className="w-full sm:w-auto flex">
                    <InputNumber
                      placeholder="Min Workspaces"
                      value={minWorkspaces}
                      onChange={setMinWorkspaces}
                      style={{ width: '50%' }}
                    />
                    <InputNumber
                      placeholder="Max Workspaces"
                      value={maxWorkspaces}
                      onChange={setMaxWorkspaces}
                      style={{ width: '50%' }}
                    />
                  </Space.Compact>
                  <Select
                    allowClear
                    placeholder="Personal Workspace"
                    className="w-full sm:w-auto"
                    style={{ minWidth: 180 }}
                    value={personalWorkspaceActive}
                    onChange={setPersonalWorkspaceActive}
                    options={[
                      { label: 'Yes', value: 'yes' },
                      { label: 'No', value: 'no' },
                    ]}
                  />
                </div>
              </div>
            </div>
          ),
        }}
      >
        <Spin spinning={loading || error != null}>
          <Table
            pagination={{ defaultPageSize: 10 }}
            dataSource={filteredTenants}
            size="small"
            rowKey="userid"
            rowSelection={{
              selectedRowKeys,
              onChange: setSelectedRowKeys,
            }}
          >
            <Table.Column
              title="User ID"
              dataIndex="userid"
              sorter={(a: Tenant, b: Tenant) =>
                a.userid.localeCompare(b.userid)
              }
              key="userid"
              width={170}
            />
            <Table.Column
              responsive={['md', 'lg']}
              title="Name"
              dataIndex="name"
              sorter={(a: Tenant, b: Tenant) => a.name.localeCompare(b.name)}
              key="name"
              width={120}
            />
            <Table.Column
              responsive={['md', 'lg']}
              title="Surname"
              dataIndex="surname"
              sorter={(a: Tenant, b: Tenant) =>
                a.surname.localeCompare(b.surname)
              }
              key="surname"
              width={120}
            />
            <Table.Column
              responsive={['sm', 'md', 'lg']}
              title="Email"
              dataIndex="email"
              ellipsis={true}
              key="email"
              width={150}
            />
            <Table.Column
              responsive={['md', 'lg']}
              title="Creation date"
              dataIndex="creationDate"
              render={(date: string) =>
                date ? dayjs(date).format('YYYY-MM-DD') : 'N/A'
              }
              sorter={(a: Tenant, b: Tenant) =>
                (a.creationDate || '').localeCompare(b.creationDate || '')
              }
              key="creationDate"
              width={140}
            />
            <Table.Column
              responsive={['lg']}
              title="Last login"
              dataIndex="lastLogin"
              render={(date: string) =>
                date ? dayjs(date).format('YYYY-MM-DD HH:mm') : 'N/A'
              }
              sorter={(a: Tenant, b: Tenant) =>
                (a.lastLogin || '').localeCompare(b.lastLogin || '')
              }
              key="lastLogin"
              width={150}
            />
            <Table.Column
              responsive={['sm', 'md', 'lg']}
              title="Workspaces"
              dataIndex="workspaces"
              render={(workspaces: Tenant['workspaces']) =>
                workspaces ? workspaces.length : 0
              }
              sorter={(a: Tenant, b: Tenant) =>
                (a.workspaces ? a.workspaces.length : 0) -
                (b.workspaces ? b.workspaces.length : 0)
              }
              key="workspaces"
              width={120}
            />
            <Table.Column
              responsive={['sm', 'md', 'lg']}
              title="Personal Workspace"
              dataIndex="personalWorkspace"
              render={(active: boolean) => (active ? 'Yes' : 'No')}
              sorter={(a: Tenant, b: Tenant) =>
                (a.personalWorkspace === b.personalWorkspace) ? 0 : a.personalWorkspace ? -1 : 1
              }
              key="personalWorkspace"
              width={140}
            />
            <Table.Column
              title="Actions"
              key="actions"
              width={100}
              render={(tenant: Tenant) => (
                <Space>
                  <Tooltip title="Edit tenant">
                    <Button
                      type="text"
                      icon={<EditOutlined />}
                      onClick={() => navigate('/tenants/' + tenant.userid)}
                    />
                  </Tooltip>
                  <Tooltip title="Delete tenant">
                    <Popconfirm
                      title="You are going to delete this user. Are you sure?"
                      onConfirm={() => handleDeleteTenant(tenant.userid)}
                      okText="Yes"
                      cancelText="No"
                    >
                      <Button type="text" danger icon={<DeleteOutlined />} />
                    </Popconfirm>
                  </Tooltip>
                </Space>
              )}
            />
          </Table>
        </Spin>
      </Box>
    </Col>
  );
}
