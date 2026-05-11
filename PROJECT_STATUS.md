# White-Blue Protocol 项目总结

## 项目概述

一条专为企业发币而生的区块链。白皮书：~/white_blue_protocol_whitepaper.pdf（PDF 需用 pypdf 读取）。

**仓库**：/Users/guoyixuan03/white-blue-protocol
**远程**：https://github.com/Hataka-Mori/white-blue-protocol.git (main 分支)

---

## 核心概念

- **白币 (White Coin, WC)**：链原生货币，PoS 出块产生，总量 10 亿，年衰减 10%
- **蓝币 (Blue Coin)**：企业代币，每种固定 100 万枚，通过独立 AMM 池子(x*y=k)与白币交易
- **协议哲学**：无运营主体、不触碰法币、不做撮合、规则固定、无治理

---

## 当前已完成 (main 分支)

### 功能模块
- PoS 共识（单验证者，15s 出块）
- 白币挖矿（50WC 初始奖励，年衰减 10%）
- 蓝币发行（deploy + vesting 锁仓释放）
- AMM 交易（恒定乘积，0.1% 费用销毁）
- 白币/蓝币转账（手续费 max(0.001WC, 0.1%) 销毁）
- CLI 全部命令 + HTTP API (7 个端点)
- P2P 网络（libp2p + GossipSub + mDNS + 区块同步）
- 非验证者全节点模式 (--no-validator)
- 可配置端口 (--api-port, --p2p-port, --api-url)

### 技术栈
- Go 1.26 + Cobra CLI + BoltDB + libp2p + GossipSub
- ECDSA P-256 签名
- JSON 序列化

### 关键参数
```
MaxWhiteSupply  = 1,000,000,000 (10亿)
BlockInterval   = 15 seconds
InitialReward   = 50 WC/block
AnnualDecayRate = 10%
BlueCoinFixedSupply = 1,000,000/token
FeeRate         = 0.1% (min 0.001 WC)
GenesisPremine  = 10,000 WC
P2P Port        = 30303
API Port        = 8080
```

---

## 下一步：Mainnet Beta 升级 (v2 — 三轮 Review 后)

### 详细方案文档位置
`/Users/guoyixuan03/.takumi/plans/parsed-jumping-handler.md`

### 三轮 Review 关键发现 (38 项)
- R1: 11 个原始 MVP 问题全部确认
- R2: 20 项代码级新发现（uint64 溢出、AMM K 不验证、内存安全、P2P DoS 等）
- R3: 18 项设计级新发现（slot 偷槽、无 fork choice、验证者集排序、心跳垄断等）

### 待做事项总览

| Phase | 内容 | 是否破坏性 | 预估 LOC |
|-------|------|-----------|---------|
| 0 | 原子区块应用 + SafeMath + 错误传播 + Nonce 集中化 + Speculative Validation | 否 | ~400 |
| 1A | 强制签名验证 + Hash 重算 + Low-S 规范化 | **是** | ~150 |
| 1B | 区块签名 + 创世时间戳硬编码 | **是** | ~150 |
| 2A | 钱包加密 Keystore（scrypt + AES-256-GCM）| 否 | ~260 |
| 2B | 交易回执 + 交易级失败机制 | 否 | ~200 |
| 2C | Mempool 修复（known map + 切片 GC）| 否 | ~10 |
| 2D | 网络加固（HTTP 错误 + P2P 限流 + FloodPublish 移除 + 私钥不打印）| 否 | ~60 |
| 2E | 死代码清理 + TxVestingUnlock 删除 + 未用字段移除 | 否 | ~40 |
| 3 | 多验证者 PoS（含逐级跳过、单块 reorg、心跳自证、申诉窗口）| **是** | ~1100 |
| 4A | Swap 滑点保护 (--min-out) | 否 | ~40 |
| 4B | 真实多签 N-of-M（含 domain separator）| 否 | ~340 |
| 4C | 蓝币转账手续费修复（固定 MinFee）| 否 | ~10 |

