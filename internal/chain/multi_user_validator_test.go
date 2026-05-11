package chain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/token"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func findPoWNonce(address string, height uint64) uint64 {
	for nonce := uint64(0); ; nonce++ {
		data := fmt.Sprintf("%s:%d:%d", address, height, nonce)
		hash := sha256.Sum256([]byte(data))
		if hash[0] == 0 {
			return nonce
		}
	}
}

func resetDevMode() {
	types.DevBlockInterval = types.BlockInterval
	types.DevUptimeBlocks = types.UptimeBlocks
	types.DevSuspendBlocks = types.SuspendBlocks
	types.DevEvictBlocks = types.EvictBlocks
	types.DevPoWDifficulty = types.PoWDifficulty
}

func TestFiveValidatorLifecycle(t *testing.T) {
	types.SetDevMode()
	defer resetDevMode()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	validators := make([]*types.KeyPair, 5)
	for i := range validators {
		validators[i] = genKeyPair(t)
	}

	createTestChain(t, db, validators[0])
	st := state.New(db)

	setupValidatorSet(t, db,
		types.ValidatorRecord{
			Address:             validators[0].Address,
			PublicKey:           validators[0].PublicKey,
			Status:              types.ValidatorStatusActive,
			JoinHeight:          0,
			LastHeartbeatHeight: 1,
		},
	)

	for i := 1; i <= 4; i++ {
		acct := db.GetOrCreateAccount(validators[i].Address)
		acct.WhiteBalance = 5_000_000_000
		acct.PublicKey = validators[i].PublicKey
		db.SaveAccount(acct)
	}

	joinHeight := uint64(1)
	for i := 1; i <= 4; i++ {
		nonce := findPoWNonce(validators[i].Address, joinHeight)
		payload, _ := json.Marshal(struct {
			Nonce uint64 `json:"nonce"`
		}{Nonce: nonce})

		joinTx := types.Transaction{
			Type:      types.TxValidatorJoin,
			From:      validators[i].Address,
			Amount:    joinHeight,
			Nonce:     1,
			Payload:   payload,
			PublicKey: validators[i].PublicKey,
		}
		signTx(t, &joinTx, validators[i].PrivateKey)

		block := buildBlock(t, db, validators[0].Address, []types.Transaction{joinTx})
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatalf("failed to apply join block for validator%d: %v", i, err)
		}

		receipt, _ := db.GetReceipt(joinTx.Hash)
		if receipt.Status != "success" {
			t.Fatalf("validator%d join should succeed, got %s (err: %s)", i, receipt.Status, receipt.Error)
		}
	}

	vs := db.GetValidatorSet()
	active := vs.ActiveValidators()
	if len(active) != 5 {
		t.Fatalf("expected 5 active validators, got %d", len(active))
	}

	hbHeight := uint64(6)
	for i := 0; i < 5; i++ {
		if i == 2 {
			continue
		}
		hbTx := types.Transaction{
			Type:      types.TxHeartbeat,
			From:      validators[i].Address,
			PublicKey: validators[i].PublicKey,
			Amount:    hbHeight,
			Hash:      crypto.SHA256Hex([]byte(fmt.Sprintf("hb-init:%s:%d", validators[i].Address, hbHeight))),
		}
		block := buildBlock(t, db, validators[0].Address, []types.Transaction{hbTx})
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatalf("failed to apply heartbeat block for validator%d: %v", i, err)
		}
	}

	suspendTarget := hbHeight + types.GetSuspendBlocks() + 1
	for h := db.GetLatestHeight() + 1; h <= suspendTarget; h++ {
		var hbTxs []types.Transaction
		for i := 0; i < 5; i++ {
			if i == 2 {
				continue
			}
			hbTx := types.Transaction{
				Type:      types.TxHeartbeat,
				From:      validators[i].Address,
				PublicKey: validators[i].PublicKey,
				Amount:    h,
				Hash:      crypto.SHA256Hex([]byte(fmt.Sprintf("hb:%s:%d", validators[i].Address, h))),
			}
			hbTxs = append(hbTxs, hbTx)
		}
		block := buildBlock(t, db, validators[0].Address, hbTxs)
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatalf("failed to apply block at height targeting suspend: %v", err)
		}
	}

	vs = db.GetValidatorSet()
	rec2 := vs.FindRecord(validators[2].Address)
	if rec2 == nil || rec2.Status != types.ValidatorStatusSuspended {
		status := "nil"
		if rec2 != nil {
			status = rec2.Status
		}
		t.Fatalf("validator2 should be suspended, got %s", status)
	}

	currentHeight := db.GetLatestHeight() + 1
	recoveryHbTx := types.Transaction{
		Type:      types.TxHeartbeat,
		From:      validators[2].Address,
		PublicKey: validators[2].PublicKey,
		Amount:    currentHeight,
		Hash:      crypto.SHA256Hex([]byte(fmt.Sprintf("hb-recover:%s:%d", validators[2].Address, currentHeight))),
	}
	block := buildBlock(t, db, validators[0].Address, []types.Transaction{recoveryHbTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatalf("failed to apply recovery heartbeat block: %v", err)
	}

	vs = db.GetValidatorSet()
	rec2 = vs.FindRecord(validators[2].Address)
	if rec2 == nil || rec2.Status != types.ValidatorStatusActive {
		status := "nil"
		if rec2 != nil {
			status = rec2.Status
		}
		t.Fatalf("validator2 should be active after heartbeat, got %s", status)
	}

	latestH := db.GetLatestHeight()
	for i := 0; i < 5; i++ {
		if i == 3 {
			continue
		}
		hbTx := types.Transaction{
			Type:      types.TxHeartbeat,
			From:      validators[i].Address,
			PublicKey: validators[i].PublicKey,
			Amount:    latestH + 1,
			Hash:      crypto.SHA256Hex([]byte(fmt.Sprintf("hb-reset3:%s:%d", validators[i].Address, latestH+1))),
		}
		block := buildBlock(t, db, validators[0].Address, []types.Transaction{hbTx})
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatalf("failed to apply reset heartbeat for validator%d: %v", i, err)
		}
	}

	evictTarget := latestH + 1 + types.GetEvictBlocks() + 1
	for h := db.GetLatestHeight() + 1; h <= evictTarget; h++ {
		var hbTxs []types.Transaction
		for i := 0; i < 5; i++ {
			if i == 3 {
				continue
			}
			hbTx := types.Transaction{
				Type:      types.TxHeartbeat,
				From:      validators[i].Address,
				PublicKey: validators[i].PublicKey,
				Amount:    h,
				Hash:      crypto.SHA256Hex([]byte(fmt.Sprintf("hb-evict:%s:%d", validators[i].Address, h))),
			}
			hbTxs = append(hbTxs, hbTx)
		}
		block := buildBlock(t, db, validators[0].Address, hbTxs)
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatalf("failed to apply block targeting eviction: %v", err)
		}
	}

	vs = db.GetValidatorSet()
	rec3 := vs.FindRecord(validators[3].Address)
	if rec3 == nil || rec3.Status != types.ValidatorStatusRemoved {
		status := "nil"
		if rec3 != nil {
			status = rec3.Status
		}
		t.Fatalf("validator3 should be removed (evicted), got %s", status)
	}

	acct3, _ := db.GetAccount(validators[3].Address)
	if acct3.StakedBalance != 0 {
		t.Fatalf("evicted validator3 stake should be 0, got %d", acct3.StakedBalance)
	}
}

