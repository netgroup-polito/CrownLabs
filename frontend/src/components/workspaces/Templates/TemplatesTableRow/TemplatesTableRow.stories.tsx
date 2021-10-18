import TemplatesTableRow, {
  ITemplatesTableRowProps,
} from './TemplatesTableRow';
import { Story, Meta } from '@storybook/react';
import { someKeysOf, WorkspaceRole } from '../../../../utils';

export default {
  title: 'Components/workspaces/Templates/TemplatesTableRow',
  component: TemplatesTableRow,
  argTypes: {
    id: { table: { disable: true } },
    editTemplate: { table: { disable: true } },
    deleteTemplate: { table: { disable: true } },
  },
} as Meta;

const defaultArgs: someKeysOf<ITemplatesTableRowProps> = {
  id: '0_1',
  name: 'Ubuntu VM',
  gui: true,
  role: WorkspaceRole.manager,
  activeInstances: 2,
  editTemplate: () => null,
  deleteTemplate: () => new Promise(() => null),
  persistent: true,
  resources: { cpu: 2, memory: '8', disk: '16' },
};

const Template: Story<ITemplatesTableRowProps> = args => (
  <TemplatesTableRow {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
