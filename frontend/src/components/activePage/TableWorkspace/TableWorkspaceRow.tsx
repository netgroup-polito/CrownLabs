import { Badge, Space, Typography } from 'antd';
import { useMemo, type FC } from 'react';
import type { Template } from '../../../utils';

const { Text } = Typography;
export interface ITableWorkspaceRowProps {
  title: string;
  templates: Template[];
  id: string;
  expandRow: (rowId: string) => void;
}

const TableWorkspaceRow: FC<ITableWorkspaceRowProps> = ({ ...props }) => {
  const { title, templates, id, expandRow } = props;

  const { nTotal, nRunning } = useMemo(() => {
    const nTotal = templates.reduce((acc, t) => acc + t.instances.length, 0);
    const nRunning = templates.reduce(
      (acc, t) => acc + t.instances.filter(i => i.running).length,
      0,
    );
    return { nTotal, nRunning };
  }, [templates]);

  return (
    <div
      className="w-full flex justify-between cursor-pointer"
      onClick={() => expandRow(id)}
    >
      <Space size="middle">
        <Badge
          count={`${nRunning}/${nTotal}`}
          className="mx-0"
          color="magenta"
        />
        <Text className="font-bold w-48 xs:w-56 sm:w-max" ellipsis>
          {title}
        </Text>
      </Space>
    </div>
  );
};

export default TableWorkspaceRow;
