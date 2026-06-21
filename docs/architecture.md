# Architecture

## Overview

chelly is built with [Cobra](https://github.com/spf13/cobra) for CLI and [Viper](https://github.com/spf13/viper) for config.
`main.go` is only the binary entrypoint; application behavior lives under `internal`.

## Package structure

The implementation is split by feature under `internal`:

| Package              | Responsibility                                                     |
|----------------------|--------------------------------------------------------------------|
| `internal/cmd`       | Cobra command tree and wiring for `build`, `run`, and `config`     |
| `internal/config`    | Config struct, loading from file/environment, formatting, updates  |
| `internal/container` | Container argument construction, TTY detection, command execution  |

## Config loading

`LoadConfigFrom(configDir string)` creates a Viper instance that:

1. Reads `<configDir>/config.toml`
2. Binds `CHELLY_*` environment variables (env overrides config file)
3. Applies defaults

The public `LoadConfig()` resolves the config directory with `DefaultConfigDir()` and delegates.

## Args construction

Arg-building logic is separated from execution:

- `container.BuildArgs(cfg, noCache)` returns the `docker build ...` argument slice
- `container.RunArgs(cfg, wd, isTTY, userArgs)` returns the `docker run ...` argument slice
- `podman_options.run` is inserted only when `container_cmd` resolves to a basename of `podman`

This makes unit testing straightforward: tests call these functions directly and assert the returned slices without any subprocess mocking.

## Execution semantics

`build` runs the container runtime as a normal child process. `run` replaces the `chelly` process with the container runtime so runtime behavior is delegated directly to the caller-facing process.

## Container command detection

`DetectContainerCmd()` probes `PATH` for `container`, `podman`, `docker` in that order and returns the first found. It is used as the default for `container_cmd` when not configured.

## Container image naming

- Image: `chelly:latest` (hardcoded)

## TTY detection

`IsTTY(f *os.File) bool` checks `os.ModeCharDevice` on the file's stat mode. Both stdin and stdout must be TTYs for the `run` command to add `--interactive --tty`.
