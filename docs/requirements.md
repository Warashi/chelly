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

### `chelly runit [command [args...]]`

Start an interactive shell (or run a command) in the chelly container.

- Same as `run`, but always forces `--interactive --tty`
- Defaults to `sh` if no command is given

## Configuration

Config file: `$XDG_CONFIG_HOME/chelly/config.toml` (default: `$HOME/.config/chelly/config.toml`)

Environment variables override config file values.

| TOML key              | Environment variable          | Default                      | Description                                          |
|-----------------------|-------------------------------|------------------------------|------------------------------------------------------|
| `container_cmd`       | `CHELLY_CONTAINER_CMD`        | auto-detect                  | Container runtime command (`container`/`podman`/`docker`) |
| `config_home`         | `CHELLY_CONFIG_HOME`          | `$XDG_CONFIG_HOME/chelly`    | Build context directory                              |
| `workdir`             | `CHELLY_WORKDIR`              | current directory            | Working directory inside the container               |
| `additional_mounts`   | `CHELLY_ADDITIONAL_MOUNTS`    | (empty)                      | Additional volume mounts (`host:container` format, comma-separated for env var) |
| `container_setup_cmd` | `CHELLY_CONTAINER_SETUP_CMD`  | (empty)                      | Shell command to run inside the container before the main command |

### Example config file

```toml
container_cmd = "podman"
workdir = "/workspace"
additional_mounts = ["/home/user/.ssh:/home/user/.ssh"]
container_setup_cmd = "source /etc/profile"
```
