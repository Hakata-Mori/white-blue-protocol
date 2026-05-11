package chain

import (
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func deployBlueToken(t *testing.T, db *storage.DB, st *state.StateDB, deployerKP *types.KeyPair, name, symbol string, poolRatio, teamRatio uint8, initWhite uint64) string {
	t.Helper()
	deployParams := token.DeployParams{
		Name:           name,
		Symbol:         symbol,
		PoolRatio:      poolRatio,
		TeamRatio:      teamRatio,
		InitWhite:      initWhite,
		ReleaseMonthly: 10_000,
		MultiSigAddr:   "0xmultisig",
	}
	payload, _ := json.Marshal(deployParams)

	deployerAcct, _ := db.GetAccount(deployerKP.Address)
	nonce := deployerAcct.Nonce + 1

	prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())

	deployTx := types.Transaction{
		Type:      types.TxDeployBlue,
		From:      deployerKP.Address,
		Nonce:     nonce,
		Payload:   payload,
		PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)

	block, err := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(deployTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("deploy should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	tokenID := token.GenerateTokenID(deployerKP.Address, name, nonce)
	pool, err := db.GetPool(tokenID)
	if err != nil {
		t.Fatalf("pool not found after deploy: %v", err)
	}
	if pool.WhiteReserve == 0 || pool.BlueReserve == 0 {
		t.Fatal("pool reserves should be non-zero after deploy")
	}
	return tokenID
}

func fundUser(t *testing.T, db *storage.DB, kp *types.KeyPair, white uint64) {
	t.Helper()
	acct := db.GetOrCreateAccount(kp.Address)
	acct.WhiteBalance = white
	acct.PublicKey = kp.PublicKey
	acct.Nonce = 0
	if err := db.SaveAccount(acct); err != nil {
		t.Fatal(err)
	}
}

func swapWhiteToBlue(t *testing.T, db *storage.DB, st *state.StateDB, validatorAddr string, userKP *types.KeyPair, tokenID string, amount uint64) {
	t.Helper()
	acct, _ := db.GetAccount(userKP.Address)
	prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())

	swapTx := types.Transaction{
		Type:      types.TxSwapWhiteBlue,
		From:      userKP.Address,
		TokenID:   tokenID,
		Amount:    amount,
		Nonce:     acct.Nonce + 1,
		PublicKey: userKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &swapTx, userKP.PrivateKey)

	block, err := CreateBlock(prevBlock, validatorAddr, []types.Transaction{swapTx}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(swapTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("swap should succeed for %s, got %s (err: %s)", userKP.Address, receipt.Status, receipt.Error)
	}
}

func TestMultiUserSwapPriceCurve(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "CurveCoin", "CRV", 70, 30, 1_000_000_000)

	users := make([]*types.KeyPair, 8)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 1_000_000_000)
	}

	pool, _ := db.GetPool(tokenID)
	initialK := new(big.Int).Mul(
		new(big.Int).SetUint64(pool.WhiteReserve),
		new(big.Int).SetUint64(pool.BlueReserve),
	)

	var blueReceived []uint64
	for i := 0; i < 8; i++ {
		acctBefore, _ := db.GetAccount(users[i].Address)
		blueBefore := acctBefore.BlueBalances[tokenID]

		swapWhiteToBlue(t, db, st, validatorKP.Address, users[i], tokenID, 100_000_000)

		acctAfter, _ := db.GetAccount(users[i].Address)
		blueAfter := acctAfter.BlueBalances[tokenID]
		got := blueAfter - blueBefore
		blueReceived = append(blueReceived, got)

		poolAfter, _ := db.GetPool(tokenID)
		currentK := new(big.Int).Mul(
			new(big.Int).SetUint64(poolAfter.WhiteReserve),
			new(big.Int).SetUint64(poolAfter.BlueReserve),
		)
		if currentK.Cmp(initialK) < 0 {
			t.Fatalf("K invariant violated after swap %d: currentK=%s < initialK=%s", i, currentK, initialK)
		}
	}

	for i := 1; i < len(blueReceived); i++ {
		if blueReceived[i] >= blueReceived[i-1] {
			t.Fatalf("price curve should increase: user %d got %d blue but user %d got %d",
				i, blueReceived[i], i-1, blueReceived[i-1])
		}
	}
}