**总计 ~2,760 LOC 新增 / ~700 LOC 修改。10 周实施周期。**

### 多验证者关键设计 (Scheme F v2)

- **入场**：连续在线 24h（5760 块）→ 自动成为验证者
- **质押**：余额 ≥ 1000 WC → 自动锁定 stake；心跳需 ≥ 100 WC（防 Sybil）
- **出块**：按加入时间排序（同 JoinHeight 按地址字典序），逐级跳过出块
- **掉线**：逐级跳过（每 15s 轮一人），24h 无心跳可被 evict
- **退出**：7 天解锁后取回质押
- **Eviction 申诉**：被举报后有 2h（480 块）申诉窗口可反证
- **心跳自证**：候选人可在 TxValidatorJoin Payload 中自带签名心跳，不依赖出块者打包
- **验证者集生效**：区块 H 中的变更从 H+1 生效
- **分叉选择**：同高度冲突块取最低 block hash 胜出（单块 reorg）
- **确认数**：10 块确认（~2.5 分钟）
- **确定性共识**：On-Chain Heartbeat + 确定性排序
- **冷启动兜底**：活跃验证者 < 3 或高度 < 5760 时放宽 24h 要求

### 破坏性变更
Phase 1A + 1B + 3 需要清空 genesis，chainID 升级为 `wblue-mainnet-beta-1`。

### 实施顺序
```
Week 1:  Phase 0（原子化 + SafeMath + 错误传播 + nonce + speculative validation）
Week 2:  Phase 2C + 2D + 2E（mempool + 网络加固 + 死代码清理）
Week 3:  Phase 1A + 1B → 重置 testnet
Week 4:  Phase 2A + 2B（keystore + 交易回执）
Week 5:  Phase 3A + 3B（validator types + heartbeat + 自证）
Week 6:  Phase 3C + 3D（状态机 + 逐级跳过出块）
Week 7:  Phase 3E + 3F（分叉选择 + 确认数）+ 集成测试
Week 8:  Phase 4A + 4C（滑点保护 + fee 修复）
Week 9:  Phase 4B（多签）
Week 10: 全面审计 → Mainnet Beta 发布
```

---

## 项目结构

```
white-blue-protocol/
├── main.go
├── Makefile
├── cmd/
│   ├── root.go          — CLI entry, start 命令, 全局 flags
│   ├── wallet.go        — wallet create/list/info
│   ├── transfer.go      — transfer white/blue
│   ├── amm.go           — amm swap/pool-info/price
│   ├── bluecoin.go      — bluecoin deploy/list/info
│   ├── chain.go         — chain status
│   └── helpers.go       — loadWalletByAddress
├── internal/
│   ├── node/node.go     — Node 生命周期, Config, P2P 集成
│   ├── consensus/pos.go — PoS 出块引擎 (BlockCh)
│   ├── chain/
│   │   ├── blockchain.go — CreateBlock, ApplyBlock, VerifyBlockHash/MerkleRoot
│   │   └── rewards.go   — CalcReward (年衰减)
│   ├── state/state.go   — ValidateTransaction, ApplyTransaction (7 types)
│   ├── storage/
│   │   ├── db.go        — BoltDB wrapper
│   │   ├── block_repo.go — SaveBlock, GetBlockByHeight/Hash
│   │   └── state_repo.go — Account, Pool, BlueCoin persistence
│   ├── txpool/mempool.go — Mempool (OnAdd callback, RemoveTxs, MaxSize=10000)
│   ├── token/
│   │   ├── deploy.go    — BlueCoin deploy (接受 blockTime 参数)
│   │   └── vesting.go   — 按月自动释放
│   ├── amm/
│   │   ├── executor.go  — ExecuteSwap (直接写 DB)
│   │   └── math.go      — GetAmountOut (x*y=k + 0.1% fee)
│   ├── crypto/
│   │   ├── keypair.go   — ECDSA P-256
│   │   ├── address.go   — SHA256(pubkey)[:20]
│   │   ├── hash.go      — SHA256Hex, MerkleRoot
│   │   └── sign.go      — Sign, Verify
│   ├── api/server.go    — HTTP API (7 endpoints)
│   └── p2p/
│       ├── host.go      — libp2p Host, GossipSub, 区块同步, 验证
│       ├── messages.go  — Envelope, BlockMsg, TxMsg, StatusMsg, SyncReq/Res
│       ├── discovery.go — mDNS + seed 节点拨号
│       └── dedup.go     — ring-buffer 去重
└── scripts/demo.sh      — 完整功能演示脚本
```

