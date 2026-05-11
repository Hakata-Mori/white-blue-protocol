# White & Blue Protocol

**Let anyone issue their own token. A coffee shop, a startup, a family — anyone.**

White & Blue Protocol is a dual-token PoS blockchain where every organization can issue its own tradeable token (Blue Coin) backed by the chain's native currency (White Coin) through built-in AMM pools.

> Think of it as: **Every organization gets its own stock-like token, instantly tradeable, with zero listing fees.**

---

## Why Does This Exist?

Traditional fundraising is broken for small organizations:
- IPOs cost millions and take years
- Crypto token launches require smart contract expertise
- Crowdfunding platforms take huge cuts and give investors nothing in return

**White & Blue Protocol fixes this.** Any organization — from a neighborhood coffee shop to a tech startup — can issue its own Blue Coin in one transaction. Supporters buy Blue Coins to show support and potentially profit as the organization grows.

---

## Real-World Examples

### A Coffee Shop

Latte Lab opens a new location and needs community support:

1. Deploy **LatteCoin (LTC)** with 70% in the AMM pool, 30% locked for the team
2. Inject 500 WC as initial liquidity
3. Regulars buy LatteCoin through the built-in swap
4. As more people buy in, the price naturally rises (constant product AMM)
5. Team's 30% releases monthly — used for expansion, new equipment, etc.
6. If the shop thrives, early supporters profit. If not, the pool drains naturally — no messy delisting

### A Tech Startup

A 3-person startup wants seed funding without giving up equity:

1. The founders set up a **2-of-3 multisig wallet**
2. Deploy **StartupCoin (STC)** with the multisig controlling the team allocation
3. Investors buy StartupCoin on the AMM — the price reflects market confidence
4. Any team spending requires 2 of 3 founders to agree (multisig)
5. Monthly token releases fund development milestones

### A Family

Yes, even a family:

1. The Johnsons deploy **JohnsonCoin (JSN)** as a fun family experiment
2. Family members and friends buy in
3. The "market cap" becomes a running joke at Thanksgiving dinner
4. But underneath, it's a real token with real economics — AMM pricing, burn mechanics, the works

### Your Imagination

- A YouTuber issues **FanCoin** — supporters invest in their growth
- A local farm issues **FarmCoin** — buy tokens, get discounts on produce
- A school club issues **ClubCoin** — fundraise transparently
- A band issues **BandCoin** — fans become stakeholders

**If it has supporters, it can have a Blue Coin.**

---

## How It Works

### Two Tokens

| | White Coin (WC) | Blue Coin |
|---|---|---|
| What | Native chain currency | Organization-specific token |
| Supply | 1 billion (mined via PoS) | 1 million per token (fixed) |
| How to get | Run a validator node | Buy on AMM / receive transfer |
| Purpose | Gas, liquidity, staking | Investment, community, loyalty |

### Token Economics

- **AMM Pool**: Constant product formula (`x * y = k`). Price moves with supply and demand
- **2% Burn on Swap**: Every trade burns 2% of Blue Coins, creating deflation
- **Team Vesting**: Team allocation releases monthly, not all at once
- **Multisig Control**: Team funds can be locked behind N-of-M multisig for accountability
- **Natural Death**: If nobody trades, the pool drains to zero. No delisting needed

### Validator Economics

| Parameter | Value |
|-----------|-------|
| Join cost | Free (24h online + proof-of-work) |
| Block reward | 50 WC per block |
| Block interval | 15 seconds |
| Auto-staking | Rewards lock as stake until 1,000 WC |
| Fee sharing | 50% of tx fees go to block producer |
| Offline penalty | Suspended after 24h, evicted after 72h |
| Double-sign | Permanently banned + stake destroyed |

---

## Live Testnet

