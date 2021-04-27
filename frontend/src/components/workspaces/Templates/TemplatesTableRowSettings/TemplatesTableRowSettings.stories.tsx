import TemplatesTableRowSettings, {
  ITemplatesTableRowSettingsProps,
} from './TemplatesTableRowSettings';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../../utils';

export default {
  title: 'Components/workspaces/Templates/TemplatesTableRowSettings',
  component: TemplatesTableRowSettings,
  argTypes: {},
} as Meta;

const defaultArgs: someKeysOf<ITemplatesTableRowSettingsProps> = {
  id: '0_1',
  editTemplate: () => null,
  deleteTemplate: () => null,
};

const Template: Story<ITemplatesTableRowSettingsProps> = args => (
  <TemplatesTableRowSettings {...args} />
);

export const Default = Template.bind({});

Default.args = defaultArgs;
