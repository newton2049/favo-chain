package favobft

import (
	"sync"

	"github.com/newton2049/favo-chain/types"
)

var setupHeaderHashFuncOnce sync.Once

// favoBFTHeaderHash defines the custom implementation for getting the header hash,
// because of the extraData field
func setupHeaderHashFunc() {
	setupHeaderHashFuncOnce.Do(func() {
		originalHeaderHash := types.HeaderHash

		types.HeaderHash = func(h *types.Header) types.Hash {
			// when hashing the block for signing we have to remove from
			// the extra field the seal and committed seal items
			extra, err := GetIbftExtraClean(h.ExtraData)
			if err != nil {
				return types.ZeroHash
			}

			// override extra data without seals and committed seal items
			hh := h.Copy()
			hh.ExtraData = extra

			return originalHeaderHash(hh)
		}
	})
}
