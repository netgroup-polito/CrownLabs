import InstancesTable, { IInstancesTableProps } from './InstancesTable';
import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';
import { someKeysOf } from '../../../../utils';
import { instances } from '../../tempData';

export default {
  title: 'Components/ActivePage/Instances/InstancesTable',
  component: InstancesTable,
  parameters: {
    docs: {
      page: () => {
        <>
          <Title>Instances table</Title>
          <Description />
        </>;
      },
    },
  },
};

const defaultArgs: someKeysOf<IInstancesTableProps> = {
  instances: instances,
  isManaged: false,
};

const Template: Story<IInstancesTableProps> = args => (
  <InstancesTable {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Active Instances Table</Title>
          <Description>
            This table contains all the active intaces available per tempalte.
            Each row is ivided in two columns: one for the icons and one for the
            actial content.
          </Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};
