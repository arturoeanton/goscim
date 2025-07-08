# GoSCIM スタートガイド

GoSCIMへようこそ！このガイドでは、数分で独自のSCIM 2.0サーバーを立ち上げて実行できるようになります。

## GoSCIMとは？

GoSCIMは、Goで構築された軽量、高速、かつ柔軟なSCIM 2.0プロトコルの実装です。以下を目的として設計されています：

- 🚀 **アイデンティティ管理の簡素化** 複数システム間での
- 🔧 **簡単な統合** 既存のインフラストラクチャとの
- 📈 **スケール** 小規模なスタートアップから大企業まで
- 🛡️ **保護** 業界のベストプラクティスでアイデンティティデータを

## クイックスタート（2分）

### オプション1：Docker（推奨）

GoSCIMを試す最速の方法：

```bash
# リポジトリをクローン
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# Docker Composeで開始
docker-compose up -d

# サービスの開始を待つ（約30秒）
sleep 30

# SCIMサーバーをテスト
curl http://localhost:8080/ServiceProviderConfig
```

### オプション2：ソースからビルド

ソースからビルドしたい場合：

```bash
# 前提条件：Go 1.16+とCouchbase
git clone https://github.com/arturoeanton/goscim.git
cd goscim

# 環境変数を設定
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="admin123"

# サーバーを実行
go run main.go
```

## 最初のSCIM操作

サーバーが実行されたら、これらの基本操作を試してください：

### 1. サーバー設定の確認
```bash
curl http://localhost:8080/ServiceProviderConfig
```

### 2. 最初のユーザーを作成
```bash
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
    "userName": "jane.doe@example.com",
    "name": {
      "familyName": "田中",
      "givenName": "花子"
    },
    "emails": [{
      "value": "jane.doe@example.com",
      "type": "work",
      "primary": true
    }],
    "active": true
  }'
```

### 3. ユーザーの検索
```bash
curl "http://localhost:8080/scim/v2/Users?filter=userName sw 'jane'"
```

### 4. 利用可能なリソースの一覧
```bash
curl http://localhost:8080/ResourceTypes
```

## レスポンスの理解

ユーザーを作成すると、このようなレスポンスが得られます：

```json
{
  "schemas": ["urn:ietf:params:scim:schemas:core:2.0:User"],
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "userName": "jane.doe@example.com",
  "name": {
    "familyName": "田中",
    "givenName": "花子",
    "formatted": "田中花子"
  },
  "emails": [{
    "value": "jane.doe@example.com",
    "type": "work",
    "primary": true
  }],
  "active": true,
  "meta": {
    "resourceType": "User",
    "created": "2023-12-01T10:30:00Z",
    "lastModified": "2023-12-01T10:30:00Z",
    "version": "W/\"a330bc54f0671c9\"",
    "location": "/scim/v2/Users/550e8400-e29b-41d4-a716-446655440000"
  }
}
```

キー要素：
- **`id`**: サーバーが生成する一意の識別子
- **`meta`**: 作成時間と場所を含むメタデータ
- **`schemas`**: このリソースに使用されるSCIMスキーマ

## 一般的な使用例

### 1. 従業員のオンボーディング
```bash
# 新しい従業員を作成
curl -X POST http://localhost:8080/scim/v2/Users \
  -H "Content-Type: application/json" \
  -d '{
    "schemas": [
      "urn:ietf:params:scim:schemas:core:2.0:User",
      "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User"
    ],
    "userName": "john.smith@company.com",
    "name": {
      "familyName": "スミス",
      "givenName": "ジョン"
    },
    "emails": [{
      "value": "john.smith@company.com",
      "type": "work"
    }],
    "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
      "employeeNumber": "EMP001",
      "department": "エンジニアリング",
      "manager": {
        "value": "manager-id-here"
      }
    }
  }'
```

### 2. 高度なフィルタリング
```bash
# エンジニアリング部門のアクティブユーザーを検索
curl "http://localhost:8080/scim/v2/Users?filter=active eq true and urn:ietf:params:scim:schemas:extension:enterprise:2.0:User:department eq 'エンジニアリング'"

# 会社のメールを持つユーザーを検索
curl "http://localhost:8080/scim/v2/Users?filter=emails[type eq 'work' and value ew '@company.com']"

# 最近変更されたユーザーを検索
curl "http://localhost:8080/scim/v2/Users?filter=meta.lastModified gt '2023-01-01T00:00:00Z'"
```

## エラーハンドリング

GoSCIMは標準のHTTPステータスコードとSCIMエラーレスポンスを返します：

```json
{
  "schemas": ["urn:ietf:params:scim:api:messages:2.0:Error"],
  "detail": "属性 'userName' は必須です",
  "status": "400",
  "scimType": "invalidValue"
}
```

一般的なエラー：
- **400 Bad Request**: 無効なデータまたは必須フィールドの欠如
- **404 Not Found**: リソースが存在しない
- **409 Conflict**: リソースが既に存在する（重複するuserNameなど）
- **500 Internal Server Error**: サーバー側の問題

## 設定の基本

### 環境変数
```bash
# データベース接続
export SCIM_ADMIN_USER="Administrator"
export SCIM_ADMIN_PASSWORD="your-secure-password"
export SCIM_COUCHBASE_URL="couchbase-server.example.com"

# サーバー設定
export SCIM_PORT=":8080"
export SCIM_LOG_LEVEL="info"

# セキュリティ（本番環境用）
export SCIM_TLS_ENABLED="true"
export SCIM_TLS_CERT_FILE="/path/to/cert.pem"
export SCIM_TLS_KEY_FILE="/path/to/key.pem"
```

## 次のステップ

GoSCIMが実行されたので、次のステップをご紹介します：

1. **🔐 [認証の設定](security.md)** - OAuth 2.0またはJWT認証を追加
2. **📊 [モニタリングの設定](operations.md)** - メトリクスとログを設定
3. **🔌 [統合の追加](integrations.md)** - Active Directory、LDAP、またはSaaSアプリに接続
4. **🏗️ [アーキテクチャの学習](architecture.md)** - GoSCIMの内部動作を理解
5. **👩‍💻 [貢献](development.md)** - GoSCIMの改善に協力

## ヘルプの取得

- 📚 **ドキュメント**: [完全なドキュメント](README.md)をご確認ください
- 🐛 **問題**: [GitHub Issues](https://github.com/arturoeanton/goscim/issues)でバグを報告
- 💬 **ディスカッション**: [GitHub Discussions](https://github.com/arturoeanton/goscim/discussions)で質問
- 📖 **例**: より多くの使用例については`httpexamples/`ディレクトリを参照

GoSCIMコミュニティへようこそ！🎉