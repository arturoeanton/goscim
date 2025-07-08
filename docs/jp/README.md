# GoSCIM 技術文書

## プロジェクト概要

**GoSCIM** は、Go言語で構築された完全なSCIM 2.0（System for Cross-domain Identity Management）実装です。分散環境におけるアイデンティティ管理のための堅牢でスケーラブルなソリューションを提供し、異種アイデンティティシステムの統合に特化して設計されています。

## 主要機能

### SCIM 2.0 準拠
- ✅ 完全なCRUD操作（作成、読取、更新、削除）
- ✅ SCIMフィルタによる高度検索
- ✅ ページネーションとソート
- ✅ 拡張可能でカスタマイズ可能なスキーマ
- ✅ 複数リソースタイプサポート
- ✅ バルク操作（開発中）

### 技術アーキテクチャ
- **言語**: Go 1.16+
- **Webフレームワーク**: Gin（高性能）
- **データベース**: Couchbase（分散NoSQL）
- **パーサー**: SCIMフィルタ用ANTLR v4
- **データ形式**: ネイティブJSON

## システムアーキテクチャ

### コアコンポーネント

#### 1. サーバーコア (`main.go`)
```go
// サーバー初期化
func main() {
    scim.InitDB()                    // Couchbase接続
    r := gin.Default()               // HTTPルーター
    scim.ReadResourceType(config, r) // 動的スキーマ読み込み
    r.Run(port)                      // HTTPサーバー
}
```

#### 2. 設定管理 (`scim/config.go`)
- **動的スキーマ読み込み** JSONファイルから
- **リソースタイプベースの自動エンドポイント登録**
- **サーバー起動時のスキーマ検証**

#### 3. データベース統合 (`scim/couchbase.go`)
- **認証付きセキュア接続**
- **リソースタイプごとの自動バケット作成**
- **カスタマイズ可能なバケット設定**
- **自動プライマリインデックス**

#### 4. フィルタパーサー (`scim/parser/`)
- **SCIMフィルタのANTLR文法**
- **N1QLクエリへの自動変換**
- **SCIMオペレーターの完全サポート**

### SCIM 操作

#### 作成 (POST)
```http
POST /scim/v2/Users
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "userName": "user@example.com",
  "name": {
    "familyName": "姓",
    "givenName": "名"
  }
}
```

#### 読取 (GET)
```http
GET /scim/v2/Users/12345
```

#### 更新 (PATCH)
```http
PATCH /scim/v2/Users/12345
Content-Type: application/json

{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:PatchOp"],
  "Operations": [
    {
      "op": "replace",
      "path": "name.familyName",
      "value": "新姓"
    }
  ]
}
```

#### 検索 (GET)
```http
GET /scim/v2/Users?filter=userName sw "admin"&sortBy=userName&sortOrder=ascending&startIndex=1&count=10
```

#### 削除 (DELETE)
```http
DELETE /scim/v2/Users/12345
```

## システム設定

### 環境変数

#### 必須
```bash
export SCIM_ADMIN_USER="Administrator"     # Couchbase管理者ユーザー
export SCIM_ADMIN_PASSWORD="admin123"     # Couchbase管理者パスワード
```

#### オプション
```bash
export SCIM_COUCHBASE_URL="localhost"     # CouchbaseサーバーURL
export SCIM_PORT=":8080"                  # SCIMサーバーポート
```

### 設定構造

```
config/
├── schemas/                    # SCIMスキーマ定義
│   ├── urn+ietf+params+scim+schemas+core+2.0+User.json
│   ├── urn+ietf+params+scim+schemas+core+2.0+Group.json
│   └── urn+ietf+params+scim+schemas+extension+*.json
├── resourceType/              # リソースタイプ
│   ├── User.json
│   ├── Group.json
│   └── Element.json
├── bucketSettings/            # Couchbaseバケット設定
│   ├── User.json
│   ├── Group.json
│   └── Element.json
└── serviceProviderConfig/     # プロバイダー設定
    └── sp_config.json
```

## スキーマと拡張

### 基本ユーザースキーマ
```json
{
  "id": "urn:ietf:params:scim:schemas:core:2.0:User",
  "name": "User",
  "description": "ユーザーアカウント",
  "attributes": [
    {
      "name": "userName",
      "type": "string",
      "required": true,
      "uniqueness": "server"
    },
    {
      "name": "name",
      "type": "complex",
      "subAttributes": [
        {
          "name": "familyName",
          "type": "string"
        },
        {
          "name": "givenName", 
          "type": "string"
        }
      ]
    }
  ]
}
```

