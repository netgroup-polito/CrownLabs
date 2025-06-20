import { MenuOutlined } from '@ant-design/icons';
import { Divider, Drawer, Layout, Typography } from 'antd';
import { Button } from 'antd';
import { type FC, useContext, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { TenantContext } from '../../../contexts/TenantContext';
import {
  LinkPosition,
  type RouteData,
  type RouteDescriptor,
} from '../../../utils';
import ThemeSwitcher from '../../misc/ThemeSwitcher';
import Logo from '../Logo';
import { LogoutButton } from '../LogoutButton';
import './Navbar.less';
import NavbarMenu from './NavbarMenu';

const Header = Layout.Header;
const { Title } = Typography;

export interface INavbarProps {
  routes: Array<RouteDescriptor>;
  transparent?: boolean;
}

const Navbar: FC<INavbarProps> = ({ ...props }) => {
  const { routes, transparent } = props;
  const routesData = routes.map(r => r.route);
  const {
    data,
    loading: tenantLoading,
    error: tenantError,
  } = useContext(TenantContext);
  const [show, setShow] = useState(false);

  const currentPath = useLocation().pathname;

  const currentName = routesData.find(r => r.path === currentPath)?.name || '';

  const buttons = routes.map((b, i) => {
    const routeData = b.route;
    const isExtLink = routeData.path.indexOf('http') === 0;
    return {
      linkPosition: b.linkPosition,
      content: (
        <Link
          key={i}
          to={{ pathname: isExtLink ? '' : routeData.path }}
          rel={isExtLink ? 'noopener noreferrer' : ''}
        >
          <Button
            onClick={() =>
              isExtLink ? window.open(routeData.path, '_blank') : setShow(false)
            }
            ghost={currentPath !== routeData.path}
            className={
              'w-full flex justify-center my-3 ' +
              (routes.length <= 4
                ? 'lg:mx-4 md:mx-2 md:w-28 lg:w-36 xl:w-52 2xl:w-72 '
                : 'lg:mx-2 lg:w-28 xl:w-32 2xl:w-48') +
              (currentPath !== routeData.path ? ' navbar-button ' : '')
            }
            size="large"
            type={currentPath !== routeData.path ? 'default' : 'primary'}
            shape="round"
          >
            {routeData.name}
          </Button>
        </Link>
      ),
    };
  });

  return (
    <>
      <Header
        className={
          'flex h-auto pr-6 pl-8 justify-between ' +
          (transparent ? 'navbar-bg-transparent' : 'navbar-bg shadow-lg')
        }
      >
        <div className="flex flex-none items-center w-24 ">
          <div className="flex h-full items-center">
            <Logo widthPx={55} />
          </div>
          <h2
            className={
              'flex whitespace-nowrap py-0 my-0 ml-4 navbar-title ' +
              (routes.length > 4 ? 'lg:hidden' : 'md:hidden')
            }
          >
            {currentName}
          </h2>
        </div>
        <div
          className={
            'hidden justify-around ' +
            (routes.length > 4 ? 'lg:flex' : 'md:flex')
          }
        >
          {buttons
            .filter(b => b.linkPosition === LinkPosition.NavbarButton)
            .map(b => b.content)}
        </div>
        <div
          className={
            'w-full hidden sm:flex justify-end ' +
            (routes.length > 4 ? 'lg:hidden' : 'md:hidden')
          }
        >
          {buttons
            .map(b => b.content)
            .filter((x, i) => (i < 2 ? x : null))
            .map((b, i) => (
              <div key={i} className="w-28  mr-3">
                {b}
              </div>
            ))}
        </div>
        <div
          className={
            'flex items-center justify-end w-auto ' +
            (routes.length > 4
              ? 'lg:flex-none lg:w-24'
              : 'md:flex:none md:w-24')
          }
        >
          <div
            className={
              'hidden flex items-center justify-end ' +
              (routes.length > 4 ? 'lg:flex' : 'md:flex')
            }
          >
            <ThemeSwitcher />

            {!tenantLoading && !tenantError && (
              <>
                <Divider className="ml-4 mr-0" type="vertical" />
                <NavbarMenu
                  routes={routes
                    .filter(r => r.linkPosition === LinkPosition.MenuButton)
                    .map(r => r.route)}
                />
              </>
            )}
          </div>
          <Button
            className={
              'flex items-center ' +
              (routes.length > 4 ? 'lg:hidden' : 'md:hidden')
            }
            shape="round"
            size="large"
            type="primary"
            onClick={() => setShow(true)}
            icon={<MenuOutlined />}
          />
        </div>
      </Header>
      <Drawer
        className={
          'cl-navbar block ' + (routes.length > 4 ? 'lg:hidden' : 'md:hidden')
        }
        styles={{
          body: {
            paddingBottom: '0px',
            backgroundColor: 'var(--bg-cl-navbar)',
          },
        }}
        placement="top"
        open={show}
        onClose={() => setShow(false)}
        height={76 + 52 * routes.length + 25}
        closeIcon={null}
      >
        <div className="px-4 mt-2">
          <div className="flex mb-6 justify-between items-center">
            <ThemeSwitcher />
            {data?.tenant && (
              <Title
                className="mb-0"
                level={5}
              >{`${data?.tenant?.metadata?.name}`}</Title>
            )}
            <LogoutButton
              iconStyle={{ fontSize: '24px' }}
              buttonStyle={{ width: '48px' }}
              className="justify-end"
            />
          </div>
          {buttons.map(x => x.content)}
        </div>
      </Drawer>
    </>
  );
};

export default Navbar;
export type { RouteData };
