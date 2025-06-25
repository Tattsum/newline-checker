# check-new-line

ファイルが改行文字で終わっているかをチェックし、必要に応じて自動修正するGoツールです。

## 概要

このツールは、指定されたディレクトリ内のテキストファイルをすべてチェックし、ファイルが改行文字（`\n`）で終わっていない場合に検出します。`-fix`フラグを使用することで、自動的に末尾に改行文字を追加することも可能です。

## 機能

- ディレクトリを再帰的に処理
- バイナリファイルの自動検出・除外
- 隠しファイルや一般的なバイナリ拡張子のスキップ
- ドライランモード（チェックのみ）と修正モード
- 処理結果の詳細な統計情報表示

## インストール

### Go環境がある場合

```bash
go build -o check-new-line main.go
```

### ソースからビルド

```bash
git clone <repository-url>
cd check-new-line
go build -o check-new-line main.go
```

## 使用方法

### 基本的な使用方法

```bash
# チェックのみ（修正はしない）
./check-new-line /path/to/directory

# 自動修正する場合
./check-new-line -fix /path/to/directory
```

### 使用例

```bash
# 現在のディレクトリをチェック
./check-new-line .

# プロジェクトディレクトリを修正
./check-new-line -fix ./my-project

# 特定のディレクトリをチェック
./check-new-line /home/user/src/my-project
```

## 出力例

### チェックモード
```
=== Summary ===
Total files checked: 45
Files skipped: 12
Files missing newline: 3

Files that don't end with newline:
  - src/main.js
  - config/settings.json
  - docs/api.md

Run with -fix flag to automatically add newlines
```

### 修正モード
```
Fixed: src/main.js
Fixed: config/settings.json
Fixed: docs/api.md

=== Summary ===
Total files checked: 45
Files skipped: 12
Files fixed: 3
```

## スキップされるファイル

### 隠しファイル・ディレクトリ
- `.git/`, `.vscode/`, `.idea/` など、ドット（`.`）で始まるもの

### バイナリファイル拡張子
- 実行ファイル: `.exe`, `.dll`, `.so`, `.dylib`, `.a`, `.o`
- 画像ファイル: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.ico`, `.svg`
- メディアファイル: `.mp3`, `.mp4`, `.avi`, `.mov`, `.wav`
- アーカイブファイル: `.zip`, `.tar`, `.gz`, `.bz2`, `.7z`, `.rar`
- ドキュメント: `.pdf`, `.doc`, `.docx`, `.xls`, `.xlsx`
- コンパイル済みファイル: `.pyc`, `.pyo`, `.class`, `.jar`
- データベース: `.db`, `.sqlite`, `.sqlite3`

### バイナリファイル判定
- NULL文字（`\0`）を含むファイル
- 非印字文字が30%以上を占めるファイル

## コマンドラインオプション

| オプション | 説明 |
|-----------|------|
| `-fix` | ファイルの末尾に改行文字を自動追加する |

## 技術的詳細

- **言語**: Go
- **依存関係**: 標準ライブラリのみ
- **ファイル権限**: 修正時は`0644`で書き込み
- **対応OS**: クロスプラットフォーム（Windows, Linux, macOS）

## ライセンス

MIT License

## 貢献

プルリクエストやIssueは歓迎します。バグ報告や機能要望などお気軽にご連絡ください。

## 注意事項

- バックアップを取ってから`-fix`フラグを使用することを推奨します
- 大きなファイルやバイナリファイルは自動的にスキップされますが、重要なファイルは事前に確認してください
- シンボリックリンクは通常のファイルとして処理されます 
