# Solidity Version Support

This document describes the Solidity versions supported by Favo Chain and provides guidance on how to upgrade EVM support.

## Current Support

### Supported EVM Hard Forks

Favo Chain currently supports the following Ethereum hard forks (enabled from block 0 by default):

| Hard Fork | Status | Key Features |
|-----------|--------|--------------|
| Homestead | ✅ Supported | DELEGATECALL, CREATE gas cost changes |
| EIP-150 (Tangerine Whistle) | ✅ Supported | Gas cost repricing |
| EIP-155 | ✅ Supported | Simple replay attack protection |
| EIP-158 (Spurious Dragon) | ✅ Supported | State trie clearing, EXP gas cost |
| Byzantium | ✅ Supported | REVERT, RETURNDATASIZE, RETURNDATACOPY, STATICCALL |
| Constantinople | ✅ Supported | SHL, SHR, SAR, EXTCODEHASH, CREATE2 |
| Petersburg | ✅ Supported | Removes EIP-1283 (net gas metering) |
| Istanbul | ✅ Supported | CHAINID, SELFBALANCE, EIP-1884 gas repricing |
| London | ✅ Supported | EIP-1559 (partially), gas cost changes |

### Supported Solidity Versions

Based on the supported EVM forks, Favo Chain is compatible with:

- **Recommended**: Solidity `^0.8.0` to `^0.8.19`
- **Maximum tested**: Solidity `0.8.19` (London-compatible)
- **Minimum supported**: Solidity `0.5.0`

> **Note**: When compiling Solidity contracts, use the `--evm-version london` flag to ensure compatibility.

### Compiler Settings

To compile contracts for Favo Chain, use the following settings:

```bash
# Using solc directly
solc --evm-version london --optimize --bin --abi YourContract.sol

# Using Hardhat (hardhat.config.js)
module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: { enabled: true, runs: 200 },
      evmVersion: "london"
    }
  }
};

# Using Foundry (foundry.toml)
[profile.default]
solc_version = "0.8.19"
evm_version = "london"
```

## Unsupported Features

The following EVM features are **not yet supported**:

| Feature | EIP | Hard Fork | Required for |
|---------|-----|-----------|--------------|
| BASEFEE opcode (0x48) | EIP-3198 | London (full) | `block.basefee` in Solidity |
| PREVRANDAO/DIFFICULTY opcode (0x44) | EIP-4399 | Paris (The Merge) | `block.prevrandao` in Solidity (replaces DIFFICULTY) |
| PUSH0 opcode (0x5F) | EIP-3855 | Shanghai | Optimized bytecode in Solidity ≥0.8.20 |
| Warm/Cold access lists | EIP-2929/2930 | Berlin | Access list transactions |

> **Note**: PREVRANDAO (EIP-4399) reuses the same opcode slot as DIFFICULTY (0x44). After Paris upgrade, the DIFFICULTY opcode returns the PREVRANDAO value instead.

### Solidity Versions to Avoid

- **Solidity ≥0.8.20**: Uses PUSH0 opcode (Shanghai), which is not supported
- **Contracts using `block.basefee`**: BASEFEE opcode not implemented
- **Contracts using `block.prevrandao`**: PREVRANDAO opcode not implemented

## How to Upgrade EVM Support

To add support for newer Solidity versions, you need to implement additional EVM opcodes and hard fork features.

### Step 1: Add New Fork to Chain Parameters

Edit `chain/params.go` to add the new fork:

```go
// Forks specifies when each fork is activated
type Forks struct {
    Homestead      *Fork `json:"homestead,omitempty"`
    Byzantium      *Fork `json:"byzantium,omitempty"`
    Constantinople *Fork `json:"constantinople,omitempty"`
    Petersburg     *Fork `json:"petersburg,omitempty"`
    Istanbul       *Fork `json:"istanbul,omitempty"`
    London         *Fork `json:"london,omitempty"`
    Paris          *Fork `json:"paris,omitempty"`     // Add new fork
    Shanghai       *Fork `json:"shanghai,omitempty"`  // Add new fork
    EIP150         *Fork `json:"EIP150,omitempty"`
    EIP158         *Fork `json:"EIP158,omitempty"`
    EIP155         *Fork `json:"EIP155,omitempty"`
}
```

