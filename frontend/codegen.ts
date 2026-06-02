import type { CodegenConfig } from '@graphql-codegen/cli';

const schemaUrl =
  (process.env.GRAPHQL_URL ||
    process.env.VITE_APP_CROWNLABS_GRAPHQL_URL ||
    'https://graphql.ng.crownlabs.polito.it') +
  '/schema';

const config: CodegenConfig = {
  schema: {
    [schemaUrl]: { handleAsSDL: true },
  },
  documents: ['./src/**/*.ts', './src/**/*.graphql'],
  generates: {
    './src/generated-types.tsx': {
      config: {
        preResolveTypes: true,
      },
      plugins: [
        'typescript',
        'typescript-operations',
        'typescript-react-apollo',
      ],
    },
  },
  config: {},
};

export default config;
