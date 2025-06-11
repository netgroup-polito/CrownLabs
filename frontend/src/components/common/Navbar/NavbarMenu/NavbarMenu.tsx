import { CaretDownOutlined, LogoutOutlined } from '@ant-design/icons';
import { Dropdown } from 'antd';
import { Button } from 'antd';
import { type FC, useContext, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { TenantContext } from '../../../../contexts/TenantContext';
import { generateAvatarUrl } from '../../../../utils';
import { type RouteData } from '../Navbar';
import type { MenuItemType } from 'antd/lib/menu/interface';
import { AuthContext } from '../../../../contexts/AuthContext';

export interface INavbarMenuProps {
  routes: Array<RouteData>;
}

const NavbarMenu: FC<INavbarMenuProps> = ({ ...props }) => {
  const { routes } = props;
  const { data, displayName } = useContext(TenantContext);
  const { logout } = useContext(AuthContext);
  const tenantId = data?.tenant?.metadata?.name;
  const currentPath = useLocation().pathname;

  const [visible, setVisible] = useState(false);

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
    <Dropdown
      overlayClassName="pt-1 pr-2 2xl:pr-0"
      open={visible}
      onOpenChange={handleVisibleChange}
      placement="bottom"
      trigger={['click']}
      menu={{
        items: [
          {
            type: 'item',
            key: 'welcome',
            className: 'pointer-events-none text-center',
            label: (
              <>
                Logged in as <b>{tenantId}</b>
              </>
            ),
          },
          {
            type: 'divider',
          },

          ...routes.map(r => {
            const isExtLink = r.path.startsWith('http');
            return {
              type: 'item',
              key: r.path,
              title: 'pef',
              onClick: () => isExtLink && window.open(r.path, '_blank'),
              icon: r.navbarMenuIcon,
              className: currentPath === r.path ? 'primary-color-bg' : '',
              label: (
                <Link
                  target={isExtLink ? '_blank' : ''}
                  to={{ pathname: isExtLink ? '' : r.path }}
                  rel={isExtLink ? 'noopener noreferrer' : ''}
                >
                  {r.name}
                </Link>
              ),
            } as MenuItemType;
          }),
          {
            type: 'divider',
          },
          {
            type: 'item',
            key: 'logout',
            label: 'Logout',
            danger: true,
            icon: <LogoutOutlined />,
            onClick: logout,
          },
        ],
      }}
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
  );
};

export default NavbarMenu;
