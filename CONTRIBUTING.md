# Contributing to tea-eyes

Thanks for your interest. A few ground rules.

## Side-project SLA

tea-eyes is a side project. Issues and pull requests are reviewed on a
best-effort basis — there is **no SLA**. If a thread goes quiet for a few
weeks, a polite ping is welcome.

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/) are
encouraged. Common prefixes: `feat:`, `fix:`, `docs:`, `refactor:`,
`test:`, `chore:`.

## Before pushing

```sh
make test lint
```

Both must pass.

## Changelog

Pull requests that affect the MCP tool surface (add/remove/rename a tool,
change a tool parameter, change a return shape) **must** include an entry
under `## [Unreleased]` in `CHANGELOG.md`. Other PRs may add a changelog
entry but are not required to.

## Code of conduct

This project follows the
[Contributor Covenant 2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).
By participating you agree to abide by its terms.

## Licensing

By contributing, you agree that your contributions will be licensed under
the project's MIT license (see `LICENSE`).
