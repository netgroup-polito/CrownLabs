import InstanceActions from './InstanceActions';
import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';

export default {
  title: 'Components/ActivePage/Instances/InstanceActions',
  component: InstanceActions,
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

const defaultArgs = {
  isManaged: false,
  displayName: 'Display Name',
  tenantId: 's123456',
  tenantDisplayName: 'Name Surname',
};

const Template: Story = args => <InstanceActions {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Instance row Buttons</Title>
          <Description>
            Buttons for the different actions available for each active
            instance.
          </Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};
