# White & Blue Protocol

A blockchain built for enterprise token issuance. Any enterprise can issue its own token (Blue Coin) and trade it against the chain's native currency (White Coin) via an on-chain AMM.

一条专为企业发币而生的区块链。任何企业可发行自己的代币（蓝币），通过链上 AMM 与原生货币（白币）交易。

---

## Overview / 概述

**White Coin (WC)** - Native chain token, produced by PoS block rewards. Total supply: 1 billion.

**Blue Coin** - Enterprise tokens. Each enterprise issues one, fixed supply of 1,000,000 per token. Traded via independent AMM pools against White Coin.

**白币** - 链原生代币，PoS 出块产生，总量 10 亿。

**蓝币** - 企业代币。每家企业一种，固定总量 100 万，通过独立 AMM 池子与白币交易。

---

## Quick Start / 快速开始

### Build / 编译

```bash
git clone https://github.com/white-blue-protocol/wblue.git
cd wblue
make build
```

### Run Node / 启动节点

```bash
# Validator node (default) / 验证者节点（默认）
./wblue start

# Full node (non-validator) / 全节点（不出块）
./wblue start --no-validator

# Single-node mode (no P2P, backward compatible) / 单节点模式（无 P2P，向后兼容）
./wblue start --no-p2p
```

This will:
- Generate a validator address (auto-saved to wallet)
- Create the genesis block (validator receives 10,000 WC premine)
- Start producing blocks every 15 seconds
- Start HTTP API on port 8080
- Start P2P networking on port 30303 (with mDNS discovery)

启动后将：
- 自动生成验证者地址（保存到钱包）
- 创建创世块（验证者获得 10,000 白币预挖）
- 每 15 秒出一个块
- 在 8080 端口启动 HTTP API
- 在 30303 端口启动 P2P 网络（支持 mDNS 自动发现）

### Multi-Node Setup / 多节点部署

```bash
# Terminal 1: Start validator node A / 启动验证者节点 A
./wblue start --api-port 8080 --p2p-port 30303

# Terminal 2: Start full node B (auto-discovers A via mDNS on LAN)
# 启动全节点 B（局域网内通过 mDNS 自动发现 A）
./wblue start --no-validator --api-port 8081 --p2p-port 30304 --data-dir ~/.wblue/data2

# Or connect to a remote seed node / 或连接远程种子节点
./wblue start --no-validator --seeds "/ip4/1.2.3.4/tcp/30303/p2p/12D3KooW..."
```

---

## Commands / 命令

### Node / 节点

```bash
wblue start                              # Start validator + P2P / 启动验证者+P2P
wblue start --no-validator               # Full node, no block production / 全节点不出块
wblue start --no-p2p                     # Single node, no networking / 单节点无网络
wblue start --api-port 8081              # Custom API port / 自定义 API 端口
wblue start --p2p-port 30304             # Custom P2P port / 自定义 P2P 端口
wblue start --seeds "/ip4/.../p2p/..."   # Connect to seed nodes / 连接种子节点
wblue start --mdns=false                 # Disable mDNS discovery / 禁用 mDNS
```

### Wallet / 钱包

```bash
wblue wallet create              # Create new wallet / 创建钱包
wblue wallet list                # List wallets / 列出钱包
wblue wallet info <address>      # Show balance / 查看余额
```

### Transfer / 转账

```bash
wblue transfer white --from <addr> --to <addr> --amount 100
wblue transfer blue --from <addr> --to <addr> --token <tokenId> --amount 500
```

### Blue Coin / 蓝币

```bash
# Deploy / 发币
wblue bluecoin deploy \
  --from <addr> \
  --name "FooCoffee" \
  --symbol "FOO" \
  --pool-ratio 20 \
  --team-ratio 80 \
  --init-white 1000 \
  --release-monthly 20000 \
  --multisig <addr>

# Query / 查询
wblue bluecoin list              # List all blue coins / 列出所有蓝币
wblue bluecoin info <tokenId>    # Show details / 查看详情
```

### AMM Trading / AMM 交易

```bash
# Buy blue coin with white coin / 用白币买蓝币
wblue amm swap --from <addr> --token <tokenId> --direction white-to-blue --amount-in 100

# Sell blue coin for white coin / 卖蓝币换白币
wblue amm swap --from <addr> --token <tokenId> --direction blue-to-white --amount-in 500

# Check pool / 查看池子
wblue amm pool-info <tokenId>
wblue amm price <tokenId>
```

### Chain / 链查询

```bash
wblue chain status               # Chain status / 链状态
```

### Global Flags / 全局参数

```bash
wblue --api-url http://localhost:8081 chain status   # Connect to a different node / 连接不同节点
wblue --data-dir /path/to/data wallet list           # Custom data dir / 自定义数据目录
```

