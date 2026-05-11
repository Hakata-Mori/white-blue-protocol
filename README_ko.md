[English](README.md) | [中文](README_zh.md) | [日本語](README_ja.md) | [한국어](README_ko.md)

# 화이트 & 블루 프로토콜 (White & Blue Protocol)

**누구나 자신만의 토큰을 발행할 수 있습니다. 카페, 스타트업, 가족 — 누구든지.**

화이트 & 블루 프로토콜은 듀얼 토큰 PoS 블록체인으로, 모든 조직이 체인의 기본 통화(White Coin)를 기반으로 내장 AMM 풀을 통해 자체 거래 가능한 토큰(Blue Coin)을 발행할 수 있습니다.

> 쉽게 말해: **모든 조직이 주식과 같은 자체 토큰을 즉시 거래 가능하게 발행할 수 있으며, 상장 수수료는 제로입니다.**

---

## 왜 만들어졌나요?

전통적인 자금 조달은 소규모 조직에게 불리합니다:
- IPO는 수백만 달러의 비용이 들고 수년이 걸립니다
- 암호화폐 토큰 발행에는 스마트 컨트랙트 전문 지식이 필요합니다
- 크라우드펀딩 플랫폼은 막대한 수수료를 가져가면서 투자자에게 돌아가는 것은 없습니다

**화이트 & 블루 프로토콜이 이 문제를 해결합니다.** 동네 카페부터 기술 스타트업까지 — 어떤 조직이든 단 한 번의 트랜잭션으로 자체 Blue Coin을 발행할 수 있습니다. 지지자들은 Blue Coin을 구매해 응원하고, 조직이 성장함에 따라 수익을 얻을 수 있습니다.

---

## 실제 사례

### 카페

라떼랩이 새 매장을 오픈하면서 커뮤니티의 지원이 필요합니다:

1. **LatteCoin (LTC)** 을 배포하고, 70%는 AMM 풀에, 30%는 팀 물량으로 잠금
2. 초기 유동성으로 500 WC를 투입
3. 단골 손님들이 내장 스왑을 통해 LatteCoin을 구매
4. 더 많은 사람이 구매할수록 가격이 자연스럽게 상승 (상수곱 AMM)
5. 팀의 30%는 매월 해제 — 확장, 새 장비 구매 등에 사용
6. 매장이 번창하면 초기 지지자들이 수익을 얻고, 그렇지 않으면 풀이 자연스럽게 소진됩니다 — 복잡한 상장 폐지 절차가 필요 없습니다

### 기술 스타트업

3인 스타트업이 지분 희석 없이 시드 펀딩을 원합니다:

1. 창업자들이 **2-of-3 멀티시그 지갑**을 설정
2. 멀티시그가 팀 배분을 관리하는 **StartupCoin (STC)** 을 배포
3. 투자자들이 AMM에서 StartupCoin을 구매 — 가격은 시장 신뢰도를 반영
4. 팀의 모든 지출에는 3명의 창업자 중 2명의 동의가 필요 (멀티시그)
5. 매월 토큰 해제로 개발 마일스톤 자금을 조달

### 가족

네, 가족도 가능합니다:

1. 김씨 가족이 재미있는 가족 실험으로 **JohnsonCoin (JSN)** 을 배포
2. 가족 구성원과 친구들이 참여
3. "시가총액"은 명절 저녁 식사의 유쾌한 대화 소재가 됩니다
4. 하지만 그 이면에는 AMM 가격 책정, 소각 메커니즘 등 실제 경제 원리가 작동하는 진짜 토큰입니다

### 여러분의 상상력

- 유튜버가 **FanCoin**을 발행 — 구독자들이 성장에 투자
- 지역 농장이 **FarmCoin**을 발행 — 토큰을 구매하고 농산물 할인을 받기
- 학교 동아리가 **ClubCoin**을 발행 — 투명한 모금 활동
- 밴드가 **BandCoin**을 발행 — 팬이 이해관계자가 됨

**지지자가 있다면, Blue Coin을 가질 수 있습니다.**

---

## 작동 방식

### 두 가지 토큰

