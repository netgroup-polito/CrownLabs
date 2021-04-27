import ModalCreateTemplate, {
  IModalCreateTemplateProps,
} from './ModalCreateTemplate';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/ModalCreateTemplate',
  component: ModalCreateTemplate,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<IModalCreateTemplateProps> = {
  showmodal: true,
};

const Template: Story<IModalCreateTemplateProps> = args => (
  <ModalCreateTemplate {...args} />
);

export const Default = Template.bind({});
Default.args = defaultArgs;
