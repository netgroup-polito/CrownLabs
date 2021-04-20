import Example, { IExampleProps } from './Example';
import { Story, Meta } from '@storybook/react';

export default {
  title: 'Components/Example',
  component: Example
} as Meta;

const defaultArgs: IExampleProps = {
  text: 'Example',
  disabled: false,
  onClick: () => {
    console.log('Clicked default');
  }
};

//👇 We create a “template” of how args map to rendering
const Template: Story<IExampleProps> = args => <Example {...args} />;

//👇 Each story then reuses that template
export const Default = Template.bind({});

Default.args = defaultArgs;
