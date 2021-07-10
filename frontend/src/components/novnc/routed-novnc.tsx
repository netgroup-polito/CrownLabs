import { FC } from 'react';
import { RouteComponentProps } from 'react-router-dom';
import { NoVnc } from '.';

interface MatchParams {
  target: string;
}
const RoutedNoVnc: FC<RouteComponentProps<MatchParams>> = ({ ...props }) => {
  return <NoVnc targetWS={props.match.params.target} />;
};

export default RoutedNoVnc;
