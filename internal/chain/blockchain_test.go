package chain

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func setupTestDB(t *testing.T) (*storage.DB, func()) {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	return db, func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func createTestChain(t *testing.T, db *storage.DB, kp *types.KeyPair) {
	t.Helper()
	genesis := CreateGenesisBlock(&types.GenesisConfig{
		ChainID:          "test-chain",
		GenesisValidator: kp.Address,
	})
	if err := db.SaveBlock(genesis); err != nil {
		t.Fatal(err)
	}
	account := db.GetOrCreateAccount(kp.Address)
	account.WhiteBalance = 10_000_000_000
	account.PublicKey = kp.PublicKey
	if err := db.SaveAccount(account); err != nil {
		t.Fatal(err)
	}
}

func genKeyPair(t *testing.T) *types.KeyPair {
	t.Helper()
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	return kp
}

func signTx(t *testing.T, tx *types.Transaction, privKey string) {
	t.Helper()
	txCopy := *tx
	txCopy.Signature = ""
	txCopy.Hash = ""
	txData, _ := json.Marshal(txCopy)
	tx.Hash = crypto.SHA256Hex(txData)
	sig, err := crypto.Sign(privKey, txData)
	if err != nil {
		t.Fatal(err)
	}
	tx.Signature = sig
}

func TestApplyBlockAtomic(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	bobKP := genKeyPair(t)
	bobAcct := db.GetOrCreateAccount(bobKP.Address)
	bobAcct.WhiteBalance = 1_000_000
	bobAcct.PublicKey = bobKP.PublicKey
	bobAcct.Nonce = 0
	db.SaveAccount(bobAcct)

	prevBlock, _ := db.GetBlockByHeight(0)

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	transferTx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      bobKP.Address,
		To:        "0xalice",
		Amount:    500_000,
		Nonce:     1,
		PublicKey: bobKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &transferTx, bobKP.PrivateKey)

	block, err := CreateBlock(prevBlock, validatorKP.Address, []types.Transaction{rewardTx, transferTx}, 50_000_000)
	if err != nil {
		t.Fatal(err)
	}

	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	bobAfter, _ := db.GetAccount(bobKP.Address)
	fee := types.CalcFee(500_000)
	expectedBob := uint64(1_000_000) - 500_000 - fee
	if bobAfter.WhiteBalance != expectedBob {
		t.Fatalf("bob balance: expected %d, got %d", expectedBob, bobAfter.WhiteBalance)
	}

	alice, _ := db.GetAccount("0xalice")
	if alice.WhiteBalance != 500_000 {
		t.Fatalf("alice balance: expected 500000, got %d", alice.WhiteBalance)
	}

	receipt, err := db.GetReceipt(transferTx.Hash)
	if err != nil {
		t.Fatalf("receipt not found: %v", err)
	}
	if receipt.Status != "success" {
		t.Fatalf("receipt status: expected success, got %s (err: %s)", receipt.Status, receipt.Error)
	}
}

func TestApplyBlockIdempotent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	prevBlock, _ := db.GetBlockByHeight(0)
	block, _ := CreateBlock(prevBlock, kp.Address, nil, 0)

	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal("second apply should be idempotent")
	}
}

func TestApplyBlockConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	prevBlock, _ := db.GetBlockByHeight(0)
	block1, _ := CreateBlock(prevBlock, kp.Address, nil, 0)
	if err := ApplyBlock(db, st, block1); err != nil {
		t.Fatal(err)
	}

	block2, _ := CreateBlock(prevBlock, "0xother", nil, 0)
	if err := ApplyBlock(db, st, block2); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestTxFailureDoesNotCrashBlock(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	brokeKP := genKeyPair(t)
	brokeAcct := db.GetOrCreateAccount(brokeKP.Address)
	brokeAcct.WhiteBalance = 100
	brokeAcct.PublicKey = brokeKP.PublicKey
	brokeAcct.Nonce = 0
	db.SaveAccount(brokeAcct)

	prevBlock, _ := db.GetBlockByHeight(0)

	badTx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      brokeKP.Address,
		To:        "0xsomeone",
		Amount:    999_999_999,
		Nonce:     1,
		PublicKey: brokeKP.PublicKey,
	}
	signTx(t, &badTx, brokeKP.PrivateKey)

	block, _ := CreateBlock(prevBlock, kp.Address, []types.Transaction{badTx}, 0)
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatalf("block should not fail even if tx fails: %v", err)
	}

	receipt, err := db.GetReceipt(badTx.Hash)
	if err != nil {
		t.Fatal("failed receipt should still be recorded")
	}
	if receipt.Status != "failed" {
		t.Fatalf("expected failed receipt, got %s", receipt.Status)
	}

	brokeAfter, _ := db.GetAccount(brokeKP.Address)
	if brokeAfter.WhiteBalance != 100 {
		t.Fatalf("broke balance should be unchanged, got %d", brokeAfter.WhiteBalance)
	}
}

