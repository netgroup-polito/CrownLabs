import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';
import ViewModeButton, { IViewModeButtonProps } from './ViewModeButton';
import { someKeysOf, WorkspaceRole } from '../../../../utils';

export default {
  title: 'Components/ActivePage/ActiveLanding/ViewModeButton/ViewModeButton',
  component: ViewModeButton,
  parameters: {
    docs: {
      page: () => {
        <>
          <Title />
          <Description />
        </>;
      },
    },
  },
};

const defaultArgs: someKeysOf<IViewModeButtonProps> = {
  currentView: WorkspaceRole.user,
  setCurrentView: () => null,
};

const Template: Story<IViewModeButtonProps> = args => (
  <ViewModeButton {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>View Mode Dropdown</Title>
          <Description>Selector for the manager/user view modes.</Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};
