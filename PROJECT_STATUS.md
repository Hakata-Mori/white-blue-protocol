# White-Blue Protocol 项目状态

## 项目概述

一条专为企业发币而生的区块链。任何人可以通过购买企业蓝币来"投资"任何组织。

- **仓库**: https://github.com/Hataka-Mori/white-blue-protocol
- **测试网**: http://8.217.52.231 (阿里云香港)
- **Chain ID**: `wblue-testnet-1`

---

## 核心概念

- **白币 (White Coin, WC)**: 链原生货币，PoS 出块产生，总量 10 亿，年衰减 10%
- **蓝币 (Blue Coin)**: 用户部署的代币，每种固定 100 万枚，通过独立 AMM 池 (x*y=k) 与白币交易
- **协议哲学**: 无运营主体、不触碰法币、不做撮合、规则固定、无治理

---

## 当前状态: Mainnet Beta (已部署)

所有计划中的 12 个升级阶段 (Phase 0 ~ 4C) 均已完成。链运行在阿里云香港服务器上。

### 已完成功能

**基础设施 (Phase 0)**
- 原子 BoltDB 事务 (ApplyBlock 全部在单个 Bolt Tx 内完成)
- SafeMath (溢出安全的 uint64 运算)
- 错误传播 (消除所有吞掉的 error)
- Nonce 集中化管理
- Speculative Validation (提交时预检)
- 交易回执 (Receipt)

**安全性 (Phase 1A/1B)**
- 强制交易签名验证
- 区块签名 (出块者 ECDSA 签名)
- Low-S 规范化 (消除签名可塑性)

**加固 (Phase 2A~2E)**
- Keystore 加密存储 (scrypt + AES-256-GCM)
- 交易回执查询 API
- Mempool known map GC 修复
- P2P 限流、FloodPublish 移除、私钥不打印
- 死代码清理、TxVestingUnlock 删除

**多验证者 PoS (Phase 3)**
- ValidatorSet 管理 (加入/退出/暂停/驱逐)
- 心跳自证机制
- Slot 轮转出块
- 5 种验证者交易类型 (Join/Exit/Evict/Heartbeat/SlashEvidence)
- 24h + PoW 入场门槛

**经济模型 (Phase 4A~4C + Validator Economics V3)**
- Swap 滑点保护 (--min-out)
- 真实多签 N-of-M (含执行)
- 蓝币转账手续费修复
- 动态质押: `DynamicStakeAmount = BlockReward × BlocksPerDay ÷ activeCount × StakeDays(1)`，最低 10,000 WC
- 蓝币 2% swap 销毁 + 手动 TxBlueBurn
- 验证者出块分润: 50% 交易手续费给出块者
- 免费加入 (24h 在线 + PoW)，退出销毁质押
- 双层离线惩罚 (暂停 → 驱逐)，双签永久封禁

**区块浏览器 + 钱包**
- React + TypeScript + Tailwind CSS + Vite 前端
- 11 个页面: 首页、区块列表、区块详情、交易详情、地址、蓝币列表、蓝币详情、验证者、钱包、水龙头、404
- 钱包: 创建/导入 keystore、转账 WC/Blue、Swap (AMM)、部署蓝币
- 签名: @noble/curves (P-256)，加密: @noble/ciphers (AES-256-GCM)，KDF: scrypt-js

**水龙头**
- 独立钱包地址 `0xde7a...c37`
- 每地址每 24h 领取 100 WC
- BoltDB 持久化冷却记录

**部署 & 运维**
- 硬编码种子节点 (用户直接 `./wblue start` 即可连接)
- `--genesis` 隐藏 flag (防止意外创建新链)
- Config 文件支持、版本嵌入、结构化日志 (slog)
- API 限流、CORS、输入校验、Mempool TTL/per-addr 限制
- systemd 服务 + Nginx 反向代理
- README 四语版本 (English / 中文 / 日本語 / 한국어)

---

## 链上参数

