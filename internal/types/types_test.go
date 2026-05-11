package types

import (
	"testing"
)

func TestActiveValidatorsEmpty(t *testing.T) {
	vs := &ValidatorSet{}
	active := vs.ActiveValidators()
	if len(active) != 0 {
		t.Fatal("empty set should return empty active")
	}
}

func TestActiveValidatorsFiltering(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xa", Status: ValidatorStatusActive, JoinHeight: 1},
			{Address: "0xb", Status: ValidatorStatusSuspended, JoinHeight: 2},
			{Address: "0xc", Status: ValidatorStatusActive, JoinHeight: 3},
			{Address: "0xd", Status: ValidatorStatusRemoved, JoinHeight: 4},
			{Address: "0xe", Status: ValidatorStatusSlashed, JoinHeight: 5},
		},
	}
	active := vs.ActiveValidators()
	if len(active) != 2 {
		t.Fatalf("should have 2 active, got %d", len(active))
	}
	if active[0].Address != "0xa" || active[1].Address != "0xc" {
		t.Fatal("wrong active validators returned")
	}
}

func TestActiveValidatorsSortedByJoinHeight(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xc", Status: ValidatorStatusActive, JoinHeight: 30},
			{Address: "0xa", Status: ValidatorStatusActive, JoinHeight: 10},
			{Address: "0xb", Status: ValidatorStatusActive, JoinHeight: 20},
		},
	}
	active := vs.ActiveValidators()
	if active[0].Address != "0xa" || active[1].Address != "0xb" || active[2].Address != "0xc" {
		t.Fatal("should be sorted by JoinHeight ascending")
	}
}

func TestActiveValidatorsSortTiebreaker(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xbb", Status: ValidatorStatusActive, JoinHeight: 10},
			{Address: "0xaa", Status: ValidatorStatusActive, JoinHeight: 10},
		},
	}
	active := vs.ActiveValidators()
	if active[0].Address != "0xaa" {
		t.Fatal("same JoinHeight should sort by address")
	}
}

func TestActiveValidatorsAtHeight(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xa", Status: ValidatorStatusActive, JoinHeight: 5},
			{Address: "0xb", Status: ValidatorStatusActive, JoinHeight: 15},
			{Address: "0xc", Status: ValidatorStatusActive, JoinHeight: 25},
		},
	}
	at10 := vs.ActiveValidatorsAt(10)
	if len(at10) != 1 {
		t.Fatalf("at height 10 should have 1, got %d", len(at10))
	}
	at20 := vs.ActiveValidatorsAt(20)
	if len(at20) != 2 {
		t.Fatalf("at height 20 should have 2, got %d", len(at20))
	}
	at30 := vs.ActiveValidatorsAt(30)
	if len(at30) != 3 {
		t.Fatalf("at height 30 should have 3, got %d", len(at30))
	}
}

func TestFindRecord(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xa"},
			{Address: "0xb"},
		},
	}
	r := vs.FindRecord("0xb")
	if r == nil || r.Address != "0xb" {
		t.Fatal("should find 0xb")
	}
	if vs.FindRecord("0xc") != nil {
		t.Fatal("should return nil for unknown address")
	}
}

func TestAddOrUpdateNew(t *testing.T) {
	vs := &ValidatorSet{}
	vs.AddOrUpdate(ValidatorRecord{Address: "0xa", Status: ValidatorStatusActive})
	if len(vs.Validators) != 1 {
		t.Fatal("should have 1 validator")
	}
}

func TestAddOrUpdateExisting(t *testing.T) {
	vs := &ValidatorSet{
		Validators: []ValidatorRecord{
			{Address: "0xa", Status: ValidatorStatusActive},
		},
	}
	vs.AddOrUpdate(ValidatorRecord{Address: "0xa", Status: ValidatorStatusSuspended})
	if len(vs.Validators) != 1 {
		t.Fatal("should still have 1 validator")
	}
	if vs.Validators[0].Status != ValidatorStatusSuspended {
		t.Fatal("status should be updated")
	}
}

func TestDeriveMultiSigAddressDeterministic(t *testing.T) {
	owners := []string{"0xalice", "0xbob"}
	a1 := DeriveMultiSigAddress(owners, 2)
	a2 := DeriveMultiSigAddress(owners, 2)
	if a1 != a2 {
		t.Fatal("same input should produce same address")
	}
}

func TestDeriveMultiSigAddressOrderIndependent(t *testing.T) {
	a1 := DeriveMultiSigAddress([]string{"0xalice", "0xbob"}, 2)
	a2 := DeriveMultiSigAddress([]string{"0xbob", "0xalice"}, 2)
	if a1 != a2 {
		t.Fatal("owner order should not matter (sorted internally)")
	}
}

func TestDeriveMultiSigAddressDifferentThreshold(t *testing.T) {
	a1 := DeriveMultiSigAddress([]string{"0xalice", "0xbob"}, 1)
	a2 := DeriveMultiSigAddress([]string{"0xalice", "0xbob"}, 2)
	if a1 == a2 {
		t.Fatal("different threshold should produce different address")
	}
}

func TestDeriveMultiSigAddressFormat(t *testing.T) {
	addr := DeriveMultiSigAddress([]string{"0xa", "0xb"}, 2)
	if len(addr) != 42 {
		t.Fatalf("address should be 42 chars, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Fatal("should start with 0x")
	}
}
