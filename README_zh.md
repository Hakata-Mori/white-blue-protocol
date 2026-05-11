[English](README.md) | [中文](README_zh.md) | [日本語](README_ja.md) | [한국어](README_ko.md)

# 白蓝协议 (White & Blue Protocol)

**让任何人都能发行自己的代币。一家咖啡店、一家初创公司、一个家庭——任何人。**

白蓝协议是一条双代币 PoS 区块链，每个组织都可以发行自己的可交易代币（蓝币），通过内置的 AMM 池以链的原生货币（白币）作为支撑。

> 你可以这样理解：**每个组织都能获得类似股票的代币，即时交易，零上市费用。**

---

## 为什么要做这个？

传统募资方式对小型组织来说问题重重：
- IPO 需要数百万资金，耗时数年
- 加密代币发行需要智能合约专业知识
- 众筹平台抽成巨大，投资者却得不到任何回报

**白蓝协议解决了这个问题。** 任何组织——从社区咖啡店到科技初创公司——都可以通过一笔交易发行自己的蓝币。支持者购买蓝币来表达支持，并有机会随着组织的成长而获利。

---

## 真实场景示例

### 一家咖啡店

拿铁实验室要开新店，需要社区支持：

1. 发行 **拿铁币 (LTC)**，70% 注入 AMM 池，30% 锁定给团队
2. 注入 500 WC 作为初始流动性
3. 老顾客通过内置兑换功能购买拿铁币
4. 随着越来越多人买入，价格自然上涨（恒定乘积 AMM）
5. 团队的 30% 按月释放——用于扩张、购买新设备等
6. 如果咖啡店生意兴隆，早期支持者获利。如果不行，池子自然排空——无需麻烦的退市流程

### 一家科技初创公司

一个三人团队想获得种子资金，但不想稀释股权：

1. 创始人建立一个 **2/3 多签钱包**
2. 发行 **创业币 (STC)**，由多签钱包控制团队分配部分
3. 投资者在 AMM 上购买创业币——价格反映市场信心
4. 团队的任何支出都需要 3 位创始人中的 2 位同意（多签）
5. 每月释放的代币用于资助开发里程碑

### 一个家庭

没错，连家庭也可以：

1. 张家发行 **张家币 (JSN)** 作为一次有趣的家庭实验
2. 家庭成员和朋友参与买入
3. "市值"变成每次家庭聚餐时的一个梗
4. 但实质上，它是一个拥有真实经济机制的代币——AMM 定价、销毁机制，一应俱全

### 发挥你的想象

- 一位视频博主发行 **粉丝币**——支持者投资他的成长
- 一家本地农场发行 **农场币**——买代币，享受农产品折扣
- 一个学校社团发行 **社团币**——透明地募集资金
- 一支乐队发行 **乐队币**——粉丝成为利益相关者

**只要有支持者，就可以拥有蓝币。**

---

## 运作原理

### 双代币体系

| | 白币 (WC) | 蓝币 |
|---|---|---|
| 是什么 | 链原生货币 | 组织专属代币 |
| 供应量 | 10 亿（通过 PoS 挖矿产出） | 每种代币 100 万（固定） |
| 如何获取 | 运行验证节点 | 在 AMM 上购买 / 接收转账 |
| 用途 | Gas 费、流动性、质押 | 投资、社区、忠诚度 |

### 代币经济学

- **AMM 池**：恒定乘积公式（`x * y = k`）。价格随供需变动
- **2% 交易销毁**：每笔交易销毁 2% 的蓝币，产生通缩效应
- **团队锁仓释放**：团队分配按月释放，而非一次性解锁
- **多签控制**：团队资金可通过 N-of-M 多签锁定，确保问责制
- **自然淘汰**：如果无人交易，池子归零。无需退市

### 验证者经济

| 参数 | 值 |
|-----------|----------|
| 加入成本 | 免费（24 小时在线 + 工作量证明） |
| 出块奖励 | 每块 50 WC |
| 出块间隔 | 15 秒 |
| 自动质押 | 奖励锁定为质押，直到达到 1,000 WC |
| 手续费分成 | 50% 的交易费归出块者 |
| 离线惩罚 | 24 小时后暂停，72 小时后驱逐 |
| 双签 | 永久封禁 + 质押销毁 |

---

## 在线测试网

