_default:
  @just --list

cobra-add COMMAND:
  go run github.com/spf13/cobra-cli@latest --config .cobra.yaml add {{COMMAND}}
