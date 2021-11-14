import { FC } from 'react';
import { Layout, Drawer } from 'antd';
import Button from 'antd-button-color';
import { MenuOutlined } from '@ant-design/icons';
import { Link, useLocation } from 'react-router-dom';
import { useState } from 'react';
import ThemeSwitcher from '../../misc/ThemeSwitcher';
import './Navbar.less';
import Logo from '../Logo';
import { LogoutButton } from '../LogoutButton';

const Header = Layout.Header;

type RouteData = {
  name: string;
  path: string;
};

export interface INavbarProps {
  routes: Array<RouteData>;
  transparent?: boolean;
  logoutHandler: () => void;
}

const Navbar: FC<INavbarProps> = ({ ...props }) => {
  const { routes, transparent, logoutHandler } = props;
  const [show, setShow] = useState(false);

  const currentPath = useLocation().pathname;

  const currentName = routes.find(r => r.path === currentPath)?.name || '';

  const buttons = routes.map((b, i) => {
    const isExtLink = b.path.indexOf('http') === 0;
    return (
      <Link
        target={isExtLink ? '_blank' : ''}
        key={i}
        to={{ pathname: isExtLink ? '' : b.path }}
        rel={isExtLink ? 'noopener noreferrer' : ''}
      >
        <Button
          onClick={() =>
            isExtLink ? window.open(b.path, '_blank') : setShow(false)
          }
          ghost={currentPath !== b.path}
          className={
            'w-full flex justify-center my-3 ' +
            (routes.length <= 4
              ? 'lg:mx-4 md:mx-2 md:w-28 lg:w-36 xl:w-52 2xl:w-72 '
              : 'lg:mx-2 lg:w-28 xl:w-32 2xl:w-48') +
            (currentPath !== b.path ? ' navbar-button ' : '')
          }
          size="large"
          type={currentPath !== b.path ? 'default' : 'primary'}
          shape="round"
        >
          {b.name}
        </Button>
      </Link>
    );
  });

  return (
    <>
      <Header
        className={
          'flex h-auto px-6 justify-between ' +
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
          {buttons}
        </div>
        <div
          className={
            'w-full hidden sm:flex justify-end ' +
            (routes.length > 4 ? 'lg:hidden' : 'md:hidden')
          }
        >
          {buttons
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

            <LogoutButton
              className=" justify-end"
              iconStyle={{ fontSize: '24px' }}
              logoutHandler={logoutHandler}
            />
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
        bodyStyle={{
          paddingBottom: '0px',
          backgroundColor: 'var(--bg-cl-navbar)',
        }}
        placement="top"
        visible={show}
        onClose={() => setShow(false)}
        height={76 + 52 * routes.length + 25}
        closeIcon={null}
      >
        <div className="px-4 mt-2">
          <div className="flex mb-6 justify-between items-center">
            <ThemeSwitcher />
            <LogoutButton
              logoutHandler={logoutHandler}
              iconStyle={{ fontSize: '24px' }}
              className="justify-end"
            />
          </div>
          {buttons}
        </div>
      </Drawer>
    </>
  );
};

export default Navbar;
export type { RouteData };
