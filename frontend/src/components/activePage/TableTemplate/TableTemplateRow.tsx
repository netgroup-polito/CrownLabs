import { FC, useState } from 'react';
import { Space, Menu, Dropdown, Typography, Tooltip } from 'antd';
import Button from 'antd-button-color';
import {
  DeleteOutlined,
  MoreOutlined,
  DesktopOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import Badge from '../../common/Badge';
import { ModalAlert } from '../../common/ModalAlert';
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

  const [showAlert, setShowAlert] = useState(false);

  const dropdownHandler = (key: DropDownAction) => {
    switch (key) {
      case DropDownAction.destroy_all:
        setShowAlert(true);
        break;
      default:
        break;
    }
  };
  return (
    <>
      <div
        className="w-full flex justify-between pr-2 cursor-pointer"
        onClick={e => {
          expandRow(id);
        }}
      >
        <Space size="middle">
          {gui ? (
            <DesktopOutlined
              className="primary-color-fg"
              style={{ fontSize: '24px' }}
            />
          ) : (
            <CodeOutlined
              className="primary-color-fg"
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
                    Your files won't be deleted in case of an internal
                    disservice of CrownLabs.
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
        <Button
          type="danger"
          shape="round"
          size="middle"
          icon={<DeleteOutlined />}
          className="hidden lg:inline-block border-0"
          onClick={e => {
            e.stopPropagation();
            setShowAlert(true);
          }}
        >
          Destroy All
        </Button>
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
            onClick={e => e.stopPropagation()}
            icon={
              <MoreOutlined
                className="flex items-center"
                style={{ fontSize: '20px' }}
              />
            }
          />
        </Dropdown>
      </div>
      <ModalAlert
        headTitle="Destroy All"
        show={showAlert}
        alertMessage="ATTENTION"
        alertDescription={`Are you sure do you want to destroy all the instances in ${name}. This operation is dangerous and irreversible!`}
        alertType="error"
        buttons={[
          <Button
            type="danger"
            shape="round"
            size="middle"
            icon={<DeleteOutlined />}
            className="border-0"
            onClick={() => destroyAll()}
          >
            Destroy All
          </Button>,
        ]}
        setShow={setShowAlert}
      />
    </>
  );
};

export default TableTemplateRow;
