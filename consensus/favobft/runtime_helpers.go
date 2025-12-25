package favobft

import (
	"errors"

	"github.com/newton2049/favo-chain/blockchain"
	"github.com/newton2049/favo-chain/types"
)

// isEndOfPeriod checks if an end of a period (either it be sprint or epoch)
// is reached with the current block (the parent block of the current fsm iteration)
func isEndOfPeriod(blockNumber, periodSize uint64) bool {
	return blockNumber%periodSize == 0
}

// getBlockData returns block header and extra
// Note: Retry logic is applied only for header retrieval, not for extra data parsing,
// as extra data parsing errors are permanent (malformed data won't fix itself on retry)
func getBlockData(blockNumber uint64, blockchainBackend blockchainBackend) (*types.Header, *Extra, error) {
	blockHeader, found := blockchainBackend.GetHeaderByNumber(blockNumber)
	if !found {
		return nil, nil, blockchain.ErrNoBlock
	}

	blockExtra, err := GetIbftExtra(blockHeader.ExtraData)
	if err != nil {
		return nil, nil, err
	}

	return blockHeader, blockExtra, nil
}

// isEpochEndingBlock checks if given block is an epoch ending block with improved validation
func isEpochEndingBlock(blockNumber uint64, extra *Extra, blockchain blockchainBackend) (bool, error) {
	if !extra.Validators.IsEmpty() {
		// if validator set delta is not empty, the validator set was changed in this block
		// meaning the epoch changed as well
		return true, nil
	}

	_, nextBlockExtra, err := getBlockData(blockNumber+1, blockchain)
	if err != nil {
		// Distinguish between "no block" (expected case) and other errors
		if errors.Is(err, blockchain.ErrNoBlock) {
			// No next block means we can't determine if this is epoch ending
			return false, err
		}
		// For other errors, still return them but log context would be helpful
		return false, err
	}

	// validator set delta can be empty (no change in validator set happened)
	// so we need to check if their epoch numbers are different
	return extra.Checkpoint.EpochNumber != nextBlockExtra.Checkpoint.EpochNumber, nil
}
