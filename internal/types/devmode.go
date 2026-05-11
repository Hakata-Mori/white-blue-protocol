package types

var (
	DevBlockInterval  = BlockInterval
	DevUptimeBlocks   = UptimeBlocks
	DevSuspendBlocks  = SuspendBlocks
	DevEvictBlocks    = EvictBlocks
	DevPoWDifficulty  = PoWDifficulty
)

func SetDevMode() {
	DevBlockInterval = 5
	DevUptimeBlocks = 2
	DevSuspendBlocks = 6
	DevEvictBlocks = 30
	DevPoWDifficulty = 8
}

func GetBlockInterval() int {
	return DevBlockInterval
}

func GetUptimeBlocks() uint64 {
	return uint64(DevUptimeBlocks)
}

func GetSuspendBlocks() uint64 {
	return uint64(DevSuspendBlocks)
}

func GetEvictBlocks() uint64 {
	return uint64(DevEvictBlocks)
}

func GetPoWDifficulty() int {
	return DevPoWDifficulty
}
