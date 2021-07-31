type envVarNames =
  | 'REACT_APP_CROWNLABS_GRAPHQL_URL'
  | 'REACT_APP_CROWNLABS_OIDC_PROVIDER_URL'
  | 'REACT_APP_CROWNLABS_OIDC_CLIENT_ID'
  | 'REACT_APP_CROWNLABS_OIDC_REALM'
  | 'PUBLIC_URL';

type envVarObj = { [key in envVarNames]?: string };

const getEnvVar = (envVarName: envVarNames) => {
  const envVar = process.env[envVarName] ?? (window as envVarObj)[envVarName];
  if (envVar === undefined) {
    throw new Error(`ERROR: ENV VAR ${envVarName} NOT DEFINED`);
  }
  return envVar;
};

export const REACT_APP_CROWNLABS_OIDC_PROVIDER_URL = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_PROVIDER_URL'
);
export const REACT_APP_CROWNLABS_OIDC_CLIENT_ID = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_CLIENT_ID'
);
export const REACT_APP_CROWNLABS_GRAPHQL_URL = getEnvVar(
  'REACT_APP_CROWNLABS_GRAPHQL_URL'
);
export const REACT_APP_CROWNLABS_OIDC_REALM = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_REALM'
);
export const PUBLIC_URL = getEnvVar('PUBLIC_URL');
