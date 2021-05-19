import { FC } from 'react';
import { Badge as BadgeAnt, Col, Row, Menu, Dropdown, Typography } from 'antd';
import Button from 'antd-button-color';
import { DeleteOutlined, MoreOutlined } from '@ant-design/icons';
import Badge from '../../common/Badge';

export interface IRowHeadingProps {
  text: string;
  nActive: number;
  newTempl: boolean;
  destroyAll: () => void;
}

const RowHeading: FC<IRowHeadingProps> = ({ ...props }) => {
  const { text, nActive, newTempl, destroyAll } = props;
  const { Text } = Typography;
  return (
    <Row className="items-center">
      <Col className="flex items-center flex-grow">
        {newTempl ? (
          <BadgeAnt dot offset={[-8, 2]} className="hidden lg:inline-block">
            <Badge size="small" value={nActive} />
          </BadgeAnt>
        ) : (
          <Badge size="small" value={nActive} />
        )}
        <Text className="font-bold ml-4 w-48 sm:w-max" ellipsis={true}>
          {text}
        </Text>
      </Col>
      <Col>
        <Button
          type="danger"
          shape="round"
          size="middle"
          icon={<DeleteOutlined />}
          className="hidden lg:inline-block border-0"
          onClick={e => {
            e.stopPropagation();
            destroyAll();
          }}
        >
          Destory All
        </Button>
        <Dropdown
          overlay={
            <Menu
              onClick={e => {
                e.domEvent.preventDefault();
                destroyAll();
              }}
              className="p-0 rounded-sm"
            >
              <Menu.Item
                key={1}
                icon={<DeleteOutlined className="text-lg" />}
                className="rounded-sm"
                danger
              >
                Destory All
              </Menu.Item>
            </Menu>
          }
          trigger={['click']}
        >
          <MoreOutlined
            className="lg:hidden"
            onClick={e => e.stopPropagation()}
          />
        </Dropdown>
      </Col>
    </Row>
  );
};

export default RowHeading;
