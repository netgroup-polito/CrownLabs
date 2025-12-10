import Box from '../../common/Box';
import type { FC } from 'react';
import Logo from '../../common/Logo';

const WorkspaceWelcome: FC = () => {
  return (
    <Box
      header={{
        size: 'large',
        center: (
          <div className="h-full flex justify-center items-center px-5">
            <p className="md:text-4xl text-2xl text-center mb-0">
              <b>Welcome to CrownLabs</b>
            </p>
          </div>
        ),
      }}
    >
      <div className="w-full h-full flex-grow flex items-center justify-center flex-col py-5">
        <div className="w-full pb-10 flex justify-center">
          <Logo widthPx={150} />
        </div>
        <p className="text-xl xs:text-3xl text-center px-5 xs:px-24 hidden md:flex">
          No workspace selected
        </p>
      </div>
    </Box>
  );
};

export default WorkspaceWelcome;
