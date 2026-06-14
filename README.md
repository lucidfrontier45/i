# go-project-template
A Project Template for Golang

## Linters and Formatters

[golangci-lint](https://golangci-lint.run/) is used for both linting and formatting. Install it by following the instructions on their website.

Then you can run it with:

```bash
# linter and auto-fix
golangci-lint run --fix
# formatter
golangci-lint fmt
```

## VS Code Setup

golangci-lint integrated VS Code settings are provided in `.vscode/settings.json`. Make sure to have the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) installed.