| | White Coin (WC) | Blue Coin |
|---|---|---|
| 무엇인가 | 체인의 기본 통화 | 조직별 토큰 |
| 공급량 | 10억 개 (PoS를 통해 채굴) | 토큰당 100만 개 (고정) |
| 획득 방법 | 검증자 노드 운영 | AMM에서 구매 / 전송 수신 |
| 목적 | 가스, 유동성, 스테이킹 | 투자, 커뮤니티, 로열티 |

### 토큰 경제학

- **AMM 풀**: 상수곱 공식 (`x * y = k`). 공급과 수요에 따라 가격이 변동
- **스왑 시 2% 소각**: 모든 거래에서 Blue Coin의 2%가 소각되어 디플레이션 발생
- **팀 베스팅**: 팀 배분은 한꺼번에가 아닌 매월 해제
- **멀티시그 관리**: 팀 자금은 책임성을 위해 N-of-M 멀티시그로 잠금 가능
- **자연 소멸**: 아무도 거래하지 않으면 풀은 자연스럽게 제로로 소진. 상장 폐지 불필요

### 검증자 경제학

| 항목 | 값 |
|-----------|-------|
| 참여 비용 | 무료 (24시간 온라인 + proof-of-work) |
| 블록 보상 | 블록당 50 WC |
| 블록 간격 | 15초 |
| 자동 스테이킹 | 보상이 1,000 WC까지 스테이크로 잠금 |
| 수수료 분배 | 트랜잭션 수수료의 50%가 블록 생성자에게 |
| 오프라인 패널티 | 24시간 후 정지, 72시간 후 퇴출 |
| 이중 서명 | 영구 차단 + 스테이크 몰수 |

---

## 라이브 테스트넷

