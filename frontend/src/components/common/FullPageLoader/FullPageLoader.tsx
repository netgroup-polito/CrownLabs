import { FC } from 'react';
import './FullPageLoader.less';
import CrownLoader from '../../misc/CrownLoader';

export interface IFullPageLoaderProps {
  text: string;
  subtext: string;
}

const FullPageLoader: FC<IFullPageLoaderProps> = ({ ...props }) => {
  const { text, subtext } = props;

  return (
    <div className="cl-full-page-loader text-center">
      <CrownLoader size={'min(40vw, 30vh)'} duration={3} />
      <h1>{text}</h1>
      <span>{subtext}</span>
    </div>
  );
};

export default FullPageLoader;
