type envVarNames =
  | 'BASE_URL'
  | 'VITE_APP_CROWNLABS_GRAPHQL_URL'
  | 'VITE_APP_CROWNLABS_OIDC_CLIENT_ID'
  | 'VITE_APP_CROWNLABS_OIDC_AUTHORITY';

type envVarObj = { [key in envVarNames]?: string };

const getEnvVar = (envVarName: envVarNames): string => {
  const envVar: string =
    import.meta.env[envVarName] ?? (window as envVarObj)[envVarName];
  if (envVar === undefined) {
    throw new Error(`ERROR: ENV VAR ${envVarName} NOT DEFINED`);
  }
  return envVar;
};

export const VITE_APP_CROWNLABS_OIDC_CLIENT_ID = getEnvVar(
  'VITE_APP_CROWNLABS_OIDC_CLIENT_ID',
);
export const VITE_APP_CROWNLABS_GRAPHQL_URL = getEnvVar(
  'VITE_APP_CROWNLABS_GRAPHQL_URL',
);
export const VITE_APP_CROWNLABS_OIDC_AUTHORITY = getEnvVar(
  'VITE_APP_CROWNLABS_OIDC_AUTHORITY',
);
export const BASE_URL = getEnvVar('BASE_URL');
