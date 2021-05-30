import InstanceContent, { IInstanceContentProps } from './InstanceContent';
import { Story } from '@storybook/react';
import {
  Title,
  Description,
  Stories,
  Primary,
  ArgsTable,
} from '@storybook/addon-docs/blocks';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/ActivePage/Instances/InstanceContent',
  component: InstanceContent,
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

const defaultArgs: someKeysOf<IInstanceContentProps> = {
  isManaged: false,
  displayName: 'Display Name',
  tenantId: 's123456',
  tenantDisplayName: 'Name Surname',
};

const Template: Story<IInstanceContentProps> = args => (
  <InstanceContent {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Instance row content</Title>
          <Description>
            Contents consist of three separate elements: heading, info icon and
            actions. The contents will be different depending if the instance is
            personal or managed: the former will have a more concise heading, an
            info button and standard actions while the latter will have a more
            verbose heading, no info button and specific actions.
          </Description>
          <Stories includePrimary />
          <ArgsTable />
        </>
      );
    },
  },
};

export const Managed = Template.bind({});

Managed.args = { ...defaultArgs, isManaged: true };
Managed.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Instance row content</Title>
          <Description>Managed instance with a different heading.</Description>
          <Primary />
          <ArgsTable />
        </>
      );
    },
  },
};
