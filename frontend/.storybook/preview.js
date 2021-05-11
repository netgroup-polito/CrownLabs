import '../src/theming';

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
};
