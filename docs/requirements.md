# Requirements

## Overview

chelly is a CLI tool for managing a personal container-based development environment.

## Commands

### `chelly build [--no-cache]`

Build the chelly container image.

- Reads the build context from `$XDG_CONFIG_HOME/chelly` (default: `$HOME/.config/chelly`)
- Auto-detects the container runtime: tries `container`, `podman`, `docker` in order
- `--no-cache`: disables the build cache

### `chelly run [command [args...]]`

Run a command inside the chelly container.

- Mounts the current directory at the same path inside the container
- When run inside a linked Git worktree, also mounts the parent directory of the Git common dir at the same path inside the container
- If Git metadata cannot be resolved, continues without the linked worktree mount
- Sets the container workdir to the current directory (configurable)
- Inherits configured environment variables from the `chelly run` process into the container
- Detects stdin/stdout TTY and adds `--interactive --tty` automatically
- Replaces the `chelly` process with the container runtime, so the runtime's exit status, signal handling, TTY behavior, and process ownership pass through to the caller

### `chelly config list`

Print all current effective configuration values in TOML format.

- Shows the resolved configuration after applying config file and environment variable overrides

### `chelly config get <key>`

Print the effective value of a single configuration key.

- `key`: one of `container_cmd`, `config_home`, `workdir`, `additional_mounts`, `container_setup_cmds`, `inherit_env`, or `runtime_options.<runtime>.<subcommand>`
- For `additional_mounts`, `container_setup_cmds`, `inherit_env`, and `runtime_options.<runtime>.<subcommand>`, prints values as a comma-separated string

### `chelly config set <key> <value>`

Write a key-value pair to `config.toml`.

- Creates the config file and its directory if they do not exist
- For `additional_mounts`, `container_setup_cmds`, `inherit_env`, and `runtime_options.<runtime>.<subcommand>`, `value` is a comma-separated list
- Environment variable overrides still take precedence when reading back via `list`/`get`

## Configuration

Config file: `$XDG_CONFIG_HOME/chelly/config.toml` (default: `$HOME/.config/chelly/config.toml`)

Environment variables override config file values.

| TOML key               | Environment variable           | Default                      | Description                                          |
|------------------------|--------------------------------|------------------------------|------------------------------------------------------|
| `container_cmd`        | `CHELLY_CONTAINER_CMD`         | auto-detect                  | Container runtime command (`container`/`podman`/`docker`) |
| `config_home`          | `CHELLY_CONFIG_HOME`           | `$XDG_CONFIG_HOME/chelly`    | Build context directory                              |
| `workdir`              | `CHELLY_WORKDIR`               | current directory            | Working directory inside the container               |
| `additional_mounts`    | `CHELLY_ADDITIONAL_MOUNTS`     | (empty)                      | Additional volume mounts (`host:container` format, comma-separated for env var) |
| `container_setup_cmds` | `CHELLY_CONTAINER_SETUP_CMDS`  | (empty)                      | Shell commands to run inside the container before the main command; multiple commands run in parallel with stdout redirected to stderr |
| `inherit_env`          | `CHELLY_INHERIT_ENV`           | (empty)                      | Environment variable names inherited from `chelly run` into the container |
| `runtime_options`      | `CHELLY_RUNTIME_OPTIONS_<RUNTIME>_<SUBCOMMAND>` | (empty)      | Extra arguments for a specific container runtime and subcommand (see below) |

### `runtime_options`

`runtime_options` is a list of `{runtime, subcommand, args}` entries. Each entry's
`args` are inserted into the corresponding container runtime invocation only when
`container_cmd` resolves to a basename matching `runtime`, and the chelly command
being run maps to `subcommand` (`run` for `chelly run`, `build` for `chelly build`).

- `chelly config get`/`set` address a single entry via the key
  `runtime_options.<runtime>.<subcommand>` (e.g. `runtime_options.podman.run`),
  with `value` as a comma-separated list of arguments.
- The environment variable `CHELLY_RUNTIME_OPTIONS_<RUNTIME>_<SUBCOMMAND>`
  (uppercased runtime and subcommand) overrides the corresponding config file
  entry, e.g. `CHELLY_RUNTIME_OPTIONS_PODMAN_RUN`.
- `subcommand` must be one of `run` or `build`.
- Duplicate entries for the same `runtime`/`subcommand` pair are rejected when
  `run`/`build` validate the loaded configuration.

### Example config file

```toml
container_cmd = "podman"
workdir = "/workspace"
additional_mounts = ["/home/user/.cache:/home/user/.cache"]
container_setup_cmds = ["source /etc/profile", "mise activate"]
inherit_env = ["SSH_AUTH_SOCK", "GITHUB_TOKEN"]

[[runtime_options]]
runtime = "podman"
subcommand = "run"
args = ["--userns=keep-id"]
```
