import { useContext, useMemo, useState } from 'react';
import { Table, Input, Spin, Col, Tooltip, Space, DatePicker } from 'antd';
import dayjs from 'dayjs';
import { ErrorContext } from '../../../errorHandling/ErrorContext';
import { useTenantsQuery } from '../../../generated-types';
import { makeTenantsList } from '../../../utilsLogic';
import { multiStringIncludes, type Tenant } from '../../../utils';
import Box from '../../common/Box';
import { EditOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';

export default function TenantListPage() {
  const { apolloErrorCatcher } = useContext(ErrorContext);
  const navigate = useNavigate();

  const [searchText, setSearchText] = useState('');
  const [registrationDateRange, setRegistrationDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null);
  const [lastLoginDateRange, setLastLoginDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null);
  const [labelKeyFilter, setLabelKeyFilter] = useState('');
  const [labelValueFilter, setLabelValueFilter] = useState('');

  const { data, loading, error } = useTenantsQuery({
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
  });

  const tenants = useMemo(() => makeTenantsList(data), [data]);
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
        if (registrationDateRange && registrationDateRange[0] && registrationDateRange[1]) {
          if (!tenant.creationDate) matchesReg = false;
          else {
            const date = dayjs(tenant.creationDate);
            matchesReg = date.isAfter(registrationDateRange[0].startOf('day')) && date.isBefore(registrationDateRange[1].endOf('day'));
          }
        }

        let matchesLogin = true;
        if (lastLoginDateRange && lastLoginDateRange[0] && lastLoginDateRange[1]) {
          if (!tenant.lastLogin) matchesLogin = false;
          else {
            const date = dayjs(tenant.lastLogin);
            matchesLogin = date.isAfter(lastLoginDateRange[0].startOf('day')) && date.isBefore(lastLoginDateRange[1].endOf('day'));
          }
        }

        let matchesLabel = true;
        if (labelKeyFilter || labelValueFilter) {
          if (!tenant.labels) {
            matchesLabel = false;
          } else {
            if (labelKeyFilter && labelValueFilter) {
              matchesLabel = tenant.labels[labelKeyFilter] !== undefined && 
                             tenant.labels[labelKeyFilter].toLowerCase().includes(labelValueFilter.toLowerCase());
            } else if (labelKeyFilter) {
              matchesLabel = tenant.labels[labelKeyFilter] !== undefined;
            } else if (labelValueFilter) {
              matchesLabel = Object.values(tenant.labels).some(v => v.toLowerCase().includes(labelValueFilter.toLowerCase()));
            }
          }
        }

        return searchMatches && matchesReg && matchesLogin && matchesLabel;
      }),
    [tenants, searchText, registrationDateRange, lastLoginDateRange, labelKeyFilter, labelValueFilter],
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
                  <DatePicker.RangePicker 
                    className="w-full sm:w-auto"
                    placeholder={['Reg. Start', 'Reg. End']}
                    onChange={(dates) => setRegistrationDateRange(dates as any)}
                    allowClear
                  />
                  <DatePicker.RangePicker 
                    className="w-full sm:w-auto"
                    placeholder={['Login Start', 'Login End']}
                    onChange={(dates) => setLastLoginDateRange(dates as any)}
                    allowClear
                  />
                  <Space.Compact className="w-full sm:w-auto flex">
                    <Input 
                      placeholder="Label Key" 
                      onChange={(e) => setLabelKeyFilter(e.target.value)}
                      allowClear
                      style={{ width: '40%' }}
                    />
                    <Input 
                      placeholder="Label Value" 
                      onChange={(e) => setLabelValueFilter(e.target.value)}
                      allowClear
                      style={{ width: '60%' }}
                    />
                  </Space.Compact>
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
              title="Registration Date"
              dataIndex="creationDate"
              render={(date: string) => date ? dayjs(date).format('YYYY-MM-DD') : 'N/A'}
              sorter={(a: Tenant, b: Tenant) => 
                (a.creationDate || '').localeCompare(b.creationDate || '')
              }
              key="creationDate"
              width={140}
            />
            <Table.Column
              responsive={['lg']}
              title="Last Login"
              dataIndex="lastLogin"
              render={(date: string) => date ? dayjs(date).format('YYYY-MM-DD HH:mm') : 'N/A'}
              sorter={(a: Tenant, b: Tenant) => 
                (a.lastLogin || '').localeCompare(b.lastLogin || '')
              }
              key="lastLogin"
              width={150}
            />
            <Table.Column
              title="Actions"
              key="actions"
              width={60}
              render={(tenant: Tenant) => (
                <Tooltip title="Edit tenant">
                  <EditOutlined
                    className="mr-2"
                    onClick={() => navigate('/tenants/' + tenant.userid)}
                  />
                </Tooltip>
              )}
            />
          </Table>
        </Spin>
      </Box>
    </Col>
  );
}
