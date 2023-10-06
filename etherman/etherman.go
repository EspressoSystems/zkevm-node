package etherman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/0xPolygonHermez/zkevm-node/etherman/etherscan"
	"github.com/0xPolygonHermez/zkevm-node/etherman/ethgasstation"
	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/ihotshot"
	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/matic"
	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygonzkevm"
	"github.com/0xPolygonHermez/zkevm-node/etherman/smartcontracts/polygonzkevmglobalexitroot"
	ethmanTypes "github.com/0xPolygonHermez/zkevm-node/etherman/types"
	"github.com/0xPolygonHermez/zkevm-node/hex"
	"github.com/0xPolygonHermez/zkevm-node/log"
	"github.com/0xPolygonHermez/zkevm-node/state"
	"github.com/0xPolygonHermez/zkevm-node/test/operations"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

var (
	updateGlobalExitRootSignatureHash           = crypto.Keccak256Hash([]byte("UpdateGlobalExitRoot(bytes32,bytes32)"))
	sequencedBatchesEventSignatureHash          = crypto.Keccak256Hash([]byte("SequenceBatches(uint64)"))
	verifyBatchesSignatureHash                  = crypto.Keccak256Hash([]byte("VerifyBatches(uint64,bytes32,address)"))
	verifyBatchesTrustedAggregatorSignatureHash = crypto.Keccak256Hash([]byte("VerifyBatchesTrustedAggregator(uint64,bytes32,address)"))
	setTrustedSequencerURLSignatureHash         = crypto.Keccak256Hash([]byte("SetTrustedSequencerURL(string)"))
	setTrustedSequencerSignatureHash            = crypto.Keccak256Hash([]byte("SetTrustedSequencer(address)"))
	transferOwnershipSignatureHash              = crypto.Keccak256Hash([]byte("OwnershipTransferred(address,address)"))
	setSecurityCouncilSignatureHash             = crypto.Keccak256Hash([]byte("SetSecurityCouncil(address)"))
	proofDifferentStateSignatureHash            = crypto.Keccak256Hash([]byte("ProofDifferentState(bytes32,bytes32)"))
	emergencyStateActivatedSignatureHash        = crypto.Keccak256Hash([]byte("EmergencyStateActivated()"))
	emergencyStateDeactivatedSignatureHash      = crypto.Keccak256Hash([]byte("EmergencyStateDeactivated()"))
	updateZkEVMVersionSignatureHash             = crypto.Keccak256Hash([]byte("UpdateZkEVMVersion(uint64,uint64,string)"))
	newBlocksSignatureHash                      = crypto.Keccak256Hash([]byte("NewBlocks(uint256,uint256)"))

	// Proxy events
	initializedSignatureHash    = crypto.Keccak256Hash([]byte("Initialized(uint8)"))
	adminChangedSignatureHash   = crypto.Keccak256Hash([]byte("AdminChanged(address,address)"))
	beaconUpgradedSignatureHash = crypto.Keccak256Hash([]byte("BeaconUpgraded(address)"))
	upgradedSignatureHash       = crypto.Keccak256Hash([]byte("Upgraded(address)"))

	// ErrNotFound is used when the object is not found
	ErrNotFound = errors.New("not found")
	// ErrIsReadOnlyMode is used when the EtherMan client is in read-only mode.
	ErrIsReadOnlyMode = errors.New("etherman client in read-only mode: no account configured to send transactions to L1. " +
		"please check the [Etherman] PrivateKeyPath and PrivateKeyPassword configuration")
	// ErrPrivateKeyNotFound used when the provided sender does not have a private key registered to be used
	ErrPrivateKeyNotFound = errors.New("can't find sender private key to sign tx")
	// ErrSkipBatch indicates that we have seen an old batch and should skip re-processing it
	ErrSkipBatch = errors.New("skip old batch")
)

// SequencedBatchesSigHash returns the hash for the `SequenceBatches` event.
func SequencedBatchesSigHash() common.Hash { return sequencedBatchesEventSignatureHash }

// TrustedVerifyBatchesSigHash returns the hash for the `TrustedVerifyBatches` event.
func TrustedVerifyBatchesSigHash() common.Hash { return verifyBatchesTrustedAggregatorSignatureHash }

// EventOrder is the the type used to identify the events order
type EventOrder string

const (
	// GlobalExitRootsOrder identifies a GlobalExitRoot event
	GlobalExitRootsOrder EventOrder = "GlobalExitRoots"
	// SequenceBatchesOrder identifies a VerifyBatch event
	SequenceBatchesOrder EventOrder = "SequenceBatches"
	// TrustedVerifyBatchOrder identifies a TrustedVerifyBatch event
	TrustedVerifyBatchOrder EventOrder = "TrustedVerifyBatch"
)

type ethereumClient interface {
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.TransactionSender

	bind.DeployBackend
}

type externalGasProviders struct {
	MultiGasProvider bool
	Providers        []ethereum.GasPricer
}

// Client is a simple implementation of EtherMan.
type Client struct {
	EthClient             ethereumClient
	PoE                   *polygonzkevm.Polygonzkevm
	GlobalExitRootManager *polygonzkevmglobalexitroot.Polygonzkevmglobalexitroot
	Matic                 *matic.Matic
	HotShot               *ihotshot.Ihotshot
	SCAddresses           []common.Address

	GasProviders externalGasProviders

	cfg  Config
	auth map[common.Address]bind.TransactOpts // empty in case of read-only client
}

