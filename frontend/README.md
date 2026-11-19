# CrownLabs frontend

This file describes the structure and general coding guidelines of the CrownLabs frontend.

## Setup

Before starting set the necessary environment variables either using a `.env` file (preferred on windows) or by defining them on your local machine. To setup the repo, use the following commands.

```bash
# install necessary packages
npm install

# To run the app locally
npm start

# To build the app locally
npm build-app

```

After the setup is complete, if you start the app locally and see the apiserver url on the home page everything is working fine.

## Structure

Our frontend is a [React](https://it.reactjs.org/) application. We use [antd](https://ant.design/) as the main component library and [Tailwind](https://tailwindcss.com/) utilities to handle specific css scenarios (padding and margin). We chose to use [Typescript](https://www.typescriptlang.org/) to have a bettere development experience.

The application is made to be deployed using docker and can be hosted on a custom subroute. The application takes some environment variables, each needs to have the `REACT_APP_CROWNLABS` prefix. In order to define environment variables at container run-time they need to be defined on the `window` object in the `public/config.js` file.

## CI checks

We use ESLint to enforce linting and code quality. We use Prettier to have a uniform code style. These checks are enforced by a GitHub Action run on PRs. Also locally a pre-commit hook written with [husky](https://typicode.github.io/husky/#/) uses [lint-staged](https://github.com/okonet/lint-staged) to format and check files in the staging area before each commit.

Be sure to add the following to `~/.config/husky/init.sh` if you're using `nvm` (which is strongly advised)

```sh
export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
```

## Guidelines

We define some coding guidelines to ease the collaboration in our team and to ease the review process.

### Dir structure

The components are in `src/components`. The folder has subdirs, one for each page of the app, plus `misc` for the miscellaneous UI elements (those common between all components) and a `util` dir for custom UI elements used multiple times across the code (custom dialogs, inputs, etc...).

Each component needs to have its own folder with the following structure, e.g. for an `Example` component:

- `Example.tsx` - component definition
- `Exmaple.css` - css specific code for component
- `index.ts` - utility file to shorten component import statement in other files

#### Component

Each component file needs to host a single component and needs to have the following structure:

- import declarations (when you can use non-default imports use them to reduce bundle size)
- general constant declarations for the components (if needed)
- component props interface with the name `${ComponentName}Props` (if needed)
- functional component declaration

Additionally:

- if a task (like manipulating dynamic elements, animations, resizing) can be **_easily_** implemented using tailwind/css follow that implementation instead of relying on React/TS
- have jsx only inside the return statement of the component

  - Invalid:

  ```ts
  const Example: FC<IExampleProps> = ({ ...props }) => {
    const { text, disabled, onClick, size, specialCSS } = props;
    // do not have jsx outside the return
    const content = <h5 className={specialCSS ? 'rainbow-text' : ''}>{text}</h5>

    return (
    <Button
        disabled={disabled}
        size={size}
        onClick={onClick}
        type="primary"
        className="p-10"
    >
        {content}
    </Button>
  );
  ```

  - Accepted:

  ```tsx
  const Example: FC<IExampleProps> = ({ ...props }) => {
    const { text, disabled, onClick, size, specialCSS } = props;

    return (
    <Button
        disabled={disabled}
        size={size}
        onClick={onClick}
        type="primary"
        className="p-10"
    >
        <h5 className={specialCSS ? 'rainbow-text' : ''}>{text}</h5>
    </Button>
  );
  ```

## Environment IP Resolution & SSH Access

Instance IP addresses and SSH access instructions are handled through a joint backend-frontend workflow:

### Backend IP Resolution
1. **IP Field Population**: The Instance Operator resolves the internal VMI or container Pod IP address (`GetEnvironmentResolvedIP`) and populates the `ip` field under `status.environments[]` in the `Instance` Custom Resource status.

### Frontend SSH Command Generation
1. **Direct SSH Connection**: In the SSH modal (`SSHModalContent`), the frontend reads `env.ip` and dynamically constructs the copyable SSH command:
   ```bash
   ssh -J bastion@${VITE_APP_CROWNLABS_BASTION_HOST} crownlabs@${env.ip}
   ```
2. **WebSSH Connection**: For in-browser SSH sessions, the frontend builds the URL `/instance/${namespace}/${name}/${envName}/ssh`, launching the WebSSH terminal component (`SSHTerminal`) which connects via WebSockets to the Bastion service exposed on `/webssh`.

## Kubernetes cluster templates

Workspace managers can select **Kubernetes Cluster** in the template editor,
choose a ClusterFlavor, and configure 1–10 workers with CPU/RAM **per worker**.
Users instantiate these templates through the existing **Create** action.

The flavor selector reads the ImageList configured by
`VITE_APP_CROWNLABS_IMAGELIST_CLUSTER_FLAVORS` (default:
`local-clusterflavors`). For Helm, set
`configuration.imagelist.clusterFlavors` in the frontend chart, or
`frontend.configuration.imagelist.clusterFlavors` in the umbrella chart.
The ImageList updater and Instance Operator must use the same central
ClusterFlavor namespace. Flavor names are stored verbatim in the environment's
`image` field; neither the catalog namespace nor a version tag is appended.

Install the updated Template CRD and restart qlkube before deploying this
frontend. Its GraphQL schema must expose `EnvironmentType.Cluster` and
`cluster.workersCount`. Regenerate types against that backend with:

```bash
GRAPHQL_URL=http://localhost:8085/graph npm run generate
npm test
npm run build-app
```

Template and instance resource summaries, including quota prechecks, count CPU
and RAM for all workers. These are worker-only estimates: the flavor's
control-plane and storage allocations are not exposed by the catalog, and
Kubernetes remains authoritative for resource quota enforcement.

Cluster GUI/SSH/kubeconfig access and stop/restart are not implemented by the
current backend. The editor hides unsupported environment controls; templates
containing a cluster disable inactivity cleanup, public port exposure, and
node selection, while preserving maximum-lifetime cleanup. Instance power
controls are also unavailable for mixed VM/cluster templates. Supported VM
connections remain available individually in mixed instances. Cluster
instances can still be monitored and deleted.

## Useful links

- [Useful answer for deploying react app on subroute](https://stackoverflow.com/questions/49429906/how-should-i-configure-create-react-app-to-serve-app-from-subdirectory/58508562#58508562)