func TestDoubleSignSlashWithReward(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validator0KP := genKeyPair(t)
	validator1KP := genKeyPair(t)
	validator2KP := genKeyPair(t)

	createTestChain(t, db, validator0KP)
	st := state.New(db)

	for _, kp := range []*types.KeyPair{validator1KP, validator2KP} {
		acct := db.GetOrCreateAccount(kp.Address)
		acct.WhiteBalance = 500_000_000
		acct.StakedBalance = 1_000_000_000
		acct.PublicKey = kp.PublicKey
		db.SaveAccount(acct)
	}

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validator0KP.Address, PublicKey: validator0KP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: validator1KP.Address, PublicKey: validator1KP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: validator2KP.Address, PublicKey: validator2KP.PublicKey, Status: types.ValidatorStatusActive},
	)

	h1 := types.BlockHeader{Height: 10, Validator: validator1KP.Address, Hash: crypto.SHA256Hex([]byte("double_a"))}
	sig1, _ := crypto.Sign(validator1KP.PrivateKey, []byte(h1.Hash))
	h1.Signature = sig1

	h2 := types.BlockHeader{Height: 10, Validator: validator1KP.Address, Hash: crypto.SHA256Hex([]byte("double_b"))}
	sig2, _ := crypto.Sign(validator1KP.PrivateKey, []byte(h2.Hash))
	h2.Signature = sig2

	evidence := struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}{Header1: h1, Header2: h2}
	payload, _ := json.Marshal(evidence)

	v2Before, _ := db.GetAccount(validator2KP.Address)
	v2BalBefore := v2Before.WhiteBalance

	slashTx := types.Transaction{
		Type:      types.TxSlashEvidence,
		From:      validator2KP.Address,
		Amount:    1,
		Nonce:     1,
		Payload:   payload,
		PublicKey: validator2KP.PublicKey,
		Timestamp: 1750000015,
	}
	signTx(t, &slashTx, validator2KP.PrivateKey)

	block := buildBlock(t, db, validator0KP.Address, []types.Transaction{slashTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(slashTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("slash should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	vs := db.GetValidatorSet()
	rec := vs.FindRecord(validator1KP.Address)
	if rec.Status != types.ValidatorStatusSlashed {
		t.Fatalf("validator1 should be slashed, got %s", rec.Status)
	}

	v1After, _ := db.GetAccount(validator1KP.Address)
	if v1After.StakedBalance != 0 {
		t.Fatalf("slashed validator1 stake should be 0, got %d", v1After.StakedBalance)
	}

	v2After, _ := db.GetAccount(validator2KP.Address)
	expectedBalance := v2BalBefore + types.SlashReward
	if v2After.WhiteBalance != expectedBalance {
		t.Fatalf("submitter should receive SlashReward: expected %d, got %d", expectedBalance, v2After.WhiteBalance)
	}

	types.SetDevMode()
	defer resetDevMode()

	rejoinNonce := findPoWNonce(validator1KP.Address, 2)
	rejoinPayload, _ := json.Marshal(struct {
		Nonce uint64 `json:"nonce"`
	}{Nonce: rejoinNonce})

	rejoinTx := types.Transaction{
		Type:      types.TxValidatorJoin,
		From:      validator1KP.Address,
		Amount:    2,
		Nonce:     1,
		Payload:   rejoinPayload,
		PublicKey: validator1KP.PublicKey,
	}
	signTx(t, &rejoinTx, validator1KP.PrivateKey)

	block2 := buildBlock(t, db, validator0KP.Address, []types.Transaction{rejoinTx})
	ApplyBlock(db, st, block2)

	rejoinReceipt, _ := db.GetReceipt(rejoinTx.Hash)
	if rejoinReceipt.Status != "failed" {
		t.Fatal("slashed validator should not be allowed to rejoin")
	}
	if !strings.Contains(rejoinReceipt.Error, "permanently banned") {
		t.Fatalf("expected 'permanently banned' error, got: %s", rejoinReceipt.Error)
	}
}

func TestThreeOfFiveMultiSig(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owners := make([]*types.KeyPair, 5)
	ownerAddrs := make([]string, 5)
	for i := range owners {
		owners[i] = genKeyPair(t)
		ownerAddrs[i] = owners[i].Address
		acct := db.GetOrCreateAccount(owners[i].Address)
		acct.WhiteBalance = 1_000_000_000
		acct.PublicKey = owners[i].PublicKey
		db.SaveAccount(acct)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: ownerAddrs, Threshold: 3})

	regTx := types.Transaction{
		Type:      types.TxMultiSigRegister,
		From:      owners[0].Address,
		Amount:    1,
		Nonce:     1,
		Payload:   regPayload,
		PublicKey: owners[0].PublicKey,
	}
	signTx(t, &regTx, owners[0].PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}
	receipt, _ := db.GetReceipt(regTx.Hash)
	if receipt.Status != "success" {
		t.Fatalf("register should succeed, got %s (err: %s)", receipt.Status, receipt.Error)
	}

	msAddr := types.DeriveMultiSigAddress(ownerAddrs, 3)
	ms, err := db.GetMultiSig(msAddr)
	if err != nil {
		t.Fatalf("multisig account not found: %v", err)
	}
	if ms.Threshold != 3 || len(ms.Owners) != 5 {
		t.Fatalf("multisig config wrong: threshold=%d owners=%d", ms.Threshold, len(ms.Owners))
	}

	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 2_000_000_000
	db.SaveAccount(msAcct)

	recipientKP := genKeyPair(t)
	innerTx := types.Transaction{
		Type:   types.TxTransferWhite,
		From:   msAddr,
		To:     recipientKP.Address,
		Amount: 500_000_000,
	}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx})

	propTx := types.Transaction{
		Type:      types.TxMultiSigPropose,
		From:      owners[0].Address,
		Amount:    2,
		Nonce:     2,
		Payload:   propPayload,
		PublicKey: owners[0].PublicKey,
	}
	signTx(t, &propTx, owners[0].PrivateKey)

	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block2)

	propReceipt, _ := db.GetReceipt(propTx.Hash)
	if propReceipt.Status != "success" {
		t.Fatalf("propose should succeed, got %s (err: %s)", propReceipt.Status, propReceipt.Error)
	}

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	approve1Tx := types.Transaction{
		Type:      types.TxMultiSigApprove,
		From:      owners[1].Address,
		Amount:    3,
		Nonce:     1,
		Payload:   approvePayload,
		PublicKey: owners[1].PublicKey,
	}
	signTx(t, &approve1Tx, owners[1].PrivateKey)

	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approve1Tx})
	ApplyBlock(db, st, block3)

	approve1Receipt, _ := db.GetReceipt(approve1Tx.Hash)
	if approve1Receipt.Status != "success" {
		t.Fatalf("first approval should succeed, got %s (err: %s)", approve1Receipt.Status, approve1Receipt.Error)
	}

	recipientBefore, _ := db.GetAccount(recipientKP.Address)
	if recipientBefore != nil && recipientBefore.WhiteBalance != 0 {
		t.Fatalf("recipient should have 0 before 3rd approval, got %d", recipientBefore.WhiteBalance)
	}

	approve2Tx := types.Transaction{
		Type:      types.TxMultiSigApprove,
		From:      owners[2].Address,
		Amount:    4,
		Nonce:     1,
		Payload:   approvePayload,
		PublicKey: owners[2].PublicKey,
	}
	signTx(t, &approve2Tx, owners[2].PrivateKey)

	block4 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approve2Tx})
	ApplyBlock(db, st, block4)

	approve2Receipt, _ := db.GetReceipt(approve2Tx.Hash)
	if approve2Receipt.Status != "success" {
		t.Fatalf("third approval (execution) should succeed, got %s (err: %s)", approve2Receipt.Status, approve2Receipt.Error)
	}

	recipientAfter, _ := db.GetAccount(recipientKP.Address)
	if recipientAfter == nil || recipientAfter.WhiteBalance != 500_000_000 {
		bal := uint64(0)
		if recipientAfter != nil {
			bal = recipientAfter.WhiteBalance
		}
		t.Fatalf("recipient should have 500_000_000, got %d", bal)
	}

	msAfter, _ := db.GetAccount(msAddr)
	fee := types.CalcFee(500_000_000)
	expectedMsBal := uint64(2_000_000_000) - 500_000_000 - fee
	if msAfter.WhiteBalance != expectedMsBal {
		t.Fatalf("multisig balance expected %d, got %d", expectedMsBal, msAfter.WhiteBalance)
	}

	nonOwnerKP := genKeyPair(t)
	nonOwnerAcct := db.GetOrCreateAccount(nonOwnerKP.Address)
	nonOwnerAcct.WhiteBalance = 1_000_000_000
	nonOwnerAcct.PublicKey = nonOwnerKP.PublicKey
	db.SaveAccount(nonOwnerAcct)

	innerTx2 := types.Transaction{
		Type:   types.TxTransferWhite,
		From:   msAddr,
		To:     nonOwnerKP.Address,
		Amount: 100,
	}
	propPayload2, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerTx2})

	propTx2 := types.Transaction{
		Type:      types.TxMultiSigPropose,
		From:      owners[0].Address,
		Amount:    5,
		Nonce:     3,
		Payload:   propPayload2,
		PublicKey: owners[0].PublicKey,
	}
	signTx(t, &propTx2, owners[0].PrivateKey)

	block5 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx2})
	ApplyBlock(db, st, block5)

	nonOwnerApproveTx := types.Transaction{
		Type:      types.TxMultiSigApprove,
		From:      nonOwnerKP.Address,
		Amount:    6,
		Nonce:     1,
		Payload: func() []byte {
			p, _ := json.Marshal(struct {
				ProposalID string `json:"proposalId"`
			}{ProposalID: propTx2.Hash})
			return p
		}(),
		PublicKey: nonOwnerKP.PublicKey,
	}
	signTx(t, &nonOwnerApproveTx, nonOwnerKP.PrivateKey)

	block6 := buildBlock(t, db, validatorKP.Address, []types.Transaction{nonOwnerApproveTx})
	ApplyBlock(db, st, block6)

	nonOwnerReceipt, _ := db.GetReceipt(nonOwnerApproveTx.Hash)
	if nonOwnerReceipt.Status != "failed" {
		t.Fatal("non-owner approval should fail")
	}

	owner3ApproveTx := types.Transaction{
		Type:    types.TxMultiSigApprove,
		From:    owners[3].Address,
		Amount:  7,
		Nonce:   1,
		Payload: approvePayload,
		PublicKey: owners[3].PublicKey,
	}
	signTx(t, &owner3ApproveTx, owners[3].PrivateKey)

	block7 := buildBlock(t, db, validatorKP.Address, []types.Transaction{owner3ApproveTx})
	ApplyBlock(db, st, block7)

	owner3Receipt, _ := db.GetReceipt(owner3ApproveTx.Hash)
	if owner3Receipt.Status != "failed" {
		t.Fatal("approval after execution should fail")
	}
	if !strings.Contains(owner3Receipt.Error, "proposal is not pending") {
		t.Fatalf("expected 'proposal is not pending' error, got: %s", owner3Receipt.Error)
	}
}

