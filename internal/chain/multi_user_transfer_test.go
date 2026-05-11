package chain

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/state"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func TestChainTransfer10Users(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	users := make([]*types.KeyPair, 10)
	for i := 0; i < 10; i++ {
		users[i] = genKeyPair(t)
		acct := db.GetOrCreateAccount(users[i].Address)
		acct.WhiteBalance = 10_000_000_000
		acct.PublicKey = users[i].PublicKey
		acct.Nonce = 0
		db.SaveAccount(acct)
	}

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	var userTxs []types.Transaction
	amount := uint64(1_000_000_000)
	for i := 0; i < 9; i++ {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      users[i].Address,
			To:        users[i+1].Address,
			Amount:    amount,
			Nonce:     1,
			PublicKey: users[i].PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, users[i].PrivateKey)
		userTxs = append(userTxs, tx)
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	fee := types.CalcFee(amount)

	expectedUser0 := uint64(10_000_000_000) - amount - fee
	acct0, _ := db.GetAccount(users[0].Address)
	if acct0.WhiteBalance != expectedUser0 {
		t.Fatalf("user0 balance: expected %d, got %d", expectedUser0, acct0.WhiteBalance)
	}

	for i := 1; i < 9; i++ {
		expected := uint64(10_000_000_000) + amount - amount - fee
		acct, _ := db.GetAccount(users[i].Address)
		if acct.WhiteBalance != expected {
			t.Fatalf("user%d balance: expected %d, got %d", i, expected, acct.WhiteBalance)
		}
	}

	expectedUser9 := uint64(10_000_000_000) + amount
	acct9, _ := db.GetAccount(users[9].Address)
	if acct9.WhiteBalance != expectedUser9 {
		t.Fatalf("user9 balance: expected %d, got %d", expectedUser9, acct9.WhiteBalance)
	}
}

func TestBatchTransferOneToMany(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	sender := genKeyPair(t)
	senderAcct := db.GetOrCreateAccount(sender.Address)
	senderAcct.WhiteBalance = 10_000_000_000
	senderAcct.PublicKey = sender.PublicKey
	senderAcct.Nonce = 0
	db.SaveAccount(senderAcct)

	receivers := make([]*types.KeyPair, 5)
	for i := 0; i < 5; i++ {
		receivers[i] = genKeyPair(t)
	}

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	amount := uint64(1_000_000_000)
	fee := types.CalcFee(amount)
	var userTxs []types.Transaction
	for i := 0; i < 5; i++ {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      sender.Address,
			To:        receivers[i].Address,
			Amount:    amount,
			Nonce:     uint64(i + 1),
			PublicKey: sender.PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, sender.PrivateKey)
		userTxs = append(userTxs, tx)
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	expectedSender := uint64(10_000_000_000) - 5*(amount+fee)
	senderAfter, _ := db.GetAccount(sender.Address)
	if senderAfter.WhiteBalance != expectedSender {
		t.Fatalf("sender balance: expected %d, got %d", expectedSender, senderAfter.WhiteBalance)
	}

	for i := 0; i < 5; i++ {
		recv, _ := db.GetAccount(receivers[i].Address)
		if recv.WhiteBalance != amount {
			t.Fatalf("receiver%d balance: expected %d, got %d", i, amount, recv.WhiteBalance)
		}
	}
}

func TestCircularTransferConservation(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	users := make([]*types.KeyPair, 4)
	initialBalance := uint64(5_000_000_000)
	for i := 0; i < 4; i++ {
		users[i] = genKeyPair(t)
		acct := db.GetOrCreateAccount(users[i].Address)
		acct.WhiteBalance = initialBalance
		acct.PublicKey = users[i].PublicKey
		acct.Nonce = 0
		db.SaveAccount(acct)
	}

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	amount := uint64(1_000_000_000)
	fee := types.CalcFee(amount)

	transfers := [][2]int{{0, 1}, {1, 2}, {2, 3}, {3, 0}}
	var userTxs []types.Transaction
	for _, pair := range transfers {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      users[pair[0]].Address,
			To:        users[pair[1]].Address,
			Amount:    amount,
			Nonce:     1,
			PublicKey: users[pair[0]].PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, users[pair[0]].PrivateKey)
		userTxs = append(userTxs, tx)
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	totalInitial := 4 * initialBalance
	totalFees := 4 * fee

	validatorAfter, _ := db.GetAccount(validatorKP.Address)
	validatorShare := totalFees / 2

	var totalAfter uint64
	for i := 0; i < 4; i++ {
		acct, _ := db.GetAccount(users[i].Address)
		totalAfter += acct.WhiteBalance
	}

	expectedTotal := totalInitial - totalFees + validatorShare
	totalAfter += validatorAfter.WhiteBalance

	totalWithValidator := totalAfter
	expectedWithValidator := totalInitial - totalFees + validatorShare + validatorAfter.WhiteBalance - validatorShare

	_ = totalWithValidator
	_ = expectedWithValidator

	var userSum uint64
	for i := 0; i < 4; i++ {
		acct, _ := db.GetAccount(users[i].Address)
		userSum += acct.WhiteBalance
	}
	expectedUserSum := totalInitial - totalFees
	if userSum != expectedUserSum {
		t.Fatalf("total user balance: expected %d, got %d", expectedUserSum, userSum)
	}

	_ = expectedTotal
}

