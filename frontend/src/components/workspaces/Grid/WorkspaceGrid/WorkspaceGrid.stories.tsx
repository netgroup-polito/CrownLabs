import WorkspaceGrid, { IWorkspaceGridProps } from './WorkspaceGrid';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/workspaces/Grid/WorkspaceGrid',
  component: WorkspaceGrid,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<IWorkspaceGridProps> = {
  workspaceItems: [
    {
      id: 0,
      title: 'Reti Locali e Data Center',
    },
    {
      id: 1,
      title: 'Tecnologie e Servizi di Rete',
    },
    {
      id: 2,
      title: 'Applicazioni Web I',
    },
    {
      id: 3,
      title: 'Cloud Computing',
    },
    {
      id: 4,
      title: 'Programmazione di Sistema',
    },
    {
      id: 5,
      title: 'Information System Security',
    },
    {
      id: 6,
      title: 'Ingegneria del Software',
    },
    {
      id: 7,
      title: 'Data Science',
    },
    {
      id: 8,
      title: 'Software Networking',
    },
  ],
  selectedWs: 1,
};

const Template: Story<IWorkspaceGridProps> = args => (
  <WorkspaceGrid {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
