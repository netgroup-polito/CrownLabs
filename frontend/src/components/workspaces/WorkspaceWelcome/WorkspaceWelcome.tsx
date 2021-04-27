import logo from '../../../assets/logo.svg';
import Box from '../../common/Box';
import { FC } from 'react';

export interface IWorkspaceWelcomeProps {}

const WorkspaceWelcome: FC<IWorkspaceWelcomeProps> = ({ ...args }) => {
  return (
    <Box
      header={{
        size: 'large',
        center: (
          <p className="md:text-4xl text-2xl text-center mb-0">
            <b>Welcome to CrownLabs</b>
          </p>
        ),
      }}
    >
      <div className="w-full flex-grow flex flex-wrap content-center justify-center py-5 2xl:py-52">
        <div className="w-full pb-10 flex justify-center">
          <img width={150} src={logo} alt="Logo Crownlabs" />
        </div>
        <p className="text-xl xs:text-3xl text-center px-5 xs:px-24">
          Select a Workspace <br />
          to explore Virtual Machines <br />
          and Services available
        </p>
      </div>
    </Box>
  );
};

export default WorkspaceWelcome;