// NewClient creates a new etherman.
func NewClient(cfg Config) (*Client, error) {
	// Connect to ethereum node
	ethClient, err := ethclient.Dial(cfg.URL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.URL, err)
		return nil, err
	}
	// Create smc clients
	poe, err := polygonzkevm.NewPolygonzkevm(cfg.PoEAddr, ethClient)
	if err != nil {
		return nil, err
	}
	globalExitRoot, err := polygonzkevmglobalexitroot.NewPolygonzkevmglobalexitroot(cfg.GlobalExitRootManagerAddr, ethClient)
	if err != nil {
		return nil, err
	}
	matic, err := matic.NewMatic(cfg.MaticAddr, ethClient)
	if err != nil {
		return nil, err
	}
	hotshot, err := ihotshot.NewIhotshot(cfg.HotShotAddr, ethClient)
	if err != nil {
		return nil, err
	}

	var scAddresses []common.Address
	scAddresses = append(scAddresses, cfg.PoEAddr, cfg.GlobalExitRootManagerAddr, cfg.HotShotAddr)

	gProviders := []ethereum.GasPricer{ethClient}
	if cfg.MultiGasProvider {
		if cfg.Etherscan.ApiKey == "" {
			log.Info("No ApiKey provided for etherscan. Ignoring provider...")
		} else {
			log.Info("ApiKey detected for etherscan")
			gProviders = append(gProviders, etherscan.NewEtherscanService(cfg.Etherscan.ApiKey))
		}
		gProviders = append(gProviders, ethgasstation.NewEthGasStationService())
	}

	log.Infof("Hotshot address %s", cfg.HotShotAddr.String())
	log.Infof("Genesis hotshot block number %d", cfg.GenesisHotShotBlockNumber)

	return &Client{
		EthClient:             ethClient,
		PoE:                   poe,
		Matic:                 matic,
		GlobalExitRootManager: globalExitRoot,
		HotShot:               hotshot,
		SCAddresses:           scAddresses,
		GasProviders: externalGasProviders{
			MultiGasProvider: cfg.MultiGasProvider,
			Providers:        gProviders,
		},
		cfg:  cfg,
		auth: map[common.Address]bind.TransactOpts{},
	}, nil
}

// VerifyGenBlockNumber verifies if the genesis Block Number is valid
func (etherMan *Client) VerifyGenBlockNumber(ctx context.Context, genBlockNumber uint64) (bool, error) {
	genBlock := big.NewInt(0).SetUint64(genBlockNumber)
	response, err := etherMan.EthClient.CodeAt(ctx, etherMan.cfg.PoEAddr, genBlock)
	if err != nil {
		log.Error("error getting smc code for gen block number. Error: ", err)
		return false, err
	}
	responseString := hex.EncodeToString(response)
	if responseString == "" {
		return false, nil
	}
	responsePrev, err := etherMan.EthClient.CodeAt(ctx, etherMan.cfg.PoEAddr, genBlock.Sub(genBlock, big.NewInt(1)))
	if err != nil {
		if parsedErr, ok := tryParseError(err); ok {
			if errors.Is(parsedErr, ErrMissingTrieNode) {
				return true, nil
			}
		}
		log.Error("error getting smc code for gen block number. Error: ", err)
		return false, err
	}
	responsePrevString := hex.EncodeToString(responsePrev)
	if responsePrevString != "" {
		return false, nil
	}
	return true, nil
}

// GetForks returns fork information
func (etherMan *Client) GetForks(ctx context.Context) ([]state.ForkIDInterval, error) {
	// Filter query
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(1),
		Addresses: etherMan.SCAddresses,
		Topics:    [][]common.Hash{{updateZkEVMVersionSignatureHash}},
	}
	logs, err := etherMan.EthClient.FilterLogs(ctx, query)
	if err != nil {
		return []state.ForkIDInterval{}, err
	}
	var forks []state.ForkIDInterval
	for i, l := range logs {
		zkevmVersion, err := etherMan.PoE.ParseUpdateZkEVMVersion(l)
		if err != nil {
			return []state.ForkIDInterval{}, err
		}
		var fork state.ForkIDInterval
		if i == 0 {
			fork = state.ForkIDInterval{
				FromBatchNumber: zkevmVersion.NumBatch,
				ToBatchNumber:   math.MaxUint64,
				ForkId:          zkevmVersion.ForkID,
				Version:         zkevmVersion.Version,
			}
		} else {
			forks[len(forks)-1].ToBatchNumber = zkevmVersion.NumBatch - 1
			fork = state.ForkIDInterval{
				FromBatchNumber: zkevmVersion.NumBatch,
				ToBatchNumber:   math.MaxUint64,
				ForkId:          zkevmVersion.ForkID,
				Version:         zkevmVersion.Version,
			}
		}
		forks = append(forks, fork)
	}
	log.Debugf("Forks decoded: %+v", forks)
	return forks, nil
}

// GetRollupInfoByBlockRange function retrieves the Rollup information that are included in all this ethereum blocks
// from block x to block y.
func (etherMan *Client) GetRollupInfoByBlockRange(ctx context.Context, fromBlock uint64, toBlock *uint64, prevBatch state.L2BatchInfo, usePreconfirmations bool) ([]Block, map[common.Hash][]Order, error) {
	// Filter query
	query := ethereum.FilterQuery{
		FromBlock: new(big.Int).SetUint64(fromBlock),
		Addresses: etherMan.SCAddresses,
	}
	if toBlock != nil {
		query.ToBlock = new(big.Int).SetUint64(*toBlock)
	}
	blocks, blocksOrder, err := etherMan.readEvents(ctx, prevBatch, query, usePreconfirmations)
	if err != nil {
		return nil, nil, err
	}
	return blocks, blocksOrder, nil
}