func TestOverflowAmountFeeRejected(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	attackerKP := genKeyPair(t)
	attackAcct := db.GetOrCreateAccount(attackerKP.Address)
	attackAcct.WhiteBalance = 1_000
	attackAcct.PublicKey = attackerKP.PublicKey
	attackAcct.Nonce = 0
	db.SaveAccount(attackAcct)

	prevBlock, _ := db.GetBlockByHeight(0)

	overflowTx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      attackerKP.Address,
		To:        "0xvictim",
		Amount:    ^uint64(0),
		Nonce:     1,
		PublicKey: attackerKP.PublicKey,
	}
	signTx(t, &overflowTx, attackerKP.PrivateKey)

	block, _ := CreateBlock(prevBlock, kp.Address, []types.Transaction{overflowTx}, 0)
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(overflowTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("overflow attack should produce failed receipt")
	}

	attackAfter, _ := db.GetAccount(attackerKP.Address)
	if attackAfter.WhiteBalance != 1_000 {
		t.Fatalf("attacker balance should be unchanged, got %d", attackAfter.WhiteBalance)
	}
}

func TestAMMSwapKInvariant(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	deployerKP := genKeyPair(t)
	createTestChain(t, db, deployerKP)
	st := state.New(db)

	deployParams := token.DeployParams{
		Name:           "TestCoin",
		Symbol:         "TST",
		PoolRatio:      50,
		TeamRatio:      50,
		InitWhite:      1_000_000_000,
		ReleaseMonthly: 10_000,
		MultiSigAddr:   "0xmultisig",
	}
	payload, _ := json.Marshal(deployParams)

	prevBlock, _ := db.GetBlockByHeight(0)

	deployTx := types.Transaction{
		Type:      types.TxDeployBlue,
		From:      deployerKP.Address,
		Nonce:     1,
		Payload:   payload,
		PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)

	block, _ := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	tokenID := token.GenerateTokenID(deployerKP.Address, "TestCoin", 1)
	pool, err := db.GetPool(tokenID)
	if err != nil {
		receipt, _ := db.GetReceipt(deployTx.Hash)
		t.Fatalf("pool not found: %v (receipt: %+v)", err, receipt)
	}
	if pool.WhiteReserve == 0 || pool.BlueReserve == 0 {
		t.Fatal("pool reserves should be non-zero")
	}

	deployerAcct, _ := db.GetAccount(deployerKP.Address)

	swapTx := types.Transaction{
		Type:      types.TxSwapWhiteBlue,
		From:      deployerKP.Address,
		TokenID:   tokenID,
		Amount:    100_000_000,
		Nonce:     deployerAcct.Nonce + 1,
		PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 30,
	}
	signTx(t, &swapTx, deployerKP.PrivateKey)

	prevBlock2, _ := db.GetBlockByHeight(1)
	block2, _ := CreateBlock(prevBlock2, deployerKP.Address, []types.Transaction{swapTx}, 0)
	if err := ApplyBlock(db, st, block2); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(swapTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("swap should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}
}

func TestTotalMintedTracked(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	prevBlock, _ := db.GetBlockByHeight(0)

	reward := uint64(50_000_000)
	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     kp.Address,
		Amount: reward,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	block, _ := CreateBlock(prevBlock, kp.Address, []types.Transaction{rewardTx}, reward)
	ApplyBlock(db, st, block)

	minted := db.GetTotalMinted()
	if minted != reward {
		t.Fatalf("total minted: expected %d, got %d", reward, minted)
	}
}

func TestCalcRewardCap(t *testing.T) {
	r := CalcReward(0, 0)
	if r != types.InitialReward {
		t.Fatalf("expected %d, got %d", types.InitialReward, r)
	}

	r = CalcReward(0, types.MaxWhiteSupply)
	if r != 0 {
		t.Fatalf("expected 0 when supply exhausted, got %d", r)
	}

	r = CalcReward(0, types.MaxWhiteSupply-10)
	if r != 10 {
		t.Fatalf("expected 10 remaining, got %d", r)
	}

	r = CalcReward(types.BlocksPerYear*300, 0)
	if r != 0 {
		t.Fatalf("expected 0 for year 300, got %d", r)
	}
}

func TestSpeculativeValidation(t *testing.T) {
	aliceKP := genKeyPair(t)
	cache := make(map[string]*types.Account)
	cache[aliceKP.Address] = &types.Account{
		Address:      aliceKP.Address,
		PublicKey:    aliceKP.PublicKey,
		WhiteBalance: 1_000_000,
		Nonce:        0,
		BlueBalances: make(map[string]uint64),
	}

	tx1 := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xbob",
		Amount:    400_000,
		Nonce:     1,
		PublicKey: aliceKP.PublicKey,
	}

	if err := state.ValidateTransactionWithAccount(tx1, cache[aliceKP.Address]); err != nil {
		t.Fatalf("tx1 should pass: %v", err)
	}
	state.SpeculativeApply(cache, tx1)

	tx2 := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xcharlie",
		Amount:    400_000,
		Nonce:     2,
		PublicKey: aliceKP.PublicKey,
	}

	if err := state.ValidateTransactionWithAccount(tx2, cache[aliceKP.Address]); err != nil {
		t.Fatalf("tx2 should pass: %v", err)
	}
	state.SpeculativeApply(cache, tx2)

	tx3 := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xdave",
		Amount:    400_000,
		Nonce:     3,
		PublicKey: aliceKP.PublicKey,
	}

	if err := state.ValidateTransactionWithAccount(tx3, cache[aliceKP.Address]); err == nil {
		t.Fatal("tx3 should fail: alice has insufficient balance after tx1+tx2")
	}
}