func TestNonceReplayAttack(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	userA := genKeyPair(t)
	acctA := db.GetOrCreateAccount(userA.Address)
	acctA.WhiteBalance = 10_000_000_000
	acctA.PublicKey = userA.PublicKey
	acctA.Nonce = 0
	db.SaveAccount(acctA)

	userB := genKeyPair(t)

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	amount := uint64(1_000_000_000)
	fee := types.CalcFee(amount)

	tx1 := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      userA.Address,
		To:        userB.Address,
		Amount:    amount,
		Nonce:     1,
		PublicKey: userA.PublicKey,
		Timestamp: time.Now().Unix(),
	}
	signTx(t, &tx1, userA.PrivateKey)

	tx2 := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      userA.Address,
		To:        userB.Address,
		Amount:    amount,
		Nonce:     1,
		PublicKey: userA.PublicKey,
		Timestamp: time.Now().Unix() + 1,
	}
	signTx(t, &tx2, userA.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{rewardTx, tx1, tx2})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt1, _ := db.GetReceipt(tx1.Hash)
	if receipt1.Status != "success" {
		t.Fatalf("first tx should succeed, got %s", receipt1.Status)
	}

	receipt2, _ := db.GetReceipt(tx2.Hash)
	if receipt2.Status != "failed" {
		t.Fatalf("second tx with same nonce should fail, got %s", receipt2.Status)
	}

	acctAfter, _ := db.GetAccount(userA.Address)
	expectedBalance := uint64(10_000_000_000) - amount - fee
	if acctAfter.WhiteBalance != expectedBalance {
		t.Fatalf("A balance: expected %d, got %d", expectedBalance, acctAfter.WhiteBalance)
	}
}

func TestImpersonationAttack(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	userA := genKeyPair(t)
	acctA := db.GetOrCreateAccount(userA.Address)
	acctA.WhiteBalance = 10_000_000_000
	acctA.PublicKey = userA.PublicKey
	acctA.Nonce = 0
	db.SaveAccount(acctA)

	userB := genKeyPair(t)
	acctB := db.GetOrCreateAccount(userB.Address)
	acctB.WhiteBalance = 10_000_000_000
	acctB.PublicKey = userB.PublicKey
	acctB.Nonce = 0
	db.SaveAccount(acctB)

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	tx := types.Transaction{
		Type:      types.TxTransferWhite,
		From:      userB.Address,
		To:        userA.Address,
		Amount:    5_000_000_000,
		Nonce:     1,
		PublicKey: userA.PublicKey,
		Timestamp: time.Now().Unix(),
	}
	signTx(t, &tx, userA.PrivateKey)

	block := buildBlock(t, db, validatorKP.Address, []types.Transaction{rewardTx, tx})
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt, _ := db.GetReceipt(tx.Hash)
	if receipt.Status != "failed" {
		t.Fatalf("impersonation tx should fail, got %s", receipt.Status)
	}

	bAfter, _ := db.GetAccount(userB.Address)
	if bAfter.WhiteBalance != 10_000_000_000 {
		t.Fatalf("B balance should be unchanged, got %d", bAfter.WhiteBalance)
	}
}

