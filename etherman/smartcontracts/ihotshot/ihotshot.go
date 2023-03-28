// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ihotshot

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// IhotshotMetaData contains all meta data concerning the Ihotshot contract.
var IhotshotMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"firstBlockNumber\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"numBlocks\",\"type\":\"uint256\"}],\"name\":\"NewBlocks\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"blockNumber\",\"type\":\"uint256\"}],\"name\":\"commitments\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// IhotshotABI is the input ABI used to generate the binding from.
// Deprecated: Use IhotshotMetaData.ABI instead.
var IhotshotABI = IhotshotMetaData.ABI

// Ihotshot is an auto generated Go binding around an Ethereum contract.
type Ihotshot struct {
	IhotshotCaller     // Read-only binding to the contract
	IhotshotTransactor // Write-only binding to the contract
	IhotshotFilterer   // Log filterer for contract events
}

// IhotshotCaller is an auto generated read-only Go binding around an Ethereum contract.
type IhotshotCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IhotshotTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IhotshotTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IhotshotFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IhotshotFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IhotshotSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IhotshotSession struct {
	Contract     *Ihotshot         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IhotshotCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IhotshotCallerSession struct {
	Contract *IhotshotCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// IhotshotTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IhotshotTransactorSession struct {
	Contract     *IhotshotTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// IhotshotRaw is an auto generated low-level Go binding around an Ethereum contract.
type IhotshotRaw struct {
	Contract *Ihotshot // Generic contract binding to access the raw methods on
}

// IhotshotCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IhotshotCallerRaw struct {
	Contract *IhotshotCaller // Generic read-only contract binding to access the raw methods on
}

// IhotshotTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IhotshotTransactorRaw struct {
	Contract *IhotshotTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIhotshot creates a new instance of Ihotshot, bound to a specific deployed contract.
func NewIhotshot(address common.Address, backend bind.ContractBackend) (*Ihotshot, error) {
	contract, err := bindIhotshot(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Ihotshot{IhotshotCaller: IhotshotCaller{contract: contract}, IhotshotTransactor: IhotshotTransactor{contract: contract}, IhotshotFilterer: IhotshotFilterer{contract: contract}}, nil
}

// NewIhotshotCaller creates a new read-only instance of Ihotshot, bound to a specific deployed contract.
func NewIhotshotCaller(address common.Address, caller bind.ContractCaller) (*IhotshotCaller, error) {
	contract, err := bindIhotshot(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IhotshotCaller{contract: contract}, nil
}

// NewIhotshotTransactor creates a new write-only instance of Ihotshot, bound to a specific deployed contract.
func NewIhotshotTransactor(address common.Address, transactor bind.ContractTransactor) (*IhotshotTransactor, error) {
	contract, err := bindIhotshot(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IhotshotTransactor{contract: contract}, nil
}

// NewIhotshotFilterer creates a new log filterer instance of Ihotshot, bound to a specific deployed contract.
func NewIhotshotFilterer(address common.Address, filterer bind.ContractFilterer) (*IhotshotFilterer, error) {
	contract, err := bindIhotshot(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IhotshotFilterer{contract: contract}, nil
}

// bindIhotshot binds a generic wrapper to an already deployed contract.
func bindIhotshot(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(IhotshotABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ihotshot *IhotshotRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ihotshot.Contract.IhotshotCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ihotshot *IhotshotRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ihotshot.Contract.IhotshotTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ihotshot *IhotshotRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ihotshot.Contract.IhotshotTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Ihotshot *IhotshotCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Ihotshot.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Ihotshot *IhotshotTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Ihotshot.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Ihotshot *IhotshotTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Ihotshot.Contract.contract.Transact(opts, method, params...)
}

// Commitments is a free data retrieval call binding the contract method 0x49ce8997.
//
// Solidity: function commitments(uint256 blockNumber) view returns(uint256)
func (_Ihotshot *IhotshotCaller) Commitments(opts *bind.CallOpts, blockNumber *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Ihotshot.contract.Call(opts, &out, "commitments", blockNumber)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Commitments is a free data retrieval call binding the contract method 0x49ce8997.
//
// Solidity: function commitments(uint256 blockNumber) view returns(uint256)
func (_Ihotshot *IhotshotSession) Commitments(blockNumber *big.Int) (*big.Int, error) {
	return _Ihotshot.Contract.Commitments(&_Ihotshot.CallOpts, blockNumber)
}

// Commitments is a free data retrieval call binding the contract method 0x49ce8997.
//
// Solidity: function commitments(uint256 blockNumber) view returns(uint256)
func (_Ihotshot *IhotshotCallerSession) Commitments(blockNumber *big.Int) (*big.Int, error) {
	return _Ihotshot.Contract.Commitments(&_Ihotshot.CallOpts, blockNumber)
}

// IhotshotNewBlocksIterator is returned from FilterNewBlocks and is used to iterate over the raw logs and unpacked data for NewBlocks events raised by the Ihotshot contract.
type IhotshotNewBlocksIterator struct {
	Event *IhotshotNewBlocks // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *IhotshotNewBlocksIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(IhotshotNewBlocks)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(IhotshotNewBlocks)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *IhotshotNewBlocksIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *IhotshotNewBlocksIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// IhotshotNewBlocks represents a NewBlocks event raised by the Ihotshot contract.
type IhotshotNewBlocks struct {
	FirstBlockNumber *big.Int
	NumBlocks        *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterNewBlocks is a free log retrieval operation binding the contract event 0x8203a21e4f95f72e5081d5e0929b1a8c52141e123f9a14e1e74b0260fa5f52f1.
//
// Solidity: event NewBlocks(uint256 firstBlockNumber, uint256 numBlocks)
func (_Ihotshot *IhotshotFilterer) FilterNewBlocks(opts *bind.FilterOpts) (*IhotshotNewBlocksIterator, error) {

	logs, sub, err := _Ihotshot.contract.FilterLogs(opts, "NewBlocks")
	if err != nil {
		return nil, err
	}
	return &IhotshotNewBlocksIterator{contract: _Ihotshot.contract, event: "NewBlocks", logs: logs, sub: sub}, nil
}

// WatchNewBlocks is a free log subscription operation binding the contract event 0x8203a21e4f95f72e5081d5e0929b1a8c52141e123f9a14e1e74b0260fa5f52f1.
//
// Solidity: event NewBlocks(uint256 firstBlockNumber, uint256 numBlocks)
func (_Ihotshot *IhotshotFilterer) WatchNewBlocks(opts *bind.WatchOpts, sink chan<- *IhotshotNewBlocks) (event.Subscription, error) {

	logs, sub, err := _Ihotshot.contract.WatchLogs(opts, "NewBlocks")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(IhotshotNewBlocks)
				if err := _Ihotshot.contract.UnpackLog(event, "NewBlocks", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNewBlocks is a log parse operation binding the contract event 0x8203a21e4f95f72e5081d5e0929b1a8c52141e123f9a14e1e74b0260fa5f52f1.
//
// Solidity: event NewBlocks(uint256 firstBlockNumber, uint256 numBlocks)
func (_Ihotshot *IhotshotFilterer) ParseNewBlocks(log types.Log) (*IhotshotNewBlocks, error) {
	event := new(IhotshotNewBlocks)
	if err := _Ihotshot.contract.UnpackLog(event, "NewBlocks", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