### カスタム拡張
```json
{
  "id": "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User",
  "name": "エンタープライズユーザー",
  "attributes": [
    {
      "name": "employeeNumber",
      "type": "string",
      "uniqueness": "server"
    },
    {
      "name": "department",
      "type": "string"
    }
  ]
}
```

## アクセス制御

### ロールと権限
```json
{
  "name": "sensitiveAttribute",
  "type": "string",
  "$reader": ["admin", "hr"],      # 読み取り可能なロール
  "$writer": ["admin"],            # 書き込み可能なロール
  "returned": "default"
}
```

### ロール検証
```go
// 検索時の自動検証
roles := []string{"user", "admin", "superadmin"}
element := ValidateReadRole(roles, resourceType, item)
```

## SCIM フィルタ

### サポートされる構文
```
# 基本比較
userName eq "admin"
name.familyName co "田中"
userName sw "admin"
active pr

# 時間比較
meta.lastModified gt "2023-01-01T00:00:00Z"
meta.lastModified ge "2023-01-01T00:00:00Z"

# 論理演算子
title pr and userType eq "Employee"
title pr or userType eq "Intern"
userType eq "Employee" and (emails co "company.com" or emails co "company.org")
```

### N1QL 変換
```go
// 変換例
query, _ := parser.FilterToN1QL("User", 
    "userName eq \"admin\" and active eq true")
// 結果: SELECT * FROM `User` WHERE `userName` = "admin" AND `active` = true
```

## インストールとデプロイ

### システム要件

#### 開発環境
- Go 1.16以上
- Couchbase Server 6.0+
- ANTLR 4.7（パーサー再生成用）

#### 本番環境
- CPU: 2+コア
- メモリ: 4GB+（データ量に依存）
- ストレージ: SSD推奨
- ネットワーク: 1Gbps+（高並行性）

### ローカルインストール

#### 1. リポジトリのクローン
```bash
git clone https://github.com/arturoeanton/goscim.git
cd goscim
```

#### 2. 依存関係のインストール
```bash
go mod download
```

#### 3. Couchbaseの設定
```bash
# DockerでCouchbaseを実行
docker run -d --name couchbase-scim \
  -p 8091-8094:8091-8094 \
  -p 11210:11210 \
  couchbase:latest

# WebUIでクラスターを設定
# http://localhost:8091/ui/index.html
```

#### 4. 環境変数の設定
```bash
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"
export SCIM_COUCHBASE_URL="localhost"
export SCIM_PORT=":8080"
```

#### 5. サーバーの実行
```bash
go run main.go
```

## 監視と運用

### システムログ
```bash
# 構造化ログの設定
export SCIM_LOG_LEVEL=debug
export SCIM_LOG_FORMAT=json

# ログ例
{"level":"info","timestamp":"2023-12-01T10:00:00Z","message":"GoScim v0.1 開始"}
{"level":"debug","timestamp":"2023-12-01T10:00:01Z","message":"バケット作成","bucket":"User"}
```

### 推奨メトリクス
- 毎秒リクエスト数（RPS）
- レスポンス時間パーセンタイル
- エンドポイントごとのエラー率
- アクティブなCouchbase接続数
- メモリとCPU使用率

### ヘルスチェック
```http
GET /health
{
  "status": "healthy",
  "database": "connected",
  "version": "v0.1"
}
```

## セキュリティ

### セキュリティ考慮事項

#### 認証
- OAuth 2.0 / OpenID Connectの実装
- JWTトークンサポート
- 各リクエストでのトークン検証

#### 認可
- 細粒度のロールベース制御
- リソースと操作の権限
- アクセス監査ログ

#### 通信
- 本番環境でのTLS 1.3強制
- 有効な証明書
- HTTPセキュリティヘッダー

## トラブルシューティング

### 一般的な問題

#### Couchbase接続
```bash
# 接続の確認
telnet localhost 8091

# 認証情報の確認
curl -u Administrator:admin123 http://localhost:8091/pools
```

#### スキーマエラー
```bash
# スキーマJSONの検証
jq . config/schemas/schema.json

# 構文の確認
go run main.go --validate-config
```

## コミュニティとサポート

技術サポート、バグレポート、機能リクエストについて：
- **問題**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)
- **ディスカッション**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)
- **ドキュメント**: [プロジェクトWiki](https://github.com/arturoeanton/goscim/wiki)
- **例**: `httpexamples/` ディレクトリ