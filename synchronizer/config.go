package synchronizer

import (
	"github.com/0xPolygonHermez/zkevm-node/config/types"
)

// Config represents the configuration of the synchronizer
type Config struct {
	// SyncInterval is the delay interval between reading new rollup information
	SyncInterval types.Duration `mapstructure:"SyncInterval"`
	// PreconfirmationsSyncInterval is the delay interval between reading new preconfirmations from the sequencer
	PreconfirmationsSyncInterval types.Duration `mapstructure:"PreconfirmationsSyncInterval"`

	// SyncChunkSize is the number of blocks to sync on each chunk
	SyncChunkSize uint64 `mapstructure:"SyncChunkSize"`

	GenBlockNumber uint64 `mapstructure:"GenBlockNumber"`

	IgnoreGenBlockNumberCheck bool `mapstructure:"IgnoreGenBlockNumberCheck"`
}