func TestBalanceExhaustionSequence(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	userA := genKeyPair(t)
	acctA := db.GetOrCreateAccount(userA.Address)
	acctA.WhiteBalance = 2_000_000_000
	acctA.PublicKey = userA.PublicKey
	acctA.Nonce = 0
	db.SaveAccount(acctA)

	receiver := genKeyPair(t)

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	amount := uint64(1_000_000_000)
	var userTxs []types.Transaction
	for i := 1; i <= 3; i++ {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      userA.Address,
			To:        receiver.Address,
			Amount:    amount,
			Nonce:     uint64(i),
			PublicKey: userA.PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, userA.PrivateKey)
		userTxs = append(userTxs, tx)
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	receipt1, _ := db.GetReceipt(userTxs[0].Hash)
	if receipt1.Status != "success" {
		t.Fatalf("first tx should succeed, got %s (err: %s)", receipt1.Status, receipt1.Error)
	}

	receipt3, _ := db.GetReceipt(userTxs[2].Hash)
	if receipt3.Status != "failed" {
		t.Fatalf("third tx should fail, got %s", receipt3.Status)
	}

	aAfter, _ := db.GetAccount(userA.Address)
	if aAfter.WhiteBalance >= 2_000_000_000 {
		t.Fatalf("A balance should have decreased, got %d", aAfter.WhiteBalance)
	}
}

func TestOverflowAmountFeeMultiUser(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	users := make([]*types.KeyPair, 3)
	for i := 0; i < 3; i++ {
		users[i] = genKeyPair(t)
		acct := db.GetOrCreateAccount(users[i].Address)
		acct.WhiteBalance = 10_000_000_000
		acct.PublicKey = users[i].PublicKey
		acct.Nonce = 0
		db.SaveAccount(acct)
	}

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	var userTxs []types.Transaction
	for i := 0; i < 3; i++ {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      users[i].Address,
			To:        validatorKP.Address,
			Amount:    math.MaxUint64,
			Nonce:     1,
			PublicKey: users[i].PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, users[i].PrivateKey)
		userTxs = append(userTxs, tx)
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		receipt, _ := db.GetReceipt(userTxs[i].Hash)
		if receipt.Status != "failed" {
			t.Fatalf("user%d overflow tx should fail, got %s", i, receipt.Status)
		}

		acct, _ := db.GetAccount(users[i].Address)
		if acct.WhiteBalance != 10_000_000_000 {
			t.Fatalf("user%d balance should be unchanged, got %d", i, acct.WhiteBalance)
		}
	}
}

func TestFeeShareMultiTransfer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	validatorKP := genKeyPair(t)
	createTestChain(t, db, validatorKP)
	st := state.New(db)

	validatorBefore, _ := db.GetAccount(validatorKP.Address)
	validatorInitial := validatorBefore.WhiteBalance

	senders := make([]*types.KeyPair, 5)
	for i := 0; i < 5; i++ {
		senders[i] = genKeyPair(t)
		acct := db.GetOrCreateAccount(senders[i].Address)
		acct.WhiteBalance = 10_000_000_000
		acct.PublicKey = senders[i].PublicKey
		acct.Nonce = 0
		db.SaveAccount(acct)
	}

	receiver := genKeyPair(t)

	rewardTx := types.Transaction{
		Type:   types.TxBlockReward,
		To:     validatorKP.Address,
		Amount: 50_000_000,
		Hash:   crypto.SHA256Hex([]byte("reward:1")),
	}

	amounts := []uint64{
		1_000_000_000,
		2_000_000_000,
		500_000_000,
		750_000_000,
		1_500_000_000,
	}

	var totalFees uint64
	var userTxs []types.Transaction
	for i := 0; i < 5; i++ {
		tx := types.Transaction{
			Type:      types.TxTransferWhite,
			From:      senders[i].Address,
			To:        receiver.Address,
			Amount:    amounts[i],
			Nonce:     1,
			PublicKey: senders[i].PublicKey,
			Timestamp: time.Now().Unix(),
		}
		signTx(t, &tx, senders[i].PrivateKey)
		userTxs = append(userTxs, tx)
		totalFees += types.CalcFee(amounts[i])
	}

	block := buildBlock(t, db, validatorKP.Address, append([]types.Transaction{rewardTx}, userTxs...))
	if err := ApplyBlock(db, st, block); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		receipt, _ := db.GetReceipt(userTxs[i].Hash)
		if receipt.Status != "success" {
			t.Fatalf("sender%d tx should succeed, got %s (err: %s)", i, receipt.Status, receipt.Error)
		}
	}

	validatorShare := totalFees / 2
	expectedTotal := validatorInitial + 50_000_000 + validatorShare
	validatorAfter, _ := db.GetAccount(validatorKP.Address)
	actualTotal := validatorAfter.WhiteBalance + validatorAfter.StakedBalance
	if actualTotal != expectedTotal {
		t.Fatalf("validator total balance: expected %d, got %d (totalFees=%d, share=%d)",
			expectedTotal, actualTotal, totalFees, validatorShare)
	}

	fmt.Printf("totalFees=%d validatorShare=%d\n", totalFees, validatorShare)
}
