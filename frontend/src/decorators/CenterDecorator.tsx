import { Story } from '@storybook/react';
import ThemeContextProvider from '../contexts/ThemeContext';

const CenterDecorator = (Story: Story) => {
  return (
    <ThemeContextProvider>
      <div className="flex h-screen items-center justify-center">
        <Story />
      </div>
    </ThemeContextProvider>
  );
};

export { CenterDecorator };
