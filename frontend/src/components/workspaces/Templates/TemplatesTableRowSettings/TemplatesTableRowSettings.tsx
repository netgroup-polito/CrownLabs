import { Dropdown } from 'antd';
import { Button } from 'antd';
import {
  PlayCircleOutlined,
  EditOutlined,
  DeleteOutlined,
  EllipsisOutlined,
} from '@ant-design/icons';
import type { FetchResult } from '@apollo/client';
import type { CreateInstanceMutation } from '../../../../generated-types';
import type { Template } from '../../ModalCreateTemplate/ModalCreateTemplate';

export interface ITemplatesTableRowSettingsProps {
  id: string;
  template: Template; // <-- Add this
  createInstance: (
    id: string,
  ) => Promise<
    FetchResult<
      CreateInstanceMutation,
      Record<string, unknown>,
      Record<string, unknown>
    >
  >;
  editTemplate: (template: Template) => void; // <-- Change signature
  deleteTemplate: (id: string) => void;
}
const TemplatesTableRowSettings = ({ ...props }) => {
  const { id, template, createInstance, editTemplate, deleteTemplate } = props;

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
            label: 'Edit',
            icon: <EditOutlined />,
            onClick: () => {
              // call provided handler if any
              try {
                editTemplate?.(template);
              } catch (_e) {
                /* ignore handler errors */
              }
              // also emit a global event so parents that didn't wire editTemplate can react
              try {
                window.dispatchEvent(
                  new CustomEvent('openTemplateModal', { detail: template }),
                );
              } catch (_e) {
                /* ignore dispatch errors */
              }
            },
          },
          {
            type: 'item',
            key: 3,
            label: 'Delete',
            icon: <DeleteOutlined />,
            danger: true,
            onClick: () => deleteTemplate(id),
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
