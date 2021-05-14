import '../src/theming';

const customViewports = {
  pixel2XL: {
    name: 'Mobile',
    styles: {
      width: '411px',
      height: '823px',
    },
  },
  iPad: {
    name: 'Tablet',
    styles: {
      width: '768px',
      height: '1024px',
    },
  },
};

export const parameters = {
  actions: { argTypesRegex: '^on[A-Z].*' },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
  themes: {
    default: 'dark',
    list: [
      { name: 'light', class: 'light', color: '#fafafa' },
      { name: 'dark', class: 'dark', color: '#1a1a1a' },
    ],
  },
  viewport: {
    viewports: {
      ...customViewports,
    },
  },
};