// Order contains the event order to let the synchronizer store the information following this order.
type Order struct {
	Name EventOrder
	Pos  int
}

func (etherMan *Client) readEvents(ctx context.Context, prevBatch state.L2BatchInfo, query ethereum.FilterQuery, usePreconfirmations bool) ([]Block, map[common.Hash][]Order, error) {
	logs, err := etherMan.EthClient.FilterLogs(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	var blocks []Block
	blocksOrder := make(map[common.Hash][]Order)
	for _, vLog := range logs {
		err := etherMan.processEvent(ctx, &prevBatch, vLog, &blocks, &blocksOrder, usePreconfirmations)
		if err != nil {
			log.Warnf("error processing event. Retrying... Error: %s. vLog: %+v", err.Error(), vLog)
			return nil, nil, err
		}
	}
	return blocks, blocksOrder, nil
}

func (etherMan *Client) processEvent(ctx context.Context, prevBatch *state.L2BatchInfo, vLog types.Log, blocks *[]Block, blocksOrder *map[common.Hash][]Order, usePreconfirmations bool) error {
	switch vLog.Topics[0] {
	case newBlocksSignatureHash:
		if usePreconfirmations {
			// When using preconfirmations, we get information about new blocks directly from the
			// sequencer, so we can ignore events indicating that new blocks have been received on
			// L1 (which happens later).
			return nil
		} else {
			return etherMan.newBlocksEvent(ctx, prevBatch, vLog, blocks, blocksOrder)
		}
	case updateGlobalExitRootSignatureHash:
		return etherMan.updateGlobalExitRootEvent(ctx, vLog, blocks, blocksOrder)
	case verifyBatchesTrustedAggregatorSignatureHash:
		return etherMan.verifyBatchesTrustedAggregatorEvent(ctx, vLog, blocks, blocksOrder)
	case verifyBatchesSignatureHash:
		log.Warn("VerifyBatches event not implemented yet")
		return nil
	case setTrustedSequencerURLSignatureHash:
		log.Debug("SetTrustedSequencerURL event detected")
		return nil
	case setTrustedSequencerSignatureHash:
		log.Debug("SetTrustedSequencer event detected")
		return nil
	case initializedSignatureHash:
		log.Debug("Initialized event detected")
		return nil
	case adminChangedSignatureHash:
		log.Debug("AdminChanged event detected")
		return nil
	case beaconUpgradedSignatureHash:
		log.Debug("BeaconUpgraded event detected")
		return nil
	case upgradedSignatureHash:
		log.Debug("Upgraded event detected")
		return nil
	case transferOwnershipSignatureHash:
		log.Debug("TransferOwnership event detected")
		return nil
	case setSecurityCouncilSignatureHash:
		log.Debug("SetSecurityCouncil event detected")
		return nil
	case proofDifferentStateSignatureHash:
		log.Debug("ProofDifferentState event detected")
		return nil
	case emergencyStateActivatedSignatureHash:
		log.Debug("EmergencyStateActivated event detected")
		return nil
	case emergencyStateDeactivatedSignatureHash:
		log.Debug("EmergencyStateDeactivated event detected")
		return nil
	case updateZkEVMVersionSignatureHash:
		log.Debug("UpdateZkEVMVersion event detected")
		return nil
	}
	log.Warn("Event not registered: ", vLog)
	return nil
}

func (etherMan *Client) updateGlobalExitRootEvent(ctx context.Context, vLog types.Log, blocks *[]Block, blocksOrder *map[common.Hash][]Order) error {
	log.Debug("UpdateGlobalExitRoot event detected")
	globalExitRoot, err := etherMan.GlobalExitRootManager.ParseUpdateGlobalExitRoot(vLog)
	if err != nil {
		return err
	}
	fullBlock, err := etherMan.EthClient.BlockByHash(ctx, vLog.BlockHash)
	if err != nil {
		return fmt.Errorf("error getting hashParent. BlockNumber: %d. Error: %w", vLog.BlockNumber, err)
	}
	var gExitRoot GlobalExitRoot
	gExitRoot.MainnetExitRoot = common.BytesToHash(globalExitRoot.MainnetExitRoot[:])
	gExitRoot.RollupExitRoot = common.BytesToHash(globalExitRoot.RollupExitRoot[:])
	gExitRoot.BlockNumber = vLog.BlockNumber
	gExitRoot.GlobalExitRoot = hash(globalExitRoot.MainnetExitRoot, globalExitRoot.RollupExitRoot)

	if len(*blocks) == 0 || ((*blocks)[len(*blocks)-1].BlockHash != vLog.BlockHash || (*blocks)[len(*blocks)-1].BlockNumber != vLog.BlockNumber) {
		block := prepareBlock(fullBlock)
		block.GlobalExitRoots = append(block.GlobalExitRoots, gExitRoot)
		*blocks = append(*blocks, block)
	} else if (*blocks)[len(*blocks)-1].BlockHash == vLog.BlockHash && (*blocks)[len(*blocks)-1].BlockNumber == vLog.BlockNumber {
		(*blocks)[len(*blocks)-1].GlobalExitRoots = append((*blocks)[len(*blocks)-1].GlobalExitRoots, gExitRoot)
	} else {
		log.Error("Error processing UpdateGlobalExitRoot event. BlockHash:", vLog.BlockHash, ". BlockNumber: ", vLog.BlockNumber)
		return fmt.Errorf("error processing UpdateGlobalExitRoot event")
	}
	or := Order{
		Name: GlobalExitRootsOrder,
		Pos:  len((*blocks)[len(*blocks)-1].GlobalExitRoots) - 1,
	}
	(*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash] = append((*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash], or)
	return nil
}

// WaitTxToBeMined waits for an L1 tx to be mined. It will return error if the tx is reverted or timeout is exceeded
func (etherMan *Client) WaitTxToBeMined(ctx context.Context, tx *types.Transaction, timeout time.Duration) (bool, error) {
	err := operations.WaitTxToBeMined(ctx, etherMan.EthClient, tx, timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// BuildTrustedVerifyBatchesTxData builds a []bytes to be sent to the PoE SC method TrustedVerifyBatches.
func (etherMan *Client) BuildTrustedVerifyBatchesTxData(lastVerifiedBatch, newVerifiedBatch uint64, inputs *ethmanTypes.FinalProofInputs) (to *common.Address, data []byte, err error) {
	opts, err := etherMan.generateRandomAuth()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build trusted verify batches, err: %w", err)
	}
	opts.NoSend = true
	// force nonce, gas limit and gas price to avoid querying it from the chain
	opts.Nonce = big.NewInt(1)
	opts.GasLimit = uint64(1)
	opts.GasPrice = big.NewInt(1)

	var newLocalExitRoot [32]byte
	copy(newLocalExitRoot[:], inputs.NewLocalExitRoot)

	var newStateRoot [32]byte
	copy(newStateRoot[:], inputs.NewStateRoot)

	// Referenced several times below
	finalproof := inputs.FinalProof

	proofA, err := strSliceToBigIntArray(finalproof.Proof.ProofA)
	if err != nil {
		return nil, nil, err
	}
	proofB, err := proofSlcToIntArray(finalproof.Proof.ProofB)
	if err != nil {
		return nil, nil, err
	}
	proofC, err := strSliceToBigIntArray(finalproof.Proof.ProofC)
	if err != nil {
		return nil, nil, err
	}

	const pendStateNum = 0 // TODO hardcoded for now until we implement the pending state feature

	var oldAIH [32]byte
	copy(oldAIH[:], finalproof.Public.PublicInputs.OldAccInputHash)

	var newAIH [32]byte
	copy(newAIH[:], finalproof.Public.NewAccInputHash)

	packedHotShotParams := polygonzkevm.PolygonZkEVMPackedHotShotParams{
		OldAccInputHash: oldAIH,
		NewAccInputHash: newAIH,
		CommProof:       []byte{}, // TODO: Not yet implemented
	}

	tx, err := etherMan.PoE.VerifyBatchesTrustedAggregator(
		&opts,
		pendStateNum,
		lastVerifiedBatch,
		newVerifiedBatch,
		newLocalExitRoot,
		newStateRoot,
		proofA,
		proofB,
		proofC,
		packedHotShotParams,
	)
	if err != nil {
		if parsedErr, ok := tryParseError(err); ok {
			err = parsedErr
		}
		return nil, nil, err
	}

	return tx.To(), tx.Data(), nil
}

// GetSendSequenceFee get super/trusted sequencer fee
func (etherMan *Client) GetSendSequenceFee(numBatches uint64) (*big.Int, error) {
	f, err := etherMan.PoE.GetCurrentBatchFee(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, err
	}
	fee := new(big.Int).Mul(f, new(big.Int).SetUint64(numBatches))
	return fee, nil
}

// TrustedSequencer gets trusted sequencer address
func (etherMan *Client) TrustedSequencer() (common.Address, error) {
	return etherMan.PoE.TrustedSequencer(&bind.CallOpts{Pending: false})
}

func (etherMan *Client) getMaxPreconfirmation() (uint64, error) {
	url := etherMan.cfg.HotShotQueryServiceURL + "/availability/block-height"
	response, err := http.Get(url)
	if err != nil {
		// Usually this means the hotshot query service is not yet running.
		// Returning the error here will cause the processing of the batch to be
		// retried.
		return 0, err
	}
	if response.StatusCode != 200 {
		return 0, fmt.Errorf("Query service responded with status code %d", response.StatusCode)
	}

	var blockHeight uint64
	err = json.NewDecoder(response.Body).Decode(&blockHeight)
	if err != nil {
		// It's unlikely we can recover from this error by retrying.
		panic(err)
	}
	return blockHeight, nil
}

func (etherMan *Client) GetPreconfirmations(ctx context.Context, prevBatch state.L2BatchInfo) ([]Block, map[common.Hash][]Order, error) {
	hotShotBlockHeight, err := etherMan.getMaxPreconfirmation()
	if err != nil {
		return nil, nil, err
	}

	var blocks []Block
	order := make(map[common.Hash][]Order)

	// Start fetching from the next L2 batch (prevBatch.Number + 1), adjusting batch numbers to
	// HotShot block numbers by offsetting by the HotShot block height at L2 genesis time.
	fromHotShotBlock := prevBatch.Number + 1 + etherMan.cfg.GenesisHotShotBlockNumber
	log.Infof("Getting HotShot blocks in range %d - %d", fromHotShotBlock, hotShotBlockHeight)
	for hotShotBlockNum := fromHotShotBlock; hotShotBlockNum < hotShotBlockHeight; hotShotBlockNum++ {
		var batch SequencedBatch
		err = etherMan.fetchL2Block(ctx, hotShotBlockNum, &prevBatch, &batch)
		if errors.Is(err, ErrSkipBatch) {
			continue
		} else if err != nil {
			return nil, nil, err
		}

		err = etherMan.appendSequencedBatches(ctx, []SequencedBatch{batch}, batch.BlockNumber, &blocks, &order)
		if err != nil {
			return nil, nil, err
		}
	}

	return blocks, order, nil
}

func (etherMan *Client) newBlocksEvent(ctx context.Context, prevBatch *state.L2BatchInfo, vLog types.Log, blocks *[]Block, blocksOrder *map[common.Hash][]Order) error {
	newBlocks, err := etherMan.HotShot.ParseNewBlocks(vLog)
	if err != nil {
		return err
	}
	log.Debugf("NewBlocks event detected %+v", newBlocks)

	// Read the tx for this event.
	tx, isPending, err := etherMan.EthClient.TransactionByHash(ctx, vLog.TxHash)
	if err != nil {
		return err
	} else if isPending {
		return fmt.Errorf("error tx is still pending. TxHash: %s", tx.Hash().String())
	}
	msg, err := core.TransactionToMessage(tx, types.NewLondonSigner(tx.ChainId()), big.NewInt(0))
	if err != nil {
		return err
	}
	sequences, err := etherMan.decodeSequencesHotShot(ctx, prevBatch, tx.Data(), *newBlocks, msg.From, msg.Nonce)
	if err != nil {
		return fmt.Errorf("error decoding the sequences: %v", err)
	}
	err = etherMan.appendSequencedBatches(ctx, sequences, vLog.BlockNumber, blocks, blocksOrder)
	if err != nil {
		return err
	}
	return nil
}

func (etherMan *Client) appendSequencedBatches(ctx context.Context, sequences []SequencedBatch, blockNumber uint64, blocks *[]Block, blocksOrder *map[common.Hash][]Order) error {
	if len(*blocks) == 0 || (*blocks)[len(*blocks)-1].BlockNumber != blockNumber {
		// Sanity check: if we got a new L1 block number, it should be increasing.
		if len(*blocks) > 0 && blockNumber < (*blocks)[len(*blocks)-1].BlockNumber {
			log.Fatalf("L1 block number decreased from %d to %d", (*blocks)[len(*blocks)-1].BlockNumber, blockNumber)
		}

		fullBlock, err := etherMan.EthBlockByNumber(ctx, blockNumber)
		if err != nil {
			return fmt.Errorf("error getting hashParent. BlockNumber: %d. Error: %w", blockNumber, err)
		}
		block := prepareBlock(fullBlock)
		block.SequencedBatches = append(block.SequencedBatches, sequences)
		*blocks = append(*blocks, block)
	} else {
		(*blocks)[len(*blocks)-1].SequencedBatches = append((*blocks)[len(*blocks)-1].SequencedBatches, sequences)
	}
	or := Order{
		Name: SequenceBatchesOrder,
		Pos:  len((*blocks)[len(*blocks)-1].SequencedBatches) - 1,
	}
	(*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash] = append((*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash], or)
	return nil
}

func (etherMan *Client) decodeSequencesHotShot(ctx context.Context, prevBatch *state.L2BatchInfo, txData []byte, newBlocks ihotshot.IhotshotNewBlocks, sequencer common.Address, nonce uint64) ([]SequencedBatch, error) {

	// Get number of batches by parsing transaction
	numNewBatches := newBlocks.NumBlocks.Uint64()
	firstHotShotBlockNum := newBlocks.FirstBlockNumber.Uint64()

	var sequencedBatches []SequencedBatch
	for i := uint64(0); i < numNewBatches; i++ {
		curHotShotBlockNum := firstHotShotBlockNum + i

		if curHotShotBlockNum <= etherMan.cfg.GenesisHotShotBlockNumber {
			log.Infof(
				"Hotshot block number %d not greater than genesis block number %d: skipping",
				curHotShotBlockNum,
				etherMan.cfg.GenesisHotShotBlockNumber,
			)
			continue
		}
		newBatch := SequencedBatch{}
		err := etherMan.fetchL2Block(ctx, curHotShotBlockNum, prevBatch, &newBatch)
		if errors.Is(err, ErrSkipBatch) {
			continue
		} else if err != nil {
			return nil, err
		}
		sequencedBatches = append(sequencedBatches, newBatch)
	}

	return sequencedBatches, nil
}

func (etherMan *Client) fetchL2Block(ctx context.Context, hotShotBlockNum uint64, prevBatch *state.L2BatchInfo, batch *SequencedBatch) error {
	// Get transactions and metadata from HotShot query service
	hotShotBlockNumStr := strconv.FormatUint(hotShotBlockNum, 10)

	// Retry until we get a response from the query service
	//
	// Very recent blocks may not available on the query service. If we return
	// an error immediately all the batches need to be re-processed and the
	// error will likely occur again. If instead we retry a few times here we
	// can successfully fetch all the L2 blocks.
	url := etherMan.cfg.HotShotQueryServiceURL + "/availability/block/" + hotShotBlockNumStr
	maxRetries := 10
	var response *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		response, err = http.Get(url)
		success := err == nil && response.StatusCode == 200

		// If there's no error we should close the response.
		if response != nil && !success {
			response.Body.Close()
		}

		if success {
			defer response.Body.Close()
			break
		}

		if err != nil {
			log.Warnf("Error fetching l2 block %d from query service, try %d: %v", hotShotBlockNum, i+1, err)
		} else {
			log.Warnf("Error fetching l2 block %d from query service, try %d: status code %d", hotShotBlockNum, i+1, response.StatusCode)
		}

		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// Handle the case if retries exhausted and still not successful
	if response == nil || response.StatusCode != 200 {
		return fmt.Errorf("Failed to fetch l2 block %d from query service after %d attempts", hotShotBlockNum, maxRetries)
	}

	var l2Block SequencerBlock
	err = json.NewDecoder(response.Body).Decode(&l2Block)
	if err != nil {
		// It's unlikely we can recover from this error by retrying.
		panic(err)
	}

	batchNum := hotShotBlockNum - etherMan.cfg.GenesisHotShotBlockNumber
	log.Infof("Creating batch number %d", batchNum)

	// Check that we got the expected batch.
	if batchNum < prevBatch.Number + 1 {
		// We got a batch which is older than expected. This means we have somehow already processed
		// this batch. This should not be possible, since each batch is included exactly once in the
		// batch stream, and the only case where we process the same section of the batch stream
		// twice is when retrying after an error, in which case we should have reverted any changes
		// we made to the state when processing the first time.
		//
		// If this happens, it is indicative of a programming error or a corrupt block stream, so we
		// will complain loudly. However, there is no actual harm done: since we have ostensibly
		// already processed this batch, we can just skip it and continue to make progress.
		log.Errorf("received old batch %d, prev batch is %v", batchNum, prevBatch)
		return ErrSkipBatch
	} else if batchNum > prevBatch.Number + 1 {
		// In this case we have somehow skipped a batch. This should also not be possible, because
		// we always process batches sequentially. This is indicative of a corrupt DB. All we can do
		// is return an error to trigger a retry in the caller.
		return fmt.Errorf("received batch %d from the future, prev batch is %v", batchNum, prevBatch)
	}

	// Adjust L1 block and timestamp as needed.
	//
	// This should not be necessary, since HotShot should enforce non-decreasing timestamps and L1
	// block numbers. However, since HotShot does not currently support the ValidatedState API,
	// timestamps and L1 block numbers proposed by leaders are not checked by replicas, and may
	// occasionally decrease. In this case, just use the previous value, to avoid breaking the rest
	// of the execution pipeline.
	if l2Block.L1Block < prevBatch.L1Block {
		log.Warnf("HotShot block %d has decreasing L1Block: %d-%d", l2Block.Height, prevBatch.L1Block, l2Block.L1Block)
		l2Block.L1Block = prevBatch.L1Block
	}
	if l2Block.Timestamp < prevBatch.Timestamp {
		log.Warnf("HotShot block %d has decreasing timestamp: %d-%d", l2Block.Height, prevBatch.Timestamp, l2Block.Timestamp)
		l2Block.Timestamp = prevBatch.Timestamp
	}
	*prevBatch = state.L2BatchInfo{
		Number:    batchNum,
		L1Block:   l2Block.L1Block,
		Timestamp: l2Block.Timestamp,
	}

	log.Infof(
		"Fetched L1 block %d, hotshot block: %d, timestamp %v, transactions %v",
		l2Block.L1Block,
		l2Block.Height,
		l2Block.Timestamp,
		l2Block.Transactions,
	)
	txns, err := hex.DecodeHex(l2Block.Transactions)
	if err != nil {
		// It's unlikely we can recover from this error by retrying.
		panic(err)
	}

	var ger [32]byte
	code, err := etherMan.EthClient.CodeAt(ctx, etherMan.cfg.GlobalExitRootManagerAddr, big.NewInt(int64(l2Block.L1Block)))
	if err != nil {
		return err
	}
	if len(code) == 0 {
		// Since this L2 is deployed onto an already-running HotShot sequencer, there may be HotShot
		// blocks from before the global exit root manager contract was deployed. These blocks
		// should not contain any transactions for this L2, since they were created before the L2
		// was deployed. In this case it doesn't matter what global exit root we use.
		if len(txns) != 0 {
			return fmt.Errorf("block %v (L1 block %v) contains L2 transactions from before GlobalExitRootManager was deployed", hotShotBlockNum, l2Block.L1Block)
		}
	} else {
		ger, err = etherMan.GlobalExitRootManager.GetLastGlobalExitRoot(&bind.CallOpts{BlockNumber: big.NewInt(int64(l2Block.L1Block))})
		if err != nil {
			return err
		}
	}

	newBatchData := PolygonZkEVMBatchData{
		Transactions:   txns,
		GlobalExitRoot: ger,
		Timestamp:      l2Block.Timestamp,
	}

	*batch = SequencedBatch{
		BatchNumber:           batchNum,
		BlockNumber:           l2Block.L1Block,
		PolygonZkEVMBatchData: newBatchData, // BatchData info

		// Some metadata (in particular: information about the L1 transaction which sequenced this
		// L2 batch in the rollup contract) is not available when using preconfirmations (since the
		// L2 batch _hasn't_ been sent to the rollup contract yet). Thus, we fill it with dummy
		// values. These values are not needed to compute the new L2 state or state transition
		// proof, since the rollup VM does not expose them. They are simply informational.
		SequencerAddr: common.Address{},
		TxHash:        common.Hash{},
		Coinbase:      common.Address{},
		Nonce:         0,
	}

	return nil
}

func (etherMan *Client) verifyBatchesTrustedAggregatorEvent(ctx context.Context, vLog types.Log, blocks *[]Block, blocksOrder *map[common.Hash][]Order) error {
	log.Debug("TrustedVerifyBatches event detected")
	vb, err := etherMan.PoE.ParseVerifyBatchesTrustedAggregator(vLog)
	if err != nil {
		return err
	}
	var trustedVerifyBatch VerifiedBatch
	trustedVerifyBatch.BlockNumber = vLog.BlockNumber
	trustedVerifyBatch.BatchNumber = vb.NumBatch
	trustedVerifyBatch.TxHash = vLog.TxHash
	trustedVerifyBatch.StateRoot = vb.StateRoot
	trustedVerifyBatch.Aggregator = vb.Aggregator

	if len(*blocks) == 0 || ((*blocks)[len(*blocks)-1].BlockHash != vLog.BlockHash || (*blocks)[len(*blocks)-1].BlockNumber != vLog.BlockNumber) {
		fullBlock, err := etherMan.EthClient.BlockByHash(ctx, vLog.BlockHash)
		if err != nil {
			return fmt.Errorf("error getting hashParent. BlockNumber: %d. Error: %w", vLog.BlockNumber, err)
		}
		block := prepareBlock(fullBlock)
		block.VerifiedBatches = append(block.VerifiedBatches, trustedVerifyBatch)
		*blocks = append(*blocks, block)
	} else if (*blocks)[len(*blocks)-1].BlockHash == vLog.BlockHash && (*blocks)[len(*blocks)-1].BlockNumber == vLog.BlockNumber {
		(*blocks)[len(*blocks)-1].VerifiedBatches = append((*blocks)[len(*blocks)-1].VerifiedBatches, trustedVerifyBatch)
	} else {
		log.Error("Error processing trustedVerifyBatch event. BlockHash:", vLog.BlockHash, ". BlockNumber: ", vLog.BlockNumber)
		return fmt.Errorf("error processing trustedVerifyBatch event")
	}
	or := Order{
		Name: TrustedVerifyBatchOrder,
		Pos:  len((*blocks)[len(*blocks)-1].VerifiedBatches) - 1,
	}
	(*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash] = append((*blocksOrder)[(*blocks)[len(*blocks)-1].BlockHash], or)
	return nil
}

func prepareBlock(fullBlock *types.Block) Block {
	var block Block
	block.BlockNumber = fullBlock.Number().Uint64()
	block.BlockHash = fullBlock.Hash()
	block.ParentHash = fullBlock.ParentHash()
	block.ReceivedAt = time.Unix(int64(fullBlock.Time()), 0)
	return block
}

func hash(data ...[32]byte) [32]byte {
	var res [32]byte
	hash := sha3.NewLegacyKeccak256()
	for _, d := range data {
		hash.Write(d[:]) //nolint:errcheck,gosec
	}
	copy(res[:], hash.Sum(nil))
	return res
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (etherMan *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return etherMan.EthClient.HeaderByNumber(ctx, number)
}

// EthBlockByNumber function retrieves the ethereum block information by ethereum block number.
func (etherMan *Client) EthBlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	block, err := etherMan.EthClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) || err.Error() == "block does not exist in blockchain" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return block, nil
}

// GetLastBatchTimestamp function allows to retrieve the lastTimestamp value in the smc
func (etherMan *Client) GetLastBatchTimestamp() (uint64, error) {
	return etherMan.PoE.LastTimestamp(&bind.CallOpts{Pending: false})
}

// GetLatestBatchNumber function allows to retrieve the latest proposed batch in the smc
func (etherMan *Client) GetLatestBatchNumber() (uint64, error) {
	return etherMan.PoE.LastBatchSequenced(&bind.CallOpts{Pending: false})
}

// GetLatestBlockNumber gets the latest block number from the ethereum
func (etherMan *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, nil)
	if err != nil || header == nil {
		return 0, err
	}
	return header.Number.Uint64(), nil
}

