import { FC, useContext } from 'react';
import { AppstoreAddOutlined } from '@ant-design/icons';
import { Empty } from 'antd';
import { Auth } from '../auth';

export interface ITemplatesEmptyProps {}

const TemplatesEmpty: FC<ITemplatesEmptyProps> = ({ ...props }) => {
  const auth = useContext(Auth);

  return auth ? (
    <div className="w-full flex-grow flex flex-wrap content-center justify-center py-5 2xl:py-52">
      <div className="w-full pb-10 flex justify-center">
        <p className="text-7xl text-center mb-0">
          <AppstoreAddOutlined style={{ color: '#1890ff' }} />
        </p>
      </div>
      <div className="w-full">
        <p className="text-3xl text-center px-24 block">
          <b>There are no Template yet</b>
        </p>
      </div>
      <div className="w-full">
        <p className="text-2xl text-center px-24 block">
          You can <b>Create</b> and <b>Customize</b> new ones for your Students
          clicking on the top-right Button
        </p>
      </div>
    </div>
  ) : (
    <div className="w-full flex-grow flex flex-wrap content-center justify-center py-5 2xl:py-52">
      <div className="w-full pb-10 flex justify-center">
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={false} />
      </div>
      <p className="text-3xl text-center px-24">No Templates available</p>
    </div>
  );
};

export default TemplatesEmpty;
