# AGENTS.md

このファイルは、このリポジトリ固有の作業ルールだけをまとめる。ユーザー単位の設定や一般論はここに書かず、重複する説明は必要に応じて [docs/architecture.md](docs/architecture.md) に委ねる。

## 運用方針

- 必要になったらこの `AGENTS.md` 自体も更新すること。
- 変更は「既存の形を守る」より「破壊的変更も含めて最適な形に整える」ことを優先する。
- Richard Gabriel の LoB （Locality of Behavior） を意識して、コードとドキュメントは関連する内容を近くに置くことを心がける。
- コードは原則として package by feature で整理する。
- ドキュメントは MECE を保ち、同じ説明を繰り返さない。重複が必要な場合は本文ではなくリンクで参照する。

## このリポジトリの前提

- Go のプロジェクトであり、依存関係と開発環境は Nix/flake で管理する。
- 変更時は、関連する `go test` と `golangci-lint run` の確認を前提に進める。
- 整形対象は `nix fmt` と `golangci-lint fmt` を基準にする。
- 内部構造の判断基準は LoB を優先し、実装詳細は [docs/architecture.md](docs/architecture.md) を参照する。

## 作業ルール

- git commit は意味のある最小の単位で行う。
- git commit の粒度を考慮して作業の順序や粒度を決める。
- 実装を追加・変更・削除したら、関連テストも同じ粒度で更新する。
- 作業は小さな単位で進め、各単位ごとに lint と test を通す。
- commit message は `conventional commits` に従う。

## 参照先

- 外部挙動は [docs/requirements.md](docs/requirements.md) を参照する。
- 設計意図とアーキテクチャの詳細は [docs/architecture.md](docs/architecture.md) を参照する。
- ここに書くのは、作業時に迷わないための運用ルールだけに限定する。
