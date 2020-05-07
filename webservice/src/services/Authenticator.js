import { UserManager } from 'oidc-client';

/**
 * Api class to manage authN
 *
 */

export default class Authenticator {
  constructor() {
    if (window.OIDC_CLIENT_ID === undefined) {
      window.OIDC_CLIENT_ID = OIDC_CLIENT_ID;
    }

    if (window.OIDC_PROVIDER_URL === undefined) {
      window.OIDC_PROVIDER_URL = OIDC_PROVIDER_URL;
    }

    if (window.OIDC_CLIENT_SECRET === undefined) {
      window.OIDC_CLIENT_SECRET = OIDC_CLIENT_SECRET;
    }

    if (window.OIDC_REDIRECT_URI === undefined) {
      window.OIDC_REDIRECT_URI = OIDC_REDIRECT_URI;
    }

    this.manager = new UserManager({
      automaticSilentRenew: true,
      response_type: 'code',
      filterProtocolClaims: true,
      scope: 'openid ',
      loadUserInfo: true,
      client_secret: window.OIDC_CLIENT_SECRET,
      authority: window.OIDC_PROVIDER_URL,
      client_id: window.OIDC_CLIENT_ID,
      redirect_uri: `${window.OIDC_REDIRECT_URI}/callback`,
      post_logout_redirect_uri: `${window.OIDC_REDIRECT_URI}/logout`
    });
    this.login = this.login.bind(this);
    this.logout = this.logout.bind(this);
    this.completeLogin = this.completeLogin.bind(this);
  }

  /**
   * Function to perform the login.
   * It will automatically redirect you
   * @return {Promise<void>}
   */
  login() {
    return this.manager.signinRedirect();
  }

  /**
   * Function to process response from the authN endpoint
   * @return {Promise<User>}
   */
  completeLogin() {
    return this.manager.signinRedirectCallback();
  }

  /**
   * Function to logout
   * @return {Promise<void>}
   */
  logout() {
    return this.manager.signoutRedirect();
  }
}
