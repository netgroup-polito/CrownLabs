type envVarNames =
  | 'BASE_URL'
  | 'VITE_APP_CROWNLABS_GRAPHQL_URL'
  | 'VITE_APP_CROWNLABS_OIDC_CLIENT_ID'
  | 'VITE_APP_CROWNLABS_OIDC_AUTHORITY'
  | 'VITE_APP_MYDRIVE_TEMPLATE_NAME'
  | 'VITE_APP_MYDRIVE_WORKSPACE_NAME'
  | 'VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX'
  | 'VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM';

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
export const VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX = getEnvVar(
  'VITE_APP_CROWNLABS_GROUPS_CLAIM_PREFIX',
);
export const VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM = getEnvVar(
  'VITE_APP_CROWNLABS_GROUPS_ADMIN_CLAIM',
);
export const BASE_URL = getEnvVar('BASE_URL');

export const VITE_APP_MYDRIVE_TEMPLATE_NAME = getEnvVar(
  'VITE_APP_MYDRIVE_TEMPLATE_NAME',
);
export const VITE_APP_MYDRIVE_WORKSPACE_NAME = getEnvVar(
  'VITE_APP_MYDRIVE_WORKSPACE_NAME',
);
