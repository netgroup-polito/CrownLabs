import NestedTables, { INestedTablesProps } from './NestedTables';
import { Story } from '@storybook/react';
import {
  Title,
  Description,
  Stories,
  Primary,
} from '@storybook/addon-docs/blocks';
import { someKeysOf } from '../../../utils';
import { templates, workspaces, instances } from '../tempData';

export default {
  title: 'Components/ActivePage/NestedTables/NestedTables',
  component: NestedTables,
  argTypes: {
    destroyAll: { action: 'clicked' },
  },
  parameters: {
    docs: {
      page: () => {
        <>
          <Title>Workpaces/Templates table</Title>
          <Description />
        </>;
      },
    },
  },
  decorators: [
    (Story: any) => (
      <div style={{ height: '300px' }}>
        <Story />
      </div>
    ),
  ],
};

const defaultArgs: someKeysOf<INestedTablesProps> = {
  workspaces: workspaces,
  templates: templates,
  instances: instances,
  nested: false,
};

const Template: Story<INestedTablesProps> = args => <NestedTables {...args} />;

export const UserView = Template.bind({});

UserView.args = defaultArgs;
UserView.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Active Templates Table</Title>
          <Description>
            Table containing active personal templates. This is the default view
            for any tenant and also the only view for user-privileged tenants.
            All active VM instances are stored inside the respective template.
          </Description>
          <Primary />
        </>
      );
    },
  },
};

export const ManagerView = Template.bind({});

ManagerView.args = { ...defaultArgs, nested: true };
ManagerView.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Active Workspaces Table</Title>
          <Description>
            Table containing workspaces which in turn contain a nested table of
            templates. This layout is intended to be viewed only by
            manager-level tenants to help them organize the active templates
            they own and that are used by other tenants.
          </Description>
          <Stories />
        </>
      );
    },
  },
};
