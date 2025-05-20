import { CaretDownOutlined, LogoutOutlined } from '@ant-design/icons';
import { Dropdown, Menu, Space } from 'antd';
import { Button } from 'antd';
import { type FC, useContext, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { TenantContext } from '../../../../contexts/TenantContext';
import { generateAvatarUrl } from '../../../../utils';
import { type RouteData } from '../Navbar';

export interface INavbarMenuProps {
  routes: Array<RouteData>;
  logoutHandler: () => void;
}

const NavbarMenu: FC<INavbarMenuProps> = ({ ...props }) => {
  const { routes, logoutHandler } = props;
  const { data, displayName } = useContext(TenantContext);
  const tenantId = data?.tenant?.metadata?.name!;
  const currentPath = useLocation().pathname;

  const [visible, setVisible] = useState(false);

  const handleMenuClick = (e: { key: string }) => {
    if (e.key !== 'welcome') setVisible(false);
  };

  const handleVisibleChange = (flag: boolean) => {
    setVisible(flag);
  };

  const userIcon = (
    <img
      src={generateAvatarUrl('bottts', tenantId ?? '')}
      className="anticon"
      width="35"
      height="35"
    />
  );

  return (
    <div className="flex justify-center items-center">
      <Dropdown
        overlayClassName="pt-1 pr-2 2xl:pr-0"
        open={visible}
        onOpenChange={handleVisibleChange}
        placement="bottom"
        trigger={['click']}
        overlay={
          <Menu onClick={handleMenuClick} selectedKeys={[currentPath]}>
            <Menu.Item
              key="welcome"
              className="pointer-events-none text-center"
            >
              Logged in as <b>{tenantId}</b>
            </Menu.Item>
            <Menu.Divider />
            {routes.map(r => {
              const isExtLink = r.path.indexOf('http') === 0;
              return (
                <Menu.Item
                  key={r.path}
                  className="text-center "
                  onClick={() => isExtLink && window.open(r.path, '_blank')}
                >
                  <Link
                    target={isExtLink ? '_blank' : ''}
                    key={r.path}
                    to={{ pathname: isExtLink ? '' : r.path }}
                    rel={isExtLink ? 'noopener noreferrer' : ''}
                  >
                    <Space size="small">
                      {r.navbarMenuIcon}
                      {r.name}
                    </Space>
                  </Link>
                </Menu.Item>
              );
            })}
            <Menu.Divider />
            <Menu.Item
              onClick={logoutHandler}
              className="text-center bg-opacity-60 hover:bg-opacity-100 hover:text-white bg-red-700"
            >
              <Space size="small">
                <LogoutOutlined />
                Logout
              </Space>
            </Menu.Item>
          </Menu>
        }
      >
        <Button
          className="flex justify-center items-center px-2 ml-1 "
          type={routes.find(r => r.path === currentPath) ? 'primary' : 'text'}
          shape="round"
          size="large"
          icon={userIcon}
          classNames={{ icon: 'w-8 mt-3' }}
        >
          <div className="2xl:flex hidden items-center ml-1">{displayName}</div>
          <CaretDownOutlined
            className="flex items-center ml-2"
            style={{ fontSize: '15px' }}
          />
        </Button>
      </Dropdown>
    </div>
  );
};

export default NavbarMenu;
