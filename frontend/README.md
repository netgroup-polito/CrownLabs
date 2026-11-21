# CrownLabs frontend

This file describes the structure and general coding guidelines of the CrownLabs frontend.

## Frontend Logic

### Authentication

The Frontend uses Keycloak to authenticate the users. After authentication and the OIDC token callback, the load balancer returns `/app`. From there, the user can access and manage CrownLabs instances.

```mermaid
sequenceDiagram
    actor User
    participant LB as Gateway + Load Balancer
    participant Frontend as Frontend Page<br/>/app
    participant Keycloak as Keycloak<br/>OIDC Authentication
    participant Instance as Instance

    User->>LB: Open crownlabs.polito.it/app

    opt User is not authenticated
        LB->>Keycloak: Redirect to login
        Keycloak-->>User: Show login page
        User->>Keycloak: Submit credentials
        Keycloak-->>LB: OIDC callback with authorization code
        LB->>Keycloak: Exchange the authorization code for tokens
        Keycloak-->>LB: Return ID and access tokens
        LB-->>User: Send token and cookies
        Note over User,Keycloak: Authentication successful
    end

    LB->>Frontend: Return /app
    Frontend-->>User: Display /app

    Note over User,Instance: The user can manage and access instances

```

### GraphQL

Through the `/graph` route, the end user can send queries and mutations over HTTP. The Gateway forwards them to QLKube, which accesses the Kubernetes API server.

```mermaid
sequenceDiagram
    actor User
    participant LB as Gateway + Load Balancer
    participant GraphQL as QLKube<br/>/graph
    participant APIServer as Kubernetes API Server

    alt Query or mutation
        User->>LB: Send HTTP query or mutation with Bearer token
        LB->>GraphQL: Route /graph to qlkube
        GraphQL->>APIServer: Query or update Kubernetes resources
        APIServer-->>GraphQL: Return resource data or mutation result
        GraphQL-->>LB: Return the GraphQL response
        LB-->>User: Forward the response
    end
```

### Instance access

The frontend provides three ways to access an instance:

- **Interactive remote desktop:** the browser opens the graphical environment such as VSCode or the GNOME of the Distro.
- **Direct service route:** the browser opens an application or service exposed through an instance-specific route.
- **Web SSH:** the browser opens a WebSocket connection; the Bastion service then provides SSH access to the selected instance.

```mermaid
sequenceDiagram
    actor User
    participant LB as Gateway + Load Balancer
    participant Frontend as Frontend Page<br/>/app
    participant Instance as Instance Environment
    participant Bastion as Bastion WebSSH

    alt Interactive remote desktop
        User->>Frontend: Choose the interactive remote desktop
        Frontend->>LB: Open the instance-specific desktop route
        LB->>Instance: Rewrite and route the request
        Instance-->>LB: Return the remote desktop
        LB-->>User: Display the instance desktop
    else Direct service route
        User->>Frontend: Choose a direct service exposed by the instance
        Frontend->>LB: Open the instance-specific service route
        LB->>Instance: Forward the service request
        Instance-->>LB: Return the exposed service
        LB-->>User: Display the direct service
    else Web SSH
        User->>Frontend: Choose Web SSH
        Frontend->>LB: Open WebSocket /webssh
        LB->>Bastion: Route the WebSocket connection
        Bastion->>Instance: Establish SSH access
        Instance-->>Bastion: Open the terminal session
        Bastion-->>User: Display the Web SSH terminal
    end
```

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

## Useful links

- [Useful answer for deploying react app on subroute](https://stackoverflow.com/questions/49429906/how-should-i-configure-create-react-app-to-serve-app-from-subdirectory/58508562#58508562)