**区块浏览器**：[http://8.217.52.231](http://8.217.52.231)

在浏览器中创建钱包、浏览区块、进行交易：
- 查看实时区块、交易和验证者状态
- 创建钱包并管理代币
- 转账白币和蓝币
- 部署你自己的蓝币代币
- 在内置 AMM 上兑换代币

---

## 快速开始

### 方式一：使用网页浏览器（无需安装）

访问 [http://8.217.52.231](http://8.217.52.231) 并点击 **Wallet** 即可开始。

### 方式二：运行全节点

```bash
git clone https://github.com/Hataka-Mori/white-blue-protocol.git
cd white-blue-protocol
make build

# 启动全节点（自动从测试网同步）
./wblue start --no-validator
```

### 方式三：运行验证节点（赚取 WC）

```bash
./wblue start
```

你的节点将会：
1. 自动生成一个钱包
2. 连接到测试网种子节点
3. 开始 24 小时候选期
4. 24 小时后通过 PoW 验证，开始出块并获得每块 50 WC 的奖励

---

## CLI 命令

### 钱包

```bash
wblue wallet create                # 创建新钱包（含助记词备份）
wblue wallet recover               # 通过助记词恢复钱包
wblue wallet list                  # 列出钱包
wblue wallet info <address>        # 查看余额
```

### 转账

```bash
wblue transfer white --from <addr> --to <addr> --amount 100
wblue transfer blue --from <addr> --to <addr> --token <id> --amount 500
```

### 蓝币

```bash
# 部署新代币
wblue bluecoin deploy \
  --from <addr> \
  --name "LatteCoin" \
  --symbol "LTC" \
  --pool-ratio 70 \
  --team-ratio 30 \
  --init-white 500 \
  --release-monthly 5000 \
  --multisig <multisig-addr>

# 查询
wblue bluecoin list
wblue bluecoin info <tokenId>
wblue bluecoin burn --from <addr> --token <id> --amount 1000
```

### AMM 兑换

```bash
# 用白币购买蓝币
wblue amm swap --from <addr> --token <id> --direction white-to-blue --amount-in 100

# 卖出蓝币换取白币
wblue amm swap --from <addr> --token <id> --direction blue-to-white --amount-in 500

# 查看池子信息
wblue amm pool-info <tokenId>
wblue amm price <tokenId>
```

### 验证者

```bash
wblue validator join --from <addr>
wblue validator exit --from <addr>
wblue validator status
wblue validator heartbeat --from <addr>
```

### 多签

```bash
wblue multisig register --owners addr1,addr2,addr3 --threshold 2
wblue multisig propose --multisig <ms-addr> --to <target> --amount 100
wblue multisig approve --multisig <ms-addr> --proposal-id 0
wblue multisig info --address <ms-addr>
```

### 链信息

```bash
wblue chain status
wblue chain tx <hash>
wblue version
```

---

## HTTP API

| 方法 | 端点 | 说明 |
|--------|----------|-------------|
| GET | `/health` | 健康检查 |
| GET | `/api/v1/chain/status` | 链状态 |
| GET | `/api/v1/chain/block/:height` | 按高度获取区块 |
| GET | `/api/v1/blocks?limit=20&offset=0` | 区块列表（分页） |
| GET | `/api/v1/block/hash/:hash` | 按哈希获取区块 |
| GET | `/api/v1/stats` | 网络统计 |
| GET | `/api/v1/wallet/:address` | 账户余额 |
| GET | `/api/v1/bluecoin` | 列出所有蓝币 |
| GET | `/api/v1/bluecoin/:tokenId` | 蓝币配置 |
| GET | `/api/v1/bluecoin/:tokenId/state` | 蓝币状态（已销毁、已锁定） |
| GET | `/api/v1/pool/:tokenId` | AMM 池信息 |
| GET | `/api/v1/validators` | 验证者集合 |
| GET | `/api/v1/multisig/:address` | 多签账户 |
| GET | /api/v1/address/:address/txs | 地址交易历史 |
| POST | `/api/v1/tx/submit` | 提交交易 |
| GET | `/api/v1/tx/:hash` | 交易回执 |

---

## 架构

```
wblue (单一可执行文件)
 ├── Consensus    PoS 轮转出块，15 秒区块间隔
 ├── State        账户、余额、验证者集合
 ├── Storage      BoltDB（嵌入式键值存储）
 ├── Token        蓝币部署、锁仓释放、销毁
 ├── AMM          恒定乘积兑换 + 2% 销毁
 ├── Multisig     链上 N-of-M 多签钱包
 ├── P2P          libp2p + GossipSub + mDNS
 ├── API          REST API，支持速率限制 + CORS
 ├── Explorer     React + TypeScript + Tailwind CSS
 └── CLI          Cobra 命令行工具
```

---

## 关键参数

| 参数 | 值 |
|-----------|----------|
| 白币总供应量 | 1,000,000,000 |
| 出块间隔 | 15 秒 |
| 出块奖励 | 50 WC（每年衰减 10%） |
| 蓝币供应量（每种代币） | 1,000,000 |
| 兑换销毁率 | 每笔交易销毁 2% 的蓝币 |
| 交易手续费 | max(0.001 WC, 金额的 0.1%) |
| 验证者加入 | 免费（24 小时 + PoW） |
| 自动质押上限 | 1,000 WC |
| 暂停阈值 | 离线 24 小时 |
| 驱逐阈值 | 离线 72 小时 |
| 创世预挖 | 10,000 WC |

---

## 路线图

- [x] 双代币 PoS 区块链（白币 + 蓝币）
- [x] 恒定乘积公式 AMM + 2% 销毁
- [x] 多验证者 PoS 轮转出块
- [x] 验证者经济模型（免费加入、自动质押、惩罚机制）
- [x] N-of-M 多签钱包
- [x] 团队锁仓按月释放
- [x] P2P 网络（libp2p + GossipSub）
- [x] 区块浏览器 + 网页钱包
- [x] API 速率限制、CORS、输入验证
- [x] 结构化日志（slog）
- [x] 测试网部署
- [x] 地址交易历史索引
- [x] 助记词备份（BIP39，12 个单词）
- [ ] 家庭节点 NAT 穿透
- [ ] 移动端钱包应用

---

## 许可证

MIT
