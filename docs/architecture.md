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
| `internal/git`       | Git metadata resolution for linked worktree-aware runtime mounts   |

## Config loading

`LoadConfigFrom(configDir string)` creates a Viper instance that:

1. Reads `<configDir>/config.toml`
2. Binds `CHELLY_*` environment variables (env overrides config file)
3. Applies defaults

The public `LoadConfig()` resolves the config directory with `DefaultConfigDir()` and delegates.

## Args construction

Arg-building logic is separated from execution:

- `container.BuildArgs(cfg, noCache)` returns the `docker build ...` argument slice
- `container.RunArgs(cfg, wd, isTTY, userArgs, autoMounts...)` returns the `docker run ...` argument slice
- `run` passes linked worktree auto-mounts resolved by `internal/git` into `RunArgs`; Git resolution failures produce no auto-mounts
- `container.RunConfig`/`BuildConfig.ExtraArgs` are inserted unconditionally by the `container` package; resolving *which* arguments apply for the current runtime and subcommand is the caller's responsibility (see `config.ResolveRuntimeArgs` below), so `internal/container` has no knowledge of specific runtimes
- Mounts are emitted in current directory, auto-mount, then `additional_mounts` order, with duplicate mount specs removed
- `inherit_env` is inserted as common `--env NAME` run options before the image name
- `env_files` entries are resolved by `config.ResolveEnvFiles` in `run` (tilde expansion, absolute paths must exist, missing relative paths are skipped with a warning) and inserted as `--env-file PATH` run options before the `inherit_env` flags; duplicate-variable merging is delegated to the runtime

This makes unit testing straightforward: tests call these functions directly and assert the returned slices without any subprocess mocking.

## Runtime-specific extra arguments

`Config.RuntimeOptions` holds a list of `{Runtime, Subcommand, Args}` entries
(TOML: `[[runtime_options]]`). `config.ResolveRuntimeArgs(cfg, subcommand)`
resolves the `Args` for the current `container_cmd`'s basename and the given
subcommand (`config.SubcommandRun` or `config.SubcommandBuild`):

1. `CHELLY_RUNTIME_OPTIONS_<RUNTIME>_<SUBCOMMAND>` environment variable, if set
2. Otherwise, the first `RuntimeOptions` entry matching `Runtime` and `Subcommand`

`run` and `build` each call `config.ValidateRuntimeOptions(cfg.RuntimeOptions)`
before use, rejecting unsupported subcommands and duplicate `Runtime`+`Subcommand`
entries. Because resolution is keyed purely by data (`container_cmd` basename and
subcommand name), adding a new runtime or subcommand requires no changes to
`internal/container`.

## Execution semantics

`build` runs the container runtime as a normal child process. `run` replaces the `chelly` process with the container runtime so runtime behavior is delegated directly to the caller-facing process.

## Container command detection

`DetectContainerCmd()` probes `PATH` for `container`, `podman`, `docker` in that order and returns the first found. It is used as the default for `container_cmd` when not configured.

## Container image naming

- Image: `chelly:latest` (hardcoded)

## TTY detection

`IsTTY(f *os.File) bool` checks `os.ModeCharDevice` on the file's stat mode. Both stdin and stdout must be TTYs for the `run` command to add `--interactive --tty`.
