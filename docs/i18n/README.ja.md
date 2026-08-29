# Open Novel — 世界多言語小説プラットフォーム

<div align="center">

[中文](../../README.md) · [English](README.en.md) · **日本語** · [한국어](README.ko.md) · [Русский](README.ru.md) · [Deutsch](README.de.md) · [Français](README.fr.md) · [Español](README.es.md) · [Português](README.pt.md) · [हिन्दी](README.hi.md) · [العربية](README.ar.md) · [বাংলা](README.bn.md) · [Bahasa Indonesia](README.id.md)

</div>

> **Go-Kratos** マイクロサービスアーキテクチャ + **Flutter / HarmonyOS** マルチクライアントフロントエンドに基づく、世界多言語対応の小説読書プラットフォームです。**12 以上の主要言語**に対応し、世界中のユーザーに読書・交流・検索・パーソナライズレコメンド機能を提供します。

<div align="center"><img src="../mascot.svg" alt="Open Novel のマスコット Novi" width="150"/></div>

---

## プロジェクト概要

Open Novel は、クラウドネイティブなマイクロサービスアーキテクチャの世界多言語小説プラットフォームです：

- **バックエンド**：Go-Kratos v2（gRPC / HTTP デュアルプロトコル）。マイクロサービスはドメインごとに分割（ユーザー、書籍、チャプター、コメント、検索、レコメンド）
- **フロントエンド**：Flutter オールプラットフォーム（Web / Desktop / Mobile）+ HarmonyOS NEXT ネイティブアプリ。同一のバックエンド API を共用
- **多言語対応**：i18n リソースを動的にロードし、12 以上の言語に対応（中国語、英語、日本語、韓国語、フランス語、ドイツ語、スペイン語、ロシア語、アラビア語など）
- **ストレージ**：MySQL 8（マスタースレーブ読み書き分離）+ Redis（ホットキャッシュ / セッション）+ OpenSearch（多言語検索）
- **運用**：Docker Compose によるワンクリックデプロイ、Prometheus + Grafana モニタリング、GitHub Actions 継続的インテグレーション


## 機能一覧

<p align="center"><img src="images/ja/features.svg" alt="機能アーキテクチャ図" width="860"/></p>

