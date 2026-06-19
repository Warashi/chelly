# Architecture

## Overview

chelly is built with [Cobra](https://github.com/spf13/cobra) for CLI and [Viper](https://github.com/spf13/viper) for config.

## Package structure

All commands live in the `cmd` package following a flat layout (package by feature at the package level):

| File              | Responsibility                                        |
|-------------------|-------------------------------------------------------|
| `cmd/root.go`     | Root `chelly` command registration                    |
| `cmd/config.go`   | Config struct, loading from file and environment      |
| `cmd/build.go`    | `chelly build` subcommand and args construction       |
| `cmd/run.go`      | `chelly run` and `chelly runit` subcommands and args construction |

## Config loading

`LoadConfigFrom(configDir string)` creates a Viper instance that:

1. Reads `<configDir>/config.toml`
2. Binds `CHELLY_*` environment variables (env overrides config file)
3. Applies defaults

The public `LoadConfig()` resolves the config directory from `os.UserConfigDir()` and delegates.

## Args construction

Arg-building logic is separated from execution:

- `BuildArgs(cfg, noCache)` returns the `docker build ...` argument slice
- `ContainerRunArgs(cfg, wd, forceTTY, isTTY, userArgs)` returns the `docker run ...` argument slice

This makes unit testing straightforward: tests call these functions directly and assert the returned slices without any subprocess mocking.

## Container command detection

`DetectContainerCmd()` probes `PATH` for `container`, `podman`, `docker` in that order and returns the first found. It is used as the default for `container_cmd` when not configured.

## Container image and volume naming

- Image: `chelly:latest` (hardcoded)
- Named volume for `/home`: `chelly-home` (hardcoded, persists across runs)

## TTY detection

`IsTTY(f *os.File) bool` checks `os.ModeCharDevice` on the file's stat mode. Both stdin and stdout must be TTYs for the `run` command to add `--interactive --tty`.
