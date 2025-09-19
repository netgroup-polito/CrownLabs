import { InfoOutlined } from '@ant-design/icons';
import { Col, Layout, Result, Row } from 'antd';
import { type FC, useContext, useState } from 'react';
import { BrowserRouter, Route, Routes } from 'react-router-dom';
import { TenantContext } from '../../../contexts/TenantContext';
import { BASE_URL } from '../../../env';
import { type RouteDescriptor } from '../../../utils';
import FullPageLoader from '../FullPageLoader';
import Navbar from '../Navbar';
import SidebarInfo from '../SidebarInfo';
import TooltipButton from '../TooltipButton';
import type { TooltipButtonData } from '../TooltipButton/TooltipButton';
import './AppLayout.less';
import { AuthContext } from '../../../contexts/AuthContext';

const { Content } = Layout;

export interface IAppLayoutProps {
  routes: Array<RouteDescriptor>;
  TooltipButtonData?: TooltipButtonData;
  TooltipButtonLink?: string;
  transparentNavbar?: boolean;
}

const AppLayout: FC<IAppLayoutProps> = ({ ...props }) => {
  const { profile } = useContext(AuthContext);

  const [sideLeftShow, setSideLeftShow] = useState(false);
  const { routes, transparentNavbar, TooltipButtonData, TooltipButtonLink } =
    props;

  const { data: tenantData } = useContext(TenantContext);
  const tenantNsIsReady =
    tenantData?.tenant?.status?.personalNamespace?.created ?? false;
  const firstName = profile?.given_name;

  return (
    <BrowserRouter basename={BASE_URL}>
      <Layout className="min-h-screen flex flex-col">
        <Navbar routes={routes} transparent={transparentNavbar} />
        <Content className="flex-1 overflow-hidden">
          {tenantNsIsReady ? (
            <Routes>
              {routes
                .filter(r => r.content)
                .map(r => (
                  <Route
                    key={r.route.path}
                    path={r.route.path}
                    element={
                      <div
                        style={{
                          height: '100%',
                          padding: '20px 0',
                          overflow: 'hidden',
                        }}
                      >
                        <Row style={{ height: '100%' }}>
                          <Col span={0} lg={1} xxl={2}></Col>
                          <Col
                            span={24}
                            lg={22}
                            xxl={20}
                            style={{ height: '100%' }}
                          >
                            {r.content}
                          </Col>
                          <Col span={0} lg={1} xxl={2}></Col>
                        </Row>
                      </div>
                    }
                  />
                ))}
              <Route
                element={
                  <div className="flex justify-center items-center w-full">
                    <Result
                      status="404"
                      title="404"
                      subTitle="Sorry, the page you visited does not exist."
                    />
                  </div>
                }
              />
            </Routes>
          ) : (
            <FullPageLoader
              text={firstName ? `Welcome back ${firstName}!` : 'Welcome back!'}
              subtext="Settings things back up... Hold tight!"
            />
          )}
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
