import type { FC } from 'react';
import { WorkspaceGridItem } from '../WorkspaceGridItem';

export interface IWorkspaceGridProps {
  workspaceItems: Array<{ id: number; title: string; waitingTenants?: number }>;
  selectedWs: number;
  onClick: (id: number) => void;
}

const WorkspaceGrid: FC<IWorkspaceGridProps> = ({ ...props }) => {
  const { workspaceItems, selectedWs, onClick } = props;
  return (
    <div className="grid lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 lg:grid-flow-row grid-rows-1 grid-flow-col gap-3 lg:gap-0">
      {workspaceItems.map(workspaceItem => (
        <div key={workspaceItem.id} className="h-full flex justify-center">
          <WorkspaceGridItem
            id={workspaceItem.id}
            title={workspaceItem.title}
            isActive={selectedWs === workspaceItem.id}
            badgeValue={workspaceItem.waitingTenants}
            onClick={onClick}
          />
        </div>
      ))}
      <WorkspaceGridItem
        id={-2}
        title="Add Workspace"
        previewName="+"
        isActive={selectedWs === -2}
        onClick={onClick}
      />
    </div>
  );
};

export default WorkspaceGrid;
