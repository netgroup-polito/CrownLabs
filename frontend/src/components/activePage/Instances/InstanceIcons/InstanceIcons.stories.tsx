import InstanceIcons, { IInstanceIconsProps } from './InstanceIcons';
import { Story } from '@storybook/react';
import { Title, Description, Stories } from '@storybook/addon-docs/blocks';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/ActivePage/Instances/InstanceIcons',
  component: InstanceIcons,
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

const defaultArgs: someKeysOf<IInstanceIconsProps> = {
  isGUI: false,
  phase: 'ready',
};

const Template: Story<IInstanceIconsProps> = args => (
  <InstanceIcons {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
Default.parameters = {
  docs: {
    // eslint-disable-next-line react/no-multi-comp
    page: () => {
      return (
        <>
          <Title>Instance Icons</Title>
          <Description>
            Instance icons, one for the type of VM (CLI or GUI), one for the
            status and one with a VM specific image
          </Description>
          <Stories includePrimary />
        </>
      );
    },
  },
};