**Block Explorer**: [http://8.217.52.231](http://8.217.52.231)

Create a wallet, explore blocks, and trade — all from your browser:
- View real-time blocks, transactions, and validator status
- Create a wallet and manage your tokens
- Transfer WhiteCoin and BlueCoin
- Deploy your own BlueCoin token
- Swap tokens on the built-in AMM

---

## Quick Start

### Option 1: Use the Web Explorer (No Install)

Visit [http://8.217.52.231](http://8.217.52.231) and click **Wallet** to get started.

### Option 2: Run a Full Node

```bash
git clone https://github.com/Hataka-Mori/white-blue-protocol.git
cd white-blue-protocol
make build

# Start a full node (syncs from testnet automatically)
./wblue start --no-validator
```

### Option 3: Run a Validator (Earn WC)

```bash
./wblue start
```

Your node will:
1. Generate a wallet automatically
2. Connect to the testnet seed node
3. Begin the 24-hour candidate period
4. After 24h + PoW verification, start producing blocks and earning 50 WC per block

---

## CLI Commands

### Wallet

```bash
wblue wallet create                # Create new wallet
wblue wallet list                  # List wallets
wblue wallet info <address>        # Show balance
```

### Transfer

```bash
wblue transfer white --from <addr> --to <addr> --amount 100
wblue transfer blue --from <addr> --to <addr> --token <id> --amount 500
```

### Blue Coin

```bash
# Deploy a new token
wblue bluecoin deploy \
  --from <addr> \
  --name "LatteCoin" \
  --symbol "LTC" \
  --pool-ratio 70 \
  --team-ratio 30 \
  --init-white 500 \
  --release-monthly 5000 \
  --multisig <multisig-addr>

# Query
wblue bluecoin list
wblue bluecoin info <tokenId>
wblue bluecoin burn --from <addr> --token <id> --amount 1000
```

### AMM Swap

```bash
# Buy Blue Coin with White Coin
wblue amm swap --from <addr> --token <id> --direction white-to-blue --amount-in 100

# Sell Blue Coin for White Coin
wblue amm swap --from <addr> --token <id> --direction blue-to-white --amount-in 500

# Check pool
wblue amm pool-info <tokenId>
wblue amm price <tokenId>
```

### Validator

```bash
wblue validator join --from <addr>
wblue validator exit --from <addr>
wblue validator status
wblue validator heartbeat --from <addr>
```

### Multisig

```bash
wblue multisig register --owners addr1,addr2,addr3 --threshold 2
wblue multisig propose --multisig <ms-addr> --to <target> --amount 100
wblue multisig approve --multisig <ms-addr> --proposal-id 0
wblue multisig info --address <ms-addr>
```

### Chain

```bash
wblue chain status
wblue chain tx <hash>
wblue version
```

---

## HTTP API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/chain/status` | Chain status |
| GET | `/api/v1/chain/block/:height` | Get block by height |
| GET | `/api/v1/blocks?limit=20&offset=0` | List blocks (paginated) |
| GET | `/api/v1/block/hash/:hash` | Get block by hash |
| GET | `/api/v1/stats` | Network statistics |
| GET | `/api/v1/wallet/:address` | Account balance |
| GET | `/api/v1/bluecoin` | List all Blue Coins |
| GET | `/api/v1/bluecoin/:tokenId` | Blue Coin config |
| GET | `/api/v1/bluecoin/:tokenId/state` | Blue Coin state (burned, locked) |
| GET | `/api/v1/pool/:tokenId` | AMM pool info |
| GET | `/api/v1/validators` | Validator set |
| GET | `/api/v1/multisig/:address` | Multisig account |
| POST | `/api/v1/tx/submit` | Submit transaction |
| GET | `/api/v1/tx/:hash` | Transaction receipt |

---

## Architecture

```
wblue (single binary)
 ├── Consensus    PoS with slot rotation, 15s blocks
 ├── State        Accounts, balances, validator set
 ├── Storage      BoltDB (embedded key-value store)
 ├── Token        Blue Coin deploy, vesting, burn
 ├── AMM          Constant product swap with 2% burn
 ├── Multisig     N-of-M on-chain multisig wallets
 ├── P2P          libp2p + GossipSub + mDNS
 ├── API          REST API with rate limiting + CORS
 ├── Explorer     React + TypeScript + Tailwind CSS
 └── CLI          Cobra commands
```

---

## Key Parameters

| Parameter | Value |
|-----------|-------|
| White Coin total supply | 1,000,000,000 |
| Block interval | 15 seconds |
| Block reward | 50 WC (decays 10% annually) |
| Blue Coin supply (per token) | 1,000,000 |
| Swap burn rate | 2% of Blue Coin per trade |
| Transaction fee | max(0.001 WC, 0.1% of amount) |
| Validator join | Free (24h + PoW) |
| Auto-staking cap | 1,000 WC |
| Suspend threshold | 24 hours offline |
| Evict threshold | 72 hours offline |
| Genesis premine | 10,000 WC |

---

## Roadmap

- [x] Dual-token PoS blockchain (White Coin + Blue Coin)
- [x] AMM with constant product formula + 2% burn
- [x] Multi-validator PoS with slot rotation
- [x] Validator economics (free join, auto-stake, slashing)
- [x] N-of-M multisig wallets
- [x] Team vesting with monthly release
- [x] P2P networking (libp2p + GossipSub)
- [x] Block explorer + web wallet
- [x] API rate limiting, CORS, input validation
- [x] Structured logging (slog)
- [x] Testnet deployment
- [ ] NAT traversal for home nodes
- [ ] Address transaction history index
- [ ] Mobile wallet app

---

## License

MIT
