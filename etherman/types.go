package etherman

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Block of L2 transactions produced by the HotShot sequencer
type SequencerBlock struct {
	Timestamp  			  uint64 `json:"timestamp"`
	Height 				  uint64 `json:"height"`
	L1Block 			  uint64 `json:"l1_block"`
	// Transactions are a blob of data encoded as hex
	Transactions 	      string `json:"transactions"`
}

// Block struct
type Block struct {
	BlockNumber           uint64
	BlockHash             common.Hash
	ParentHash            common.Hash
	GlobalExitRoots       []GlobalExitRoot
	ForcedBatches         []ForcedBatch
	SequencedBatches      [][]SequencedBatch
	VerifiedBatches       []VerifiedBatch
	ReceivedAt            time.Time
}

// GlobalExitRoot struct
type GlobalExitRoot struct {
	BlockNumber     uint64
	MainnetExitRoot common.Hash
	RollupExitRoot  common.Hash
	GlobalExitRoot  common.Hash
}

// SequencedBatch represents virtual batch
type SequencedBatch struct {
	BatchNumber   uint64
	BlockNumber   uint64
	SequencerAddr common.Address
	TxHash        common.Hash
	Nonce         uint64
	Coinbase      common.Address
	PolygonZkEVMBatchData
}

// ForcedBatch represents a ForcedBatch
type ForcedBatch struct {
	BlockNumber       uint64
	ForcedBatchNumber uint64
	Sequencer         common.Address
	GlobalExitRoot    common.Hash
	RawTxsData        []byte
	ForcedAt          time.Time
}

// VerifiedBatch represents a VerifiedBatch
type VerifiedBatch struct {
	BlockNumber uint64
	BatchNumber uint64
	Aggregator  common.Address
	StateRoot   common.Hash
	TxHash      common.Hash
}

// Copied from binding for a previous iteration of the contract
type PolygonZkEVMBatchData struct {
	Transactions       []byte
	GlobalExitRoot     [32]byte
	Timestamp          uint64
}