func TestMultiSigSwapExecution(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	owners := make([]*types.KeyPair, 5)
	ownerAddrs := make([]string, 5)
	for i := range owners {
		owners[i] = genKeyPair(t)
		ownerAddrs[i] = owners[i].Address
		acct := db.GetOrCreateAccount(owners[i].Address)
		acct.WhiteBalance = 1_000_000_000
		acct.PublicKey = owners[i].PublicKey
		db.SaveAccount(acct)
	}

	regPayload, _ := json.Marshal(struct {
		Owners    []string `json:"owners"`
		Threshold uint8    `json:"threshold"`
	}{Owners: ownerAddrs, Threshold: 3})

	regTx := types.Transaction{
		Type:      types.TxMultiSigRegister,
		From:      owners[0].Address,
		Amount:    1,
		Nonce:     1,
		Payload:   regPayload,
		PublicKey: owners[0].PublicKey,
	}
	signTx(t, &regTx, owners[0].PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{regTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	msAddr := types.DeriveMultiSigAddress(ownerAddrs, 3)

	msAcct := db.GetOrCreateAccount(msAddr)
	msAcct.WhiteBalance = 5_000_000_000
	db.SaveAccount(msAcct)

	deployParams := token.DeployParams{
		Name:           "SwapCoin",
		Symbol:         "SWC",
		PoolRatio:      50,
		TeamRatio:      50,
		InitWhite:      1_000_000_000,
		ReleaseMonthly: 10000,
		MultiSigAddr:   msAddr,
	}
	deployPayload, _ := json.Marshal(deployParams)

	prevBlock, _ := db.GetBlockByHeight(db.GetLatestHeight())

	deployTx := types.Transaction{
		Type:      types.TxDeployBlue,
		From:      validatorKP.Address,
		Nonce:     1,
		Payload:   deployPayload,
		PublicKey: validatorKP.PublicKey,
		Timestamp: prevBlock.Header.Timestamp + 15,
	}
	signTx(t, &deployTx, validatorKP.PrivateKey)

	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{deployTx})
	if err := ApplyBlock(db, st, block2); err != nil {
		t.Fatal(err)
	}

	deployReceipt, _ := db.GetReceipt(deployTx.Hash)
	if deployReceipt.Status != "success" {
		t.Fatalf("deploy should succeed, got %s (err: %s)", deployReceipt.Status, deployReceipt.Error)
	}

	tokenID := token.GenerateTokenID(validatorKP.Address, "SwapCoin", 1)

	innerSwapTx := types.Transaction{
		Type:    types.TxSwapWhiteBlue,
		From:    msAddr,
		TokenID: tokenID,
		Amount:  500_000_000,
	}
	propPayload, _ := json.Marshal(struct {
		MultiSigAddr string            `json:"multiSigAddr"`
		TxPayload    types.Transaction `json:"txPayload"`
	}{MultiSigAddr: msAddr, TxPayload: innerSwapTx})

	propTx := types.Transaction{
		Type:      types.TxMultiSigPropose,
		From:      owners[0].Address,
		Amount:    2,
		Nonce:     2,
		Payload:   propPayload,
		PublicKey: owners[0].PublicKey,
	}
	signTx(t, &propTx, owners[0].PrivateKey)

	block3 := buildBlock(t, db, validatorKP.Address, []types.Transaction{propTx})
	ApplyBlock(db, st, block3)

	propReceipt, _ := db.GetReceipt(propTx.Hash)
	if propReceipt.Status != "success" {
		t.Fatalf("propose swap should succeed, got %s (err: %s)", propReceipt.Status, propReceipt.Error)
	}

	approvePayload, _ := json.Marshal(struct {
		ProposalID string `json:"proposalId"`
	}{ProposalID: propTx.Hash})

	approve1Tx := types.Transaction{
		Type:      types.TxMultiSigApprove,
		From:      owners[1].Address,
		Amount:    3,
		Nonce:     1,
		Payload:   approvePayload,
		PublicKey: owners[1].PublicKey,
	}
	signTx(t, &approve1Tx, owners[1].PrivateKey)

	block4 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approve1Tx})
	ApplyBlock(db, st, block4)

	approve2Tx := types.Transaction{
		Type:      types.TxMultiSigApprove,
		From:      owners[2].Address,
		Amount:    4,
		Nonce:     1,
		Payload:   approvePayload,
		PublicKey: owners[2].PublicKey,
	}
	signTx(t, &approve2Tx, owners[2].PrivateKey)

	block5 := buildBlock(t, db, validatorKP.Address, []types.Transaction{approve2Tx})
	ApplyBlock(db, st, block5)

	approve2Receipt, _ := db.GetReceipt(approve2Tx.Hash)
	if approve2Receipt.Status != "success" {
		t.Fatalf("swap execution should succeed, got %s (err: %s)", approve2Receipt.Status, approve2Receipt.Error)
	}

	msAfter, _ := db.GetAccount(msAddr)
	if msAfter.BlueBalances == nil || msAfter.BlueBalances[tokenID] == 0 {
		t.Fatal("multisig wallet should hold blue coins after swap")
	}
}

