# golangci-lint-strict

Sane strict configuration for [golangci-lint](https://github.com/golangci/golangci-lint).

Enabled all linters/rules except ones which will likely hurt readability (e.g. because they
recommend some premature optimization) or too annoying in real projects without enough value.

This is recommended config for a new projects, but existing projects may require some fair
amount of work to start using this config.