func TestSpeculativeNonceEnforcement(t *testing.T) {
	aliceKP := genKeyPair(t)
	cache := make(map[string]*types.Account)
	cache[aliceKP.Address] = &types.Account{
		Address:      aliceKP.Address,
		PublicKey:    aliceKP.PublicKey,
		WhiteBalance: 10_000_000,
		Nonce:        0,
		BlueBalances: make(map[string]uint64),
	}

	tx1 := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xbob",
		Amount:    100,
		Nonce:     1,
		PublicKey: aliceKP.PublicKey,
	}
	state.ValidateTransactionWithAccount(tx1, cache[aliceKP.Address])
	state.SpeculativeApply(cache, tx1)

	txBadNonce := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xcharlie",
		Amount:    100,
		Nonce:     1,
		PublicKey: aliceKP.PublicKey,
	}

	if err := state.ValidateTransactionWithAccount(txBadNonce, cache[aliceKP.Address]); err == nil {
		t.Fatal("replayed nonce should be rejected")
	}

	txCorrectNonce := &types.Transaction{
		Type:      types.TxTransferWhite,
		From:      aliceKP.Address,
		To:        "0xcharlie",
		Amount:    100,
		Nonce:     2,
		PublicKey: aliceKP.PublicKey,
	}

	if err := state.ValidateTransactionWithAccount(txCorrectNonce, cache[aliceKP.Address]); err != nil {
		t.Fatalf("nonce 2 should pass: %v", err)
	}
}

func TestBlockSignature(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)

	prevBlock, _ := db.GetBlockByHeight(0)
	block, _ := CreateBlock(prevBlock, kp.Address, nil, 0)

	if err := SignBlock(block, kp.PrivateKey); err != nil {
		t.Fatal(err)
	}

	if !VerifyBlockSignature(block, kp.PublicKey) {
		t.Fatal("valid block signature should verify")
	}

	otherKP := genKeyPair(t)
	if VerifyBlockSignature(block, otherKP.PublicKey) {
		t.Fatal("block signature should not verify with wrong key")
	}
}

func TestVerifyBlockHashClearsSignature(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)

	prevBlock, _ := db.GetBlockByHeight(0)
	block, _ := CreateBlock(prevBlock, kp.Address, nil, 0)

	SignBlock(block, kp.PrivateKey)

	if !VerifyBlockHash(block) {
		t.Fatal("VerifyBlockHash should work even with Signature field set")
	}
}

func setupValidatorSet(t *testing.T, db *storage.DB, validators ...types.ValidatorRecord) {
	t.Helper()
	vs := &types.ValidatorSet{Validators: validators, UpdatedAt: 0}
	if err := db.SaveValidatorSet(vs); err != nil {
		t.Fatal(err)
	}
}

func buildBlock(t *testing.T, db *storage.DB, validator string, txs []types.Transaction) *types.Block {
	t.Helper()
	height := db.GetLatestHeight()
	prev, err := db.GetBlockByHeight(height)
	if err != nil {
		t.Fatal(err)
	}
	block, err := CreateBlock(prev, validator, txs, 0)
	if err != nil {
		t.Fatal(err)
	}
	return block
}

