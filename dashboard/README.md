# Using LiqoDash

Given the weird nature of git submodules, this document defines some common guidelines to develop the dashboard of CrownLabs using this approach.

## Environment setup

Before even cloning the Crownlabs repo to start developing the dashboard locally, you should set the following environment variables (e.g. in your `.bashrc` file or equivalent):

```bash
export APISERVER_URL="https://apiserver.crownlabs.polito.it"
export OIDC_PROVIDER_URL="https://auth.crownlabs.polito.it/auth/realms/crownlabs"
export OIDC_CLIENT_ID="k8s"
export OIDC_REDIRECT_URI="http://localhost:8000"
export OIDC_CLIENT_SECRET=
```

You should also ask the maintainers the value for the variable `OIDC_CLIENT_SECRET`.

## NPM Scripts

To help develop the dashboard locally, these [following scripts](package.json) perform some useful actions:

- `pull-submodule`: pull all the files from of the submodule (like a "clone all the files needed to run the dashboard locally in the _/dashboard_ folder).
- `setup`: prepares the local environment to develop the dashboard using the current commit of the submodule (not the latest on the main branch of LiqoDash).
- `update-version`: updates the submodule to the latest commit on master of LiqoDash. **_Note_** _that this command adds a commit to the history of this repo_, so, if the main branch of the dashboard gets updated multiple times during the life of your branch, avoid making more than one of this commit per PR, just squash them together.
- `copy-changes`: copy back the files changed when developing in the submodule. This script is needed because to view real-time changes in the dashboard the custom files need to be inside the submodule, so after having performed the necessary changes they need to be brought back for commit.

## Step-by-step process

The ideal sequence of steps should be the following:

```bash
# add custom env variables

# if the dashboard core needs to be updated
npm run update-version

npm run setup
npm start
# develop additional changes
npm test
npm run copy-changes
# commit
```
