import { useContext, useMemo, useState } from 'react';
import { Table, Input, Spin, Col, Tooltip } from 'antd';
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

  const { data, loading, error } = useTenantsQuery({
    onError: apolloErrorCatcher,
    notifyOnNetworkStatusChange: true,
  });

  const tenants = useMemo(() => makeTenantsList(data), [data]);
  const filteredTenants = useMemo(
    () =>
      tenants.filter(tenant =>
        multiStringIncludes(
          searchText,
          tenant.name,
          tenant.surname,
          tenant.userid,
        ),
      ),
    [tenants, searchText],
  );

  const handleSearch = (value: string) => {
    setSearchText(value.toLowerCase());
  };

  return (
    <Col span={24} lg={22} xxl={20} className="h-full">
      <Box
        header={{
          size: 'large',
          center: (
            <div className="h-full flex flex-col justify-center items-center gap-4">
              <p className="md:text-2xl text-lg text-center mb-0">
                <b>Manage tenant</b>
              </p>

              <Input.Search
                placeholder="Search users"
                style={{ width: 300 }}
                onSearch={handleSearch}
                enterButton
                allowClear={true}
              />
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
