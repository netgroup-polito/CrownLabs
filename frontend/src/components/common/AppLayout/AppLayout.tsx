import { FC, useState } from 'react';
import { Layout, Row, Col } from 'antd';
import Navbar from '../Navbar';
import { BrowserRouter, Route, Switch } from 'react-router-dom';
import { logout } from '../../../contexts/AuthContext';
import SidebarInfo from '../SidebarInfo';
import TooltipButton from '../TooltipButton';
import './AppLayout.less';
import { TooltipButtonData } from '../TooltipButton/TooltipButton';
import { PUBLIC_URL } from '../../../env';
import { InfoOutlined } from '@ant-design/icons';
import { RouteDescriptor } from '../../../utils';

const { Content } = Layout;

export interface IAppLayoutProps {
  routes: Array<RouteDescriptor>;
  TooltipButtonData?: TooltipButtonData;
  TooltipButtonLink?: string;
  transparentNavbar?: boolean;
}

const AppLayout: FC<IAppLayoutProps> = ({ ...props }) => {
  const [sideLeftShow, setSideLeftShow] = useState(false);
  const { routes, transparentNavbar, TooltipButtonData, TooltipButtonLink } =
    props;

  return (
    <BrowserRouter basename={PUBLIC_URL}>
      <Layout className="h-full">
        <Navbar
          logoutHandler={logout}
          routes={routes}
          transparent={transparentNavbar}
        />
        <Content className="flex">
          <Switch>
            {routes.map(r =>
              r.content ? (
                <Route exact key={r.route.path} path={r.route.path}>
                  <Row className="h-full pt-5 xs:pt-10 pb-20 flex w-full px-4">
                    <Col span={0} lg={1} xxl={2}></Col>
                    {r.content}
                    <Col span={0} lg={1} xxl={2}></Col>
                  </Row>
                </Route>
              ) : null
            )}
          </Switch>
        </Content>
        <div className="left-TooltipButton">
          <TooltipButton
            TooltipButtonData={{
              tooltipTitle: 'Show CrownLabs infos',
              tooltipPlacement: 'right',
              type: 'primary',
              icon: <InfoOutlined style={{ fontSize: '22px' }} />,
            }}
            onClick={() => setSideLeftShow(true)}
          />
        </div>
        {TooltipButtonData && (
          <div className="right-TooltipButton">
            <TooltipButton
              TooltipButtonData={{
                tooltipTitle: TooltipButtonData.tooltipTitle,
                tooltipPlacement: TooltipButtonData.tooltipPlacement,
                type: TooltipButtonData.type,
                icon: TooltipButtonData.icon,
              }}
              onClick={() => window.open(TooltipButtonLink, '_blank')}
            />
          </div>
        )}
      </Layout>
      <SidebarInfo
        show={sideLeftShow}
        setShow={setSideLeftShow}
        position="left"
      />
    </BrowserRouter>
  );
};

export default AppLayout;
