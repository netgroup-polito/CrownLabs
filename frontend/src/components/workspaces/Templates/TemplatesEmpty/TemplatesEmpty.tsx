import type { FC } from 'react';
import { AppstoreAddOutlined } from '@ant-design/icons';
import { Empty } from 'antd';
import { WorkspaceRole } from '../../../../utils';

export interface ITemplatesEmptyProps {
  role: WorkspaceRole;
}

const TemplatesEmpty: FC<ITemplatesEmptyProps> = ({ ...props }) => {
  const { role } = props;
  return role === WorkspaceRole.manager ? (
    <div className="w-full h-full flex-grow flex flex-wrap content-center justify-center py-5 ">
      <div className="w-full pb-10 flex justify-center">
        <p className="text-6xl md:text-7xl text-center mb-0 primary-color-fg">
          <AppstoreAddOutlined />
        </p>
      </div>
      <div className="w-full">
        <p className="text-2xl md:text-3xl text-center px-5 px-10 md:px-24 block">
          <b>There are no Template yet</b>
        </p>
      </div>
      <div className="w-full">
        <p className="text-xl md:text-2xl text-center px-5 px-10 md:px-24 block">
          You can <b>Create</b> and <b>Customize</b> new ones for your Users
          clicking on the top-right Button
        </p>
      </div>
    </div>
  ) : (
    <div className="w-full h-full flex-grow flex flex-wrap content-center justify-center py-5 ">
      <div className="w-full pb-10 flex justify-center">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={false} />
      </div>
      <p className="text-xl xs:text-3xl text-center px-5 xs:px-24">
        No Templates available
      </p>
    </div>
  );
};

export default TemplatesEmpty;
