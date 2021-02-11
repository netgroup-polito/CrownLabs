# Using LiqoDash

Given the weird intrinsic nature of git submodules, this document defines some common guidelines to develop the dashboard of Crownlabs using this approach.

## Scripts

- `pull-submodule`: pull all the files from of the submodule (think of it more as a "clone all the files needed to run the dashboard locally in the _/dashboard_ folder)
- `update-version`: point the submodule to the latest commit on master of LiqoDash. **Note** this command adds a commit to the history of this repo, so, if the main branch of the dashboard get updated multiple times during the life of your branch, avoid making more than one of this commit per PR, just squash them together.
