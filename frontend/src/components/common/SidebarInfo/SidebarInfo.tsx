import { FC } from 'react';
import { Drawer, List } from 'antd';
import {
  GithubOutlined,
  LinkOutlined,
  MailOutlined,
  NotificationOutlined,
  SlackOutlined,
  YoutubeFilled,
} from '@ant-design/icons';
import Logo from '../Logo';

export interface ISidebarInfoProps {
  onClose: () => void;
  position: 'left' | 'right';
  visible: boolean;
}

const SidebarInfo: FC<ISidebarInfoProps> = ({ ...props }) => {
  const { onClose, visible, position } = props;

  return (
    <Drawer
      headerStyle={{
        borderBottom: `solid #1c7afd`,
        borderBottomWidth: '1px',
      }}
      contentWrapperStyle={
        position === 'left'
          ? {
              borderRight: `solid #1c7afd`,
              borderRightWidth: '1px',
            }
          : { borderLeft: `solid #1c7afd`, borderLeftWidth: '1px' }
      }
      bodyStyle={{ padding: '0px' }}
      footerStyle={{ padding: '0px' }}
      title={
        <>
          <div className="flex">
            <div className="flex-none flex items-center w-12">
              <Logo />
            </div>
            <div className="h-full flex flex-grow justify-center items-center px-5">
              <p className="md:text-2xl text-center mb-0">
                <b>CrownLabs</b>
              </p>
            </div>
            <div className="flex-none w-12"></div>
          </div>
        </>
      }
      placement={position}
      closable={true}
      onClose={onClose}
      visible={visible}
      width={350}
      footer={
        <>
          <div className="m-3.5 text-center flex justify-between px-8">
            <b>
              This software has been proudly developed at{' '}
              <a target="_blank" rel="noreferrer" href="https://www.polito.it">
                Politecnico di Torino
              </a>
            </b>
          </div>
        </>
      }
    >
      <div className="mx-4 mt-5 text-sm text-center">
        <p>
          CrownLabs provides immediate access to your{' '}
          <b>remote computing labs</b>, without any special requirements: just a{' '}
          <b>browser</b>!
        </p>
        <p className="mb-2">
          Do not install on your laptop the tools required by your subjects:
          connect to your remote environment, with{' '}
          <b>everithing already set up</b>.
        </p>
      </div>
      <List
        size="large"
        dataSource={[
          {
            icon: <LinkOutlined style={{ fontSize: '25px' }} />,
            title: null,
            link: 'http://crownlabs.polito.it',
            linktext: 'http://crownlabs.polito.it',
          },
          {
            icon: <NotificationOutlined style={{ fontSize: '25px' }} />,
            title: 'Telegram',
            link: 'https://t.me/crownlabsNews',
            linktext: 'crownlabsNews',
          },
          {
            icon: <GithubOutlined style={{ fontSize: '25px' }} />,
            title: 'GitHub',
            link: 'https://github.com/netgroup-polito/CrownLabs',
            linktext: 'netgroup-polito/Crownlabs',
          },
          {
            icon: <SlackOutlined style={{ fontSize: '25px' }} />,
            title: 'Slack',
            link: 'https://crown-team-group.slack.com/',
            linktext: 'crown-team-group',
          },
          {
            icon: <MailOutlined style={{ fontSize: '25px' }} />,
            title: 'Email',
            link: 'crownlabs@polito.it',
            linktext: 'crownlabs@polito.it',
          },
          {
            icon: <YoutubeFilled style={{ fontSize: '25px' }} />,
            title: 'YouTube',
            link:
              'https://www.youtube.com/playlist?list=PLTAfidx4guQIIPZVaEn8H_hfSTFJ5VQDu',
            linktext: 'CrownLabs videos',
          },
        ]}
        renderItem={item => (
          <List.Item className="flex justify-start pl-8 pr-0">
            {item.icon}
            <div className="ml-4 flex items-end">
              {item.title && <h3 className="mr-2 m-0">{item.title}</h3>}
              <a target="_blank" rel="noreferrer" href={item.link}>
                {item.linktext}
              </a>
            </div>
          </List.Item>
        )}
      />
    </Drawer>
  );
};

export default SidebarInfo;
