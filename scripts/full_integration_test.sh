#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

WBLUE="./wblue"
DA="$HOME/.wblue/test_a"
DB="$HOME/.wblue/test_b"
DC="$HOME/.wblue/test_c"
DD="$HOME/.wblue/test_d"
PW="testpw"
export WBLUE_VALIDATOR_PASSWORD="$PW"

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; NC='\033[0m'
pass() { echo -e "  ${GREEN}[PASS]${NC} $1"; PASSES=$((PASSES+1)); }
fail() { echo -e "  ${RED}[FAIL]${NC} $1"; FAILURES=$((FAILURES+1)); }
section() { echo -e "\n${YELLOW}=== $1 ===${NC}"; }
FAILURES=0; PASSES=0

api() { curl -s "http://localhost:$1$2" 2>/dev/null; }
jv() { python3 -c "import sys,json; d=json.load(sys.stdin); print($1)" 2>/dev/null; }
get_balance() { api "$1" "/api/v1/wallet/$2" | jv "d.get('whiteBalance',0)"; }
get_staked() { api "$1" "/api/v1/wallet/$2" | jv "d.get('stakedBalance',0)"; }
get_height() { api "$1" "/api/v1/chain/status" | jv "d['height']"; }
get_val_status() { api "$1" "/api/v1/validators" | python3 -c "import sys,json; vs=json.load(sys.stdin); [print(v['status']) for v in vs.get('validators',[]) if v['address']=='$2']" 2>/dev/null | head -1; }
wallet_addr() { ls "$1/wallets/" 2>/dev/null | head -1 | sed 's/.json//'; }
kill_port() { pgrep -f "p2p-port $1" | xargs kill 2>/dev/null; sleep 2; }

cleanup() {
    echo -e "\n=== Cleanup ==="
    pkill -f "wblue start" 2>/dev/null || true
    sleep 2
    rm -rf "$DA" "$DB" "$DC" "$DD"
}
trap cleanup EXIT

echo "============================================"
echo "  Full Integration Test (4 users, 9 phases)"
echo "============================================"
cleanup 2>/dev/null
make build

BI=5  # dev block interval

# ========== Phase 1 ==========
section "Phase 1: Cold Start + Basic Transfer"

$WBLUE start --data-dir "$DA" --api-port 8080 --p2p-port 30303 --mdns=false --dev &>/tmp/test_a.log &
sleep $((BI*2))

ALICE=$(wallet_addr "$DA")
echo "  Alice: $ALICE"

H=$(get_height 8080)
[ "$H" -ge 1 ] && pass "Alice producing blocks (height=$H)" || fail "Not producing blocks"

S=$(get_staked 8080 "$ALICE")
[ "$S" -gt 0 ] && pass "Auto-staking (staked=$S)" || fail "Auto-staking failed"

P2P_ADDR=$(grep "\[P2P\] Listening on /ip4/127" /tmp/test_a.log | head -1 | sed 's/.*Listening on //')

PWFILE=$(mktemp); echo "$PW" > "$PWFILE"; echo "$PW" >> "$PWFILE"
BOB=$($WBLUE --data-dir "$DA" wallet create < "$PWFILE" 2>&1 | grep -o '0x[0-9a-f]\{40\}' | head -1)
CHARLIE=$($WBLUE --data-dir "$DA" wallet create < "$PWFILE" 2>&1 | grep -o '0x[0-9a-f]\{40\}' | head -1)
DAVE=$($WBLUE --data-dir "$DA" wallet create < "$PWFILE" 2>&1 | grep -o '0x[0-9a-f]\{40\}' | head -1)
rm -f "$PWFILE"
echo "  Bob=$BOB  Charlie=$CHARLIE  Dave=$DAVE"

$WBLUE --data-dir "$DA" --api-url http://localhost:8080 transfer white --from "$ALICE" --to "$BOB" --amount 2000 >/dev/null 2>&1
sleep $((BI+1))
$WBLUE --data-dir "$DA" --api-url http://localhost:8080 transfer white --from "$ALICE" --to "$CHARLIE" --amount 2000 >/dev/null 2>&1
sleep $((BI+1))
$WBLUE --data-dir "$DA" --api-url http://localhost:8080 transfer white --from "$ALICE" --to "$DAVE" --amount 2000 >/dev/null 2>&1
sleep $((BI+1))