// GetLatestBlockTimestamp gets the latest block timestamp from the ethereum
func (etherMan *Client) GetLatestBlockTimestamp(ctx context.Context) (uint64, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, nil)
	if err != nil || header == nil {
		return 0, err
	}
	return header.Time, nil
}

// GetLatestVerifiedBatchNum gets latest verified batch from ethereum
func (etherMan *Client) GetLatestVerifiedBatchNum() (uint64, error) {
	return etherMan.PoE.LastVerifiedBatch(&bind.CallOpts{Pending: false})
}

// GetTx function get ethereum tx
func (etherMan *Client) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	return etherMan.EthClient.TransactionByHash(ctx, txHash)
}

// GetTxReceipt function gets ethereum tx receipt
func (etherMan *Client) GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return etherMan.EthClient.TransactionReceipt(ctx, txHash)
}

// ApproveMatic function allow to approve tokens in matic smc
func (etherMan *Client) ApproveMatic(ctx context.Context, account common.Address, maticAmount *big.Int, to common.Address) (*types.Transaction, error) {
	opts, err := etherMan.getAuthByAddress(account)
	if err == ErrNotFound {
		return nil, errors.New("can't find account private key to sign tx")
	}
	if etherMan.GasProviders.MultiGasProvider {
		opts.GasPrice = etherMan.GetL1GasPrice(ctx)
	}
	tx, err := etherMan.Matic.Approve(&opts, etherMan.cfg.PoEAddr, maticAmount)
	if err != nil {
		if parsedErr, ok := tryParseError(err); ok {
			err = parsedErr
		}
		return nil, fmt.Errorf("error approving balance to send the batch. Error: %w", err)
	}

	return tx, nil
}

