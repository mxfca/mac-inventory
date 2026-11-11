# mac-inventory

## Required binaries

Every commit runs through [pre-commit](https://pre-commit.com) hooks. They need these tools, all installed with Homebrew:

| Binary | What it does in the hooks | Hook | Check it is installed |
| --- | --- | --- | --- |
| `pre-commit` | Runs all the hooks on each commit | all | `pre-commit --version` |
| `gitleaks` | Finds secrets and private keys in staged changes | `gitleaks` | `gitleaks version` |
| `go` | Runs the tests | `go-test` | `go version` |
| `golangci-lint` | Lints Go code, including `go vet`, `gosec` and `gofmt` | `golangci-lint` | `golangci-lint version` |
| `govulncheck` | Finds known vulnerabilities in dependencies the code calls | `govulncheck` | `govulncheck -version` |

`govulncheck` needs network access to download the vulnerability database.

The basic hooks (`trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-merge-conflict`) need no binary: `pre-commit` builds their environment itself on the first run.

## How to install hooks

1. Install the tools:

   ```sh
   brew install pre-commit gitleaks go golangci-lint govulncheck
   ```

2. Activate the hooks in your clone, once:

   ```sh
   pre-commit install
   ```

3. Check that everything passes:

   ```sh
   pre-commit run --all-files
   ```

From now on, `git commit` runs the hooks and stops the commit when one fails.

## How to configure pre-commit

### Where the config lives

- `.pre-commit-config.yaml`: the list of hooks.
- `.golangci.yml`: the `golangci-lint` config (v2 format). It keeps the default linters, enables `gosec` and turns on the `gofmt` formatter.

### Hooks

| Hook id | What it catches |
| --- | --- |
| `trailing-whitespace` | Spaces at the end of a line. Fixes the file and fails, so you stage the fix. |
| `end-of-file-fixer` | A file without a final newline. Fixes the file and fails. |
| `check-yaml` | Invalid YAML. |
| `check-merge-conflict` | Leftover conflict markers. It runs on every commit, so a line that is exactly `=======` (for example a Markdown heading underlined with seven `=`) also fails it. |
| `gitleaks` | Secrets and private keys in staged changes. The secret value is redacted in the output. It only reads what is staged, so `pre-commit run --all-files` with nothing staged checks nothing; use `gitleaks dir .` to scan the whole tree. |
| `golangci-lint` | Lint problems, `go vet` issues, `gosec` security findings and badly formatted Go code. |
| `go-test` | Failing tests (`go test ./...`). |
| `govulncheck` | Vulnerable dependencies the code actually calls. |

### When the Go hooks run

`golangci-lint`, `go-test` and `govulncheck` check the whole module (`./...`), but only run when the commit touches a `.go`, `go.mod` or `go.sum` file. Otherwise they show as `Skipped`.

### Add or disable a hook

- **Add:** append an entry under `hooks:` in `.pre-commit-config.yaml`, either from a remote repo (with its `rev`) or as a `local` hook with `language: system`.
- **Disable:** delete or comment out its entry.

### Skip a hook for one commit

Set `SKIP` to the hook id (comma-separated for several):

```sh
SKIP=golangci-lint git commit -m "..."
```

When you are offline, `govulncheck` fails because it cannot reach its database. Skip it for that commit:

```sh
SKIP=govulncheck git commit -m "..."
```

### Update versions

- **Remote hooks** (`pre-commit-hooks`) are pinned by `rev`. Update them with:

  ```sh
  pre-commit autoupdate
  ```

- **Local tools** (`gitleaks`, `go`, `golangci-lint`, `govulncheck`) are not pinned in the repo: the hooks use whatever version Homebrew installed. Update them with `brew upgrade`.
