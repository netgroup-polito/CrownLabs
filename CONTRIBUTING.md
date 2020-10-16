# Contributing guidelines for CrownLabs

## Team structure

The CrownLabs development team is distributed among many Specific Interest Groups (SIG):

| sig name         | area of work                                              |
| ---------------- | --------------------------------------------------------- |
| api              | APIs design and development of backend applications       |
| auth             | Users' authentication and authorization aspects           |
| community        | External relations and open-source goals                  |
| devops           | Design and delivery of CI/CD pipelines                    |
| operations       | Kubernetes cluster and infrastructural services operation |
| ui               | Front-end logic and design                                |
| user-environment | Definition and generation of the end-user environments    |

Before contributing to the project try to understand the target area of work of your modifications. You are not restricted to work on a single area but encouraged to go beyond your current knowledge and acquire more skills.

## Coding guidelines

When creating PRs and issues follow the repo's guidelines. Use meaningful messages for your commits to have a clean and clear code history. CrownLabs is opened to new developers who want to grow and experiment, maintaining always a high level of quality. Feel free to enter our [Slack workspace](https://crown-team-group.slack.com/) and meet everyone else.

## PR merging guidelines

Before a PR can be merged to the master branch, two approving reviews and all successful checks are required. Then, once a PR is ready to be merged, this action can be performed by any contributor with write access to the repository, either manually or issuing a `/merge` comment to the PR itself.

If your PR is behind master and it fails to merge, **don't** use the GitHub button to merge the current master in your PR. Instead, perform a rebase on your local machine or issue a `/rebase` comment to the PR itself. This will force the bot to do the rebase for you (if no conflicts are present).
