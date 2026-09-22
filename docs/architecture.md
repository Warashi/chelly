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
- `container.RunArgs(cfg, wd, isTTY, userArgs, autoMounts...)` returns the `docker run ...` argument slice and always keeps stdin attached
- When `userArgs` is empty, `RunArgs` appends nothing after the image name so the runtime uses the image's `ENTRYPOINT` and `CMD`
- `run` passes linked worktree auto-mounts resolved by `internal/git` into `RunArgs`; Git resolution failures produce no auto-mounts
- `container.RunConfig`/`BuildConfig.ExtraArgs` are inserted unconditionally by the `container` package; resolving *which* arguments apply for the current runtime and subcommand is the caller's responsibility (see `config.ResolveRuntimeArgs` below). The only runtime-specific knowledge inside `internal/container` is the local-image policy table described below
- `RunArgs` adds `--init` unconditionally as part of the fixed `run --rm --interactive [--tty] --init` prefix. The image entrypoint exec-chains into the user command, which then becomes PID 1 and only waits for its own children, so without an init the orphans of a long session accumulate as zombies. Podman, Docker, and Apple container (verified against 1.4.1, which wraps the process in `vminitd`) all spell the flag `--init`, so no runtime table is involved and the line can be removed on its own
- `RunArgs` inserts the runtime's pull-forbidding option (`--pull=never` on podman/docker) right after that fixed prefix, before `ExtraArgs`
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

## Local-image-only policy for `run`

`internal/container/image.go` is a self-contained unit that keeps `chelly run` on the locally built image. It holds one table keyed by the `container_cmd` basename with two facts per runtime:

| Runtime     | Pull-forbidding run option | Existence check                                   |
|-------------|----------------------------|---------------------------------------------------|
| `podman`    | `--pull=never`             | `image ls -q chelly:latest` (empty output = missing) |
| `docker`    | `--pull=never`             | same as podman                                    |
| `container` | none (Apple container 1.4.1 `run` has no pull policy) | `image inspect chelly:latest` (non-zero exit = missing) |

The two facts play different roles: the run option is the guarantee (the runtime itself refuses to pull), while `CheckImage` only exists so `run` can fail before `Exec` with a `chelly build` hint, since after `Exec` chelly can no longer shape the error. An existence check alone is not a guarantee because the image can disappear between the check and the start; on Apple container that gap is accepted and documented in requirements.

`CheckImage` runs the check with stderr passed through and stdout captured, so runtime errors stay visible and stdout stays clean. `ValidateRunExtraArgs` rejects `--pull`/`--pull=…` in resolved run arguments so `runtime_options` cannot override the policy. Runtimes missing from the table are rejected by `CheckImage` rather than silently running with the default pull policy. `run` calls both before resolving mounts and env files.

Unit tests inject the command runner (`checkImageWith`) instead of invoking a runtime; `internal/cmd` tests exercise the wiring with a fake `podman` script.

## Execution semantics

`build` runs the container runtime as a normal child process. `run` replaces the `chelly` process with the container runtime so runtime behavior is delegated directly to the caller-facing process. `run` sets Cobra's `SilenceUsage` so pre-exec failures print only the error and never write usage text to stdout.

## Container command detection

`DetectContainerCmd()` probes `PATH` for `container`, `podman`, `docker` in that order and returns the first found. It is used as the default for `container_cmd` when not configured.

## Container image naming

- Image: `chelly:latest` (hardcoded); `build` tags it and `run` references it through the same `container.ImageName` constant, which is also what the local-image policy checks

## TTY detection

`IsTTY(f *os.File) bool` uses `term.IsTerminal` (an isatty-style termios ioctl) rather than `os.ModeCharDevice`, because `/dev/null` is also a character device and must not count as a terminal. `run` always adds `--interactive` so piped protocols can keep reading stdin, and adds `--tty` only when both stdin and stdout are TTYs.
