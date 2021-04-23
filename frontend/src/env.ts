type envVarNames =
  | 'REACT_APP_CROWNLABS_APISERVER_URL'
  | 'REACT_APP_CROWNLABS_OIDC_PROVIDER_URL'
  | 'REACT_APP_CROWNLABS_OIDC_CLIENT_ID'
  | 'REACT_APP_CROWNLABS_OIDC_CLIENT_SECRET'
  | 'REACT_APP_CROWNLABS_OIDC_REDIRECT_URI'
  | 'PUBLIC_URL';

type envVarObj = { [key in envVarNames]?: string };

const getEnvVar = (envVarName: envVarNames) => {
  const envVar = process.env[envVarName] ?? (window as envVarObj)[envVarName];
  if (envVar === undefined) {
    throw new Error(`ERROR: ENV VAR ${envVarName} NOT DEFINED`);
  }
  return envVar;
};

export const REACT_APP_CROWNLABS_APISERVER_URL = getEnvVar(
  'REACT_APP_CROWNLABS_APISERVER_URL'
);
export const REACT_APP_CROWNLABS_OIDC_PROVIDER_URL = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_PROVIDER_URL'
);
export const REACT_APP_CROWNLABS_OIDC_CLIENT_ID = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_CLIENT_ID'
);
export const REACT_APP_CROWNLABS_OIDC_CLIENT_SECRET = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_CLIENT_SECRET'
);
export const REACT_APP_CROWNLABS_OIDC_REDIRECT_URI = getEnvVar(
  'REACT_APP_CROWNLABS_OIDC_REDIRECT_URI'
);
export const PUBLIC_URL = getEnvVar('PUBLIC_URL');