// GetTrustedSequencerURL Gets the trusted sequencer url from rollup smc
func (etherMan *Client) GetTrustedSequencerURL() (string, error) {
	return etherMan.PoE.TrustedSequencerURL(&bind.CallOpts{Pending: false})
}

// GetL2ChainID returns L2 Chain ID
func (etherMan *Client) GetL2ChainID() (uint64, error) {
	return etherMan.PoE.ChainID(&bind.CallOpts{Pending: false})
}

// GetL2ForkID returns current L2 Fork ID
func (etherMan *Client) GetL2ForkID() (uint64, error) {
	// TODO: implement this
	return 1, nil
}

// GetL2ForkIDIntervals return L2 Fork ID intervals
func (etherMan *Client) GetL2ForkIDIntervals() ([]state.ForkIDInterval, error) {
	// TODO: implement this
	return []state.ForkIDInterval{{FromBatchNumber: 0, ToBatchNumber: math.MaxUint64, ForkId: 1}}, nil
}

// GetL1GasPrice gets the l1 gas price
func (etherMan *Client) GetL1GasPrice(ctx context.Context) *big.Int {
	// Get gasPrice from providers
	gasPrice := big.NewInt(0)
	for i, prov := range etherMan.GasProviders.Providers {
		gp, err := prov.SuggestGasPrice(ctx)
		if err != nil {
			log.Warnf("error getting gas price from provider %d. Error: %s", i+1, err.Error())
		} else if gasPrice.Cmp(gp) == -1 { // gasPrice < gp
			gasPrice = gp
		}
	}
	log.Debug("gasPrice chose: ", gasPrice)
	return gasPrice
}

