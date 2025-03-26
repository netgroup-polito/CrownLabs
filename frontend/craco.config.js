const CracoLessPlugin = require('craco-less');

module.exports = {
  style: {
    postcss: {
      plugins: [require('tailwindcss')],
    },
  },
  plugins: [
    {
      plugin: CracoLessPlugin,
      options: {
        lessLoaderOptions: {
          lessOptions: {
            modifyVars: {
              '@primary-color': '#1c7afd',
              '@secondary-color': '#FF7C11',
            },
            javascriptEnabled: true,
          },
        },
      },
    },
  ],
};
