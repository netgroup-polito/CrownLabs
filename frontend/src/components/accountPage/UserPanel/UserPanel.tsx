import { FC } from 'react';
import { Row, Col, Avatar, Tabs, Table } from 'antd';
import { UserOutlined } from '@ant-design/icons';
import UserInfo from '../UserInfo/UserInfo';

const { TabPane } = Tabs;
const { Column } = Table;
export interface IUserPanelProps {
  firstName: string;
  lastName: string;
  username: string;
  email: string;
  avatar?: string;
  sshKeys?: { name: string; key: string }[];
}

const UserPanel: FC<IUserPanelProps> = props => {
  const { avatar, sshKeys, ...otherInfo } = props;

  return (
    <Row className="p-4">
      <Col xs={24} sm={8} className="text-center">
        <Avatar size="large" icon={avatar ?? <UserOutlined />} />
        <p>
          {otherInfo.firstName} {otherInfo.lastName}
          <br />
          <strong>{otherInfo.username}</strong>
        </p>
      </Col>
      <Col xs={24} sm={16} className="px-4 ">
        <Tabs>
          <TabPane tab="Info" key="1">
            <UserInfo {...otherInfo} />
          </TabPane>
          <TabPane tab="SSH Keys" key="2">
            <Table dataSource={sshKeys}>
              <Column title="Name" dataIndex="name" key="name" />
              <Column title="Key" dataIndex="key" key="key" />
            </Table>
          </TabPane>
        </Tabs>
      </Col>
    </Row>
  );
};

export default UserPanel;
