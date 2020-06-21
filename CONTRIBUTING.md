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

An important note on our PR management. We use a bot to perform merges into master so there is no need to do it yourself. Once a PR with the *automerge* label (added by default to every PR) has 2 approving reviews and all check are successful, it will be automatically merged into master. In case this behavior is not desired, just remove the *automerge* label or override it with the *hold* label. If your PR is behind master and it fails to merge, **don't** use the GitHub button to merge the current master in your PR. Instead, perform a rebase on your local machine or add a `/rebase` line to your PR description or in a follow-up comment, this will force the bot to do the rebase for you (if no conflicts are present).
