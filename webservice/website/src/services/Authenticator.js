import { Log, UserManager } from 'oidc-client';

export default class Authenticator {
    constructor() {
        this.manager = new UserManager({
            authority: OIDC_PROVIDER_URL,
            client_id: OIDC_CLIENT_ID,
            redirect_uri: OIDC_REDIRECT_URI + "/callback",
            post_logout_redirect_uri: OIDC_REDIRECT_URI + '/logout',
            response_type: 'id_token',
            scope: 'openid',
            loadUserInfo: true
        });
        this.login = this.login.bind(this);
        this.logout = this.logout.bind(this);
        this.completeLogin = this.completeLogin.bind(this);
        this.completeLogout = this.completeLogout.bind(this);
        Log.logger = console;
        Log.level = Log.INFO;
    }
    login() {
        this.manager.signinRedirect();
    }
    completeLogin() {
        this.manager.signinRedirectCallback();
    }
    logout() {
        this.manager.signoutRedirect();
    }
    completeLogout() {
        this.manager.signoutCallback();
    }
}