- **ユーザーセンター**：登録・ログイン（JWT）、個人ブックシェルフ、端末をまたいだ読書進捗の同期、多言語プロフィール
- **読書体験**：章ごとの読書、フォント・文字サイズの切替、ライト/ダークテーマ、オフラインキャッシュ、ページめくりアニメーション
- **書籍コンテンツ**：書籍メタデータ、章管理、カテゴリタグ、連載更新、多言語翻訳
- **交流コミュニティ**：コメント・書評、いいね、お気に入り、通報・審査
- **検索・発見**：多言語形態素解析検索、ホットキーワードランキング、検索サジェスト（クライアント側ローカル履歴 20 件 + 200ms デバウンスのサジェスト）、AI レコメンド、カテゴリ閲覧
- **管理バックエンド**：コンテンツ審査、ユーザー管理、データ統計（ダッシュボード / DAU / ランキング / 行動分析 /api/stats/behavior）、設定管理（カテゴリタグ）、機械翻訳ワークフロー（DeepL、/api/admin/translate/*、管理側「翻訳」ページ + 手動編集）、監査ログ照会（/api/admin/audit-logs）
- **決済とVIP**：11 つの決済プロバイダー（Stripe、NOWPayments（USDT）、Razorpay、KOMOJU、PortOne、Mercado Pago、Xendit、PayPal、Alipay、WeChat Pay Global、UnionPay）による多チャネル決済、VIPプラン購読と更新、言語別の決済手段ルーティング（WeChat Pay Global は接続済み、国内の WeChat Pay は未接続、中国の加盟店資格が必要）

## システムアーキテクチャ

<p align="center"><img src="images/ja/architecture.svg" alt="システムアーキテクチャ図" width="860"/></p>

全体は Go-Kratos マイクロサービスアーキテクチャです。Flutter / HarmonyOS クライアントは Nginx + CDN を経由して API ゲートウェイとやり取りし、ゲートウェイはドメインごとにユーザー、書籍、チャプター、コメント、検索、レコメンドなどのバックエンドサービスへルーティングします。データ層は MySQL マスタースレーブ（読み書き分離）+ Redis キャッシュ + OpenSearch 検索インデックスです。サービス間は gRPC で通信し、外部向け HTTP インターフェースは統一プレフィックス `/api` を使用します。

その他の設計図：プロジェクト全景 [docs/project.svg](../../docs/project.svg) · リクエストサイクル [docs/request-cycle.svg](../../docs/request-cycle.svg) · セキュリティアーキテクチャ [docs/security.svg](../../docs/security.svg) · プロジェクト構成 [docs/structure.svg](../../docs/structure.svg)。

## プロジェクト全景

<p align="center"><img src="images/ja/project.svg" alt="プロジェクト全体像図" width="860"/></p>

## リクエストサイクル

<p align="center"><img src="images/ja/request-cycle.svg" alt="リクエストサイクル図" width="860"/></p>

## セキュリティアーキテクチャ

<p align="center"><img src="images/ja/security.svg" alt="セキュリティアーキテクチャ図" width="860"/></p>

---

## ディレクトリ構成

```
open-novel/
├─ apps/                     # マルチクライアントフロントエンド
│  ├─ flutter/               #   Flutter オールプラットフォーム（Web / Desktop / Mobile）、i18n 多言語対応
│  └─ harmonyos/             #   HarmonyOS NEXT ネイティブアプリ（ArkTS / ArkUI）
├─ kratos/                   # Go-Kratos フレームワークソース（上流フレームワーク、そのまま保持、変更禁止）
│  └─ backend/               #   本プロジェクトの業務バックエンド：cmd/server エントリ + api/ + internal/ + sql/ + opensearch/
├─ docs/                     # プロジェクトドキュメント（計画、アーキテクチャ図、i18n README、寄付用 QR コード）
├─ scripts/                  # ビルド・デプロイスクリプト（post-push.sh 自動リリース、smoke.sh）
├─ docker-compose.yml        # ローカル依存スタック：MySQL 8 + Redis 7 + OpenSearch 2
├─ CLAUDE.md                 # プロジェクト協力規範
└─ README.md                 # プロジェクト説明ドキュメント
```

<p align="center"><img src="images/ja/structure.svg" alt="プロジェクト構成図" width="860"/></p>

> 注意：`kratos/` は Kratos フレームワークのソースコード（README / LICENSE 付属）です。業務コードはすべて `kratos/backend/` にあります。

## 技術スタック

| レイヤー | 技術選定 |
| :--- | :--- |
| クライアント | Flutter（Web / Desktop / Mobile）、HarmonyOS NEXT（ArkTS / ArkUI） |
| ゲートウェイ | Nginx + CDN、Go-Kratos API ゲートウェイ（gRPC / HTTP デュアルプロトコル） |
| サーバーサイド | Go 1.22+、Kratos v2、protobuf / gRPC |
| ストレージ | MySQL 8.0（マスタースレーブ）、Redis 7.x（Cluster）、OpenSearch 2.x、Redis の上層の ristretto プロセス内 L1 キャッシュ（30 秒 TTL） |
| 可観測性 | Prometheus、Grafana、ELK、OpenTelemetry リンクトレーシング |
| 運用 | Docker Compose、GitHub Actions CI/CD |

## データベース

- データベース名：`novel`
- テーブルプレフィックス：`novel_`（例：`novel_user`、`novel_book`、`novel_chapter`、`novel_comment` など）

```sql
CREATE DATABASE novel DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

テーブル作成スクリプト：`kratos/backend/sql/init.sql`（Docker Compose 初回起動時に自動実行）。詳細なテーブル設計と読み書き分離戦略については [docs/novel-project-planning.md](../novel-project-planning.md) を参照してください。

## API プレフィックス

バックエンドの HTTP インターフェースはすべて `/api` で始まり、バージョンはリクエストヘッダー `X-Api-Version: v1` でネゴシエーションされます（URL には記載しません）。エンドポイントはドメインごとにグループ化されています：

| ドメイン | サンプルルート | proto 定義 |
| :--- | :--- | :--- |
| ユーザー | `/api/users` など | `kratos/backend/api/user/v1` |
| 書籍 | `/api/books`、`/api/books/{id}`、`/api/categories`、`/api/tags` | `kratos/backend/api/book/v1` |
| チャプター | `/api/...` | `kratos/backend/api/chapter/v1` |
| コメント | `/api/...` | `kratos/backend/api/comment/v1` |
| 検索 | `/api/...` | `kratos/backend/api/search/v1` |
| レコメンド | `/api/...` | `kratos/backend/api/recommendation/v1` |

詳細なルートについては、各 proto ファイルの `option (google.api.http)` 宣言を参照してください。

## クイックスタート

```bash
# 1. 依存スタックを起動（MySQL / Redis / OpenSearch、初回起動時に kratos/backend/sql/init.sql を自動実行してテーブル作成）
docker compose up -d

# 2. バックエンドサービスを起動（Kratos 業務ディレクトリ、HTTP :8000 / gRPC :9000）
cd kratos/backend && go mod tidy && go run ./cmd/server

# 3. Flutter クライアントを起動（デフォルトで localhost:8000 に接続、追加設定不要）
cd apps/client/flutter && flutter pub get && flutter run -d chrome
```

- 依存スタックのポートマッピング：MySQL `3307`、Redis `6380`、OpenSearch `9200`（ホストの 3306/6379 はローカルサービスが使用中のため、docker-compose.yml のコメント参照）。
- バックエンドのアドレスとシークレットは `kratos/backend/config/` で設定し、環境変数による上書きに対応（例：`PORT`、`OPENSEARCH_ADDR`）。
- Flutter を別のバックエンドに接続する場合：`flutter run -d chrome --dart-define=API_BASE_URL=http://<host>:8000`。

詳細は [apps/README.md](../../apps/README.md) と [apps/client/flutter/README.md](../../apps/client/flutter/README.md) を参照してください。

## リリースフロー

- **自動**：`main` をプッシュすると、GitHub Actions（[.github/workflows/release.yml](../../.github/workflows/release.yml)）が最新の `v*` タグに基づいて patch バージョンを自動インクリメントし、タグを作成してプッシュした後、増分チェンジログ付きで GitHub Release を作成します。HEAD がすでにバージョンタグを保持している場合はスキップされます。初回リリースは `v1.0.0` から始まります。
- **手動フォールバック**：[scripts/post-push.sh](../../scripts/post-push.sh) を実行（`gh` の認証が必要）：`echo "x y refs/heads/main z" | scripts/post-push.sh`。
- **手動**：

  ```bash
  git tag -a v1.0.1 -m "release v1.0.1"
  git push origin v1.0.1
  gh release create v1.0.1 --generate-notes
  ```

## ロードマップ

| フェーズ | 期間 | タスクの重点 |
| :--- | :--- | :--- |
| Phase 1 | 2〜3 週間 | Kratos バックエンド基盤サービス + MySQL / Redis / OpenSearch 統合 |
| Phase 2 | 3〜4 週間 | Flutter / HarmonyOS マルチクライアントフロントエンド + 多言語 ARB 作成 |
| Phase 3 | 2 週間 | セキュリティ強化（JWT / RBAC / レート制限）+ ストレステスト |
| Phase 4 | 1〜2 週間 | 全リンク結合テスト + CDN 高速化設定 |
| Phase 5 | 継続 | AI レコメンドアルゴリズム導入、ユーザー行動分析トラッキング |

すべてのタスクチェーンが完了しました。

## サポートと寄付

このプロジェクトがお役に立ったなら、ぜひ **Star** や **Fork** でサポートしてください。また、QR コードからスキャンして寄付していただくことも歓迎します。皆様のサポートが継続的なメンテナンスとアップデートの原動力です。ご支援ありがとうございます！

<div align="center">

**WeChat 寄付** ｜ **Alipay 寄付**

<img src="../weixinpay.png" width="130" height="130" alt="WeChat 寄付用 QR コード" />　<img src="../alipay.png" width="130" height="130" alt="Alipay 寄付用 QR コード" />

</div>

### グローバル送金による寄付（クロスボーダー送金）

【受取人情報】

- 受取人氏名：WANG KEXUN
- 受取口座番号：881015918251

【受取銀行】

- ZA Bank SWIFT Code：AABLHKHHXXX
- 銀行名：ZA Bank Limited
- 銀行番号：387
- 銀行所在地：Core F, Cyberport 3, 100 Cyberport Road, Hong Kong

【クロスボーダー送金代理銀行（必要な場合）】

> ご注意ください。これはクロスボーダー送金代理銀行（中継銀行）の情報であり、受取銀行の情報ではありません。送金銀行にクロスボーダー送金代理銀行の情報が必要かどうかお問い合わせください。

**香港ドル、人民元、米ドルの入金時の代理銀行は Citibank です**

- 銀行名：Citibank N.A. Hong Kong
- SWIFT Code：CITIHKHXXXX
- 銀行番号：006
- 支店名：Hong Kong Branch
- 支店番号：391
- 銀行所在地：Citibank Tower, Citibank Plaza, 3 Garden Road, Central, Hong Kong

**その他の通貨の入金時の代理銀行は BNY Mellon です**

- 銀行名：THE BANK OF NEW YORK MELLON
- SWIFT Code：IRVTUS3NXXX
- 銀行所在地：THE BANK OF NEW YORK MELLON, 240 GREENWICH STREET, NEW YORK, United States

### 仮想通貨の寄付 (Crypto Donation)

このプロジェクトがお役に立ったら、QRコードをスキャンして寄付してください。ありがとうございます！

| ネットワーク (Network) | QRコード (QR Code) | ウォレットアドレス (Wallet Address) |
|---|---|---|
| BNB Smart Chain (BEP20) | [<img src="../coin/1.jpg" width="150" alt="BNB Smart Chain (BEP20)">](../coin/1.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Tron (TRC20) | [<img src="../coin/2.jpg" width="150" alt="Tron (TRC20)">](../coin/2.jpg) | `TEdDHWLajt1XvqtPDWmQctdrJaC3pzZZzz` |
| Ethereum (ERC20) | [<img src="../coin/3.jpg" width="150" alt="Ethereum (ERC20)">](../coin/3.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Aptos | [<img src="../coin/4.jpg" width="150" alt="Aptos">](../coin/4.jpg) | `0x836e3780edfc3f7b2372b39e2a1a3a5d7adfaccd96c726f21cfde1b50dd68030` |
| Plasma | [<img src="../coin/5.jpg" width="150" alt="Plasma">](../coin/5.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Polygon POS | [<img src="../coin/6.jpg" width="150" alt="Polygon POS">](../coin/6.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| Solana | [<img src="../coin/7.jpg" width="150" alt="Solana">](../coin/7.jpg) | `2hfhboHdmdrYsY25XfQSsEWxq5ip4EQsR7f4AzSRMUyr` |
| The Open Network (TON) | [<img src="../coin/8.jpg" width="150" alt="The Open Network (TON)">](../coin/8.jpg) | `UQB9kFQohzmXUir9QSSZq01iwl9aQZIDdBpNmDklljRtCoGK` |
| Arbitrum One | [<img src="../coin/9.jpg" width="150" alt="Arbitrum One">](../coin/9.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |
| AVAX C-Chain | [<img src="../coin/10.jpg" width="150" alt="AVAX C-Chain">](../coin/10.jpg) | `0x355d429f97511897ccb4e271ec888205f9ab6629` |

---

## License と連絡先

- **License**：リポジトリのルートには独立した LICENSE はありません。`kratos/` は Kratos フレームワークの上流ソースであり、その [MIT License](../../kratos/LICENSE) に従います。業務コードのライセンス方式は今後のプロジェクト発表に委ねられます。
- **連絡先**：GitHub Issues / PR で交流できます。寄付は上記「サポートと寄付」をご覧ください。
