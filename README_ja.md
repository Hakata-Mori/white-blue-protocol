[English](README.md) | [中文](README_zh.md) | [日本語](README_ja.md) | [한국어](README_ko.md)

# ホワイト＆ブルー プロトコル (White & Blue Protocol)

**誰でも自分のトークンを発行できる。カフェでも、スタートアップでも、家族でも。**

White & Blue Protocolは、デュアルトークンPoSブロックチェーンです。あらゆる組織が、チェーンのネイティブ通貨（White Coin）を裏付けとした独自の取引可能なトークン（Blue Coin）を、組み込みのAMMプールを通じて発行できます。

> つまり: **すべての組織が株式のようなトークンを手に入れ、即座に取引可能で、上場手数料はゼロ。**

---

## なぜこのプロトコルが存在するのか？

従来の資金調達は、小規模な組織にとって機能していません：
- IPOには数百万ドルのコストと数年の時間がかかる
- 暗号通貨のトークン発行にはスマートコントラクトの専門知識が必要
- クラウドファンディングプラットフォームは高額な手数料を取り、投資家には何も還元しない

**White & Blue Protocolはこれを解決します。** 街のカフェからテックスタートアップまで、あらゆる組織がたった1回のトランザクションで独自のBlue Coinを発行できます。支援者はBlue Coinを購入して応援し、組織の成長に伴う利益を得られる可能性があります。

---

## 実際のユースケース

### カフェの場合

ラテラボが新店舗をオープンし、コミュニティの支援を求めています：

1. **LatteCoin (LTC)** をデプロイ — 70%をAMMプールに、30%をチーム用にロック
2. 初期流動性として500 WCを注入
3. 常連客が組み込みのスワップ機能でLatteCoinを購入
4. 購入者が増えるにつれ、価格が自然に上昇（定積AMMの仕組み）
5. チームの30%は毎月リリース — 店舗拡大や新設備の購入などに使用
6. 店が繁盛すれば、初期の支援者は利益を得る。うまくいかなければ、プールは自然に枯渇 — 面倒な上場廃止は不要

### テックスタートアップの場合

3人のスタートアップが、株式を手放さずにシード資金を調達したい場合：

1. 創業者が **2-of-3マルチシグウォレット** を設定
2. マルチシグでチーム配分を管理する **StartupCoin (STC)** をデプロイ
3. 投資家がAMMでStartupCoinを購入 — 価格は市場の信頼を反映
4. チームの支出には3人の創業者のうち2人の同意が必要（マルチシグ）
5. 毎月のトークンリリースで開発マイルストーンに資金を供給

### 家族の場合

はい、家族でも：

1. 田中家が **TanakaCoin (JSN)** を楽しい家族の実験としてデプロイ
2. 家族や友人が購入
3. 「時価総額」がお正月の食卓での話のネタになる
4. しかし中身は本物のトークンで、本物の経済原理が働く — AMM価格決定、バーンメカニクスなど、すべて揃っている

### あなたのアイデア次第

- YouTuberが **FanCoin** を発行 — 支援者がその成長に投資
- 地元の農家が **FarmCoin** を発行 — トークンを買えば農産物の割引
- 学校のクラブが **ClubCoin** を発行 — 透明性のある資金調達
- バンドが **BandCoin** を発行 — ファンがステークホルダーに

**支援者がいるなら、Blue Coinを持てる。**

---

## 仕組み

### 2つのトークン

| | White Coin (WC) | Blue Coin |
|---|---|---|
| 概要 | チェーンのネイティブ通貨 | 組織固有のトークン |
| 供給量 | 10億枚（PoSによるマイニング） | トークンあたり100万枚（固定） |
| 入手方法 | バリデータノードを運用 | AMMで購入 / 送金を受け取る |
| 用途 | ガス代、流動性、ステーキング | 投資、コミュニティ、ロイヤルティ |

### トークンエコノミクス

