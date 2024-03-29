---
weight: 1000
---

# Contributing to Zulu

Thank you so much for contributing to Zulu. We appreciate your time and help.
Here are some guidelines to help you get started.

## Code of Conduct

Be kind and respectful to the members of the community. Take time to educate
others who are seeking help. Harassment of any kind will not be tolerated.

## Questions

If you have questions regarding Zulu, feel free to raise a ticket to ask a
question.

## Filing a bug or feature

1. Before filing an issue, please check the existing issues to see if a
   similar one was already opened. If there is one already opened, feel free
   to comment on it.
2. If you believe you've found a bug, please provide detailed steps of
   reproduction, the version of Zulu and anything else you believe will be
   useful to help troubleshoot it (e.g. OS environment, environment variables,
   etc.). Also state the current behaviour vs. the expected behaviour.
3. If you'd like to see a feature or an enhancement, please open an issue with
   a clear title and description of what the feature is and why it would be
   beneficial to the project and its users.

## Submitting changes

1. CLA: Upon submitting a Pull Request (PR), contributors will be prompted to
   sign a CLA. Please sign the CLA :slightly_smiling_face:
2. You should use [`pre-commit`](https://pre-commit.com) which will ensure all
   follow required formatting.
3. Tests: If you are submitting code, please ensure you have adequate tests
   for the feature. Tests can be run via `go test ./...` or `make test`.
4. Ensure the new code is properly formatted to ensure code consistency.
   Run `make all`.

### Quick steps to contribute

1. Fork the project.
2. Clone your fork to your PC (`git clone https://github.com/your_username/zulu && cd zulu`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and run tests (`make test`)
5. Add them to staging (`git add .`)
6. Commit your changes (`git commit -m 'Add some feature'`)
7. Push to the branch (`git push origin my-new-feature`)
8. Create new pull request
