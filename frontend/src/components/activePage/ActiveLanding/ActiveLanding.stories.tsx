import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';
import ActiveLanding, { IActiveLandingProps } from './ActiveLanding';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/ActivePage/ActiveLanding/ActiveLanding',
  component: ActiveLanding,
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
  decorators: [
    (Story: any) => (
      <div style={{ height: '500px' }}>
        <Story />
      </div>
    ),
  ],
};

const defaultArgs: someKeysOf<IActiveLandingProps> = {
  isTenantManager: false,
};

const Template: Story<IActiveLandingProps> = args => (
  <ActiveLanding {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Active landing page</Title>
          <Description>
            This component showcases the dynamic toggle between **user view**
            and **manager view**. The switch is going to be visible only for
            tenants with manager privileges.
          </Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};

export const Manager = Template.bind({});

Manager.args = { ...defaultArgs, isTenantManager: true };
Manager.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Active landing page - manager</Title>
          <Stories />
        </>
      );
    },
  },
};
