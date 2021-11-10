# CrownLabs frontend

This file describes the structure and general coding guidelines of the CrownLabs frontend.

## Setup

Before starting set the necessary environment variables either using a `.env` file (preferred on windows) or by defining them on your local machine. To setup the repo, use the following commands ([yarn](https://yarnpkg.com/cli/install) use is mandatory).

```bash
# If you don't have yarn installed
npm install -g yarn

# install necessary packages
yarn install

# To run the app locally
yarn start

# To run the storybook locally
yarn storybook

# To build the app locally
yarn build-app

# To build the storybook locally
yarn build-storybook
```

After the setup is complete, if you start the app locally and see the apiserver url on the home page everything is working fine.

## Structure

Our frontend is a [React](https://it.reactjs.org/) application. We use [antd](https://ant.design/) as the main component library and [Tailwind](https://tailwindcss.com/) utilities to handle specific css scenarios (padding and margin). We also use [Storybook](https://storybook.js.org/) to ease teamwork. We chose to use [Typescript](https://www.typescriptlang.org/) to have a bettere development experience.

The application is made to be deployed using docker and can be hosted on a custom subroute. The application takes some environment variables, each needs to have the `REACT_APP_CROWNLABS` prefix. In order to define environment variables at container run-time they need to be defined on the `window` object in the `public/config.js` file.

## CI checks

We use ESLint to enforce linting and code quality. We use Prettier to have a uniform code style. These checks are enforced by a GitHub Action run on PRs. Also locally a pre-commit hook written with [husky](https://typicode.github.io/husky/#/) uses [lint-staged](https://github.com/okonet/lint-staged) to format and check files in the staging area before each commit.

## Guidelines

We define some coding guidelines to ease the collaboration in our team and to ease the review process.

### Dir structure

The components are in `src/components`. The folder has subdirs, one for each page of the app, plus `misc` for the miscellaneous UI elements (those common between all components) and a `util` dir for custom UI elements used multiple times across the code (custom dialogs, inputs, etc...).

Each component needs to have its own folder with the following structure, e.g. for an `Example` component:

- `Example.tsx` - component definition
- `Example.stories.ts` - storybook stories to show component usage
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

#### Storybook

Each storybook file needs to have the following structure:

- import declarations
- default export containing necessary story info (title, component, argTypes,...)
- default component args declaration (use `someKeysOf` type function)
- template declaration
- component stories

Additionally:

- Include in the `title` of the default export the path to create a hierarchy for your components. The path should be `Components/${subComponentsFolders}/${ComponentName}`
- If you component needs external contexts and/or side-loaded data. You can use [decorators](https://storybook.js.org/docs/react/writing-stories/decorators) in you stories to provide them

## Useful links

- [Useful answer for deploying react app on subroute](https://stackoverflow.com/a/58508562/11143279)