- **AMMプール**: 定積公式（`x * y = k`）。価格は需給に応じて変動
- **スワップ時2%バーン**: 取引ごとにBlue Coinの2%がバーンされ、デフレーションを生む
- **チームベスティング**: チーム配分は一括ではなく毎月リリース
- **マルチシグ管理**: チーム資金はアカウンタビリティのためN-of-Mマルチシグでロック可能
- **自然消滅**: 誰も取引しなければ、プールは自然にゼロへ。上場廃止は不要

### バリデータエコノミクス

| パラメータ | 値 |
|-----------|----------|
| 参加コスト | 無料（24時間オンライン + Proof-of-Work） |
| ブロック報酬 | 1ブロックあたり50 WC |
| ブロック間隔 | 15秒 |
| 自動ステーキング | 報酬は1,000 WCになるまでステークとしてロック |
| 手数料分配 | トランザクション手数料の50%がブロック生成者へ |
| オフラインペナルティ | 24時間後に停止、72時間後に除外 |
| 二重署名 | 永久追放 + ステーク没収 |

---

## ライブテストネット

**ブロックエクスプローラー**: [http://8.217.52.231](http://8.217.52.231)

ウォレットの作成、ブロックの閲覧、トークンの取引 — すべてブラウザから：
- リアルタイムのブロック、トランザクション、バリデータのステータスを表示
- ウォレットを作成してトークンを管理
- WhiteCoinとBlueCoinを送金
- 独自のBlueCoinトークンをデプロイ
- 組み込みAMMでトークンをスワップ

---

## クイックスタート

### 方法1: Webエクスプローラーを使う（インストール不要）

[http://8.217.52.231](http://8.217.52.231) にアクセスして **Wallet** をクリックすれば始められます。

### 方法2: フルノードを実行する

```bash
git clone https://github.com/Hataka-Mori/white-blue-protocol.git
cd white-blue-protocol
make build

# フルノードを起動（テストネットから自動同期）
./wblue start --no-validator
```

### 方法3: バリデータを実行する（WCを獲得）

```bash
./wblue start
```

ノードは以下の動作を行います：
1. ウォレットを自動生成
2. テストネットのシードノードに接続
3. 24時間の候補期間を開始
4. 24時間 + PoW検証の後、ブロック生成を開始し、1ブロックあたり50 WCを獲得

---

## CLIコマンド

### ウォレット

```bash
wblue wallet create                # 新しいウォレットを作成
wblue wallet list                  # ウォレット一覧を表示
wblue wallet info <address>        # 残高を表示
```

### 送金

```bash
wblue transfer white --from <addr> --to <addr> --amount 100
wblue transfer blue --from <addr> --to <addr> --token <id> --amount 500
```

### Blue Coin

```bash
# 新しいトークンをデプロイ
wblue bluecoin deploy \
  --from <addr> \
  --name "LatteCoin" \
  --symbol "LTC" \
  --pool-ratio 70 \
  --team-ratio 30 \
  --init-white 500 \
  --release-monthly 5000 \
  --multisig <multisig-addr>

# クエリ
wblue bluecoin list
wblue bluecoin info <tokenId>
wblue bluecoin burn --from <addr> --token <id> --amount 1000
```

### AMMスワップ

```bash
# White CoinでBlue Coinを購入
wblue amm swap --from <addr> --token <id> --direction white-to-blue --amount-in 100

# Blue CoinをWhite Coinに売却
wblue amm swap --from <addr> --token <id> --direction blue-to-white --amount-in 500

# プール情報を確認
wblue amm pool-info <tokenId>
wblue amm price <tokenId>
```

### バリデータ

```bash
wblue validator join --from <addr>
wblue validator exit --from <addr>
wblue validator status
wblue validator heartbeat --from <addr>
```

### マルチシグ

```bash
wblue multisig register --owners addr1,addr2,addr3 --threshold 2
wblue multisig propose --multisig <ms-addr> --to <target> --amount 100
wblue multisig approve --multisig <ms-addr> --proposal-id 0
wblue multisig info --address <ms-addr>
```

### チェーン

```bash
wblue chain status
wblue chain tx <hash>
wblue version
```

---

## HTTP API

| メソッド | エンドポイント | 説明 |
|--------|----------|-------------|
| GET | `/health` | ヘルスチェック |
| GET | `/api/v1/chain/status` | チェーンステータス |
| GET | `/api/v1/chain/block/:height` | 高さでブロックを取得 |
| GET | `/api/v1/blocks?limit=20&offset=0` | ブロック一覧（ページネーション対応） |
| GET | `/api/v1/block/hash/:hash` | ハッシュでブロックを取得 |
| GET | `/api/v1/stats` | ネットワーク統計 |
| GET | `/api/v1/wallet/:address` | アカウント残高 |
| GET | `/api/v1/bluecoin` | 全Blue Coin一覧 |
| GET | `/api/v1/bluecoin/:tokenId` | Blue Coin設定 |
| GET | `/api/v1/bluecoin/:tokenId/state` | Blue Coinステート（バーン済み、ロック済み） |
| GET | `/api/v1/pool/:tokenId` | AMMプール情報 |
| GET | `/api/v1/validators` | バリデータセット |
| GET | `/api/v1/multisig/:address` | マルチシグアカウント |
| POST | `/api/v1/tx/submit` | トランザクション送信 |
| GET | `/api/v1/tx/:hash` | トランザクションレシート |

---

## アーキテクチャ

```
wblue (単一バイナリ)
 ├── Consensus    スロットローテーション付きPoS、15秒ブロック
 ├── State        アカウント、残高、バリデータセット
 ├── Storage      BoltDB（組み込みキーバリューストア）
 ├── Token        Blue Coinデプロイ、ベスティング、バーン
 ├── AMM          2%バーン付き定積スワップ
 ├── Multisig     N-of-Mオンチェーンマルチシグウォレット
 ├── P2P          libp2p + GossipSub + mDNS
 ├── API          レート制限 + CORS付きREST API
 ├── Explorer     React + TypeScript + Tailwind CSS
 └── CLI          Cobraコマンド
```

---

## 主要パラメータ

| パラメータ | 値 |
|-----------|----------|
| White Coin総供給量 | 1,000,000,000 |
| ブロック間隔 | 15秒 |
| ブロック報酬 | 50 WC（年10%減衰） |
| Blue Coin供給量（トークンあたり） | 1,000,000 |
| スワップバーン率 | 取引あたりBlue Coinの2% |
| トランザクション手数料 | max(0.001 WC, 金額の0.1%) |
| バリデータ参加 | 無料（24時間 + PoW） |
| 自動ステーキング上限 | 1,000 WC |
| 停止しきい値 | オフライン24時間 |
| 除外しきい値 | オフライン72時間 |
| ジェネシスプレマイン | 10,000 WC |

---

## ロードマップ

- [x] デュアルトークンPoSブロックチェーン（White Coin + Blue Coin）
- [x] 定積公式 + 2%バーン付きAMM
- [x] スロットローテーション付きマルチバリデータPoS
- [x] バリデータエコノミクス（無料参加、自動ステーク、スラッシング）
- [x] N-of-Mマルチシグウォレット
- [x] 毎月リリース付きチームベスティング
- [x] P2Pネットワーキング（libp2p + GossipSub）
- [x] ブロックエクスプローラー + Webウォレット
- [x] APIレート制限、CORS、入力バリデーション
- [x] 構造化ログ（slog）
- [x] テストネットデプロイ
- [ ] ホームノード向けNATトラバーサル
- [ ] アドレストランザクション履歴インデックス
- [ ] モバイルウォレットアプリ

---

## ライセンス

MIT