---

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/chain/status` | Chain status / 链状态 |
| GET | `/api/v1/chain/block/:height` | Get block / 查区块 |
| GET | `/api/v1/wallet/:address` | Get balance / 查余额 |
| GET | `/api/v1/bluecoin` | List blue coins / 列出蓝币 |
| GET | `/api/v1/bluecoin/:tokenId` | Blue coin info / 蓝币详情 |
| GET | `/api/v1/pool/:tokenId` | Pool info / 池子信息 |
| POST | `/api/v1/tx/submit` | Submit transaction / 提交交易 |

---

## P2P Networking / P2P 网络

Built on [libp2p](https://libp2p.io/) with GossipSub for message propagation.

基于 [libp2p](https://libp2p.io/) 构建，使用 GossipSub 进行消息传播。

### Features / 功能

- **Block broadcast**: new blocks propagate to all connected nodes / 新区块广播到所有节点
- **Transaction relay**: submitted transactions forwarded across the network / 提交的交易在全网转发
- **Block sync**: new nodes automatically download historical blocks from peers / 新节点自动从对等节点同步历史区块
- **mDNS discovery**: automatic peer discovery on local network / 局域网自动发现节点
- **Seed nodes**: bootstrap from known nodes on the internet / 通过已知种子节点接入网络

### Protocol / 协议

| Component | Details |
|-----------|---------|
| GossipSub topics | `wblue/blocks/1`, `wblue/txs/1` |
| Sync stream | `/wblue/sync/1.0.0` (JSON newline-delimited) |
| Node identity | Ed25519 key (persisted at `<dataDir>/node.key`) |
| Default port | 30303 |
| Max peers | 50 (managed by ConnManager) |
| Message size limit | 1 MB |

---

## How It Works / 工作原理

### Block Production / 出块

- PoS consensus, 15-second block interval
- Block reward: 50 WC (decreases 10% annually)
- Transaction fee: max(0.001 WC, 0.1% of amount), burned
- AMM swap fee: 0.1% (built into pool math, burned)

- PoS 共识，15 秒出块
- 区块奖励：50 白币（每年衰减 10%）
- 交易手续费：max(0.001 WC, 金额×0.1%)，销毁
- AMM 交易费：0.1%（内建于池子数学，销毁）

### Blue Coin Issuance / 蓝币发行

- Fixed supply: 1,000,000 per blue coin
- Parameters set at deploy, immutable after
- Pool ratio: % of supply that goes into AMM pool
- Team ratio: % locked, released monthly to team address
- Deployer must inject White Coins to seed the pool

- 固定总量：每种蓝币 100 万
- 参数部署时设定，之后不可更改
- 池子比例：多少进 AMM 池子
- 团队比例：锁仓，按月释放到团队地址
- 发币者需注入白币作为池子初始流动性

### AMM Pool / AMM 池子

- Constant product formula: `x * y = k`
- Each blue coin has an independent pool
- Price determined by supply and demand
- 0.1% swap fee burned
- Pool drains to zero = natural death, no delisting needed

- 恒定乘积公式：`x * y = k`
- 每种蓝币独立池子
- 价格由供需自动决定
- 0.1% 交易费销毁
- 池子归零 = 自然死亡，无需退市

---

## Architecture / 架构

```
wblue (single binary)
├── P2P Network (libp2p, GossipSub, mDNS, block sync)
├── Consensus (PoS, 15s blocks)
├── State Machine (accounts, balances, pools)
├── Storage (BoltDB)
├── Token Engine (deploy, vesting)
├── AMM Engine (constant product swap)
├── HTTP API (:8080)
└── CLI (cobra)
```

---

## Key Parameters / 关键参数

| Parameter | Value |
|-----------|-------|
| White Coin total supply / 白币总量 | 1,000,000,000 |
| Block interval / 出块间隔 | 15 seconds |
| Initial block reward / 初始出块奖励 | 50 WC |
| Annual decay / 年衰减率 | 10% |
| Blue Coin supply (per token) / 蓝币总量 | 1,000,000 |
| Transaction fee / 交易手续费 | max(0.001 WC, 0.1%) burned |
| AMM swap fee / AMM 交易费 | 0.1% burned |
| Genesis premine / 创世预挖 | 10,000 WC |
| P2P default port / P2P 默认端口 | 30303 |
| API default port / API 默认端口 | 8080 |

---

## Roadmap

- [x] Single-node prototype / 单节点原型
- [x] White Coin (PoS mining) / 白币（PoS 挖矿）
- [x] Blue Coin issuance / 蓝币发行
- [x] AMM trading / AMM 交易
- [x] Vesting schedule / 锁仓释放
- [x] CLI + HTTP API
- [x] P2P networking / P2P 网络
- [ ] Multi-validator PoS / 多验证者 PoS
- [ ] Web UI / 网页界面
- [ ] Smart contract VM / 智能合约虚拟机

---

## License

MIT