| 参数 | 值 | 说明 |
|------|-----|------|
| MaxWhiteSupply | 1,000,000,000 WC | 10 亿 |
| BlockInterval | 15 秒 | |
| InitialReward | 50 WC/块 | |
| AnnualDecayRate | 10% | |
| FeeRate | 0.1% (min 0.001 WC) | 转账手续费 |
| BlueCoinFixedSupply | 1,000,000/种 | |
| BlueBurnRate | 2% | AMM swap 销毁 |
| GenesisPremine | 10,000 WC | |
| MinStakeAmount | 10,000 WC | 动态质押下限 |
| UptimeBlocks | 5,760 (24h) | 入场在线要求 |
| SuspendBlocks | 5,760 (24h) | 离线暂停阈值 |
| EvictBlocks | 17,280 (72h) | 驱逐阈值 |
| PoWDifficulty | 24 bits | 入场 PoW |
| ConfirmationBlocks | 10 (~2.5 min) | |

---

## 交易类型

| TxType | 编号 | 说明 |
|--------|------|------|
| Transfer | 0 | 白币转账 |
| BlueDeploy | 1 | 部署蓝币 |
| BlueTransfer | 2 | 蓝币转账 |
| Swap | 3 | AMM 交易 |
| Reward | 5 | 出块奖励 |
| Fee | 6 | 手续费 |
| Heartbeat | 8 | 验证者心跳 |
| ValidatorJoin | 9 | 加入验证者 |
| ValidatorExit | 10 | 退出验证者 |
| ValidatorEvict | 11 | 驱逐验证者 |
| SlashEvidence | 13 | 双签举报 |
| BlueBurn | 14 | 蓝币销毁 |
| MultiSigRegister | 20 | 注册多签钱包 |
| MultiSigPropose | 21 | 多签提案 |
| MultiSigApprove | 22 | 多签批准 |

---

## API 端点 (16 个)

| Method | Endpoint | 说明 |
|--------|----------|------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/chain/status` | 链状态 |
| GET | `/api/v1/chain/block/:height` | 按高度查块 |
| GET | `/api/v1/blocks?limit=&offset=` | 区块列表 |
| GET | `/api/v1/block/hash/:hash` | 按哈希查块 |
| GET | `/api/v1/stats` | 网络统计 |
| GET | `/api/v1/wallet/:address` | 账户余额 |
| GET | `/api/v1/bluecoin` | 蓝币列表 |
| GET | `/api/v1/bluecoin/:tokenId` | 蓝币配置 |
| GET | `/api/v1/bluecoin/:tokenId/state` | 蓝币状态 |
| GET | `/api/v1/pool/:tokenId` | AMM 池信息 |
| GET | `/api/v1/validators` | 验证者集 |
| GET | `/api/v1/multisig/:address` | 多签账户 |
| POST | `/api/v1/tx/submit` | 提交交易 |
| GET | `/api/v1/tx/:hash` | 交易回执 |
| POST | `/api/v1/faucet` | 领取测试币 |

---

## CLI 命令

```
wblue start                          启动节点
wblue version                        版本信息

wblue wallet create                  创建钱包
wblue wallet list                    列出钱包
wblue wallet info [address]          查询余额

wblue transfer white                 白币转账
wblue transfer blue                  蓝币转账

wblue bluecoin deploy                部署蓝币
wblue bluecoin list                  列出蓝币
wblue bluecoin info [tokenId]        蓝币详情
wblue bluecoin burn                  销毁蓝币

wblue amm swap                       AMM 交易
wblue amm pool-info [tokenId]        池子信息
wblue amm price [tokenId]            当前价格

wblue validator join                 加入验证者
wblue validator exit                 退出验证者
wblue validator status               验证者状态
wblue validator heartbeat            发送心跳

wblue multisig register              注册多签
wblue multisig propose               多签提案
wblue multisig approve               批准提案
wblue multisig info [address]        多签信息

