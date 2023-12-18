import Box from '../../common/Box';
import { FC } from 'react';
import { WorkspacesListLogic } from './WorkspacesListLogic';

export interface IWorkspaceAddProps {}

const WorkspaceAdd: FC<IWorkspaceAddProps> = ({ ...args }) => {
  return (
    <Box
      header={{
        size: 'large',
        center: (
          <div className="h-full flex justify-center items-center px-5">
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Join a new Workspace</b>
            </p>
          </div>
        ),
      }}
    >
      <WorkspacesListLogic />
    </Box>
  );
};

export default WorkspaceAdd;