func TestValidatorAutoStaking(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	setupValidatorSet(t, db,
		types.ValidatorRecord{
			Address:   validatorKP.Address,
			PublicKey: validatorKP.PublicKey,
			Status:    types.ValidatorStatusActive,
		},
	)

	acct := db.GetOrCreateAccount(validatorKP.Address)
	acct.StakedBalance = 0
	acct.WhiteBalance = 0
	db.SaveAccount(acct)

	rewardPerBlock := uint64(50_000_000)
	blocksNeeded := types.StakeAmount / rewardPerBlock
	totalBlocks := blocksNeeded + 5

	for i := uint64(0); i < totalBlocks; i++ {
		height := db.GetLatestHeight()
		prev, _ := db.GetBlockByHeight(height)

		rewardTx := types.Transaction{
			Type:   types.TxBlockReward,
			To:     validatorKP.Address,
			Amount: rewardPerBlock,
			Hash:   crypto.SHA256Hex([]byte(fmt.Sprintf("reward:%d", height+1))),
		}

		block, err := CreateBlock(prev, validatorKP.Address, []types.Transaction{rewardTx}, rewardPerBlock)
		if err != nil {
			t.Fatal(err)
		}
		if err := ApplyBlock(db, st, block); err != nil {
			t.Fatal(err)
		}
	}

	finalAcct, _ := db.GetAccount(validatorKP.Address)
	if finalAcct.StakedBalance != uint64(types.StakeAmount) {
		t.Fatalf("staked balance should be %d, got %d", types.StakeAmount, finalAcct.StakedBalance)
	}

	extraRewards := uint64(5) * rewardPerBlock
	if finalAcct.WhiteBalance != extraRewards {
		t.Fatalf("white balance should be %d (overflow after stake full), got %d", extraRewards, finalAcct.WhiteBalance)
	}
}