func TestMultiUserSwapBurnAccumulation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "BurnAccum", "BNA", 70, 30, 1_000_000_000)

	users := make([]*types.KeyPair, 6)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 1_000_000_000)
	}

	var totalBlueDistributed uint64
	for i := 0; i < 6; i++ {
		acctBefore, _ := db.GetAccount(users[i].Address)
		blueBefore := acctBefore.BlueBalances[tokenID]

		swapWhiteToBlue(t, db, st, validatorKP.Address, users[i], tokenID, 50_000_000)

		acctAfter, _ := db.GetAccount(users[i].Address)
		blueAfter := acctAfter.BlueBalances[tokenID]
		totalBlueDistributed += blueAfter - blueBefore
	}

	blueState, err := db.GetBlueCoinState(tokenID)
	if err != nil {
		t.Fatal(err)
	}
	if blueState.Burned == 0 {
		t.Fatal("burned should be > 0 after swaps")
	}

	totalBlueOut := totalBlueDistributed + blueState.Burned
	expectedBurn := totalBlueOut * 2 / 100
	diff := int64(blueState.Burned) - int64(expectedBurn)
	if diff < 0 {
		diff = -diff
	}
	tolerance := int64(totalBlueOut / 100)
	if tolerance < 10 {
		tolerance = 10
	}
	if diff > tolerance {
		t.Fatalf("burn should be ~2%% of total blue out: burned=%d expected~%d totalBlueOut=%d",
			blueState.Burned, expectedBurn, totalBlueOut)
	}
}

func TestMultiUserBlueTransferRing(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "RingCoin", "RNG", 70, 30, 1_000_000_000)

	users := make([]*types.KeyPair, 5)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 1_000_000_000)
	}

	swapWhiteToBlue(t, db, st, validatorKP.Address, users[0], tokenID, 200_000_000)

	user0After, _ := db.GetAccount(users[0].Address)
	blueAmount := user0After.BlueBalances[tokenID]
	if blueAmount == 0 {
		t.Fatal("user0 should have blue after swap")
	}

	for i := 0; i < 4; i++ {
		sender := users[i]
		receiver := users[i+1]

		senderAcct, _ := db.GetAccount(sender.Address)
		transferAmount := senderAcct.BlueBalances[tokenID]

		prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())
		transferTx := types.Transaction{
			Type:      types.TxTransferBlue,
			From:      sender.Address,
			To:        receiver.Address,
			TokenID:   tokenID,
			Amount:    transferAmount,
			Nonce:     senderAcct.Nonce + 1,
			PublicKey: sender.PublicKey,
			Timestamp: prevBlock.Header.Timestamp + 15,
		}
		signTx(t, &transferTx, sender.PrivateKey)

		block, err := CreateBlock(prevBlock, validatorKP.Address, []types.Transaction{transferTx}, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatal(err)
		}
		receipt, _ := db.GetReceipt(transferTx.Hash)
		if receipt.Status != "success" {
			t.Fatalf("transfer %d->%d should succeed, got %s (err: %s)", i, i+1, receipt.Status, receipt.Error)
		}
	}

	for i := 0; i < 4; i++ {
		acct, _ := db.GetAccount(users[i].Address)
		if acct.BlueBalances[tokenID] != 0 {
			t.Fatalf("user%d should have 0 blue, got %d", i, acct.BlueBalances[tokenID])
		}
	}

	finalAcct, _ := db.GetAccount(users[4].Address)
	if finalAcct.BlueBalances[tokenID] != blueAmount {
		t.Fatalf("user4 should hold all blue: expected %d, got %d", blueAmount, finalAcct.BlueBalances[tokenID])
	}
}

