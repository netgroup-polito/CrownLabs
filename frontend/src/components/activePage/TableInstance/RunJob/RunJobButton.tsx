import { ThunderboltOutlined } from '@ant-design/icons';
import { Badge, Button, Tooltip } from 'antd';
import type { FC } from 'react';

export interface IRunJobButtonProps {
  extended: boolean;
  activeJobs: number;
  onClick: () => void;
}

// Beside Connect, with its breakpoints: a round button, then an icon. Below
// that the row folds and "Run job" is in its menu.
const RunJobButton: FC<IRunJobButtonProps> = ({
  extended,
  activeJobs,
  onClick,
}) => (
  <>
    <div className={`hidden ${extended ? 'lg:block' : 'sm:block'}`}>
      <Badge count={activeJobs} showZero={false} size="small">
        <Button
          shape="round"
          size="middle"
          icon={<ThunderboltOutlined />}
          onClick={onClick}
        >
          Run job
        </Button>
      </Badge>
    </div>
    <div
      className={`hidden ${
        extended ? 'sm:block lg:hidden' : 'xs:block sm:hidden'
      }`}
    >
      <Tooltip placement="top" title="Run job">
        <Badge count={activeJobs} showZero={false} size="small">
          <Button
            className="flex items-center justify-center p-0 border-0"
            type="link"
            shape="circle"
            size="middle"
            aria-label="Run job"
            icon={
              <ThunderboltOutlined
                className="flex items-center justify-center"
                style={{ fontSize: '22px' }}
              />
            }
            onClick={onClick}
          />
        </Badge>
      </Tooltip>
    </div>
  </>
);

export default RunJobButton;