func TestSlashDoubleSign(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	targetKP := genKeyPair(t)
	targetAcct := db.GetOrCreateAccount(targetKP.Address)
	targetAcct.WhiteBalance = 500_000_000
	targetAcct.StakedBalance = 1_000_000_000
	targetAcct.PublicKey = targetKP.PublicKey
	db.SaveAccount(targetAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: targetKP.Address, PublicKey: targetKP.PublicKey, Status: types.ValidatorStatusActive},
	)

	submitterKP := genKeyPair(t)
	submitterAcct := db.GetOrCreateAccount(submitterKP.Address)
	submitterAcct.WhiteBalance = 100_000_000
	submitterAcct.PublicKey = submitterKP.PublicKey
	db.SaveAccount(submitterAcct)

	h1 := types.BlockHeader{Height: 10, Validator: targetKP.Address, Hash: crypto.SHA256Hex([]byte("block_a"))}
	sig1, _ := crypto.Sign(targetKP.PrivateKey, []byte(h1.Hash))
	h1.Signature = sig1

	h2 := types.BlockHeader{Height: 10, Validator: targetKP.Address, Hash: crypto.SHA256Hex([]byte("block_b"))}
	sig2, _ := crypto.Sign(targetKP.PrivateKey, []byte(h2.Hash))
	h2.Signature = sig2

	evidence := struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}{Header1: h1, Header2: h2}
	payload, _ := json.Marshal(evidence)

	slashTx := types.Transaction{
		Type:      types.TxSlashEvidence,
		From:      submitterKP.Address,
		Amount:    1,
		Nonce:     1,
		Payload:   payload,
		PublicKey: submitterKP.PublicKey,
		Timestamp: 1750000015,
	}
	signTx(t, &slashTx, submitterKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{slashTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(slashTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("slash should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	vs := db.GetValidatorSet()
	rec := vs.FindRecord(targetKP.Address)
	if rec.Status != types.ValidatorStatusSlashed {
		t.Fatalf("target should be slashed, got %s", rec.Status)
	}

	targetAfter, _ := db.GetAccount(targetKP.Address)
	if targetAfter.StakedBalance != 0 {
		t.Fatalf("target stake should be 0, got %d", targetAfter.StakedBalance)
	}

	submitterAfter, _ := db.GetAccount(submitterKP.Address)
	expectedBalance := uint64(100_000_000) + types.SlashReward
	if submitterAfter.WhiteBalance != expectedBalance {
		t.Fatalf("submitter should receive slash reward: expected %d, got %d", expectedBalance, submitterAfter.WhiteBalance)
	}
}

func TestSlashSameHash(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	targetKP := genKeyPair(t)
	targetAcct := db.GetOrCreateAccount(targetKP.Address)
	targetAcct.StakedBalance = 1_000_000_000
	targetAcct.PublicKey = targetKP.PublicKey
	db.SaveAccount(targetAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: targetKP.Address, PublicKey: targetKP.PublicKey, Status: types.ValidatorStatusActive},
	)

	sameHash := crypto.SHA256Hex([]byte("same_block"))
	sig, _ := crypto.Sign(targetKP.PrivateKey, []byte(sameHash))
	h := types.BlockHeader{Height: 10, Validator: targetKP.Address, Hash: sameHash, Signature: sig}

	evidence := struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}{Header1: h, Header2: h}
	payload, _ := json.Marshal(evidence)

	slashTx := types.Transaction{
		Type: types.TxSlashEvidence, From: validatorKP.Address,
		Amount: 1, Nonce: 1, Payload: payload, PublicKey: validatorKP.PublicKey,
	}
	signTx(t, &slashTx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{slashTx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(slashTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("identical headers should be rejected")
	}
}

func TestSlashAlreadySlashed(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	targetKP := genKeyPair(t)
	targetAcct := db.GetOrCreateAccount(targetKP.Address)
	targetAcct.PublicKey = targetKP.PublicKey
	db.SaveAccount(targetAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: targetKP.Address, PublicKey: targetKP.PublicKey, Status: types.ValidatorStatusSlashed},
	)

	h1 := types.BlockHeader{Height: 10, Validator: targetKP.Address, Hash: crypto.SHA256Hex([]byte("a"))}
	sig1, _ := crypto.Sign(targetKP.PrivateKey, []byte(h1.Hash))
	h1.Signature = sig1
	h2 := types.BlockHeader{Height: 10, Validator: targetKP.Address, Hash: crypto.SHA256Hex([]byte("b"))}
	sig2, _ := crypto.Sign(targetKP.PrivateKey, []byte(h2.Hash))
	h2.Signature = sig2

	evidence := struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}{Header1: h1, Header2: h2}
	payload, _ := json.Marshal(evidence)

	slashTx := types.Transaction{
		Type: types.TxSlashEvidence, From: validatorKP.Address,
		Amount: 1, Nonce: 1, Payload: payload, PublicKey: validatorKP.PublicKey,
	}
	signTx(t, &slashTx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{slashTx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(slashTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("already slashed should be rejected")
	}
}

func TestMultiSigRegisterProposeApprove(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owner1KP := genKeyPair(t)
	owner2KP := genKeyPair(t)
	for _, kp := range []*types.KeyPair{owner1KP, owner2KP} {
		a := db.GetOrCreateAccount(kp.Address)
		a.WhiteBalance = 1_000_000_000
		a.PublicKey = kp.PublicKey
		db.SaveAccount(a)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{owner1KP.Address, owner2KP.Address}, Threshold: 2})

	regTx := types.Transaction{
		Type: types.TxMultiSigRegister, From: owner1KP.Address,
		Amount: 1, Nonce: 1, Payload: regPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &regTx, owner1KP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}
	receipt, _ := db.GetReceipt(regTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("register should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	msAddr := types.DeriveMultiSigAddress([]string{owner1KP.Address, owner2KP.Address}, 2)
	ms, err := db.GetMultiSig(msAddr)
	if err != nil {
		t.Fatalf("multisig account not found: %v", err)
	}
	if ms.Threshold != 2 || len(ms.Owners) != 2 {
		t.Fatalf("multisig config wrong: threshold=%d owners=%d", ms.Threshold, len(ms.Owners))
	}

	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 1_000_000
	db.SaveAccount(msAcct)

	innerTx := types.Transaction{Type: types.TxTransferWhite, From: msAddr, To: validatorKP.Address, Amount: 100}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx})

	propTx := types.Transaction{
		Type: types.TxMultiSigPropose, From: owner1KP.Address,
		Amount: 2, Nonce: 2, Payload: propPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &propTx, owner1KP.PrivateKey)

	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block2)

	receipt2, _ := db.GetReceipt(propTx.Hash)
	if receipt2.Status != "success" {
		t.Fatalf("propose should succeed, got %s (err: %s)", receipt2.Status, receipt2.Error)
	}

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	approveTx := types.Transaction{
		Type: types.TxMultiSigApprove, From: owner2KP.Address,
		Amount: 3, Nonce: 1, Payload: approvePayload, PublicKey: owner2KP.PublicKey,
	}
	signTx(t, &approveTx, owner2KP.PrivateKey)

	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approveTx})
	ApplyBlock(db, st, block3)

	receipt3, _ := db.GetReceipt(approveTx.Hash)
	if receipt3.Status != "success" {
		t.Fatalf("approve should succeed, got %s (err: %s)", receipt3.Status, receipt3.Error)
	}
}

