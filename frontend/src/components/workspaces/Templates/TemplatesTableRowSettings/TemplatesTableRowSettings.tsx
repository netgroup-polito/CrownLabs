import { Menu, Dropdown, Tooltip } from 'antd';
import { Button } from 'antd';
import {
  PlayCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  EllipsisOutlined,
} from '@ant-design/icons';
import type { FetchResult } from '@apollo/client';
import type { CreateInstanceMutation } from '../../../../generated-types';

export interface ITemplatesTableRowSettingsProps {
  id: string;
  createInstance: (
    id: string,
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, any>,
      Record<string, any>
    >
  >;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}
const TemplatesTableRowSettings = ({ ...props }) => {
  const { id, createInstance, editTemplate, deleteTemplate } = props;

  return (
    <Dropdown
      overlay={
        <Menu>
          <Menu.Item
            onClick={() => createInstance(id)}
            className="xs:hidden block"
            key="1"
            icon={<PlayCircleOutlined />}
          >
            Create
          </Menu.Item>
          <Menu.Item
            disabled
            key="2"
            icon={<EditOutlined />}
            onClick={() => editTemplate(id)}
          >
            <Tooltip title="Coming soon" placement="left">
              Edit
            </Tooltip>
          </Menu.Item>
          <Menu.Item
            danger
            key="3"
            icon={<DeleteOutlined />}
            onClick={deleteTemplate}
          >
            Delete
          </Menu.Item>
        </Menu>
      }
      placement="bottom"
      trigger={['click']}
    >
      <Button
        type="text"
        size="middle"
        shape="circle"
        icon={
          <EllipsisOutlined
            className="flex justify-center"
            style={{ fontSize: '22px' }}
          />
        }
      />
    </Dropdown>
  );
};

export default TemplatesTableRowSettings;
