#!/bin/bash

WBLUE="./wblue"
DATA_A="$HOME/.wblue/test_node_a"
DATA_B="$HOME/.wblue/test_node_b"
LOG_A="/tmp/wblue_nodeA.log"
LOG_B="/tmp/wblue_nodeB.log"
PW="testpw123"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

pass() { echo -e "${GREEN}[PASS]${NC} $1"; }
fail() { echo -e "${RED}[FAIL]${NC} $1"; FAILURES=$((FAILURES+1)); }
FAILURES=0

cleanup() {
    echo ""
    echo "=== Cleanup ==="
    pkill -f "wblue start" 2>/dev/null || true
    sleep 1
    rm -rf "$DATA_A" "$DATA_B"
}
trap cleanup EXIT

echo "============================================"
echo "  Multi-Node Integration Test"
echo "============================================"
echo ""

cleanup 2>/dev/null
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"
make build

echo ""
echo "=== Step 1: Start Node A (validator) ==="
WBLUE_VALIDATOR_PASSWORD="$PW" $WBLUE start --data-dir "$DATA_A" --api-port 8080 --p2p-port 30303 --mdns=false &>"$LOG_A" &
PID_A=$!
sleep 3

VALIDATOR_A=$(ls "$DATA_A/wallets/" | head -1 | sed 's/.json//')
echo "Node A PID=$PID_A Validator=$VALIDATOR_A"