func TestMultiSigRegisterBadThreshold(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	payload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{validatorKP.Address, "0xother"}, Threshold: 5})

	tx := types.Transaction{
		Type: types.TxMultiSigRegister, From: validatorKP.Address,
		Amount: 1, Nonce: 1, Payload: payload, PublicKey: validatorKP.PublicKey,
	}
	signTx(t, &tx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{tx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(tx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("threshold > owner count should be rejected")
	}
}

func TestMultiSigRegisterOneOwner(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	payload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{validatorKP.Address}, Threshold: 1})

	tx := types.Transaction{
		Type: types.TxMultiSigRegister, From: validatorKP.Address,
		Amount: 1, Nonce: 1, Payload: payload, PublicKey: validatorKP.PublicKey,
	}
	signTx(t, &tx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{tx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(tx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("single owner should be rejected")
	}
}

func TestMultiSigDuplicateApprove(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owner1KP := genKeyPair(t)
	owner2KP := genKeyPair(t)
	for _, kp := range []*types.KeyPair{owner1KP, owner2KP} {
		a := db.GetOrCreateAccount(kp.Address)
		a.WhiteBalance = 1_000_000_000
		a.PublicKey = kp.PublicKey
		db.SaveAccount(a)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{owner1KP.Address, owner2KP.Address}, Threshold: 2})

	regTx := types.Transaction{
		Type: types.TxMultiSigRegister, From: owner1KP.Address,
		Amount: 1, Nonce: 1, Payload: regPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &regTx, owner1KP.PrivateKey)
	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	ApplyBlock(db, st, block)

	msAddr := types.DeriveMultiSigAddress([]string{owner1KP.Address, owner2KP.Address}, 2)

	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 1_000_000
	db.SaveAccount(msAcct)

	innerTx := types.Transaction{Type: types.TxTransferWhite, From: msAddr, To: validatorKP.Address, Amount: 100}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx})

	propTx := types.Transaction{
		Type: types.TxMultiSigPropose, From: owner1KP.Address,
		Amount: 2, Nonce: 2, Payload: propPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &propTx, owner1KP.PrivateKey)
	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block2)

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	dupTx := types.Transaction{
		Type: types.TxMultiSigApprove, From: owner1KP.Address,
		Amount: 3, Nonce: 3, Payload: approvePayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &dupTx, owner1KP.PrivateKey)
	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{dupTx})
	ApplyBlock(db, st, block3)

	receipt, _ := db.GetReceipt(dupTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("duplicate approval should be rejected")
	}
}

func TestAMMSlippageRejection(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	deployerKP := genKeyPair(t)
	createTestChain(t, db, deployerKP)
	st := state.New(db)

	deployParams := token.DeployParams{
		Name: "SlipCoin", Symbol: "SLP",
		PoolRatio: 50, TeamRatio: 50, InitWhite: 1_000_000_000,
		ReleaseMonthly: 10000, MultiSigAddr: "0xms",
	}
	payload, _ := json.Marshal(deployParams)
	prevBlock, _ := db.GetBlockByHeight(0)

	deployTx := types.Transaction{
		Type: types.TxDeployBlue, From: deployerKP.Address, Nonce: 1,
		Payload: payload, PublicKey: deployerKP.PublicKey, Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)
	block1, _ := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	ApplyBlock(db, st, block1)

	tokenID := token.GenerateTokenID(deployerKP.Address, "SlipCoin", 1)

	deployerAcct, _ := db.GetAccount(deployerKP.Address)

	swapTx := types.Transaction{
		Type: types.TxSwapWhiteBlue, From: deployerKP.Address,
		TokenID: tokenID, Amount: 100_000_000, MinAmountOut: ^uint64(0),
		Nonce: deployerAcct.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 30,
	}
	signTx(t, &swapTx, deployerKP.PrivateKey)

	block2 := buildBlock(t, db, deployerKP.Address, []types.Transaction{swapTx})
	ApplyBlock(db, st, block2)

	receipt, _ := db.GetReceipt(swapTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("swap with impossible min-out should fail due to slippage")
	}
}