func TestMultiUserManualBurn(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "ManBurn", "MBR", 70, 30, 1_000_000_000)

	users := make([]*types.KeyPair, 4)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 1_000_000_000)
	}

	blueBalances := make([]uint64, 4)
	for i := 0; i < 4; i++ {
		swapWhiteToBlue(t, db, st, validatorKP.Address, users[i], tokenID, 100_000_000)
		acct, _ := db.GetAccount(users[i].Address)
		blueBalances[i] = acct.BlueBalances[tokenID]
		if blueBalances[i] == 0 {
			t.Fatalf("user%d should have blue after swap", i)
		}
	}

	stateBeforeBurns, _ := db.GetBlueCoinState(tokenID)
	burnedFromSwaps := stateBeforeBurns.Burned

	var totalManualBurned uint64
	for i := 0; i < 4; i++ {
		burnAmount := blueBalances[i] / 2
		acct, _ := db.GetAccount(users[i].Address)

		prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())
		burnTx := types.Transaction{
			Type:      types.TxBlueBurn,
			From:      users[i].Address,
			TokenID:   tokenID,
			Amount:    burnAmount,
			Nonce:     acct.Nonce + 1,
			PublicKey: users[i].PublicKey,
			Timestamp: prevBlock.Header.Timestamp + 15,
		}
		signTx(t, &burnTx, users[i].PrivateKey)

		block, err := CreateBlock(prevBlock, validatorKP.Address, []types.Transaction{burnTx}, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatal(err)
		}
		receipt, _ := db.GetReceipt(burnTx.Hash)
		if receipt.Status != "success" {
			t.Fatalf("burn should succeed for user%d, got %s (err: %s)", i, receipt.Status, receipt.Error)
		}
		totalManualBurned += burnAmount
	}

	for i := 0; i < 4; i++ {
		acct, _ := db.GetAccount(users[i].Address)
		expected := blueBalances[i] - blueBalances[i]/2
		if acct.BlueBalances[tokenID] != expected {
			t.Fatalf("user%d blue balance should be halved: expected %d, got %d", i, expected, acct.BlueBalances[tokenID])
		}
	}

	blueState, _ := db.GetBlueCoinState(tokenID)
	expectedTotalBurned := burnedFromSwaps + totalManualBurned
	if blueState.Burned != expectedTotalBurned {
		t.Fatalf("total burned should be %d (swapBurns=%d + manualBurns=%d), got %d",
			expectedTotalBurned, burnedFromSwaps, totalManualBurned, blueState.Burned)
	}
}