func TestSlashedValidatorCannotRejoin(t *testing.T) {
	types.SetDevMode()
	defer resetDevMode()

	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	slashedKP := genKeyPair(t)
	slashedAcct := db.GetOrCreateAccount(slashedKP.Address)
	slashedAcct.WhiteBalance = 1_000_000_000
	slashedAcct.PublicKey = slashedKP.PublicKey
	db.SaveAccount(slashedAcct)

	setupValidatorSet(t, db,
		types.ValidatorRecord{Address: validatorKP.Address, PublicKey: validatorKP.PublicKey, Status: types.ValidatorStatusActive},
		types.ValidatorRecord{Address: slashedKP.Address, PublicKey: slashedKP.PublicKey, Status: types.ValidatorStatusActive},
	)

	h1 := types.BlockHeader{Height: 10, Validator: slashedKP.Address, Hash: crypto.SHA256Hex([]byte("slash_a"))}
	sig1, _ := crypto.Sign(slashedKP.PrivateKey, []byte(h1.Hash))
	h1.Signature = sig1

	h2 := types.BlockHeader{Height: 10, Validator: slashedKP.Address, Hash: crypto.SHA256Hex([]byte("slash_b"))}
	sig2, _ := crypto.Sign(slashedKP.PrivateKey, []byte(h2.Hash))
	h2.Signature = sig2

	evidence := struct {
		Header1 types.BlockHeader `json:"header1"`
		Header2 types.BlockHeader `json:"header2"`
	}{Header1: h1, Header2: h2}
	payload, _ := json.Marshal(evidence)

	slashTx := types.Transaction{
		Type:      types.TxSlashEvidence,
		From:      validatorKP.Address,
		Amount:    1,
		Nonce:     1,
		Payload:   payload,
		PublicKey: validatorKP.PublicKey,
		Timestamp: 1750000015,
	}
	signTx(t, &slashTx, validatorKP.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{slashTx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	slashReceipt, _ := db.GetReceipt(slashTx.Hash)
	if slashReceipt.Status != "success" {
		t.Fatalf("slash should succeed, got %s (err: %s)", slashReceipt.Status, slashReceipt.Error)
	}

	vs := db.GetValidatorSet()
	rec := vs.FindRecord(slashedKP.Address)
	if rec.Status != types.ValidatorStatusSlashed {
		t.Fatalf("should be slashed, got %s", rec.Status)
	}

	joinHeight := uint64(2)
	nonce := findPoWNonce(slashedKP.Address, joinHeight)
	joinPayload, _ := json.Marshal(struct {
		Nonce uint64 `json:"nonce"`
	}{Nonce: nonce})

	joinTx := types.Transaction{
		Type:      types.TxValidatorJoin,
		From:      slashedKP.Address,
		Amount:    joinHeight,
		Nonce:     1,
		Payload:   joinPayload,
		PublicKey: slashedKP.PublicKey,
	}
	signTx(t, &joinTx, slashedKP.PrivateKey)

	block2 := buildBlock(t, db, validatorKP.Address, []types.Transaction{joinTx})
	ApplyBlock(db, st, block2)

	joinReceipt, _ := db.GetReceipt(joinTx.Hash)
	if joinReceipt.Status != "failed" {
		t.Fatal("slashed validator should not be allowed to rejoin")
	}
	if !strings.Contains(joinReceipt.Error, "permanently banned") {
		t.Fatalf("expected 'permanently banned' error, got: %s", joinReceipt.Error)
	}
}
