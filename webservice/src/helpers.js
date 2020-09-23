import Toastr from 'toastr';

/**
 * Function to parse a JWT token
 * @param token the token received by keycloak
 * @returns {any} the decrypted token as a JSON object
 */
export function parseJWTtoken(token) {
  const base64 = token.split('.')[1].replace(/-/g, '+').replace(/_/g, '/');
  return JSON.parse(
    decodeURIComponent(
      atob(base64)
        .split('')
        .map(c => {
          return `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`;
        })
        .join('')
    )
  );
}

/**
 * Function to check the token, but encoded and decoded
 * @param parsed the decoded one
 * @return {boolean} true or false whether the token satisfies the constraints
 */
export function checkToken(parsed) {
  if (!parsed.groups || !parsed.groups.length) {
    Toastr.error('You do not belong to any namespace to see laboratories');
    return false;
  }
  if (!parsed.namespace || !parsed.namespace[0]) {
    Toastr.error(
      'You do not have your own namespace where to run laboratories'
    );
    return false;
  }
  return true;
}