func TestMixedLoadSingleBlock(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	users := make([]*types.KeyPair, 10)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 2_000_000_000)
	}

	tokenID := deployBlueToken(t, db, st, validatorKP, "MixCoin", "MIX", 70, 30, 1_000_000_000)

	swapWhiteToBlue(t, db, st, validatorKP.Address, users[7], tokenID, 100_000_000)
	user7AfterSwap, _ := db.GetAccount(users[7].Address)
	user7BlueBal := user7AfterSwap.BlueBalances[tokenID]

	var txs []types.Transaction

	prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())
	ts := prevBlock.Header.Timestamp + 15

	for i := 0; i < 3; i++ {
		sender := users[i]
		receiver := users[i+1]
		acct, _ := db.GetAccount(sender.Address)
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      sender.Address,
			To:        receiver.Address,
			Amount:    10_000_000,
			Nonce:     acct.Nonce + 1,
			PublicKey: sender.PublicKey,
			Timestamp: ts,
		}
		signTx(t, &tx, sender.PrivateKey)
		txs = append(txs, tx)
	}

	deployerKP := users[4]
	deployAcct, _ := db.GetAccount(deployerKP.Address)
	dp2 := token.DeployParams{
		Name: "SecondCoin", Symbol: "SC2",
		PoolRatio: 60, TeamRatio: 40, InitWhite: 500_000_000,
		ReleaseMonthly: 5_000, MultiSigAddr: "0xms2",
	}
	dp2Payload, _ := json.Marshal(dp2)
	deployTx2 := types.Transaction{
		Type:      types.TxDeployBlue,
		From:      deployerKP.Address,
		Nonce:     deployAcct.Nonce + 1,
		Payload:   dp2Payload,
		PublicKey: deployerKP.PublicKey,
		Timestamp: ts,
	}
	signTx(t, &deployTx2, deployerKP.PrivateKey)
	txs = append(txs, deployTx2)

	for _, idx := range []int{5, 6} {
		acct, _ := db.GetAccount(users[idx].Address)
		swapTx := types.Transaction{
			Type:      types.TxSwapWhiteBlue,
			From:      users[idx].Address,
			TokenID:   tokenID,
			Amount:    50_000_000,
			Nonce:     acct.Nonce + 1,
			PublicKey: users[idx].PublicKey,
			Timestamp: ts,
		}
		signTx(t, &swapTx, users[idx].PrivateKey)
		txs = append(txs, swapTx)
	}

	user7Acct, _ := db.GetAccount(users[7].Address)
	blueTransferTx := types.Transaction{
		Type:      types.TxTransferBlue,
		From:      users[7].Address,
		To:        users[8].Address,
		TokenID:   tokenID,
		Amount:    user7BlueBal / 2,
		Nonce:     user7Acct.Nonce + 1,
		PublicKey: users[7].PublicKey,
		Timestamp: ts,
	}
	signTx(t, &blueTransferTx, users[7].PrivateKey)
	txs = append(txs, blueTransferTx)

	burnAmt := user7BlueBal / 4
	burnTx := types.Transaction{
		Type:      types.TxBlueBurn,
		From:      users[7].Address,
		TokenID:   tokenID,
		Amount:    burnAmt,
		Nonce:     user7Acct.Nonce + 2,
		PublicKey: users[7].PublicKey,
		Timestamp: ts,
	}
	signTx(t, &burnTx, users[7].PrivateKey)
	txs = append(txs, burnTx)

	brokeKP := users[9]
	brokeAcct := db.GetOrCreateAccount(brokeKP.Address)
	brokeAcct.WhiteBalance = 100
	brokeAcct.PublicKey = brokeKP.PublicKey
	db.SaveAccount(brokeAcct)

	brokeAcctNow, _ := db.GetAccount(brokeKP.Address)
	failTx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      brokeKP.Address,
		To:        users[0].Address,
		Amount:    999_999_999,
		Nonce:     brokeAcctNow.Nonce + 1,
		PublicKey: brokeKP.PublicKey,
		Timestamp: ts,
	}
	signTx(t, &failTx, brokeKP.PrivateKey)
	txs = append(txs, failTx)

	block, err := CreateBlock(prevBlock, validatorKP.Address, txs, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	successCount := 0
	failCount := 0
	for _, tx := range txs {
		receipt, err := db.GetReceipt(tx.Hash)
		if err != nil {
			t.Fatalf("receipt not found for tx %s", tx.Hash)
		}
		if receipt.Status == "success" {
			successCount++
		} else {
			failCount++
		}
	}

	if failCount < 1 {
		t.Fatalf("expected at least 1 failed tx, got %d", failCount)
	}
	if successCount < 5 {
		t.Fatalf("expected at least 5 successful txs, got %d", successCount)
	}

	failReceipt, _ := db.GetReceipt(failTx.Hash)
	if failReceipt.Status != "failed" {
		t.Fatalf("insufficient balance tx should fail, got %s", failReceipt.Status)
	}
}

func TestAMMBuyThenSellMultiUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "BuySell", "BSL", 70, 30, 1_000_000_000)

	pool0, _ := db.GetPool(tokenID)
	initialK := new(big.Int).Mul(
		new(big.Int).SetUint64(pool0.WhiteReserve),
		new(big.Int).SetUint64(pool0.BlueReserve),
	)

	stateBeforeBuys, _ := db.GetBlueCoinState(tokenID)
	burnedBefore := stateBeforeBuys.Burned

	buyers := make([]*types.KeyPair, 4)
	for i := range buyers {
		buyers[i] = genKeyPair(t)
		fundUser(t, db, buyers[i], 1_000_000_000)
	}

	for i := 0; i < 4; i++ {
		swapWhiteToBlue(t, db, st, validatorKP.Address, buyers[i], tokenID, 100_000_000)
	}

	stateAfterBuys, _ := db.GetBlueCoinState(tokenID)
	buyBurnAccumulated := stateAfterBuys.Burned - burnedBefore
	if buyBurnAccumulated == 0 {
		t.Fatal("2% burn should have accumulated on buy-side swaps")
	}

	for i := 0; i < 4; i++ {
		acct, _ := db.GetAccount(buyers[i].Address)
		blueBal := acct.BlueBalances[tokenID]
		if blueBal == 0 {
			t.Fatalf("buyer%d should have blue after buy", i)
		}

		prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())
		sellTx := types.Transaction{
			Type:      types.TxSwapBlueWhite,
			From:      buyers[i].Address,
			TokenID:   tokenID,
			Amount:    blueBal,
			Nonce:     acct.Nonce + 1,
			PublicKey: buyers[i].PublicKey,
			Timestamp: prevBlock.Header.Timestamp + 15,
		}
		signTx(t, &sellTx, buyers[i].PrivateKey)

		block, err := CreateBlock(prevBlock, validatorKP.Address, []types.Transaction{sellTx}, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatal(err)
		}
		receipt, _ := db.GetReceipt(sellTx.Hash)
		if receipt.Status != "success" {
			t.Fatalf("sell should succeed for buyer%d, got %s (err: %s)", i, receipt.Status, receipt.Error)
		}
	}

	poolFinal, _ := db.GetPool(tokenID)
	finalK := new(big.Int).Mul(
		new(big.Int).SetUint64(poolFinal.WhiteReserve),
		new(big.Int).SetUint64(poolFinal.BlueReserve),
	)
	if finalK.Cmp(initialK) < 0 {
		t.Fatalf("K invariant violated after all operations: finalK=%s < initialK=%s", finalK, initialK)
	}
}

func TestBlueSupplyAudit(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	tokenID := deployBlueToken(t, db, st, validatorKP, "AuditCoin", "AUD", 70, 30, 1_000_000_000)

	config, err := db.GetBlueCoinConfig(tokenID)
	if err != nil {
		t.Fatal(err)
	}
	totalSupply := config.TotalSupply

	users := make([]*types.KeyPair, 8)
	for i := range users {
		users[i] = genKeyPair(t)
		fundUser(t, db, users[i], 2_000_000_000)
	}

	for i := 0; i < 5; i++ {
		swapWhiteToBlue(t, db, st, validatorKP.Address, users[i], tokenID, 100_000_000)
	}

	for i := 0; i < 3; i++ {
		sender := users[i]
		receiver := users[i+5]
		senderAcct, _ := db.GetAccount(sender.Address)
		transferAmount := senderAcct.BlueBalances[tokenID] / 2
		if transferAmount == 0 {
			continue
		}

		prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())
		transferTx := types.Transaction{
			Type:      types.TxTransferBlue,
			From:      sender.Address,
			To:        receiver.Address,
			TokenID:   tokenID,
			Amount:    transferAmount,
			Nonce:     senderAcct.Nonce + 1,
			PublicKey: sender.PublicKey,
			Timestamp: prevBlock.Header.Timestamp + 15,
		}
		signTx(t, &transferTx, sender.PrivateKey)

		block, err := CreateBlock(prevBlock, validatorKP.Address, []types.Transaction{transferTx}, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatal(err)
		}
	}

	var sumUserBlue uint64
	allAddresses := []string{validatorKP.Address, "0xmultisig"}
	for _, u := range users {
		allAddresses = append(allAddresses, u.Address)
	}

	for _, addr := range allAddresses {
		acct, err := db.GetAccount(addr)
		if err != nil {
			continue
		}
		sumUserBlue += acct.BlueBalances[tokenID]
	}

	pool, _ := db.GetPool(tokenID)
	blueState, _ := db.GetBlueCoinState(tokenID)

	accounted := sumUserBlue + pool.BlueReserve + blueState.Burned + blueState.TeamLocked
	if accounted != totalSupply {
		t.Fatalf("supply conservation violated: userBlue=%d + poolReserve=%d + burned=%d + teamLocked=%d = %d, totalSupply=%d",
			sumUserBlue, pool.BlueReserve, blueState.Burned, blueState.TeamLocked, accounted, totalSupply)
	}

	fmt.Printf("Supply audit passed: users=%d pool=%d burned=%d teamLocked=%d total=%d\n",
		sumUserBlue, pool.BlueReserve, blueState.Burned, blueState.TeamLocked, accounted)
}