sleep 18
HEIGHT_A=$(curl -s http://localhost:8080/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])")
if [ "$HEIGHT_A" -ge 1 ]; then
    pass "Node A is producing blocks (height=$HEIGHT_A)"
else
    fail "Node A not producing blocks"
fi

echo ""
echo "=== Step 2: Start Node B (full node, syncs from A) ==="
P2P_ADDR=$(grep "\[P2P\] Listening on /ip4/127" "$LOG_A" | head -1 | sed 's/.*Listening on //')
echo "Node A P2P addr: $P2P_ADDR"

$WBLUE start --data-dir "$DATA_B" --api-port 8081 --p2p-port 30304 --no-validator --mdns=false --seeds "$P2P_ADDR" &>"$LOG_B" &
PID_B=$!
echo "Node B PID=$PID_B"

sleep 20

HEIGHT_B=$(curl -s http://localhost:8081/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])" 2>/dev/null || echo "0")
echo "Node B height=$HEIGHT_B"
if [ "$HEIGHT_B" -ge 1 ]; then
    pass "Node B synced blocks from A (height=$HEIGHT_B)"
else
    fail "Node B failed to sync (height=$HEIGHT_B)"
fi

echo ""
echo "=== Step 3: Create Bob wallet on Node B, transfer from A ==="
PWFILE=$(mktemp)
echo "$PW" > "$PWFILE"
echo "$PW" >> "$PWFILE"
BOB_OUT=$($WBLUE --data-dir "$DATA_B" wallet create < "$PWFILE" 2>&1) || true
rm -f "$PWFILE"
BOB=$(echo "$BOB_OUT" | grep -o '0x[0-9a-f]\{40\}' | head -1)
if [ -z "$BOB" ]; then
    echo "wallet create output: $BOB_OUT"
    fail "Failed to create Bob wallet"
else
    echo "Bob address: $BOB"
fi

echo "$PW" | $WBLUE --data-dir "$DATA_A" --api-url http://localhost:8080 transfer white --from "$VALIDATOR_A" --to "$BOB" --amount 2000 >/dev/null 2>&1
echo "Transfer submitted: 2000 WC to Bob"

sleep 18

BOB_BAL=$(curl -s "http://localhost:8080/api/v1/wallet/$BOB" | python3 -c "import sys,json; print(json.load(sys.stdin).get('whiteBalance',0))")
if [ "$BOB_BAL" -ge 2000000000 ]; then
    pass "Bob received 2000 WC (balance=$BOB_BAL)"
else
    fail "Bob balance wrong: $BOB_BAL"
fi

BOB_BAL_B=$(curl -s "http://localhost:8081/api/v1/wallet/$BOB" | python3 -c "import sys,json; print(json.load(sys.stdin).get('whiteBalance',0))")
if [ "$BOB_BAL_B" -ge 2000000000 ]; then
    pass "Node B also sees Bob's balance (balance=$BOB_BAL_B)"
else
    fail "Node B doesn't see Bob's balance: $BOB_BAL_B"
fi

echo ""
echo "=== Step 4: Submit tx via Node B → A should pack it ==="
cp "$DATA_A/wallets/$BOB.json" "$DATA_B/wallets/$BOB.json" 2>/dev/null || true
if [ ! -f "$DATA_B/wallets/$BOB.json" ]; then
    cp "$DATA_B/wallets/$BOB.json" "$DATA_B/wallets/$BOB.json" 2>/dev/null || true
fi
echo "$PW" | $WBLUE --data-dir "$DATA_B" --api-url http://localhost:8081 transfer white --from "$BOB" --to "$VALIDATOR_A" --amount 100 >/dev/null 2>&1
echo "Transfer submitted via Node B: Bob → Validator 100 WC"

sleep 18

BOB_BAL2=$(curl -s "http://localhost:8080/api/v1/wallet/$BOB" | python3 -c "import sys,json; print(json.load(sys.stdin).get('whiteBalance',0))")
if [ "$BOB_BAL2" -lt "$BOB_BAL" ]; then
    pass "Tx from Node B was packed by Node A (Bob balance decreased: $BOB_BAL → $BOB_BAL2)"
else
    fail "Tx from Node B not packed (balance unchanged: $BOB_BAL2)"
fi

echo ""
echo "=== Step 5: Verify block signatures ==="
BLOCK_1=$(curl -s http://localhost:8080/api/v1/chain/block/1)
HAS_SIG=$(echo "$BLOCK_1" | python3 -c "import sys,json; d=json.load(sys.stdin); print('yes' if d['header'].get('signature','') != '' else 'no')")
if [ "$HAS_SIG" = "yes" ]; then
    pass "Block #1 has a signature"
else
    fail "Block #1 missing signature"
fi

echo ""
echo "=== Step 6: Heights are consistent ==="
HEIGHT_A2=$(curl -s http://localhost:8080/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])")
sleep 2
HEIGHT_B2=$(curl -s http://localhost:8081/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])")

DIFF=$((HEIGHT_A2 - HEIGHT_B2))
if [ "$DIFF" -lt 0 ]; then DIFF=$((-DIFF)); fi

if [ "$DIFF" -le 1 ]; then
    pass "Heights are consistent (A=$HEIGHT_A2, B=$HEIGHT_B2, diff=$DIFF)"
else
    fail "Heights diverged (A=$HEIGHT_A2, B=$HEIGHT_B2, diff=$DIFF)"
fi

echo ""
echo "=== Step 7: Kill Node B, wait, restart, verify resync ==="
kill $PID_B 2>/dev/null || true
sleep 20

$WBLUE start --data-dir "$DATA_B" --api-port 8081 --p2p-port 30304 --no-validator --mdns=false --seeds "$P2P_ADDR" &>"$LOG_B.2" &
PID_B=$!
sleep 15

HEIGHT_A3=$(curl -s http://localhost:8080/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])")
HEIGHT_B3=$(curl -s http://localhost:8081/api/v1/chain/status | python3 -c "import sys,json; print(json.load(sys.stdin)['height'])")

DIFF2=$((HEIGHT_A3 - HEIGHT_B3))
if [ "$DIFF2" -lt 0 ]; then DIFF2=$((-DIFF2)); fi

if [ "$DIFF2" -le 1 ]; then
    pass "Node B resynced after restart (A=$HEIGHT_A3, B=$HEIGHT_B3)"
else
    fail "Node B failed to resync (A=$HEIGHT_A3, B=$HEIGHT_B3, diff=$DIFF2)"
fi

echo ""
echo "============================================"
if [ $FAILURES -eq 0 ]; then
    echo -e "${GREEN}ALL TESTS PASSED${NC}"
else
    echo -e "${RED}$FAILURES TEST(S) FAILED${NC}"
fi
echo "============================================"
exit $FAILURES
