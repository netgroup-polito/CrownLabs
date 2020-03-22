import {UserManager} from 'oidc-client';

/**
 * Api class to manage authN
 *
 */
export default class Authenticator {
    constructor() {
        this.manager = new UserManager({
            authority: OIDC_PROVIDER_URL,
            client_id: OIDC_CLIENT_ID,
            redirect_uri: OIDC_REDIRECT_URI + "/callback",
            automaticSilentRenew: true,
            post_logout_redirect_uri: OIDC_REDIRECT_URI + '/',
            response_type: 'code',
            filterProtocolClaims: true,
            scope: 'openid ',
            loadUserInfo: true,
            client_secret: OIDC_CLIENT_SECRET
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