wblue chain status                   链状态
wblue chain tx [hash]                查询交易
```

---

## 项目结构

```
white-blue-protocol/
├── main.go
├── Makefile
├── cmd/
│   ├── root.go              CLI 入口, start/version 命令
│   ├── wallet.go            钱包命令
│   ├── transfer.go          转账命令
│   ├── bluecoin.go          蓝币命令
│   ├── amm.go               AMM 命令
│   ├── validator.go         验证者命令
│   ├── multisig.go          多签命令
│   ├── chain.go             链查询命令
│   └── helpers.go           辅助函数
├── internal/
│   ├── node/node.go         节点生命周期
│   ├── consensus/pos.go     PoS 出块引擎
│   ├── chain/
│   │   ├── blockchain.go    区块创建/验证/应用
│   │   └── rewards.go       奖励计算 (年衰减)
│   ├── state/
│   │   ├── state.go         交易执行/验证 (15 种类型)
│   │   ├── validator.go     验证者状态管理
│   │   └── multisig.go      多签状态管理
│   ├── storage/
│   │   ├── db.go            BoltDB 封装
│   │   ├── block_repo.go    区块持久化
│   │   ├── state_repo.go    账户/池/蓝币持久化
│   │   ├── validator_repo.go 验证者持久化
│   │   └── multisig_repo.go 多签持久化
│   ├── txpool/mempool.go    交易池
│   ├── token/
│   │   ├── deploy.go        蓝币部署
│   │   └── vesting.go       团队代币释放
│   ├── amm/
│   │   ├── executor.go      Swap 执行 (2% 销毁)
│   │   └── math.go          AMM 数学 (x*y=k)
│   ├── crypto/
│   │   ├── keypair.go       ECDSA P-256
│   │   ├── sign.go          签名/验签
│   │   ├── hash.go          SHA-256 / Merkle
│   │   └── address.go       地址派生
│   ├── keystore/keystore.go AES-256-GCM 加密钱包
│   ├── safemath/safemath.go 溢出安全运算
│   ├── api/
│   │   ├── server.go        REST API (16 端点)
│   │   └── faucet.go        水龙头服务
│   ├── p2p/
│   │   ├── host.go          libp2p + GossipSub
│   │   ├── messages.go      消息编解码
│   │   ├── discovery.go     mDNS + 种子节点
│   │   └── dedup.go         去重缓冲
│   ├── log/log.go           结构化日志 (slog)
│   ├── config/config.go     节点配置
│   ├── version/version.go   版本信息
│   └── types/
│       ├── block.go         区块结构
│       ├── transaction.go   交易类型定义
│       ├── account.go       账户结构
│       ├── bluecoin.go      蓝币结构
│       ├── validator.go     验证者结构 + 参数
│       ├── pool.go          AMM 池结构
│       ├── genesis.go       链参数 + 费用计算
│       ├── multisig.go      多签结构
│       ├── receipt.go       回执结构
│       └── devmode.go       开发模式配置
├── explorer/                React 前端 (区块浏览器 + 钱包)
│   └── src/
│       ├── pages/           11 个页面
│       ├── components/      UI/布局/钱包组件
│       ├── api/             API 客户端
│       ├── hooks/           React Hooks
│       ├── lib/wallet.ts    钱包加密工具
│       └── types/           TypeScript 类型
├── scripts/
│   └── full_integration_test.sh  集成测试 (23 项)
└── README.md (+README_zh/ja/ko.md)
```

---

## 技术栈

**后端**: Go 1.26 / Cobra CLI / BoltDB / libp2p + GossipSub / ECDSA P-256
**前端**: React 18 / TypeScript / Tailwind CSS / Vite / @noble/curves + @noble/ciphers
**部署**: Ubuntu (阿里云) / systemd / Nginx

---

## 测试

- **单元测试**: 12 个测试文件，139+ 测试用例 (`go test ./... -count=1`)
- **集成测试**: 23 项 (`bash scripts/full_integration_test.sh`)

---

## 测试网信息

| 项目 | 值 |
|------|-----|
| Chain ID | `wblue-testnet-1` |
| 服务器 | 阿里云香港 8.217.52.231 |
| 浏览器 | http://8.217.52.231 |
| API | http://8.217.52.231/api/v1/ |
| P2P 端口 | 30303 |
| Peer ID | `12D3KooWFp3UcCexRsAbxQ81DtMRedQGwG5hmv1QjP2FzUCXHd8S` |
| 验证者 | `0xd445b0e5352460b92bcd15e47e0f66174c430c38` |
| 水龙头 | 每地址每 24h 100 WC |

---

## 已知限制

1. 非活跃候选人在 activeCount >= 3 时无法发送心跳 (需有人先退出)
2. 无分叉选择规则 / 链重组 (PoS slot 轮转使分叉极少发生)
3. 无地址交易历史索引 (仅支持按区块查询)
4. 无助记词备份

---

## 后续计划

| 优先级 | 内容 |
|--------|------|
| 高 | 推广测试网 / 寻找测试用户 |
| 中 | 地址交易历史索引 + API |
| 中 | P2P 增强 (DHT / NAT 穿透) |
| 中 | 共识容错 (跳槽 / 分叉选择) |
| 低 | 助记词备份 |
