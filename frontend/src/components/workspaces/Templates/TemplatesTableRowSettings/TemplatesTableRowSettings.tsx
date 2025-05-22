import { Dropdown, Tooltip } from 'antd';
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
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}
const TemplatesTableRowSettings = ({ ...props }) => {
  const { id, createInstance, editTemplate, deleteTemplate } = props;

  return (
    <Dropdown
      menu={{
        items: [
          {
            type: 'item',
            key: 1,
            label: 'Create',
            icon: <PlayCircleOutlined />,
            className: 'xs:hidden block',
            onClick: () => createInstance(id),
          },
          {
            type: 'item',
            key: 2,
            label: (
              <Tooltip title="Coming soon" placement="left">
                Edit
              </Tooltip>
            ),
            icon: <EditOutlined />,
            disabled: true,
            onClick: () => editTemplate(id),
          },
          {
            type: 'item',
            key: 3,
            label: 'Delete',
            icon: <DeleteOutlined />,
            danger: true,
            onClick: deleteTemplate,
          },
        ],
      }}
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
