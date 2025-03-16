import { CodegenConfig } from '@graphql-codegen/cli';

const schemaUrl =
  (process.env.GRAPHQL_URL || 'https://graphql.preprod.crownlabs.polito.it') +
  '/schema';

const config: CodegenConfig = {
  schema: {
    [schemaUrl]: { handleAsSDL: true },
  },
  documents: ['./src/**/*.{graphql,ts}'],
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
