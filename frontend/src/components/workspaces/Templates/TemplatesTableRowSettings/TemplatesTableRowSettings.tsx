import { Menu, Dropdown, Tooltip } from 'antd';
import Button from 'antd-button-color';
import {
  PlayCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  EllipsisOutlined,
} from '@ant-design/icons';

export interface ITemplatesTableRowSettingsProps {
  id: string;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}
const TemplatesTableRowSettings = ({ ...props }) => {
  const { id, editTemplate, deleteTemplate } = props;

  return (
    <Dropdown
      overlay={
        <Menu>
          <Menu.Item
            className="xs:hidden block"
            key="1"
            icon={<PlayCircleOutlined />}
          >
            Create
          </Menu.Item>
          <Menu.Item
            disabled
            key="1"
            icon={<EditOutlined />}
            onClick={() => editTemplate(id)}
          >
            <Tooltip title="Coming soon" placement="left">
              Edit
            </Tooltip>
          </Menu.Item>
          <Menu.Item
            danger
            key="2"
            icon={<DeleteOutlined />}
            onClick={deleteTemplate}
          >
            Delete
          </Menu.Item>
        </Menu>
      }
      placement="bottomCenter"
      trigger={['click']}
    >
      <Button
        with="link"
        type="text"
        size="large"
        icon={<EllipsisOutlined style={{ fontSize: '22px' }} />}
      />
    </Dropdown>
  );
};

export default TemplatesTableRowSettings;
