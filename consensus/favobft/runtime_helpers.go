package favobft

import (
	"errors"
	"time"

	"github.com/newton2049/favo-chain/blockchain"
	"github.com/newton2049/favo-chain/types"
)

// isEndOfPeriod checks if an end of a period (either it be sprint or epoch)
// is reached with the current block (the parent block of the current fsm iteration)
func isEndOfPeriod(blockNumber, periodSize uint64) bool {
	return blockNumber%periodSize == 0
}

// getBlockData returns block header and extra with retry logic for transient errors
func getBlockData(blockNumber uint64, blockchainBackend blockchainBackend) (*types.Header, *Extra, error) {
	const (
		maxRetries   = 3
		retryDelayMs = 50
	)

	var (
		blockHeader *types.Header
		blockExtra  *Extra
		err         error
		found       bool
	)

	for attempt := 0; attempt < maxRetries; attempt++ {
		blockHeader, found = blockchainBackend.GetHeaderByNumber(blockNumber)
		if !found {
			return nil, nil, blockchain.ErrNoBlock
		}

		blockExtra, err = GetIbftExtra(blockHeader.ExtraData)
		if err == nil {
			return blockHeader, blockExtra, nil
		}

		// Only retry on potential transient errors, not on permanent decode errors
		if attempt < maxRetries-1 {
			time.Sleep(time.Duration(retryDelayMs) * time.Millisecond)
		}
	}

	return nil, nil, err
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
