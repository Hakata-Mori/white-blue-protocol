package types

import "sort"

const (
	ValidatorStatusCandidate  = "candidate"
	ValidatorStatusActive     = "active"
	ValidatorStatusSuspended  = "suspended"
	ValidatorStatusRemoved    = "removed"
	ValidatorStatusSlashed    = "slashed"

	StakeAmount        = 1_000_000_000
	UptimeBlocks       = 5760
	SuspendBlocks      = 5760
	EvictBlocks        = 17280
	ConfirmationBlocks = 10
	PoWDifficulty      = 24
	SlashReward        = 10_000_000
)

type ValidatorRecord struct {
	Address              string `json:"address"`
	PublicKey            string `json:"publicKey"`
	JoinHeight           uint64 `json:"joinHeight"`
	FirstHeartbeatHeight uint64 `json:"firstHeartbeatHeight"`
	LastHeartbeatHeight  uint64 `json:"lastHeartbeatHeight"`
	Status               string `json:"status"`
	SuspendedAt          uint64 `json:"suspendedAt,omitempty"`
}

type ValidatorSet struct {
	Validators []ValidatorRecord `json:"validators"`
	UpdatedAt  uint64            `json:"updatedAt"`
}

func (vs *ValidatorSet) ActiveValidators() []ValidatorRecord {
	var active []ValidatorRecord
	for _, v := range vs.Validators {
		if v.Status == ValidatorStatusActive {
			active = append(active, v)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		if active[i].JoinHeight != active[j].JoinHeight {
			return active[i].JoinHeight < active[j].JoinHeight
		}
		return active[i].Address < active[j].Address
	})
	return active
}

func (vs *ValidatorSet) ActiveValidatorsAt(height uint64) []ValidatorRecord {
	var active []ValidatorRecord
	for _, v := range vs.Validators {
		if v.Status == ValidatorStatusActive && v.JoinHeight <= height {
			active = append(active, v)
		}
	}
	sort.Slice(active, func(i, j int) bool {
		if active[i].JoinHeight != active[j].JoinHeight {
			return active[i].JoinHeight < active[j].JoinHeight
		}
		return active[i].Address < active[j].Address
	})
	return active
}

func (vs *ValidatorSet) FindRecord(address string) *ValidatorRecord {
	for i := range vs.Validators {
		if vs.Validators[i].Address == address {
			return &vs.Validators[i]
		}
	}
	return nil
}

func (vs *ValidatorSet) AddOrUpdate(rec ValidatorRecord) {
	for i := range vs.Validators {
		if vs.Validators[i].Address == rec.Address {
			vs.Validators[i] = rec
			return
		}
	}
	vs.Validators = append(vs.Validators, rec)
}
