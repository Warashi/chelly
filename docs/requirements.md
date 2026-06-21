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
- Sets the container workdir to the current directory (configurable)
- Detects stdin/stdout TTY and adds `--interactive --tty` automatically

### `chelly config list`

Print all current effective configuration values in TOML format.

- Shows the resolved configuration after applying config file and environment variable overrides

### `chelly config get <key>`

Print the effective value of a single configuration key.

- `key`: one of `container_cmd`, `config_home`, `workdir`, `additional_mounts`, `container_setup_cmds`, `podman_options.run`
- For `additional_mounts`, `container_setup_cmds`, and `podman_options.run`, prints values as a comma-separated string

### `chelly config set <key> <value>`

Write a key-value pair to `config.toml`.

- Creates the config file and its directory if they do not exist
- For `additional_mounts`, `container_setup_cmds`, and `podman_options.run`, `value` is a comma-separated list
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
| `podman_options.run`   | `CHELLY_PODMAN_OPTIONS_RUN`    | (empty)                      | Additional `podman run` options used only when the container command is `podman` |

### Example config file

```toml
container_cmd = "podman"
workdir = "/workspace"
additional_mounts = ["/home/user/.cache:/home/user/.cache"]
container_setup_cmds = ["source /etc/profile", "mise activate"]

[podman_options]
run = ["--userns=keep-id"]
```