func TestAMMBlueToWhiteSwap(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	deployerKP := genKeyPair(t)
	createTestChain(t, db, deployerKP)
	st := state.New(db)

	deployParams := token.DeployParams{
		Name: "RevCoin", Symbol: "REV",
		PoolRatio: 50, TeamRatio: 50, InitWhite: 1_000_000_000,
		ReleaseMonthly: 10000, MultiSigAddr: "0xms",
	}
	payload, _ := json.Marshal(deployParams)
	prevBlock, _ := db.GetBlockByHeight(0)

	deployTx := types.Transaction{
		Type: types.TxDeployBlue, From: deployerKP.Address, Nonce: 1,
		Payload: payload, PublicKey: deployerKP.PublicKey, Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)
	block1, _ := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	ApplyBlock(db, st, block1)

	tokenID := token.GenerateTokenID(deployerKP.Address, "RevCoin", 1)

	deployerAcct, _ := db.GetAccount(deployerKP.Address)
	swapW2B := types.Transaction{
		Type: types.TxSwapWhiteBlue, From: deployerKP.Address,
		TokenID: tokenID, Amount: 100_000_000,
		Nonce: deployerAcct.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 30,
	}
	signTx(t, &swapW2B, deployerKP.PrivateKey)
	block2 := buildBlock(t, db, deployerKP.Address, []types.Transaction{swapW2B})
	ApplyBlock(db, st, block2)

	r1, _ := db.GetReceipt(swapW2B.Hash)
	if r1.Status != "success" {
		t.Fatalf("white-to-blue should succeed: %s", r1.Error)
	}

	deployerAcct2, _ := db.GetAccount(deployerKP.Address)
	blueBal := deployerAcct2.BlueBalances[tokenID]
	if blueBal == 0 {
		t.Fatal("should have blue balance after swap")
	}

	swapB2W := types.Transaction{
		Type: types.TxSwapBlueWhite, From: deployerKP.Address,
		TokenID: tokenID, Amount: blueBal / 2,
		Nonce: deployerAcct2.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 45,
	}
	signTx(t, &swapB2W, deployerKP.PrivateKey)
	block3 := buildBlock(t, db, deployerKP.Address, []types.Transaction{swapB2W})
	ApplyBlock(db, st, block3)

	r2, _ := db.GetReceipt(swapB2W.Hash)
	if r2.Status != "success" {
		t.Fatalf("blue-to-white should succeed: %s", r2.Error)
	}

	deployerAcct3, _ := db.GetAccount(deployerKP.Address)
	if deployerAcct3.WhiteBalance <= deployerAcct2.WhiteBalance {
		t.Fatal("white balance should increase after blue-to-white swap")
	}
}

func TestValidatorJoinAlreadyActive(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
	)

	joinTx := types.Transaction{
		Type: types.TxValidatorJoin, From: validatorKP.Address,
		Amount: 1, Nonce: 1, PublicKey: validatorKP.PublicKey,
		Payload: []byte(`{"nonce":0}`),
	}
	signTx(t, &joinTx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{joinTx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(joinTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("joining when already active should be rejected")
	}
}

func TestValidatorJoinBanned(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	bannedKP := genKeyPair(t)
	bannedAcct := db.GetOrCreateAccount(bannedKP.Address)
	bannedAcct.WhiteBalance = 1_000_000_000
	bannedAcct.PublicKey = bannedKP.PublicKey
	db.SaveAccount(bannedAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: bannedKP.Address, PublicKey: bannedKP.PublicKey, Status: types.ValidatorStatusSlashed},
	)

	joinTx := types.Transaction{
		Type: types.TxValidatorJoin, From: bannedKP.Address,
		Amount: 1, Nonce: 1, PublicKey: bannedKP.PublicKey,
		Payload: []byte(`{"nonce":0}`),
	}
	signTx(t, &joinTx, bannedKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{joinTx})
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(joinTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("slashed validator should not be allowed to rejoin")
	}
}

