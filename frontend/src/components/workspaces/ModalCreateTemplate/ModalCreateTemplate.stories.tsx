import ModalCreateTemplate, {
  IModalCreateTemplateProps,
  Template,
} from './ModalCreateTemplate';
import { Story, Meta } from '@storybook/react';
import { DialogOpenDecorator } from '../../../decorators/DialogOpenDecorator';

export default {
  title: 'Components/workspaces/ModalCreateTemplate',
  component: ModalCreateTemplate,
  decorators: [DialogOpenDecorator],
  argTypes: {
    submitHandler: { table: { disable: true } },
    show: { table: { disable: true } },
    setShow: { table: { disable: true } },
  },
} as Meta;

const TemplateStorybook: Story<IModalCreateTemplateProps> = args => (
  <ModalCreateTemplate {...args} />
);

export const Create = TemplateStorybook.bind({});
Create.args = {
  submitHandler: (t?: Template) => {
    alert(JSON.stringify(t));
    return Promise.resolve({ data: undefined });
  },
  diskInterval: { min: 1, max: 32 },
  ramInterval: { min: 4, max: 16 },
  cpuInterval: { min: 1, max: 4 },
  images: [
    {
      name: 'Ubuntu',
      registry: 'registry',
      vmorcontainer: ['Container', 'VM'],
    },
    {
      name: 'Windows',
      registry: 'registry',
      vmorcontainer: ['Container'],
    },
  ],
};

export const Modify = TemplateStorybook.bind({});
Modify.args = {
  ...Create.args,
  template: {
    name: 'Existing Template',
    image: 'Ubuntu',
    registry: 'registry',
    vmorcontainer: 'Container',
    diskMode: false,
    gui: true,
    cpu: 2,
    ram: 16,
    disk: 24,
  },
};
