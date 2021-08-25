import { Story } from '@storybook/react';

const FixedHeightDecorator = (Story: Story) => {
  return (
    <div style={{ height: '300px' }}>
      <Story />
    </div>
  );
};

export { FixedHeightDecorator };
