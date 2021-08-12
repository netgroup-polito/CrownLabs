import { FC } from 'react';
import './Logo.less';
import { ReactComponent as SvgLogo } from '../../../assets/logo.svg';

export interface ILogoProps {
  widthPx?: number;
  className?: string;
}

const Logo: FC<ILogoProps> = ({ ...props }) => {
  const { widthPx, className } = props;
  return (
    <div className={'flex items-center justify-center m-0 p-0 ' + className}>
      <SvgLogo
        width={widthPx ? `${widthPx}px` : '100%'}
        className={'logo-color'}
      />
    </div>
  );
};

export default Logo;