Add the corresponding helper methods and update `ForksInTime` struct.

### Step 2: Add New Opcodes

Edit `state/runtime/evm/opcodes.go` to define new opcodes:

```go
// Example: Adding PUSH0 (EIP-3855)
PUSH0 = 0x5F

// Example: Adding BASEFEE (EIP-3198)
BASEFEE = 0x48
```

### Step 3: Implement Opcode Logic

Edit `state/runtime/evm/instructions.go` to implement the opcode behavior:

```go
// Example: PUSH0 implementation
func opPush0(c *state) {
    if !c.config.Shanghai {
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().SetUint64(0)
}

// Example: BASEFEE implementation
func opBaseFee(c *state) {
    if !c.config.LondonFull {  // Requires full London support with EIP-3198
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().Set(c.host.GetTxContext().BaseFee)
}
```

### Step 4: Register Opcodes in Dispatch Table

Edit `state/runtime/evm/dispatch_table.go`:

```go
register(PUSH0, handler{opPush0, 0, 2})
register(BASEFEE, handler{opBaseFee, 0, 2})
```

### Step 5: Update Transaction Context (if needed)

For opcodes that require new context data (like BASEFEE), update `state/runtime/runtime.go`:

```go
type TxContext struct {
    // ... existing fields ...
    BaseFee *big.Int // Add for EIP-3198
}
```

### Step 6: Update All Fork Configurations

Update `AllForksEnabled` in `chain/params.go`:

```go
var AllForksEnabled = &Forks{
    // ... existing forks ...
    Paris:    NewFork(0),
    Shanghai: NewFork(0),
}
```

### Step 7: Add Tests

Create tests in `state/runtime/evm/instructions_test.go` for new opcodes:

```go
func TestOpPush0(t *testing.T) {
    // Test PUSH0 opcode
}
```

### Step 8: Update Genesis Configuration

For new chains, update the genesis configuration to enable new forks:

```json
{
  "params": {
    "forks": {
      "homestead": 0,
      "byzantium": 0,
      "constantinople": 0,
      "petersburg": 0,
      "istanbul": 0,
      "london": 0,
      "paris": 0,
      "shanghai": 0
    }
  }
}
```

## Example: Adding Shanghai Fork (PUSH0)

Here's a complete example of adding Shanghai fork support:

### 1. Update `chain/params.go`

```go
type Forks struct {
    // ... existing fields ...
    Shanghai *Fork `json:"shanghai,omitempty"`
}

func (f *Forks) IsShanghai(block uint64) bool {
    return f.active(f.Shanghai, block)
}

func (f *Forks) At(block uint64) ForksInTime {
    return ForksInTime{
        // ... existing fields ...
        Shanghai: f.active(f.Shanghai, block),
    }
}

type ForksInTime struct {
    // ... existing fields ...
    Shanghai bool
}

var AllForksEnabled = &Forks{
    // ... existing forks ...
    Shanghai: NewFork(0),
}
```

### 2. Update `state/runtime/evm/opcodes.go`

```go
// PUSH0 pushes 0 onto the stack (EIP-3855)
PUSH0 = 0x5F
```

### 3. Update `state/runtime/evm/instructions.go`

```go
func opPush0(c *state) {
    if !c.config.Shanghai {
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().SetUint64(0)
}
```

### 4. Update `state/runtime/evm/dispatch_table.go`

```go
register(PUSH0, handler{opPush0, 0, 2})
```

## References

- [Ethereum EVM Opcodes](https://www.evm.codes/)
- [Solidity Release Notes](https://github.com/ethereum/solidity/releases)
- [Ethereum Hard Fork History](https://ethereum.org/en/history/)
- [EIP Repository](https://eips.ethereum.org/)

## Testing Compatibility

To test if your contract is compatible:

1. Compile with `--evm-version london`
2. Check for unsupported opcodes in the bytecode
3. Deploy to a test network and verify functionality

```bash
# Check compiled bytecode for unsupported opcodes (PUSH0=0x5f, BASEFEE=0x48)
# Note: This is a simplified check - actual bytecode analysis requires proper disassembly
solc --evm-version london --bin YourContract.sol
```
