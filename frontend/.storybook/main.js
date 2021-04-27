const path = require('path');

module.exports = {
  stories: ['../src/**/*.stories.mdx', '../src/**/*.stories.@(js|jsx|ts|tsx)'],
  addons: [
    '@storybook/addon-links',
    '@storybook/addon-essentials',
    {
      // allows to theme antd in the storybook
      name: '@storybook/preset-ant-design',
      options: {
        lessOptions: {
          modifyVars: {
            '@primary-color': '#1c7afd',
            '@secondary-color': '#FF7C11',
          },
        },
      },
    },
    {
      // allows to use less in the storybook in create-react-app (required for antd theming)
      name: '@storybook/preset-create-react-app',
      options: {
        craOverrides: {
          fileLoaderExcludes: ['less'],
        },
      },
    },
  ],
  // allows to use tailwind utilities in the storybook
  webpackFinal: async config => {
    config.module.rules.push({
      test: /\.css$/,
      use: [
        {
          loader: 'postcss-loader',
          options: {
            ident: 'postcss',
            plugins: [require('tailwindcss'), require('autoprefixer')],
          },
        },
      ],
      include: path.resolve(__dirname, '../'),
    });
    return config;
  },
};
