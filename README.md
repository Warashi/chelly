# chelly

chelly は作業用 container を簡単に作成するためのツールです。

## インストール

```sh
go install github.com/Warashi/chelly@latest
```

## 使い方

chelly は `$XDG_CONFIG_HOME/chelly`、未設定の場合は `$HOME/.config/chelly` を build context として container image を作成します。

まず、build context に `Dockerfile` を用意します。

```sh
mkdir -p ~/.config/chelly
cat > ~/.config/chelly/Dockerfile <<'EOF'
FROM alpine:latest
RUN apk add --no-cache bash git
EOF
```

image を build します。

```sh
chelly build
```

現在のディレクトリを container 内の同じパスに mount して command を実行します。

```sh
chelly run -- sh
chelly run -- git --version
```

設定は `chelly config` で確認・変更できます。

```sh
chelly config list
chelly config set container_cmd podman
chelly config set inherit_env SSH_AUTH_SOCK,GITHUB_TOKEN
chelly config get container_cmd
```

## ドキュメント

- 外部挙動: [docs/requirements.md](docs/requirements.md)
- 設計意図と内部構造: [docs/architecture.md](docs/architecture.md)

## License

Apache-2.0
