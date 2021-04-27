import { FC } from 'react';
import { Space, Typography } from 'antd';
import Button from 'antd-button-color';
import {
  CheckCircleOutlined,
  PoweroffOutlined,
  WarningOutlined,
} from '@ant-design/icons';

const { Text } = Typography;

export interface IInstancesTableRowProps {
  idInstance: number;
  idTemplate: string;
  name: string;
  ip: string;
  status: boolean;
  destroyInstance: (idInstance: number, idTemplate: string) => void;
}

const InstancesTableRow: FC<IInstancesTableRowProps> = ({ ...props }) => {
  const { idInstance, idTemplate, name, ip, status, destroyInstance } = props;

  return (
    <>
      <div className="w-full flex justify-between py-0 pl-4">
        <Space size={'middle'}>
          {status ? (
            <CheckCircleOutlined
              style={{ color: '#28a745', fontSize: '24px' }}
            />
          ) : (
            <WarningOutlined style={{ color: '#eca52b', fontSize: '24px' }} />
          )}

          {`${name}`}
        </Space>
        <Space>
          <Text strong type="secondary">
            {ip}
          </Text>
        </Space>
        <Space size={'small'}>
          <Button
            with="link"
            type="danger"
            shape="round"
            size="large"
            icon={<PoweroffOutlined />}
            onClick={() => destroyInstance(idInstance, idTemplate)}
          >
            Destroy
          </Button>
          <Button type="success" shape="round" size={'large'}>
            Connect
          </Button>
        </Space>
      </div>
    </>
  );
};

export default InstancesTableRow;