// SendTx sends a tx to L1
func (etherMan *Client) SendTx(ctx context.Context, tx *types.Transaction) error {
	return etherMan.EthClient.SendTransaction(ctx, tx)
}

// CurrentNonce returns the current nonce for the provided account
func (etherMan *Client) CurrentNonce(ctx context.Context, account common.Address) (uint64, error) {
	return etherMan.EthClient.NonceAt(ctx, account, nil)
}

// SuggestedGasPrice returns the suggest nonce for the network at the moment
func (etherMan *Client) SuggestedGasPrice(ctx context.Context) (*big.Int, error) {
	suggestedGasPrice := etherMan.GetL1GasPrice(ctx)
	if suggestedGasPrice.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.New("failed to get the suggested gas price")
	}
	return suggestedGasPrice, nil
}

// EstimateGas returns the estimated gas for the tx
func (etherMan *Client) EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error) {
	return etherMan.EthClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	})
}

// CheckTxWasMined check if a tx was already mined
func (etherMan *Client) CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error) {
	receipt, err := etherMan.EthClient.TransactionReceipt(ctx, txHash)
	if errors.Is(err, ethereum.NotFound) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}

	return true, receipt, nil
}

// SignTx tries to sign a transaction accordingly to the provided sender
func (etherMan *Client) SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error) {
	auth, err := etherMan.getAuthByAddress(sender)
	if err == ErrNotFound {
		return nil, ErrPrivateKeyNotFound
	}
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}
	return signedTx, nil
}

