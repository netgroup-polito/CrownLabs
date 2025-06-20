import type { FC } from 'react';
import Box from '../../common/Box';
export interface IUserPanelContainerProps {
  children: React.ReactNode;
}

const UserPanelContainer: FC<IUserPanelContainerProps> = props => {
  const { children } = props;
  return (
    <Box
      header={{
        size: 'middle',
        center: (
          <div className="h-full flex justify-center items-center px-5">
            <p className="md:text-2xl text-lg text-center mb-0">
              <b>Account details</b>
            </p>
          </div>
        ),
      }}
    >
      {children}
    </Box>
  );
};

export default UserPanelContainer;