---

## 已知 MVP 问题（三轮 Review 汇总：38 项）

### R1: 原始 11 项
1. **签名验证空隙** — `state.go:83` 无 publicKey 时跳过验签
2. **区块无签名** — 任何人可伪造区块
3. **钱包明文** — 私钥 JSON 明文存盘
4. **状态非原子** — ApplyBlock 中途崩溃状态不一致
5. **无交易回执** — 提交后不知道确认状态
6. **Swap 无滑点保护** — 可能被抢跑
7. **蓝币手续费不合理** — CalcFee(blueAmount) 逻辑错误
8. **多签未实现** — 只存了地址字段
9. **Mempool known map 泄漏** — 只增不减
10. **HTTP server 吞错误** — 端口冲突时静默失败
11. **TxVestingUnlock bug** — ApplyTransaction 无该 case

### R2: 代码级 20 项
12. uint64 溢出可盗币
13. AMM K 不验证 newK ≥ oldK
14. big.Int → uint64 截断未检查
15. GetBlockByHash 悬垂指针
16. P2P sync 无超时/限流
17. 私钥打印到控制台
18. tx hash 不重算
19. VerifyWithAddress 从未调用
20. Nonce 递增分散两处
21. 创世时间戳非确定性
22. FloodPublish 带宽放大
23. token deploy K 溢出
24. ProcessVesting 错误丢弃
25. SetTotalMinted 错误丢弃
26. json.Marshal 错误被吞
27. vesting 全量扫描
28. account 双重读取
29. RemoveTxs 切片阻 GC
30. 死代码/未用字段
31. validator[:10] panic

### R3: 设计级 18 项
32. Slot 公式允许偷槽
33. 无 fork choice → 永久分裂
34. 验证者集排序无二级键
35. 验证者集生效时机未定
36. 心跳打包可被垄断
37. Eclipse + Eviction 武器化
38. ApplyBlock 不 re-validate
39. 同 sender 多 tx 一致性
40. tx 失败崩溃整块
41. VerifyBlockHash 需清零 Signature
42. ECDSA 签名可塑性
43. 多签地址命名空间冲突
44. 心跳 Sybil
45. MEV / 三明治攻击
46. TxType 编号冲突
47. 私钥 padding 不足
48. 无 finality 概念
49. Reward decay 无 cap

---

## 使用方法

```bash
# 编译
make build

# 启动验证者节点
./wblue start

# 启动全节点（不出块，自动同步）
./wblue start --no-validator --api-port 8081 --p2p-port 30304 --data-dir ~/.wblue/data2

# 单节点模式（无 P2P，向后兼容）
./wblue start --no-p2p

# CLI 连接指定节点
./wblue --api-url http://localhost:8081 chain status
```

---

## 白皮书核心设计要点

- 买蓝币 = 投资企业，持有蓝币 = 消费 + 分享增长
- 企业发币参数部署后永不可改
- 池子里的白币不归企业，谁都拿不走
- 池子归零 = 企业自然死亡
- 协议只提供规则，不做裁判
- 无运营主体、不触碰法币、不做撮合、不存管、不审核