**블록 익스플로러**: [http://8.217.52.231](http://8.217.52.231)

지갑을 생성하고, 블록을 탐색하고, 거래까지 — 모두 브라우저에서 가능합니다:
- 실시간 블록, 트랜잭션, 검증자 상태 조회
- 지갑 생성 및 토큰 관리
- WhiteCoin과 BlueCoin 전송
- 나만의 BlueCoin 토큰 배포
- 내장 AMM에서 토큰 스왑

---

## 빠른 시작

### 옵션 1: 웹 익스플로러 사용 (설치 불필요)

[http://8.217.52.231](http://8.217.52.231)에 접속하여 **Wallet**을 클릭하면 시작됩니다.

### 옵션 2: 풀 노드 실행

```bash
git clone https://github.com/Hataka-Mori/white-blue-protocol.git
cd white-blue-protocol
make build

# 풀 노드 시작 (테스트넷에서 자동 동기화)
./wblue start --no-validator
```

### 옵션 3: 검증자 실행 (WC 획득)

```bash
./wblue start
```

노드가 수행하는 작업:
1. 자동으로 지갑 생성
2. 테스트넷 시드 노드에 연결
3. 24시간 후보 기간 시작
4. 24시간 + PoW 검증 후, 블록 생성 및 블록당 50 WC 획득 시작

---

## CLI 명령어

### 지갑

```bash
wblue wallet create                # 새 지갑 생성 (니모닉 백업 포함)
wblue wallet recover               # 니모닉 문구로 지갑 복구
wblue wallet list                  # 지갑 목록 조회
wblue wallet info <address>        # 잔액 조회
```

### 전송

```bash
wblue transfer white --from <addr> --to <addr> --amount 100
wblue transfer blue --from <addr> --to <addr> --token <id> --amount 500
```

### Blue Coin

```bash
# 새 토큰 배포
wblue bluecoin deploy \
  --from <addr> \
  --name "LatteCoin" \
  --symbol "LTC" \
  --pool-ratio 70 \
  --team-ratio 30 \
  --init-white 500 \
  --release-monthly 5000 \
  --multisig <multisig-addr>

# 조회
wblue bluecoin list
wblue bluecoin info <tokenId>
wblue bluecoin burn --from <addr> --token <id> --amount 1000
```

### AMM 스왑

```bash
# White Coin으로 Blue Coin 구매
wblue amm swap --from <addr> --token <id> --direction white-to-blue --amount-in 100

# Blue Coin을 White Coin으로 판매
wblue amm swap --from <addr> --token <id> --direction blue-to-white --amount-in 500

# 풀 확인
wblue amm pool-info <tokenId>
wblue amm price <tokenId>
```

### 검증자

```bash
wblue validator join --from <addr>
wblue validator exit --from <addr>
wblue validator status
wblue validator heartbeat --from <addr>
```

### 멀티시그

```bash
wblue multisig register --owners addr1,addr2,addr3 --threshold 2
wblue multisig propose --multisig <ms-addr> --to <target> --amount 100
wblue multisig approve --multisig <ms-addr> --proposal-id 0
wblue multisig info --address <ms-addr>
```

### 체인

```bash
wblue chain status
wblue chain tx <hash>
wblue version
```

---

## HTTP API

| 메서드 | 엔드포인트 | 설명 |
|--------|----------|------|
| GET | `/health` | 헬스 체크 |
| GET | `/api/v1/chain/status` | 체인 상태 |
| GET | `/api/v1/chain/block/:height` | 높이로 블록 조회 |
| GET | `/api/v1/blocks?limit=20&offset=0` | 블록 목록 (페이지네이션) |
| GET | `/api/v1/block/hash/:hash` | 해시로 블록 조회 |
| GET | `/api/v1/stats` | 네트워크 통계 |
| GET | `/api/v1/wallet/:address` | 계정 잔액 |
| GET | `/api/v1/bluecoin` | 모든 Blue Coin 목록 |
| GET | `/api/v1/bluecoin/:tokenId` | Blue Coin 설정 |
| GET | `/api/v1/bluecoin/:tokenId/state` | Blue Coin 상태 (소각, 잠금) |
| GET | `/api/v1/pool/:tokenId` | AMM 풀 정보 |
| GET | `/api/v1/validators` | 검증자 세트 |
| GET | `/api/v1/multisig/:address` | 멀티시그 계정 |
| GET | /api/v1/address/:address/txs | 주소별 트랜잭션 히스토리 |
| POST | `/api/v1/tx/submit` | 트랜잭션 제출 |
| GET | `/api/v1/tx/:hash` | 트랜잭션 영수증 |

---

## 아키텍처

```
wblue (단일 바이너리)
 ├── Consensus    슬롯 로테이션 기반 PoS, 15초 블록
 ├── State        계정, 잔액, 검증자 세트
 ├── Storage      BoltDB (임베디드 키-값 저장소)
 ├── Token        Blue Coin 배포, 베스팅, 소각
 ├── AMM          2% 소각이 포함된 상수곱 스왑
 ├── Multisig     N-of-M 온체인 멀티시그 지갑
 ├── P2P          libp2p + GossipSub + mDNS
 ├── API          속도 제한 + CORS가 포함된 REST API
 ├── Explorer     React + TypeScript + Tailwind CSS
 └── CLI          Cobra 명령어
```

---

## 주요 파라미터

| 항목 | 값 |
|-----------|-------|
| White Coin 총 공급량 | 1,000,000,000 |
| 블록 간격 | 15초 |
| 블록 보상 | 50 WC (연간 10% 감소) |
| Blue Coin 공급량 (토큰당) | 1,000,000 |
| 스왑 소각률 | 거래당 Blue Coin의 2% |
| 트랜잭션 수수료 | max(0.001 WC, 금액의 0.1%) |
| 검증자 참여 | 무료 (24시간 + PoW) |
| 자동 스테이킹 한도 | 1,000 WC |
| 정지 임계값 | 오프라인 24시간 |
| 퇴출 임계값 | 오프라인 72시간 |
| 제네시스 프리마인 | 10,000 WC |

---

## 로드맵

- [x] 듀얼 토큰 PoS 블록체인 (White Coin + Blue Coin)
- [x] 상수곱 공식 + 2% 소각이 적용된 AMM
- [x] 슬롯 로테이션 기반 멀티 검증자 PoS
- [x] 검증자 경제학 (무료 참여, 자동 스테이크, 슬래싱)
- [x] N-of-M 멀티시그 지갑
- [x] 월별 해제 기반 팀 베스팅
- [x] P2P 네트워킹 (libp2p + GossipSub)
- [x] 블록 익스플로러 + 웹 지갑
- [x] API 속도 제한, CORS, 입력 검증
- [x] 구조화된 로깅 (slog)
- [x] 테스트넷 배포
- [x] 주소별 트랜잭션 히스토리 인덱스
- [x] 니모닉 백업 (BIP39, 12단어)
- [ ] 홈 노드를 위한 NAT 트래버설
- [ ] 모바일 지갑 앱

---

## 라이선스

MIT