[ "$(get_balance 8080 $BOB)" -ge 2000000000 ] && pass "Bob got 2000 WC" || fail "Bob balance wrong"
[ "$(get_balance 8080 $CHARLIE)" -ge 2000000000 ] && pass "Charlie got 2000 WC" || fail "Charlie balance wrong"
[ "$(get_balance 8080 $DAVE)" -ge 2000000000 ] && pass "Dave got 2000 WC" || fail "Dave balance wrong"

# ========== Phase 2 ==========
section "Phase 2: Validator Join (cold start)"

mkdir -p "$DB/wallets"; cp "$DA/wallets/$BOB.json" "$DB/wallets/"
$WBLUE start --data-dir "$DB" --api-port 8081 --p2p-port 30304 --no-validator --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_b.log &
sleep $((BI*3))

$WBLUE --data-dir "$DB" --api-url http://localhost:8081 validator join --from "$BOB" >/dev/null 2>&1
sleep $((BI*2))

BS=$(get_val_status 8080 "$BOB")
[ "$BS" = "active" ] && pass "Bob joined (status=$BS)" || fail "Bob join failed (status=$BS)"

# Restart Bob as validator
kill_port 30304
$WBLUE start --data-dir "$DB" --api-port 8081 --p2p-port 30304 --validator "$BOB" --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_b2.log &
sleep $((BI*3))

mkdir -p "$DC/wallets"; cp "$DA/wallets/$CHARLIE.json" "$DC/wallets/"
$WBLUE start --data-dir "$DC" --api-port 8082 --p2p-port 30305 --no-validator --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_c.log &
sleep $((BI*3))

$WBLUE --data-dir "$DC" --api-url http://localhost:8082 validator join --from "$CHARLIE" >/dev/null 2>&1
sleep $((BI*2))

CS=$(get_val_status 8080 "$CHARLIE")
[ "$CS" = "active" ] && pass "Charlie joined (status=$CS)" || fail "Charlie join failed (status=$CS)"

kill_port 30305
$WBLUE start --data-dir "$DC" --api-port 8082 --p2p-port 30305 --validator "$CHARLIE" --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_c2.log &
sleep $((BI*3))

VC=$(api 8080 "/api/v1/validators" | python3 -c "import sys,json; vs=json.load(sys.stdin); print(len([v for v in vs.get('validators',[]) if v['status']=='active']))" 2>/dev/null)
[ "$VC" = "3" ] && pass "3 active validators" || fail "Expected 3 active, got $VC"

# ========== Phase 3 ==========
section "Phase 3: BlueCoin + AMM"

mkdir -p "$DB/wallets"; cp "$DA/wallets/$DAVE.json" "$DB/wallets/" 2>/dev/null
$WBLUE --data-dir "$DB" --api-url http://localhost:8081 bluecoin deploy \
    --from "$DAVE" --name "FooCoin" --symbol "FOO" \
    --pool-ratio 50 --team-ratio 50 --init-white 500 \
    --release-monthly 10000 --multisig "$DAVE" >/dev/null 2>&1
sleep $((BI*2))

TOKEN=$(api 8080 "/api/v1/bluecoin" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d[0]['tokenId'] if d else '')" 2>/dev/null)
[ -n "$TOKEN" ] && pass "BlueCoin deployed ($TOKEN)" || fail "Deploy failed"

$WBLUE --data-dir "$DB" --api-url http://localhost:8081 amm swap \
    --from "$DAVE" --token "$TOKEN" --direction white-to-blue --amount-in 100 --min-out 1 >/dev/null 2>&1
sleep $((BI*2))

DB_BLUE=$(api 8080 "/api/v1/wallet/$DAVE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(sum(d.get('blueBalances',{}).values()))" 2>/dev/null)
[ "$DB_BLUE" -gt 0 ] && pass "Swap white→blue OK (blue=$DB_BLUE)" || fail "Swap failed"

$WBLUE --data-dir "$DB" --api-url http://localhost:8081 transfer blue \
    --from "$DAVE" --to "$ALICE" --amount 1000 --token "$TOKEN" >/dev/null 2>&1
sleep $((BI*2))

AB=$(api 8080 "/api/v1/wallet/$ALICE" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('blueBalances',{}).get('$TOKEN',0))" 2>/dev/null)
[ "$AB" -gt 0 ] && pass "Blue transfer OK" || fail "Blue transfer failed"

# ========== Phase 4 (moved after Phase 5 exit) ==========
section "Phase 5: Bob Exits"

