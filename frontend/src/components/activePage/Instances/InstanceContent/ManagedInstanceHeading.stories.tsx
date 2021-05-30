import ManagedInstanceHeading, {
  IManagedInstanceHeadingProps,
} from './ManagedInstanceHeading';
import { Story } from '@storybook/react';
import {
  Title,
  Description,
  Stories,
  ArgsTable,
} from '@storybook/addon-docs/blocks';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/ActivePage/Instances/ManagedInstanceHeading',
  component: ManagedInstanceHeading,
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
};

const defaultArgs: someKeysOf<IManagedInstanceHeadingProps> = {
  displayName: 'Display Name',
  tenantId: 's123456',
  tenantDisplayName: 'Name Surname',
};

const Template: Story<IManagedInstanceHeadingProps> = args => (
  <ManagedInstanceHeading {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Instance Heading</Title>
          <Description>Specific heading for managed instances</Description>
          <Stories includePrimary />
          <ArgsTable />
        </>
      );
    },
  },
};
