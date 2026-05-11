package token

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/white-blue-protocol/wblue/internal/storage"
)

func setupDB(t *testing.T) (*storage.DB, func()) {
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

func TestDeployBasic(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "TestCoin", Symbol: "TST",
		PoolRatio: 60, TeamRatio: 40,
		InitWhite: 1_000_000, ReleaseMonthly: 5000,
		MultiSigAddr: "0xms",
	}

	cfg, err := Deploy(db, "0xcreator", params, 1, 1000)
	if err != nil {
		t.Fatalf("deploy failed: %v", err)
	}
	if cfg.TokenID == "" {
		t.Fatal("token ID should not be empty")
	}
	if cfg.Name != "TestCoin" {
		t.Fatal("name mismatch")
	}

	pool, err := db.GetPool(cfg.TokenID)
	if err != nil {
		t.Fatalf("pool not found: %v", err)
	}
	if pool.WhiteReserve != 1_000_000 {
		t.Fatalf("white reserve wrong: %d", pool.WhiteReserve)
	}
}

func TestDeployBadRatio(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Bad", Symbol: "BAD",
		PoolRatio: 60, TeamRatio: 60,
		InitWhite: 1000, ReleaseMonthly: 100,
		MultiSigAddr: "0xms",
	}

	_, err := Deploy(db, "0xcreator", params, 1, 1000)
	if err == nil {
		t.Fatal("ratio != 100 should fail")
	}
}

func TestDeployEmptyName(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "", Symbol: "X",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1000, ReleaseMonthly: 100,
		MultiSigAddr: "0xms",
	}

	_, err := Deploy(db, "0xcreator", params, 1, 1000)
	if err == nil {
		t.Fatal("empty name should fail")
	}
}

func TestDeployZeroInitWhite(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Zero", Symbol: "ZRO",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 0, ReleaseMonthly: 100,
		MultiSigAddr: "0xms",
	}

	_, err := Deploy(db, "0xcreator", params, 1, 1000)
	if err == nil {
		t.Fatal("zero initWhite should fail")
	}
}

func TestDeployDuplicate(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Dup", Symbol: "DUP",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1000, ReleaseMonthly: 100,
		MultiSigAddr: "0xms",
	}

	_, err := Deploy(db, "0xcreator", params, 1, 1000)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Deploy(db, "0xcreator", params, 1, 2000)
	if err == nil {
		t.Fatal("duplicate deploy should fail")
	}
}

func TestGenerateTokenIDDeterministic(t *testing.T) {
	id1 := GenerateTokenID("0xcreator", "Foo", 1)
	id2 := GenerateTokenID("0xcreator", "Foo", 1)
	if id1 != id2 {
		t.Fatal("same input should produce same token ID")
	}
	if id1 == GenerateTokenID("0xcreator", "Bar", 1) {
		t.Fatal("different name should produce different ID")
	}
}

func TestProcessVestingNoTokens(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	err := ProcessVesting(db, 1000)
	if err != nil {
		t.Fatalf("vesting with no tokens should not fail: %v", err)
	}
}

func TestProcessVestingOneMonth(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Vest", Symbol: "VST",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1_000_000, ReleaseMonthly: 5000,
		MultiSigAddr: "0xteam",
	}
	cfg, _ := Deploy(db, "0xcreator", params, 1, 1000)

	stateBefore, _ := db.GetBlueCoinState(cfg.TokenID)
	lockedBefore := stateBefore.TeamLocked

	err := ProcessVesting(db, 1000+SecondsPerMonth+1)
	if err != nil {
		t.Fatalf("vesting failed: %v", err)
	}

	stateAfter, _ := db.GetBlueCoinState(cfg.TokenID)
	if stateAfter.TeamLocked != lockedBefore-5000 {
		t.Fatalf("locked should decrease by 5000, got %d (was %d)", stateAfter.TeamLocked, lockedBefore)
	}
	if stateAfter.TeamReleased != 5000 {
		t.Fatalf("released should be 5000, got %d", stateAfter.TeamReleased)
	}

	teamAcct, _ := db.GetAccount("0xteam")
	if teamAcct.BlueBalances[cfg.TokenID] != 5000 {
		t.Fatalf("team should have 5000 blue tokens, got %d", teamAcct.BlueBalances[cfg.TokenID])
	}
}

func TestProcessVestingMultipleMonths(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Multi", Symbol: "MLT",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1_000_000, ReleaseMonthly: 1000,
		MultiSigAddr: "0xteam2",
	}
	cfg, _ := Deploy(db, "0xcreator", params, 1, 1000)

	err := ProcessVesting(db, 1000+SecondsPerMonth*3+1)
	if err != nil {
		t.Fatalf("vesting failed: %v", err)
	}

	state, _ := db.GetBlueCoinState(cfg.TokenID)
	if state.TeamReleased != 3000 {
		t.Fatalf("should release 3 months worth (3000), got %d", state.TeamReleased)
	}
}

func TestProcessVestingNotEnoughTime(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Early", Symbol: "ERL",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1000, ReleaseMonthly: 100,
		MultiSigAddr: "0xteam3",
	}
	cfg, _ := Deploy(db, "0xcreator", params, 1, 1000)

	ProcessVesting(db, 1000+SecondsPerMonth/2)

	state, _ := db.GetBlueCoinState(cfg.TokenID)
	if state.TeamReleased != 0 {
		t.Fatalf("should not release before 1 month, got %d", state.TeamReleased)
	}
}

func TestProcessVestingCapAtLocked(t *testing.T) {
	db, cleanup := setupDB(t)
	defer cleanup()

	params := &DeployParams{
		Name: "Cap", Symbol: "CAP",
		PoolRatio: 50, TeamRatio: 50,
		InitWhite: 1000, ReleaseMonthly: 999_999_999_999,
		MultiSigAddr: "0xteam4",
	}
	cfg, _ := Deploy(db, "0xcreator", params, 1, 1000)

	stateBefore, _ := db.GetBlueCoinState(cfg.TokenID)
	totalLocked := stateBefore.TeamLocked

	ProcessVesting(db, 1000+SecondsPerMonth+1)

	stateAfter, _ := db.GetBlueCoinState(cfg.TokenID)
	if stateAfter.TeamLocked != 0 {
		t.Fatalf("should drain all locked, got %d remaining", stateAfter.TeamLocked)
	}
	if stateAfter.TeamReleased != totalLocked {
		t.Fatalf("released should equal original locked %d, got %d", totalLocked, stateAfter.TeamReleased)
	}
}