func TestValidatorExitWhileSuspended(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	suspKP := genKeyPair(t)
	suspAcct := db.GetOrCreateAccount(suspKP.Address)
	suspAcct.StakedBalance = 1_000_000_000
	suspAcct.PublicKey = suspKP.PublicKey
	db.SaveAccount(suspAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: suspKP.Address, PublicKey: suspKP.PublicKey, Status: types.ValidatorStatusSuspended},
	)

	exitTx := types.Transaction{
		Type: types.TxValidatorExit, From: suspKP.Address,
		Amount: 1, Nonce: 1, PublicKey: suspKP.PublicKey,
	}
	signTx(t, &exitTx, suspKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{exitTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(exitTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("exit from suspended should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	vs := db.GetValidatorSet()
	rec := vs.FindRecord(suspKP.Address)
	if rec.Status != types.ValidatorStatusRemoved {
		t.Fatalf("should be removed, got %s", rec.Status)
	}

	after, _ := db.GetAccount(suspKP.Address)
	if after.StakedBalance != 0 {
		t.Fatalf("stake should be burned, got %d", after.StakedBalance)
	}
}

func TestSignatureRequired(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	kp := genKeyPair(t)
	createTestChain(t, db, kp)
	st := state.New(db)

	prevBlock, _ := db.GetBlockByHeight(0)

	noSigTx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      kp.Address,
		To:        "0xbob",
		Amount:    100,
		Nonce:     1,
		PublicKey: kp.PublicKey,
		Hash:      "fakehash",
	}

	block, _ := CreateBlock(prevBlock, kp.Address, []types.Transaction{noSigTx}, 0)
	ApplyBlock(db, st, block)

	receipt, _ := db.GetReceipt(noSigTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("tx without signature should fail")
	}
}

func TestHeartbeatPassesValidation(t *testing.T) {
	acct := &types.Account{
		Address:      "0xcandidate",
		WhiteBalance: 0,
		Nonce:        0,
		BlueBalances: make(map[string]uint64),
	}
	tx := &types.Transaction{
		Type:      types.TxHeartbeat,
		From:      "0xcandidate",
		PublicKey: "04abcd",
		Amount:    100,
	}
	if err := state.ValidateTransactionWithAccount(tx, acct); err != nil {
		t.Fatalf("heartbeat should pass ValidateTransactionWithAccount: %v", err)
	}
}

func TestMultiSigExecutesTransfer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owner1KP := genKeyPair(t)
	owner2KP := genKeyPair(t)
	for _, kp := range []*types.KeyPair{owner1KP, owner2KP} {
		a := db.GetOrCreateAccount(kp.Address)
		a.WhiteBalance = 1_000_000_000
		a.PublicKey = kp.PublicKey
		db.SaveAccount(a)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{owner1KP.Address, owner2KP.Address}, Threshold: 2})

	regTx := types.Transaction{
		Type: types.TxMultiSigRegister, From: owner1KP.Address,
		Amount: 1, Nonce: 1, Payload: regPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &regTx, owner1KP.PrivateKey)
	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	ApplyBlock(db, st, block)

	msAddr := types.DeriveMultiSigAddress([]string{owner1KP.Address, owner2KP.Address}, 2)

	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 500_000_000
	db.SaveAccount(msAcct)

	receiver := "0xreceiver_test"
	innerTx := types.Transaction{
		Type: types.TxTransferWhite, From: msAddr, To: receiver, Amount: 100_000_000,
	}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx})

	propTx := types.Transaction{
		Type: types.TxMultiSigPropose, From: owner1KP.Address,
		Amount: 2, Nonce: 2, Payload: propPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &propTx, owner1KP.PrivateKey)
	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block2)

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	approveTx := types.Transaction{
		Type: types.TxMultiSigApprove, From: owner2KP.Address,
		Amount: 3, Nonce: 1, Payload: approvePayload, PublicKey: owner2KP.PublicKey,
	}
	signTx(t, &approveTx, owner2KP.PrivateKey)
	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approveTx})
	if err := ApplyBlock(db, st, block3); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(approveTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("approve should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	receiverAcct, _ := db.GetAccount(receiver)
	if receiverAcct.WhiteBalance != 100_000_000 {
		t.Fatalf("receiver should have 100_000_000, got %d", receiverAcct.WhiteBalance)
	}

	msAfter, _ := db.GetAccount(msAddr)
	if msAfter.WhiteBalance >= 500_000_000 {
		t.Fatalf("multisig balance should decrease, got %d", msAfter.WhiteBalance)
	}
}

func TestMultiSigApproveExpired(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owner1KP := genKeyPair(t)
	owner2KP := genKeyPair(t)
	for _, kp := range []*types.KeyPair{owner1KP, owner2KP} {
		a := db.GetOrCreateAccount(kp.Address)
		a.WhiteBalance = 1_000_000_000
		a.PublicKey = kp.PublicKey
		db.SaveAccount(a)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: []string{owner1KP.Address, owner2KP.Address}, Threshold: 2})

	regTx := types.Transaction{
		Type: types.TxMultiSigRegister, From: owner1KP.Address,
		Amount: 1, Nonce: 1, Payload: regPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &regTx, owner1KP.PrivateKey)
	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	ApplyBlock(db, st, block)

	msAddr := types.DeriveMultiSigAddress([]string{owner1KP.Address, owner2KP.Address}, 2)
	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 1_000_000
	db.SaveAccount(msAcct)

	innerTx := types.Transaction{Type: types.TxTransferWhite, From: msAddr, To: validatorKP.Address, Amount: 100}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx})

	propTx := types.Transaction{
		Type: types.TxMultiSigPropose, From: owner1KP.Address,
		Amount: 100, Nonce: 2, Payload: propPayload, PublicKey: owner1KP.PublicKey,
	}
	signTx(t, &propTx, owner1KP.PrivateKey)
	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block2)

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	approveTx := types.Transaction{
		Type: types.TxMultiSigApprove, From: owner2KP.Address,
		Amount: 100 + 40320 + 1, Nonce: 1, Payload: approvePayload, PublicKey: owner2KP.PublicKey,
	}
	signTx(t, &approveTx, owner2KP.PrivateKey)
	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approveTx})
	ApplyBlock(db, st, block3)

	receipt, _ := db.GetReceipt(approveTx.Hash)
	if receipt.Status != "failed" {
		t.Fatal("expired proposal should be rejected")
	}
}

