import Course, { ICourseProps } from './Course';
import { Story, Meta } from '@storybook/react';
import { someKeysOf } from '../../../utils';

export default {
  title: 'Components/Course',
  component: Course,
  argTypes: { onClick: { action: 'clicked' } },
} as Meta;

const defaultArgs: someKeysOf<ICourseProps> = {
  title: 'Reti locali e data center',
  selected: true,
};

const Template: Story<ICourseProps> = args => <Course {...args} />;

export const Default = Template.bind({});

Default.args = defaultArgs;
