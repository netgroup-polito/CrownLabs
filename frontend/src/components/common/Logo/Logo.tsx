import { FC } from 'react';
import './Logo.less';
import { ReactComponent as SvgLogo } from '../../../assets/logo.svg';

export interface ILogoProps {
  widthPx?: number;
  className?: string;
  color?: string;
}

const Logo: FC<ILogoProps> = ({ ...props }) => {
  const { widthPx, className, color } = props;
  return (
    <SvgLogo
      width={widthPx ? `${widthPx}px` : '100%'}
      className={className + (!color ? ' logo-color' : '')}
      style={{ fill: color }}
    />
  );
};

export default Logo;
