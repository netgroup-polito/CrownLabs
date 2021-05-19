import { FC, Dispatch, SetStateAction } from 'react';
import { Radio } from 'antd';

export interface IViewModeButtonProps {
  managerView: boolean;
  setManagerView: Dispatch<SetStateAction<boolean>>;
}

const ViewModeButton: FC<IViewModeButtonProps> = ({ ...props }) => {
  const { setManagerView, managerView } = props;

  return (
    <Radio.Group
      value={managerView}
      onChange={e => setManagerView(e.target.value)}
    >
      <Radio.Button value={false}>Personal Instances</Radio.Button>
      <Radio.Button value={true}>Managed Instances</Radio.Button>
    </Radio.Group>
  );
};

export default ViewModeButton;
