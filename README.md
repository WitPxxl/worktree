# worktree

A tiny CLI that creates a `git worktree` for the current repo and runs a
project-specific bootstrap workflow declared in a `.worktree.yaml` file.

The original use case: spin up isolated copies of a repo for parallel AI/agent
sessions, each on its own branch, with the per-project `composer install` /
`npm install` / `docker compose build` already done for you.

## Install

```sh
go install github.com/WitPxxl/worktree@latest
```

Make sure `$(go env GOPATH)/bin` (usually `~/go/bin`) is in your `PATH`.

## Configure a repo

Drop a `.worktree.yaml` at the root of any repo you want to manage. Example
for a PHP + Node project bootstrapped through Docker Compose:

```yaml
params:
  TOKEN:
    description: "GitHub token forwarded to docker build"
    required: true

copy:
  - .env

pre_create:
  - git pull --ff-only

post_create:
  - docker compose build --build-arg TOKEN=${TOKEN}
  - docker compose run --rm php composer install
  - docker compose run --rm php npm install
  - docker compose run --rm php composer build

pre_remove:
  - docker compose down -v --remove-orphans
```

### Sections

| Section       | Where it runs | What it's for                                                  |
| ------------- | ------------- | -------------------------------------------------------------- |
| `params`      | n/a           | Declared inputs. `required: true` and/or `default:` supported. |
| `copy`        | n/a           | Files copied from source repo into the new worktree.           |
| `pre_create`  | source repo   | Commands run **before** `git worktree add` (e.g. `git pull`).  |
| `post_create` | new worktree  | Commands run **after** the worktree is created.                |
| `pre_remove`  | worktree      | Commands run before `git worktree remove` (failures warn-only).|

### Variables

Inside any string in the config you can reference:

- `${BRANCH}` – the branch name passed to `--branch`
- `${TARGET}` – the absolute worktree path (`~/project/worktree/<branch>`)
- `${SOURCE}` – the absolute source repo path
- Any user-declared param (e.g. `${TOKEN}`)
- Any environment variable (fallback if no param of that name exists)

Branch names containing `/` are normalized for the directory name only
(`feature/foo` → `~/project/worktree/feature_foo`); the git branch keeps its
original name.

## Usage

Run from the root of a repo that contains a `.worktree.yaml`.

Create:

```sh
worktree create --branch feature/foo --param TOKEN=ghp_xxx
# aliases: add, new
```

Remove (runs `pre_remove`, then `git worktree remove`):

```sh
worktree rm --branch feature/foo
worktree rm --branch feature/foo --force   # also discard local changes
# aliases: remove, delete
```

## How it works

1. Loads `.worktree.yaml` from the current directory.
2. Resolves params (defaults from config, overridden by `-p KEY=VAL`).
3. Runs `pre_create` commands in the source repo.
4. Creates the worktree:
   - If the branch already exists locally → `git worktree add <target> <branch>` (checks it out as-is).
   - Otherwise → `git worktree add -b <branch> <target>` (creates it from the current `HEAD`).
5. Copies files declared in `copy` into the worktree.
6. Runs `post_create` commands inside the worktree.

Commands execute through `sh -c`, so pipes, redirects and environment
variables work as expected.

## Project layout

```
cmd/                 # cobra subcommands (create, rm)
internal/config/     # .worktree.yaml loader + param resolution
internal/strategy/   # Step interface + RunStep / CopyStep (Strategy pattern)
internal/runner/     # orchestrates the create/remove flows
internal/shell/      # exec helper that streams stdio
```

Adding a new kind of step (e.g. `GitStep`, `TemplateStep`) means writing a
struct that implements `strategy.Step` and wiring it into the runner; the
existing pipeline stays untouched.

## Development

```sh
go build ./...
go vet ./...
go test ./...
```

## License

MIT – see [LICENSE](./LICENSE).