func TestBlueBurn(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	deployerKP := genKeyPair(t)
	createTestChain(t, db, deployerKP)
	st := state.New(db)

	deployParams := token.DeployParams{
		Name: "BurnCoin", Symbol: "BRN",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1_000_000_000, ReleaseMonthly: 10000,
		MultiSigAddr: "0xms",
	}
	payload, _ := json.Marshal(deployParams)
	prevBlock, _ := db.GetBlockByHeight(0)

	deployTx := types.Transaction{
		Type: types.TxDeployBlue, From: deployerKP.Address, Nonce: 1,
		Payload: payload, PublicKey: deployerKP.PublicKey, Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)
	block1, _ := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	ApplyBlock(db, st, block1)

	tokenID := token.GenerateTokenID(deployerKP.Address, "BurnCoin", 1)

	deployerAcct, _ := db.GetAccount(deployerKP.Address)
	swapTx := types.Transaction{
		Type: types.TxSwapWhiteBlue, From: deployerKP.Address,
		TokenID: tokenID, Amount: 100_000_000,
		Nonce: deployerAcct.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 30,
	}
	signTx(t, &swapTx, deployerKP.PrivateKey)
	block2 := buildBlock(t, db, deployerKP.Address, []types.Transaction{swapTx})
	ApplyBlock(db, st, block2)

	afterSwap, _ := db.GetAccount(deployerKP.Address)
	blueBal := afterSwap.BlueBalances[tokenID]
	if blueBal == 0 {
		t.Fatal("should have blue balance after swap")
	}

	burnAmount := blueBal / 2
	burnTx := types.Transaction{
		Type: types.TxBlueBurn, From: deployerKP.Address,
		TokenID: tokenID, Amount: burnAmount,
		Nonce: afterSwap.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 45,
	}
	signTx(t, &burnTx, deployerKP.PrivateKey)
	block3 := buildBlock(t, db, deployerKP.Address, []types.Transaction{burnTx})
	ApplyBlock(db, st, block3)

	r, _ := db.GetReceipt(burnTx.Hash)
	if r.Status != "success" {
		t.Fatalf("burn should succeed, got %s (err: %s)", r.Status, r.Error)
	}

	afterBurn, _ := db.GetAccount(deployerKP.Address)
	if afterBurn.BlueBalances[tokenID] != blueBal-burnAmount {
		t.Fatalf("blue balance should decrease by burn amount")
	}

	blueState, _ := db.GetBlueCoinState(tokenID)
	if blueState.Burned < burnAmount {
		t.Fatalf("burned should include burn amount, got %d", blueState.Burned)
	}
}

func TestSwapBurnsMechanism(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	deployerKP := genKeyPair(t)
	createTestChain(t, db, deployerKP)
	st := state.New(db)

	deployParams := token.DeployParams{
		Name: "SwapBurn", Symbol: "SWB",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1_000_000_000, ReleaseMonthly: 10000,
		MultiSigAddr: "0xms",
	}
	payload, _ := json.Marshal(deployParams)
	prevBlock, _ := db.GetBlockByHeight(0)

	deployTx := types.Transaction{
		Type: types.TxDeployBlue, From: deployerKP.Address, Nonce: 1,
		Payload: payload, PublicKey: deployerKP.PublicKey, Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, deployerKP.PrivateKey)
	block1, _ := CreateBlock(prevBlock, deployerKP.Address, []types.Transaction{deployTx}, 0)
	ApplyBlock(db, st, block1)

	tokenID := token.GenerateTokenID(deployerKP.Address, "SwapBurn", 1)

	stateBefore, _ := db.GetBlueCoinState(tokenID)
	burnedBefore := stateBefore.Burned

	deployerAcct, _ := db.GetAccount(deployerKP.Address)
	swapTx := types.Transaction{
		Type: types.TxSwapWhiteBlue, From: deployerKP.Address,
		TokenID: tokenID, Amount: 500_000_000,
		Nonce: deployerAcct.Nonce + 1, PublicKey: deployerKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 30,
	}
	signTx(t, &swapTx, deployerKP.PrivateKey)
	block2 := buildBlock(t, db, deployerKP.Address, []types.Transaction{swapTx})
	ApplyBlock(db, st, block2)

	stateAfter, _ := db.GetBlueCoinState(tokenID)
	if stateAfter.Burned <= burnedBefore {
		t.Fatalf("swap should increase burned amount: before=%d after=%d", burnedBefore, stateAfter.Burned)
	}
}
