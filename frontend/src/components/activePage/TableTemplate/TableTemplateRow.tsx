import { FC } from 'react';
import { Space, Menu, Dropdown, Typography, Popconfirm, Tooltip } from 'antd';
import Button from 'antd-button-color';
import {
  DeleteOutlined,
  MoreOutlined,
  DesktopOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import Badge from '../../common/Badge';
import { ReactComponent as SvgInfinite } from '../../../assets/infinite.svg';
import { Template } from '../../../utils';
import { DropDownAction } from '../../../utilsLogic';

const { Text } = Typography;
export interface ITableTemplateRowProps {
  template: Template;
  nActive: number;
  destroyAll: () => void;
  expandRow: (rowId: string) => void;
}

const TableTemplateRow: FC<ITableTemplateRowProps> = ({ ...props }) => {
  const { template, nActive, destroyAll, expandRow } = props;

  const { id, name, persistent, gui } = template;

  const dropdownHandler = (key: DropDownAction) => {
    switch (key) {
      case DropDownAction.destroy_all:
        // TODO: Popconfirm not work maybe we should use a modal for the confirmation
        break;
      default:
        break;
    }
  };
  return (
    <div
      className="w-full flex justify-between pr-2"
      onClick={e => {
        expandRow(id);
      }}
    >
      <Space size={'middle'}>
        {gui ? (
          <DesktopOutlined
            className={'primary-color-fg'}
            style={{ fontSize: '24px' }}
          />
        ) : (
          <CodeOutlined
            className={'primary-color-fg'}
            style={{ fontSize: '24px' }}
          />
        )}
        <Badge size="small" value={nActive} className="mx-0" />
        <Text className="font-bold w-28 xs:w-48 sm:w-max" ellipsis>
          {name}
        </Text>
        {persistent && (
          <Tooltip
            title={
              <>
                <div className="text-center">
                  These Instances can be stopped and restarted without being
                  deleted.
                </div>
                <div className="text-center">
                  Your files won't be deleted in case of an internal misservice
                  of CrownLabs.
                </div>
              </>
            }
          >
            <div className="success-color-fg flex items-center">
              <SvgInfinite width="22px" />
            </div>
          </Tooltip>
        )}
      </Space>

      <Popconfirm
        placement="left"
        title="You are about to delete all VMs in this. Are you sure?"
        okText="Yes"
        cancelText="No"
        onConfirm={e => {
          e?.stopPropagation();
          destroyAll();
        }}
        onCancel={e => e?.stopPropagation()}
      >
        <Button
          type="danger"
          shape="round"
          size="middle"
          icon={<DeleteOutlined />}
          className="hidden lg:inline-block border-0"
          onClick={e => e?.stopPropagation()}
        >
          Destroy All
        </Button>
      </Popconfirm>
      <Dropdown
        trigger={['click']}
        overlay={
          <Menu onClick={({ key }) => dropdownHandler(key as DropDownAction)}>
            <Menu.Item
              key="destroy_all"
              icon={<DeleteOutlined className="text-lg" />}
              danger
            >
              Destroy All
            </Menu.Item>
          </Menu>
        }
      >
        <Button
          className="lg:hidden flex justify-center"
          type="default"
          with="link"
          shape="circle"
          size="middle"
          icon={
            <MoreOutlined
              className={'flex items-center'}
              style={{ fontSize: '20px' }}
            />
          }
        />
      </Dropdown>
    </div>
  );
};

export default TableTemplateRow;
