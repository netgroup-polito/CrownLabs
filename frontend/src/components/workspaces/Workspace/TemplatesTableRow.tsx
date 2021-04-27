import { FC, useContext, useState } from 'react';
import { Space, Spin, notification, Dropdown, Menu } from 'antd';
import Button from 'antd-button-color';
import {
  DesktopOutlined,
  CodeOutlined,
  DeleteOutlined,
  EditOutlined,
  SettingOutlined,
  LoadingOutlined,
} from '@ant-design/icons';
import Badge from '../common/Badge/Badge';
import { ModalCreateInstance } from '../Modal';
import { Auth } from '../auth';

export interface ITemplatesTableRowProps {
  id: string;
  name: string;
  gui: boolean;
  activeInstances: number;
  createInstance: (id: string) => void;
  editTemplate: (id: string) => void;
  deleteTemplate: (id: string) => void;
}

const TemplatesTableRow: FC<ITemplatesTableRowProps> = ({ ...props }) => {
  const {
    id,
    name,
    gui,
    activeInstances,
    createInstance,
    editTemplate,
    deleteTemplate,
  } = props;
  const auth = useContext(Auth);

  const [showModal, setShowModal] = useState(false);
  const [loadingVm, setLoadingVm] = useState(false);

  const createVm = () => {
    setLoadingVm(oldloading => {
      setShowModal(true);
      if (oldloading === false) {
        setTimeout(() => {
          setLoadingVm(false);
          const key = `open${Date.now()}`;
          const notificationbtn = (
            <Button
              onClick={() => notification.close(key)}
              type="success"
              size="small"
            >
              Go to active
            </Button>
          );
          notification['success']({
            message: name,
            description: 'An instance of this vm have been created',
            btn: notificationbtn,
            key,
          });
          createInstance(id);
        }, 500);
        return true;
      } else {
        return false;
      }
    });
  };

  const settings = (id: string) => {
    return (
      <Menu>
        <Menu.Item
          key="1"
          icon={<EditOutlined />}
          onClick={() => editTemplate(id)}
        >
          Edit
        </Menu.Item>
        <Menu.Item
          danger
          key="2"
          icon={<DeleteOutlined />}
          onClick={() => deleteTemplate(id)}
        >
          Delete
        </Menu.Item>
      </Menu>
    );
  };

  return (
    <>
      <div className="w-full flex justify-between py-0">
        <Space size={'middle'}>
          {gui ? (
            <DesktopOutlined style={{ color: '#1c7afd', fontSize: '24px' }} />
          ) : (
            <CodeOutlined style={{ color: '#1c7afd', fontSize: '24px' }} />
          )}
          {name}
        </Space>
        <Space size={'small'}>
          <Badge value={activeInstances} />
          <Button with="link" type={'warning'} size={'large'}>
            Info
          </Button>
          {auth ? (
            <Dropdown
              overlay={settings(id)}
              placement="bottomCenter"
              trigger={['click']}
            >
              <Button
                with="link"
                type="default"
                size="large"
                icon={<SettingOutlined />}
                disabled={false}
              />
            </Dropdown>
          ) : (
            ''
          )}
          <Button
            type="primary"
            shape="round"
            size={'large'}
            onClick={createVm}
            disabled={loadingVm}
          >
            {!loadingVm ? 'Create' : 'Wait'}
            {loadingVm ? (
              <Spin
                className="ml-3"
                indicator={
                  <LoadingOutlined
                    style={{ fontSize: 20, color: 'white' }}
                    spin
                  />
                }
              />
            ) : null}
          </Button>
        </Space>
        <ModalCreateInstance
          headTitle={name}
          showmodal={showModal}
          setshowmodal={setShowModal}
          loadingVm={loadingVm}
        />
      </div>
    </>
  );
};

export default TemplatesTableRow;
