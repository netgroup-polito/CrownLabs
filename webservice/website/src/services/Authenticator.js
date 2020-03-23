import {UserManager} from 'oidc-client';

/**
 * Api class to manage authN
 * (use the commented field if ImplicitFlow disabled and if the server hosting the website allows CORS)
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
        this.completeLogout = this.completeLogout.bind(this);
    }

    login() {
        return this.manager.signinRedirect({});
    }

    completeLogin() {
        return this.manager.signinRedirectCallback();
    }

    logout() {
        return this.manager.signoutRedirect();
    }

    completeLogout() {
        return this.manager.signoutCallback();
    }
}