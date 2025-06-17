import type { FC } from 'react';
import './FullPageLoader.less';
import CrownLoader from '../../misc/CrownLoader';
import { Layout } from 'antd';

export interface IFullPageLoaderProps {
  text?: string;
  subtext?: string;
  layoutWrap?: boolean;
}

const FullPageLoader: FC<IFullPageLoaderProps> = ({ ...props }) => {
  const { text, subtext, layoutWrap } = props;

  const cont = (
    <div className="cl-full-page-loader text-center">
      <CrownLoader size={'min(40vw, 30vh)'} duration={3} />
      <h1>{text || 'Loading...'}</h1>
      <span>{subtext || 'Hold tight!'}</span>
    </div>
  );

  if (!layoutWrap) return cont;

  return (
    <Layout className="h-full">
      <Layout.Content>{cont}</Layout.Content>
    </Layout>
  );
};

export default FullPageLoader;
