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

- Uses only the locally built `chelly:latest` of the configured runtime and never pulls it from a registry
  - Podman and Docker: passes `--pull=never`; a `--pull` option in `runtime_options` for `run` is rejected as an error
  - Apple container: `run` has no pull policy option (verified against 1.4.1), so only the pre-exec existence check below applies and an image removed between the check and the start could still be pulled
  - Container commands other than `container`, `podman`, and `docker` are rejected instead of running with the runtime's default pull policy
- Checks that the image exists before starting; when it is missing, fails without touching a registry and tells the user to run `chelly build` (no automatic build, no fallback image)
  - Podman and Docker use `image ls -q chelly:latest`, so a failing check (daemon unreachable, permission denied) is reported as that failure, not as a missing image
  - Apple container uses `image inspect chelly:latest`; a non-zero exit is reported as a missing image with the runtime's stderr shown (exit code semantics not yet verified on macOS)
  - The check writes nothing to stdout
- Prints only the error, not the command usage, when `run` fails before starting the container
- When no command is given, delegates command resolution to the container runtime, preserving the image's `ENTRYPOINT` and `CMD`
- Mounts the current directory at the same path inside the container
- When run inside a linked Git worktree, also mounts the parent directory of the Git common dir at the same path inside the container
- If Git metadata cannot be resolved, continues without the linked worktree mount
- Sets the container workdir to the current directory (configurable)
- Inherits configured environment variables from the `chelly run` process into the container
- Passes configured dotenv files to the container runtime as `--env-file` flags, in configuration order (see `env_files` below)
- Keeps stdin attached for both terminal and piped protocols; adds `--tty` only when stdin and stdout are TTYs
- Always passes `--init` so the runtime's init process is PID 1 and reaps orphaned processes; the same flag is used on Podman, Docker, and Apple container, and there is no option to turn it off. When the host has no init binary, the runtime's own start failure is shown as is
- Replaces the `chelly` process with the container runtime, so the runtime's exit status, signal handling, TTY behavior, and process ownership pass through to the caller

### `chelly config list`

Print all current effective configuration values in TOML format.

- Shows the resolved configuration after applying config file and environment variable overrides

### `chelly config get <key>`

Print the effective value of a single configuration key.

- `key`: one of `container_cmd`, `config_home`, `workdir`, `additional_mounts`, `inherit_env`, `env_files`, or `runtime_options.<runtime>.<subcommand>`
- For `additional_mounts`, `inherit_env`, `env_files`, and `runtime_options.<runtime>.<subcommand>`, prints values as a comma-separated string

### `chelly config set <key> <value>`

Write a key-value pair to `config.toml`.

- Creates the config file and its directory if they do not exist
- For `additional_mounts`, `inherit_env`, `env_files`, and `runtime_options.<runtime>.<subcommand>`, `value` is a comma-separated list
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
| `inherit_env`          | `CHELLY_INHERIT_ENV`           | (empty)                      | Environment variable names inherited from `chelly run` into the container |
| `env_files`            | `CHELLY_ENV_FILES`             | (empty)                      | Dotenv files passed to the container runtime via `--env-file` (see below) |
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
- `run` entries must not contain `--pull` (see `chelly run` above).

### `env_files`

`env_files` is a list of dotenv file paths passed to the container runtime as
`--env-file` flags, in configuration order.

- Paths starting with `~` are expanded to the user home directory and treated as absolute paths.
- Relative paths are resolved against the current working directory of `chelly run`.
- A missing file referenced by an absolute path aborts `chelly run` with an error.
- A missing file referenced by a relative path is skipped with a warning on stderr.
- Merge semantics for a variable defined in multiple files are delegated to the
  container runtime; docker and podman apply files in order, so the last file wins.
- Variables listed in `inherit_env` take precedence over `env_files` values,
  because container runtimes prioritize `--env` over `--env-file`.

### Example config file

```toml
container_cmd = "podman"
workdir = "/workspace"
additional_mounts = ["/home/user/.cache:/home/user/.cache"]
inherit_env = ["SSH_AUTH_SOCK", "GITHUB_TOKEN"]
env_files = [".env", "~/.config/chelly/common.env"]

[[runtime_options]]
runtime = "podman"
subcommand = "run"
args = ["--userns=keep-id"]
```