BSB=$(get_staked 8080 "$BOB")
$WBLUE --data-dir "$DB" --api-url http://localhost:8081 validator exit --from "$BOB" >/dev/null 2>&1
sleep $((BI*2))

BS2=$(get_val_status 8080 "$BOB")
[ "$BS2" = "removed" ] && pass "Bob exited (removed)" || fail "Bob exit failed ($BS2)"

BSA=$(get_staked 8080 "$BOB")
[ "$BSA" = "0" ] && pass "Bob stake burned (was=$BSB)" || fail "Stake not burned ($BSA)"

section "Phase 4: Dave Joins (after Bob exit, activeCount=2<3)"

mkdir -p "$DD/wallets"; cp "$DA/wallets/$DAVE.json" "$DD/wallets/"
$WBLUE start --data-dir "$DD" --api-port 8083 --p2p-port 30306 --no-validator --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_d.log &
sleep $((BI*3))

$WBLUE --data-dir "$DD" --api-url http://localhost:8083 validator join --from "$DAVE" >/dev/null 2>&1
sleep $((BI*2))

DS=$(get_val_status 8080 "$DAVE")
[ "$DS" = "active" ] && pass "Dave joined" || fail "Dave join failed (status=$DS)"

kill_port 30306
$WBLUE start --data-dir "$DD" --api-port 8083 --p2p-port 30306 --validator "$DAVE" --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_d2.log &
sleep $((BI*3))

AC=$(api 8080 "/api/v1/validators" | python3 -c "import sys,json; vs=json.load(sys.stdin); print(len([v for v in vs.get('validators',[]) if v['status']=='active']))" 2>/dev/null)
[ "$AC" = "3" ] && pass "3 active (Alice+Charlie+Dave)" || fail "Expected 3, got $AC"

# ========== Phase 6 ==========
section "Phase 6: Charlie Suspended + Recovery"

kill_port 30305
echo "  Killed Charlie"

sleep $((BI*20))

CS2=$(get_val_status 8080 "$CHARLIE")
[ "$CS2" = "suspended" ] && pass "Charlie suspended" || fail "Not suspended ($CS2)"

$WBLUE start --data-dir "$DC" --api-port 8082 --p2p-port 30305 --validator "$CHARLIE" --mdns=false --dev --seeds "$P2P_ADDR" &>/tmp/test_c3.log &
sleep $((BI*10))

CS3=$(get_val_status 8080 "$CHARLIE")
[ "$CS3" = "active" ] && pass "Charlie recovered" || fail "Not recovered ($CS3)"

# ========== Phase 7 ==========
section "Phase 7: Charlie Evicted"

kill_port 30305
echo "  Killed Charlie again"

sleep $((BI*45))

CS4=$(get_val_status 8080 "$CHARLIE")
[ "$CS4" = "removed" ] && pass "Charlie evicted" || fail "Not evicted ($CS4)"

CST=$(get_staked 8080 "$CHARLIE")
[ "$CST" = "0" ] && pass "Charlie stake burned" || fail "Stake not burned ($CST)"

# ========== Phase 8 ==========
section "Phase 8: Auto-Staking Overflow"

AS=$(get_staked 8080 "$ALICE")
AB2=$(get_balance 8080 "$ALICE")
echo "  Alice: staked=$AS balance=$AB2"
[ "$AS" -ge 1000000000 ] && pass "Alice staked full (1000 WC)" || fail "Stake not full ($AS)"
[ "$AB2" -gt 0 ] && pass "Overflow into balance" || fail "No overflow"

# ========== Phase 9 ==========
section "Phase 9: Final"

FA=$(api 8080 "/api/v1/validators" | python3 -c "import sys,json; vs=json.load(sys.stdin); print(len([v for v in vs.get('validators',[]) if v['status']=='active']))" 2>/dev/null)
[ "$FA" = "2" ] && pass "Final: 2 active (Alice+Dave)" || fail "Expected 2, got $FA"

FH=$(get_height 8080)
pass "Final height: $FH"

echo ""
echo "============================================"
echo "  Results: $PASSES passed, $FAILURES failed"
if [ $FAILURES -eq 0 ]; then echo -e "  ${GREEN}ALL TESTS PASSED${NC}"
else echo -e "  ${RED}$FAILURES FAILED${NC}"; fi
echo "============================================"
exit $FAILURES
