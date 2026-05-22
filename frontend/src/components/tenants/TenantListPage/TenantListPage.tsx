import React, { useContext, useEffect, useMemo, useRef, useState } from 'react';
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
  theme,
} from 'antd';
import type { FilterDropdownProps } from 'antd/es/table/interface';
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
import { useSearchParams } from 'react-router-dom';

export default function TenantListPage() {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const { token } = theme.useToken();
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

  const searchText = searchParams.get('q') ?? '';
  const regStart = searchParams.get('regStart');
  const regEnd = searchParams.get('regEnd');
  const loginStart = searchParams.get('loginStart');
  const loginEnd = searchParams.get('loginEnd');
  const labelKeyFilter = searchParams.get('labelKey') ?? '';
  const labelValueFilter = searchParams.get('labelVal') ?? '';
  const selectedOperators = searchParams.getAll('op');
  const minWorkspaces = searchParams.get('minWs')
    ? Number(searchParams.get('minWs'))
    : null;
  const maxWorkspaces = searchParams.get('maxWs')
    ? Number(searchParams.get('maxWs'))
    : null;
  const personalWorkspaceActive = searchParams.get('pw');

  const registrationDateRange: [dayjs.Dayjs, dayjs.Dayjs] | null =
    regStart && regEnd ? [dayjs(regStart), dayjs(regEnd)] : null;
  const lastLoginDateRange: [dayjs.Dayjs, dayjs.Dayjs] | null =
    loginStart && loginEnd ? [dayjs(loginStart), dayjs(loginEnd)] : null;

  const setParam = (key: string, value: string | null) => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev);
      if (value !== null && value !== '') next.set(key, value);
      else next.delete(key);
      return next;
    });
  };

  const setMultiParam = (key: string, values: string[]) => {
    setSearchParams(prev => {
      const next = new URLSearchParams(prev);
      next.delete(key);
      values.forEach(v => next.append(key, v));
      return next;
    });
  };

  const { data, loading, error, refetch } = useTenantsQuery({
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
    variables: { retrieveWorkspaces: true },
  });

  const [deleteTenantMutation, { loading: deleteLoading }] =
    useDeleteTenantMutation({ onError: apolloErrorCatcher });

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

  const tableBodyRef = useRef<HTMLDivElement>(null);
  const [tableHeight, setTableHeight] = useState(400);
  useEffect(() => {
    const calculateHeight = () => {
      const el = tableBodyRef.current;
      if (!el) return;
      const top = el.getBoundingClientRect().top;
      setTableHeight(window.innerHeight - top - 170);
    };
    calculateHeight();
    window.addEventListener('resize', calculateHeight);
    return () => window.removeEventListener('resize', calculateHeight);
  }, []);

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
        if (
          !multiStringIncludes(searchText, tenant.name, tenant.surname, tenant.userid)
        )
          return false;

        if (registrationDateRange) {
          if (!tenant.creationDate) return false;
          const date = dayjs(tenant.creationDate);
          if (
            !date.isAfter(registrationDateRange[0].startOf('day')) ||
            !date.isBefore(registrationDateRange[1].endOf('day'))
          )
            return false;
        }

        if (lastLoginDateRange) {
          if (!tenant.lastLogin) return false;
          const date = dayjs(tenant.lastLogin);
          if (
            !date.isAfter(lastLoginDateRange[0].startOf('day')) ||
            !date.isBefore(lastLoginDateRange[1].endOf('day'))
          )
            return false;
        }

        if (labelKeyFilter || labelValueFilter) {
          if (!tenant.labels) return false;
          if (labelKeyFilter && labelValueFilter) {
            if (
              tenant.labels[labelKeyFilter] === undefined ||
              !tenant.labels[labelKeyFilter]
                .toLowerCase()
                .includes(labelValueFilter.toLowerCase())
            )
              return false;
          } else if (labelKeyFilter) {
            if (tenant.labels[labelKeyFilter] === undefined) return false;
          } else {
            if (
              !Object.values(tenant.labels).some(v =>
                v.toLowerCase().includes(labelValueFilter.toLowerCase()),
              )
            )
              return false;
          }
        }

        if (selectedOperators.length > 0) {
          const val = tenant.labels?.['crownlabsPolitoItOperatorSelector'];
          if (!val || !selectedOperators.includes(val)) return false;
        }

        const wsCount = tenant.workspaces?.length ?? 0;
        if (minWorkspaces !== null && wsCount < minWorkspaces) return false;
        if (maxWorkspaces !== null && wsCount > maxWorkspaces) return false;

        if (personalWorkspaceActive === 'yes' && !tenant.personalWorkspace)
          return false;
        if (personalWorkspaceActive === 'no' && tenant.personalWorkspace)
          return false;

        return true;
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

  const dateRangeDropdown =
    (
      value: [dayjs.Dayjs, dayjs.Dayjs] | null,
      startKey: string,
      endKey: string,
    ) =>
    ({ confirm }: FilterDropdownProps) =>
      (
        <div style={{ padding: 8, display: 'flex', flexDirection: 'column', gap: 8 }}>
          <DatePicker.RangePicker
            value={value}
            onChange={dates => {
              setSearchParams(prev => {
                const next = new URLSearchParams(prev);
                if (dates?.[0]) next.set(startKey, dates[0].toISOString());
                else next.delete(startKey);
                if (dates?.[1]) next.set(endKey, dates[1].toISOString());
                else next.delete(endKey);
                return next;
              });
            }}
          />
          <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
            <Button
              size="small"
              onClick={() => {
                setSearchParams(prev => {
                  const next = new URLSearchParams(prev);
                  next.delete(startKey);
                  next.delete(endKey);
                  return next;
                });
                confirm();
              }}
            >
              Reset
            </Button>
            <Button type="primary" size="small" onClick={() => confirm()}>
              OK
            </Button>
          </div>
        </div>
      );

  const workspacesDropdown = ({ confirm }: FilterDropdownProps) => (
    <div style={{ padding: 8, display: 'flex', flexDirection: 'column', gap: 8 }}>
      <Space.Compact>
        <InputNumber
          placeholder="Min"
          value={minWorkspaces}
          onChange={v => setParam('minWs', v !== null ? String(v) : null)}
          min={0}
          style={{ width: 100 }}
        />
        <InputNumber
          placeholder="Max"
          value={maxWorkspaces}
          onChange={v => setParam('maxWs', v !== null ? String(v) : null)}
          min={0}
          style={{ width: 100 }}
        />
      </Space.Compact>
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
        <Button
          size="small"
          onClick={() => {
            setSearchParams(prev => {
              const next = new URLSearchParams(prev);
              next.delete('minWs');
              next.delete('maxWs');
              return next;
            });
            confirm();
          }}
        >
          Reset
        </Button>
        <Button type="primary" size="small" onClick={() => confirm()}>
          OK
        </Button>
      </div>
    </div>
  );

  const personalWorkspaceDropdown = ({ confirm }: FilterDropdownProps) => (
    <div style={{ padding: 8, display: 'flex', flexDirection: 'column', gap: 8 }}>
      <Select
        allowClear
        placeholder="Any"
        style={{ width: 140 }}
        value={personalWorkspaceActive}
        onChange={v => setParam('pw', v ?? null)}
        options={[
          { label: 'Yes', value: 'yes' },
          { label: 'No', value: 'no' },
        ]}
      />
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8 }}>
        <Button
          size="small"
          onClick={() => {
            setParam('pw', null);
            confirm();
          }}
        >
          Reset
        </Button>
        <Button type="primary" size="small" onClick={() => confirm()}>
          OK
        </Button>
      </div>
    </div>
  );

  return (
    <Col span={24} lg={22} xxl={20} className="h-full">
      <Box
        header={{
          className: 'py-4 md:py-6 h-auto',
          center: (
            <div className="flex flex-col items-center gap-4 w-full px-2 sm:px-4">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Manage users</b>
              </p>

              <div
                className="grid grid-cols-1 sm:grid-cols-3 gap-x-4 gap-y-3 w-full p-3"
                style={{
                  background: token.colorFillAlter,
                  border: `1px solid ${token.colorBorderSecondary}`,
                  borderRadius: token.borderRadiusLG,
                }}
              >
                {/* Search + delete button — full width */}
                <div className="col-span-1 sm:col-span-3 flex flex-wrap items-center gap-2">
                  <Input.Search
                    placeholder="Search by name, surname, user ID"
                    className="flex-1 min-w-48"
                    value={searchText}
                    onSearch={v => setParam('q', v)}
                    onChange={e => setParam('q', e.target.value)}
                    enterButton
                    allowClear
                  />
                  {selectedRowKeys.length > 0 && (
                    <Popconfirm
                      title={`Delete ${selectedRowKeys.length} users?`}
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
                        Delete ({selectedRowKeys.length})
                      </Button>
                    </Popconfirm>
                  )}
                </div>

                <div>
                  <span
                    className="text-xs font-medium block mb-1"
                    style={{ color: token.colorTextSecondary }}
                  >
                    Operator
                  </span>
                  <Select
                    mode="multiple"
                    allowClear
                    placeholder="Any"
                    className="w-full"
                    value={selectedOperators}
                    onChange={vals => setMultiParam('op', vals)}
                    options={operatorSelectorValues.map(v => ({ label: v, value: v }))}
                  />
                </div>

                <div>
                  <span
                    className="text-xs font-medium block mb-1"
                    style={{ color: token.colorTextSecondary }}
                  >
                    Label Key
                  </span>
                  <Input
                    placeholder="e.g. crownlabs.polito.it/role"
                    value={labelKeyFilter}
                    onChange={e => setParam('labelKey', e.target.value)}
                    allowClear
                  />
                </div>

                <div>
                  <span
                    className="text-xs font-medium block mb-1"
                    style={{ color: token.colorTextSecondary }}
                  >
                    Label Value
                  </span>
                  <Input
                    placeholder="e.g. admin"
                    value={labelValueFilter}
                    onChange={e => setParam('labelVal', e.target.value)}
                    allowClear
                  />
                </div>
              </div>
            </div>
          ),
        }}
      >
        <Spin spinning={loading || error != null}>
          <div ref={tableBodyRef} style={{ height: '100%' }}>
            <Table
              pagination={{
                defaultPageSize: 10,
                pageSizeOptions: ['10', '20', '50', '100'],
                showSizeChanger: true,
              }}
              dataSource={filteredTenants}
              size="small"
              rowKey="userid"
              rowSelection={{
                selectedRowKeys,
                onChange: setSelectedRowKeys,
              }}
              scroll={{ y: tableHeight }}
              style={{ height: '100%' }}
            >
              <Table.Column
                title="User ID"
                dataIndex="userid"
                sorter={(a: Tenant, b: Tenant) => a.userid.localeCompare(b.userid)}
                key="userid"
                width={100}
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
                sorter={(a: Tenant, b: Tenant) => a.surname.localeCompare(b.surname)}
                key="surname"
                width={120}
              />
              <Table.Column
                responsive={['sm', 'md', 'lg']}
                title="Email"
                dataIndex="email"
                ellipsis
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
                  (a.creationDate ?? '').localeCompare(b.creationDate ?? '')
                }
                key="creationDate"
                width={160}
                filterDropdown={dateRangeDropdown(
                  registrationDateRange,
                  'regStart',
                  'regEnd',
                )}
                filteredValue={registrationDateRange ? ['active'] : []}
              />
              <Table.Column
                responsive={['lg']}
                title="Last login"
                dataIndex="lastLogin"
                render={(date: string) =>
                  date ? dayjs(date).format('YYYY-MM-DD HH:mm') : 'N/A'
                }
                sorter={(a: Tenant, b: Tenant) =>
                  (a.lastLogin ?? '').localeCompare(b.lastLogin ?? '')
                }
                key="lastLogin"
                width={170}
                filterDropdown={dateRangeDropdown(
                  lastLoginDateRange,
                  'loginStart',
                  'loginEnd',
                )}
                filteredValue={lastLoginDateRange ? ['active'] : []}
              />
              <Table.Column
                responsive={['sm', 'md', 'lg']}
                title="Workspaces"
                dataIndex="workspaces"
                render={(workspaces: Tenant['workspaces']) =>
                  workspaces ? workspaces.length : 0
                }
                sorter={(a: Tenant, b: Tenant) =>
                  (a.workspaces?.length ?? 0) - (b.workspaces?.length ?? 0)
                }
                key="workspaces"
                width={130}
                filterDropdown={workspacesDropdown}
                filteredValue={
                  minWorkspaces !== null || maxWorkspaces !== null ? ['active'] : []
                }
              />
              <Table.Column
                responsive={['sm', 'md', 'lg']}
                title="P. Workspace"
                dataIndex="personalWorkspace"
                render={(active: boolean) => (active ? 'Yes' : 'No')}
                sorter={(a: Tenant, b: Tenant) =>
                  a.personalWorkspace === b.personalWorkspace
                    ? 0
                    : a.personalWorkspace
                    ? -1
                    : 1
                }
                key="personalWorkspace"
                width={160}
                filterDropdown={personalWorkspaceDropdown}
                filteredValue={personalWorkspaceActive ? ['active'] : []}
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
                        onClick={() =>
                          window.open('/tenants/' + tenant.userid, '_blank')
                        }
                      />
                    </Tooltip>
                    <Tooltip title="Delete tenant">
                      <Popconfirm
                        title="Delete this user?"
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
          </div>
        </Spin>
      </Box>
    </Col>
  );
}
