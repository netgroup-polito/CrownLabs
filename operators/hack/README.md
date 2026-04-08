# Copyright and License information

## How to update the Copyright year

- Edit the `./boilerplate.go.txt` file
- Edit the `CrownLabs/operators/LICENSE` file
- Run `make generate` and `make fmt` in the `CrownLabs/operators` folder

## How the correct year is enforced

The LICENSE file, which is shown on the GitHub project page, should always be up to date, but this is not enforced.

When auto-generating the CRD go files, `make generate` uses the `./boilerplate.go.txt` as header.

The GitHub Action that handles linting executes `golangci-lint` with specific configuration, that is the `.golangci.yml` file. The `goheader` linter checks if the file starts with the license, and that it includes the current year.

When fixing linting issues in local with `make fmt`, it also calls `addlicense` which adds the license as header with the current year to each .go file only if it doesn't have one. After this, `golangci-lint` is called to fix all linting issues, including possible copyright years that have not been updated.