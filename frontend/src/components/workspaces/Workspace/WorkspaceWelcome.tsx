import logo from '../img/logo.svg';
import Box from '../common/Box';
import { FC } from 'react';

export interface IWorkspaceWelcomeProps {}

const WorkspaceWelcome: FC<IWorkspaceWelcomeProps> = ({ ...args }) => {
  return (
    <Box headTitle={'Welcome to CrownLabs'}>
      <div className="w-full flex-grow flex flex-wrap content-center justify-center py-5 2xl:py-52">
        <div className="w-full pb-10 flex justify-center">
          <img
            width={150}
            src={logo}
            style={{ color: '#1890ff' }}
            alt="Logo Crownlabs"
          />
        </div>
        <p className="text-3xl text-center px-24">
          Select a Course <br />
          to explore Virtual Machines <br />
          and Services available
        </p>
      </div>
    </Box>
  );
};

export default WorkspaceWelcome;
