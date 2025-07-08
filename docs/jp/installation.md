# インストールガイド

## クイックインストール

### オプション1：Docker（推奨）

GoSCIMを実行する最速の方法：

```bash
# リポジトリをクローン
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Docker Composeで開始
docker-compose up -d

# インストールの確認
curl http://localhost:8080/ServiceProviderConfig
```

### オプション2：ソースからビルド

```bash
# 前提条件：Go 1.16+
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# ビルドして実行
go build -o goscim main.go
./goscim
```

## システム要件

### 最小要件
- **CPU**: 1コア
- **メモリ**: 512MB RAM
- **ディスク**: 100MB 利用可能容量
- **OS**: Linux, macOS, Windows

### 推奨構成
- **CPU**: 2+ コア
- **メモリ**: 2GB+ RAM  
- **ディスク**: 1GB+ 利用可能容量（データ用）
- **ネットワーク**: 1Gbps イーサネット

## 環境変数

### 必須
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase管理者ユーザー
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase管理者パスワード
```

### オプション
```bash
export SCIM_COUCHBASE_URL="localhost"     # CouchbaseサーバーURL
export SCIM_PORT=":8080"                  # サーバーポート
export SCIM_LOG_LEVEL="info"              # ログレベル
```

## データベースセットアップ

### Couchbaseインストール

```bash
# DockerでCouchbaseを実行
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# クラスター設定（Web UI経由）
# アクセス: http://localhost:8091/ui/index.html
```

### 本番デプロイ

```bash
# 環境変数で設定
export SCIM_ADMIN_USER="your-admin-user"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="your-couchbase-server.com"

# GoSCIMを開始
./goscim
```

## インストールの確認

```bash
# サーバー状態をチェック
curl http://localhost:8080/ServiceProviderConfig

# リソースタイプをチェック
curl http://localhost:8080/ResourceTypes

# スキーマをチェック
curl http://localhost:8080/Schemas
```

## トラブルシューティング

### 一般的な問題

1. **接続失敗**
   ```bash
   # Couchbaseが動作しているかチェック
   telnet localhost 8091
   ```

2. **認証エラー**
   ```bash
   # 環境変数を確認
   echo $SCIM_ADMIN_USER
   echo $SCIM_ADMIN_PASSWORD
   ```

3. **ポート競合**
   ```bash
   # ポートを変更
   export SCIM_PORT=":8081"
   ```