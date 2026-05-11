export interface BlockHeader {
  height: number;
  prevHash: string;
  merkleRoot: string;
  timestamp: number;
  validator: string;
  reward: number;
  hash: string;
  signature?: string;
}

export interface Block {
  header: BlockHeader;
  transactions: Transaction[];
}

export interface BlockSummary {
  height: number;
  hash: string;
  prevHash: string;
  timestamp: number;
  validator: string;
  txCount: number;
  reward: number;
}

export interface Transaction {
  hash: string;
  type: number;
  from: string;
  to: string;
  amount: number;
  tokenId: string;
  fee: number;
  nonce: number;
  payload: string;
  publicKey: string;
  signature: string;
  timestamp: number;
  minAmountOut?: number;
}

export interface TxReceipt {
  txHash: string;
  blockHeight: number;
  blockHash: string;
  status: string;
  error?: string;
}

export interface Account {
  address: string;
  publicKey: string;
  whiteBalance: number;
  blueBalances: Record<string, number>;
  nonce: number;
  createdAt: number;
  stakedBalance?: number;
}

export interface BlueCoinConfig {
  tokenId: string;
  name: string;
  symbol: string;
  creator: string;
  totalSupply: number;
  poolRatio: number;
  teamRatio: number;
  initWhite: number;
  releaseMonthly: number;
  multiSigAddr: string;
  sourceUrls: string[];
  deployedAt: number;
  deployTxHash: string;
}

export interface BlueCoinState {
  tokenId: string;
  totalMinted: number;
  poolAllocation: number;
  teamLocked: number;
  teamReleased: number;
  lastUnlockTime: number;
  burned: number;
}

export interface AMMPool {
  tokenId: string;
  whiteReserve: number;
  blueReserve: number;
  k: string;
  totalFeeBurned: number;
  createdAt: number;
  lastTradedAt: number;
}

export interface ValidatorRecord {
  address: string;
  publicKey: string;
  joinHeight: number;
  firstHeartbeatHeight: number;
  lastHeartbeatHeight: number;
  status: string;
  suspendedAt?: number;
}

export interface ValidatorSet {
  validators: ValidatorRecord[];
  updatedAt: number;
}

export interface BlocksResponse {
  blocks: BlockSummary[];
  total: number;
}

export interface NetworkStats {
  height: number;
  totalMinted: number;
  totalTxs: number;
  activeValidators: number;
  avgBlockTime: number;
  chainId: string;
}

export interface TxSummary {
  hash: string;
  type: number;
  from: string;
  to: string;
  amount: number;
  fee: number;
  tokenId?: string;
  blockHeight: number;
  timestamp: number;
  status: string;
}

export interface AddressTxsResponse {
  txs: TxSummary[];
  total: number;
}