// GetRevertMessage tries to get a revert message of a transaction
func (etherMan *Client) GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error) {
	if tx == nil {
		return "", nil
	}

	receipt, err := etherMan.GetTxReceipt(ctx, tx.Hash())
	if err != nil {
		return "", err
	}

	if receipt.Status == types.ReceiptStatusFailed {
		revertMessage, err := operations.RevertReason(ctx, etherMan.EthClient, tx, receipt.BlockNumber)
		if err != nil {
			return "", err
		}
		return revertMessage, nil
	}
	return "", nil
}

// AddOrReplaceAuth adds an authorization or replace an existent one to the same account
func (etherMan *Client) AddOrReplaceAuth(auth bind.TransactOpts) error {
	log.Infof("added or replaced authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth
	return nil
}

// LoadAuthFromKeyStore loads an authorization from a key store file
func (etherMan *Client) LoadAuthFromKeyStore(path, password string) (*bind.TransactOpts, error) {
	auth, err := newAuthFromKeystore(path, password, etherMan.cfg.L1ChainID)
	if err != nil {
		return nil, err
	}

	log.Infof("loaded authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth
	return &auth, nil
}

// newKeyFromKeystore creates an instance of a keystore key from a keystore file
func newKeyFromKeystore(path, password string) (*keystore.Key, error) {
	if path == "" && password == "" {
		return nil, nil
	}
	keystoreEncrypted, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	log.Infof("decrypting key from: %v", path)
	key, err := keystore.DecryptKey(keystoreEncrypted, password)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// newAuthFromKeystore an authorization instance from a keystore file
func newAuthFromKeystore(path, password string, chainID uint64) (bind.TransactOpts, error) {
	log.Infof("reading key from: %v", path)
	key, err := newKeyFromKeystore(path, password)
	if err != nil {
		return bind.TransactOpts{}, err
	}
	if key == nil {
		return bind.TransactOpts{}, nil
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey, new(big.Int).SetUint64(chainID))
	if err != nil {
		return bind.TransactOpts{}, err
	}
	return *auth, nil
}

// getAuthByAddress tries to get an authorization from the authorizations map
func (etherMan *Client) getAuthByAddress(addr common.Address) (bind.TransactOpts, error) {
	auth, found := etherMan.auth[addr]
	if !found {
		return bind.TransactOpts{}, ErrNotFound
	}
	return auth, nil
}

// generateRandomAuth generates an authorization instance from a
// randomly generated private key to be used to estimate gas for PoE
// operations NOT restricted to the Trusted Sequencer
func (etherMan *Client) generateRandomAuth() (bind.TransactOpts, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return bind.TransactOpts{}, errors.New("failed to generate a private key to estimate L1 txs")
	}
	chainID := big.NewInt(0).SetUint64(etherMan.cfg.L1ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return bind.TransactOpts{}, errors.New("failed to generate a fake authorization to estimate L1 txs")
	}

	return *auth, nil
}
