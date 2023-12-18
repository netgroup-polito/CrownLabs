import { FC } from 'react';
import { Row, Col } from 'antd';
import './WorkspaceGridItem.less';
import Badge from '../../../common/Badge';

export interface IWorkspaceGridItemProps {
  id: number;
  title: string;
  isActive: boolean;
  badgeValue?: number;
  previewName?: string;
  onClick: (id: number) => void;
}

const WorkspaceGridItem: FC<IWorkspaceGridItemProps> = ({ ...props }) => {
  const { id, title, isActive, badgeValue, previewName, onClick } = props;

  const preview = (
    previewName ? previewName : title[0] + title[1]
  ).toUpperCase();

  return (
    <Row className="sm:px-0 md:px-4">
      <Col span={24} className="flex justify-center pb-2">
        <button
          className={`cursor-pointer row shadow-lg h-24 w-24 2xl:h-28 2xl:w-28 workspaceitem ${
            isActive ? 'active' : ''
          }`}
          onClick={() => {
            onClick(id);
          }}
        >
          <label
            className="cursor-pointer font-mono font-semibold flex justify-center items-center pt-2 "
            style={{ fontSize: '32pt', color: '#141414' }}
          >
            {preview}
          </label>
          {badgeValue && (
            <Badge
              value={badgeValue}
              size="middle"
              className="absolute top-0 right-0"
            />
          )}
        </button>
      </Col>
      <Col span={24} className="flex justify-center pb-0">
        <p className="w-28 h-6 lg:h-8 2xl:h-12 2xl:w-32 text-center text-xs 2xl:text-sm pb-0">
          {title}
        </p>
      </Col>
    </Row>
  );
};

export default WorkspaceGridItem;
