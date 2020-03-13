"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const tslib_1 = require("tslib");
const execa = require("execa");
const fs = require("fs");
const path = require("path");
const yaml = require("js-yaml");
const shelljs = require("shelljs");
const api = require("./api");
const cloud_auth_1 = require("./cloud_auth");
const config_types_1 = require("./config_types");
const exec_auth_1 = require("./exec_auth");
const file_auth_1 = require("./file_auth");
const oidc_auth_1 = require("./oidc_auth");
// fs.existsSync was removed in node 10
function fileExists(filepath) {
    try {
        fs.accessSync(filepath);
        return true;
        // tslint:disable-next-line:no-empty
    }
    catch (ignore) { }
    return false;
}
class KubeConfig {
    getContexts() {
        return this.contexts;
    }
    getClusters() {
        return this.clusters;
    }
    getUsers() {
        return this.users;
    }
    getCurrentContext() {
        return this.currentContext;
    }
    setCurrentContext(context) {
        this.currentContext = context;
    }
    getContextObject(name) {
        if (!this.contexts) {
            return null;
        }
        return findObject(this.contexts, name, 'context');
    }
    getCurrentCluster() {
        const context = this.getCurrentContextObject();
        if (!context) {
            return null;
        }
        return this.getCluster(context.cluster);
    }
    getCluster(name) {
        return findObject(this.clusters, name, 'cluster');
    }
    getCurrentUser() {
        const ctx = this.getCurrentContextObject();
        if (!ctx) {
            return null;
        }
        return this.getUser(ctx.user);
    }
    getUser(name) {
        return findObject(this.users, name, 'user');
    }
    loadFromFile(file) {
        const rootDirectory = path.dirname(file);
        this.loadFromString(fs.readFileSync(file, 'utf8'));
        this.makePathsAbsolute(rootDirectory);
    }
    applytoHTTPSOptions(opts) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const user = this.getCurrentUser();
            yield this.applyOptions(opts);
            if (user && user.username) {
                opts.auth = `${user.username}:${user.password}`;
            }
        });
    }
    applyToRequest(opts) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const cluster = this.getCurrentCluster();
            const user = this.getCurrentUser();
            yield this.applyOptions(opts);
            if (cluster && cluster.skipTLSVerify) {
                opts.strictSSL = false;
            }
            if (user && user.username) {
                opts.auth = {
                    password: user.password,
                    username: user.username,
                };
            }
        });
    }
    loadFromString(config) {
        const obj = yaml.safeLoad(config);
        if (obj.apiVersion !== 'v1') {
            throw new TypeError('unknown version: ' + obj.apiVersion);
        }
        this.clusters = config_types_1.newClusters(obj.clusters);
        this.contexts = config_types_1.newContexts(obj.contexts);
        this.users = config_types_1.newUsers(obj.users);
        this.currentContext = obj['current-context'];
    }
    loadFromOptions(options) {
        this.clusters = options.clusters;
        this.contexts = options.contexts;
        this.users = options.users;
        this.currentContext = options.currentContext;
    }
    loadFromClusterAndUser(cluster, user) {
        this.clusters = [cluster];
        this.users = [user];
        this.currentContext = 'loaded-context';
        this.contexts = [
            {
                cluster: cluster.name,
                user: user.name,
                name: this.currentContext,
            },
        ];
    }
    loadFromCluster(pathPrefix = '') {
        const host = process.env.KUBERNETES_SERVICE_HOST;
        const port = process.env.KUBERNETES_SERVICE_PORT;
        const clusterName = 'inCluster';
        const userName = 'inClusterUser';
        const contextName = 'inClusterContext';
        let scheme = 'https';
        if (port === '80' || port === '8080' || port === '8001') {
            scheme = 'http';
        }
        this.clusters = [
            {
                name: clusterName,
                caFile: `${pathPrefix}${Config.SERVICEACCOUNT_CA_PATH}`,
                server: `${scheme}://${host}:${port}`,
                skipTLSVerify: false,
            },
        ];
        this.users = [
            {
                name: userName,
                authProvider: {
                    name: 'tokenFile',
                    config: {
                        tokenFile: `${pathPrefix}${Config.SERVICEACCOUNT_TOKEN_PATH}`,
                    },
                },
            },
        ];
        this.contexts = [
            {
                cluster: clusterName,
                name: contextName,
                user: userName,
            },
        ];
        this.currentContext = contextName;
    }
    mergeConfig(config) {
        this.currentContext = config.currentContext;
        config.clusters.forEach((cluster) => {
            this.addCluster(cluster);
        });
        config.users.forEach((user) => {
            this.addUser(user);
        });
        config.contexts.forEach((ctx) => {
            this.addContext(ctx);
        });
    }
    addCluster(cluster) {
        this.clusters.forEach((c, ix) => {
            if (c.name === cluster.name) {
                throw new Error(`Duplicate cluster: ${c.name}`);
            }
        });
        this.clusters.push(cluster);
    }
    addUser(user) {
        this.users.forEach((c, ix) => {
            if (c.name === user.name) {
                throw new Error(`Duplicate user: ${c.name}`);
            }
        });
        this.users.push(user);
    }
    addContext(ctx) {
        this.contexts.forEach((c, ix) => {
            if (c.name === ctx.name) {
                throw new Error(`Duplicate context: ${c.name}`);
            }
        });
        this.contexts.push(ctx);
    }
    loadFromDefault() {
        if (process.env.KUBECONFIG && process.env.KUBECONFIG.length > 0) {
            const files = process.env.KUBECONFIG.split(path.delimiter);
            this.loadFromFile(files[0]);
            for (let i = 1; i < files.length; i++) {
                const kc = new KubeConfig();
                kc.loadFromFile(files[i]);
                this.mergeConfig(kc);
            }
            return;
        }
        const home = findHomeDir();
        if (home) {
            const config = path.join(home, '.kube', 'config');
            if (fileExists(config)) {
                this.loadFromFile(config);
                return;
            }
        }
        if (process.platform === 'win32' && shelljs.which('wsl.exe')) {
            // TODO: Handle if someome set $KUBECONFIG in wsl here...
            try {
                const result = execa.sync('wsl.exe', ['cat', shelljs.homedir() + '/.kube/config']);
                if (result.code === 0) {
                    this.loadFromString(result.stdout);
                    return;
                }
            }
            catch (err) {
                // Falling back to alternative auth
            }
        }
        if (fileExists(Config.SERVICEACCOUNT_TOKEN_PATH)) {
            this.loadFromCluster();
            return;
        }
        this.loadFromClusterAndUser({ name: 'cluster', server: 'http://localhost:8080' }, { name: 'user' });
    }
    makeApiClient(apiClientType) {
        const cluster = this.getCurrentCluster();
        if (!cluster) {
            throw new Error('No active cluster!');
        }
        const apiClient = new apiClientType(cluster.server);
        apiClient.setDefaultAuthentication(this);
        return apiClient;
    }
    makePathsAbsolute(rootDirectory) {
        this.clusters.forEach((cluster) => {
            if (cluster.caFile) {
                cluster.caFile = makeAbsolutePath(rootDirectory, cluster.caFile);
            }
        });
        this.users.forEach((user) => {
            if (user.certFile) {
                user.certFile = makeAbsolutePath(rootDirectory, user.certFile);
            }
            if (user.keyFile) {
                user.keyFile = makeAbsolutePath(rootDirectory, user.keyFile);
            }
        });
    }
    getCurrentContextObject() {
        return this.getContextObject(this.currentContext);
    }
    applyHTTPSOptions(opts) {
        const cluster = this.getCurrentCluster();
        const user = this.getCurrentUser();
        if (!user) {
            return;
        }
        if (cluster != null && cluster.skipTLSVerify) {
            opts.rejectUnauthorized = false;
        }
        const ca = cluster != null ? bufferFromFileOrString(cluster.caFile, cluster.caData) : null;
        if (ca) {
            opts.ca = ca;
        }
        const cert = bufferFromFileOrString(user.certFile, user.certData);
        if (cert) {
            opts.cert = cert;
        }
        const key = bufferFromFileOrString(user.keyFile, user.keyData);
        if (key) {
            opts.key = key;
        }
    }
    applyAuthorizationHeader(opts) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            const user = this.getCurrentUser();
            if (!user) {
                return;
            }
            const authenticator = KubeConfig.authenticators.find((elt) => {
                return elt.isAuthProvider(user);
            });
            if (!opts.headers) {
                opts.headers = [];
            }
            if (authenticator) {
                yield authenticator.applyAuthentication(user, opts);
            }
            if (user.token) {
                opts.headers.Authorization = `Bearer ${user.token}`;
            }
        });
    }
    applyOptions(opts) {
        return tslib_1.__awaiter(this, void 0, void 0, function* () {
            this.applyHTTPSOptions(opts);
            yield this.applyAuthorizationHeader(opts);
        });
    }
}
exports.KubeConfig = KubeConfig;
KubeConfig.authenticators = [
    new cloud_auth_1.CloudAuth(),
    new exec_auth_1.ExecAuth(),
    new file_auth_1.FileAuth(),
    new oidc_auth_1.OpenIDConnectAuth(),
];
// This class is deprecated and will eventually be removed.
class Config {
    static fromFile(filename) {
        return Config.apiFromFile(filename, api.CoreV1Api);
    }
    static fromCluster() {
        return Config.apiFromCluster(api.CoreV1Api);
    }
    static defaultClient() {
        return Config.apiFromDefaultClient(api.CoreV1Api);
    }
    static apiFromFile(filename, apiClientType) {
        const kc = new KubeConfig();
        kc.loadFromFile(filename);
        return kc.makeApiClient(apiClientType);
    }
    static apiFromCluster(apiClientType) {
        const kc = new KubeConfig();
        kc.loadFromCluster();
        const cluster = kc.getCurrentCluster();
        if (!cluster) {
            throw new Error('No active cluster!');
        }
        const k8sApi = new apiClientType(cluster.server);
        k8sApi.setDefaultAuthentication(kc);
        return k8sApi;
    }
    static apiFromDefaultClient(apiClientType) {
        const kc = new KubeConfig();
        kc.loadFromDefault();
        return kc.makeApiClient(apiClientType);
    }
}
exports.Config = Config;
Config.SERVICEACCOUNT_ROOT = '/var/run/secrets/kubernetes.io/serviceaccount';
Config.SERVICEACCOUNT_CA_PATH = Config.SERVICEACCOUNT_ROOT + '/ca.crt';
Config.SERVICEACCOUNT_TOKEN_PATH = Config.SERVICEACCOUNT_ROOT + '/token';
function makeAbsolutePath(root, file) {
    if (!root || path.isAbsolute(file)) {
        return file;
    }
    return path.join(root, file);
}
exports.makeAbsolutePath = makeAbsolutePath;
// This is public really only for testing.
function bufferFromFileOrString(file, data) {
    if (file) {
        return fs.readFileSync(file);
    }
    if (data) {
        return Buffer.from(data, 'base64');
    }
    return null;
}
exports.bufferFromFileOrString = bufferFromFileOrString;
// Only public for testing.
function findHomeDir() {
    if (process.env.HOME) {
        try {
            fs.accessSync(process.env.HOME);
            return process.env.HOME;
            // tslint:disable-next-line:no-empty
        }
        catch (ignore) { }
    }
    if (process.platform !== 'win32') {
        return null;
    }
    if (process.env.HOMEDRIVE && process.env.HOMEPATH) {
        const dir = path.join(process.env.HOMEDRIVE, process.env.HOMEPATH);
        try {
            fs.accessSync(dir);
            return dir;
            // tslint:disable-next-line:no-empty
        }
        catch (ignore) { }
    }
    if (process.env.USERPROFILE) {
        try {
            fs.accessSync(process.env.USERPROFILE);
            return process.env.USERPROFILE;
            // tslint:disable-next-line:no-empty
        }
        catch (ignore) { }
    }
    return null;
}
exports.findHomeDir = findHomeDir;
// Only really public for testing...
function findObject(list, name, key) {
    for (const obj of list) {
        if (obj.name === name) {
            if (obj[key]) {
                return obj[key];
            }
            return obj;
        }
    }
    return null;
}
exports.findObject = findObject;
//# sourceMappingURL=config.js.map