#!/bin/bash
set -e

WBLUE="./wblue"
echo "=== White & Blue Protocol Demo ==="
echo ""

# Clean up
rm -rf ~/.wblue/data
echo "[1] Starting node..."
$WBLUE start &
NODE_PID=$!
sleep 3

echo ""
echo "[2] Checking chain status..."
$WBLUE chain status
echo ""

# Get validator address from wallet list
VALIDATOR=$($WBLUE wallet list | head -1)
echo "[3] Validator: $VALIDATOR"
echo ""

echo "[4] Checking validator balance..."
$WBLUE wallet info $VALIDATOR
echo ""

echo "[5] Creating second wallet (Bob)..."
BOB_OUTPUT=$($WBLUE wallet create)
BOB=$(echo "$BOB_OUTPUT" | grep "Address:" | awk '{print $2}')
echo "Bob: $BOB"
echo ""

echo "[6] Transferring 500 WC from Validator to Bob..."
$WBLUE transfer white --from $VALIDATOR --to $BOB --amount 500
echo ""
echo "Waiting for next block..."
sleep 16

echo "[7] Checking Bob's balance..."
$WBLUE wallet info $BOB
echo ""

echo "[8] Deploying Blue Coin 'FooCoffee' (FOO)..."
$WBLUE bluecoin deploy \
  --from $VALIDATOR \
  --name "FooCoffee" \
  --symbol "FOO" \
  --pool-ratio 20 \
  --team-ratio 80 \
  --init-white 1000 \
  --release-monthly 20000 \
  --multisig $VALIDATOR
echo ""
echo "Waiting for next block..."
sleep 16

echo "[9] Listing blue coins..."
$WBLUE bluecoin list
echo ""

TOKEN=$($WBLUE bluecoin list | awk '{print $1}')
echo "[10] Pool info for $TOKEN..."
$WBLUE amm pool-info $TOKEN
echo ""

echo "[11] Bob buys FOO with 100 WC..."
$WBLUE amm swap --from $BOB --token $TOKEN --direction white-to-blue --amount-in 100
echo ""
echo "Waiting for next block..."
sleep 16

echo "[12] Pool after buy (price should increase)..."
$WBLUE amm pool-info $TOKEN
echo ""

echo "[13] Bob's balance (should have FOO coins)..."
$WBLUE wallet info $BOB
echo ""

echo "[14] Bob sells 5000 FOO back to white..."
$WBLUE amm swap --from $BOB --token $TOKEN --direction blue-to-white --amount-in 5000
echo ""
echo "Waiting for next block..."
sleep 16

echo "[15] Pool after sell (price should decrease)..."
$WBLUE amm pool-info $TOKEN
echo ""

echo "[16] Bob's final balance..."
$WBLUE wallet info $BOB
echo ""

echo "[17] Final chain status..."
$WBLUE chain status
echo ""

echo "=== Demo Complete! ==="
kill $NODE_PID 2>/dev/null
