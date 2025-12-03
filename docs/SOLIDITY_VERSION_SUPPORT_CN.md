# Solidity 版本支持

本文档描述 Favo Chain 支持的 Solidity 版本，并提供 EVM 升级指南。

## 当前支持情况

### 支持的 EVM 硬分叉

Favo Chain 目前支持以下以太坊硬分叉（默认从区块 0 开始启用）：

| 硬分叉 | 状态 | 主要特性 |
|--------|------|----------|
| Homestead | ✅ 支持 | DELEGATECALL，CREATE gas 成本变更 |
| EIP-150 (Tangerine Whistle) | ✅ 支持 | Gas 成本重新定价 |
| EIP-155 | ✅ 支持 | 简单重放攻击保护 |
| EIP-158 (Spurious Dragon) | ✅ 支持 | 状态树清理，EXP gas 成本 |
| Byzantium | ✅ 支持 | REVERT, RETURNDATASIZE, RETURNDATACOPY, STATICCALL |
| Constantinople | ✅ 支持 | SHL, SHR, SAR, EXTCODEHASH, CREATE2 |
| Petersburg | ✅ 支持 | 移除 EIP-1283（净 gas 计量） |
| Istanbul | ✅ 支持 | CHAINID, SELFBALANCE, EIP-1884 gas 重新定价 |
| London | ✅ 支持 | EIP-1559（部分支持），gas 成本变更 |

### 支持的 Solidity 版本

基于支持的 EVM 分叉，Favo Chain 兼容：

- **推荐版本**：Solidity `^0.8.0` 到 `^0.8.19`
- **最高测试版本**：Solidity `0.8.19`（London 兼容）
- **最低支持版本**：Solidity `0.5.0`

> **注意**：编译 Solidity 合约时，请使用 `--evm-version london` 标志以确保兼容性。

### 编译器设置

编译 Favo Chain 合约时，请使用以下设置：

```bash
# 直接使用 solc
solc --evm-version london --optimize --bin --abi YourContract.sol

# 使用 Hardhat（hardhat.config.js）
module.exports = {
  solidity: {
    version: "0.8.19",
    settings: {
      optimizer: { enabled: true, runs: 200 },
      evmVersion: "london"
    }
  }
};

# 使用 Foundry（foundry.toml）
[profile.default]
solc_version = "0.8.19"
evm_version = "london"
```

## 不支持的功能

以下 EVM 功能**尚未支持**：

| 功能 | EIP | 硬分叉 | 用途 |
|------|-----|--------|------|
| BASEFEE 操作码 (0x48) | EIP-3198 | London（完整版） | Solidity 中的 `block.basefee` |
| PREVRANDAO/DIFFICULTY 操作码 (0x44) | EIP-4399 | Paris（合并） | Solidity 中的 `block.prevrandao`（替代 DIFFICULTY） |
| PUSH0 操作码 (0x5F) | EIP-3855 | Shanghai | Solidity ≥0.8.20 的优化字节码 |
| 热/冷访问列表 | EIP-2929/2930 | Berlin | 访问列表交易 |

> **注意**：PREVRANDAO（EIP-4399）复用与 DIFFICULTY（0x44）相同的操作码槽位。Paris 升级后，DIFFICULTY 操作码返回 PREVRANDAO 值。

### 应避免的 Solidity 版本

- **Solidity ≥0.8.20**：使用 PUSH0 操作码（Shanghai），不支持
- **使用 `block.basefee` 的合约**：BASEFEE 操作码未实现
- **使用 `block.prevrandao` 的合约**：PREVRANDAO 操作码未实现

## 如何升级 EVM 支持

要添加对更新 Solidity 版本的支持，您需要实现额外的 EVM 操作码和硬分叉功能。

### 步骤 1：添加新分叉到链参数

编辑 `chain/params.go` 添加新分叉：

```go
// Forks 指定每个分叉的激活时间
type Forks struct {
    Homestead      *Fork `json:"homestead,omitempty"`
    Byzantium      *Fork `json:"byzantium,omitempty"`
    Constantinople *Fork `json:"constantinople,omitempty"`
    Petersburg     *Fork `json:"petersburg,omitempty"`
    Istanbul       *Fork `json:"istanbul,omitempty"`
    London         *Fork `json:"london,omitempty"`
    Paris          *Fork `json:"paris,omitempty"`     // 添加新分叉
    Shanghai       *Fork `json:"shanghai,omitempty"`  // 添加新分叉
    EIP150         *Fork `json:"EIP150,omitempty"`
    EIP158         *Fork `json:"EIP158,omitempty"`
    EIP155         *Fork `json:"EIP155,omitempty"`
}
```

添加相应的辅助方法并更新 `ForksInTime` 结构体。

### 步骤 2：添加新操作码

编辑 `state/runtime/evm/opcodes.go` 定义新操作码：

```go
// 示例：添加 PUSH0（EIP-3855）
PUSH0 = 0x5F

// 示例：添加 BASEFEE（EIP-3198）
BASEFEE = 0x48
```

### 步骤 3：实现操作码逻辑

编辑 `state/runtime/evm/instructions.go` 实现操作码行为：

```go
// 示例：PUSH0 实现
func opPush0(c *state) {
    if !c.config.Shanghai {
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().SetUint64(0)
}

// 示例：BASEFEE 实现
func opBaseFee(c *state) {
    if !c.config.LondonFull {  // 需要完整的 London 支持（包含 EIP-3198）
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().Set(c.host.GetTxContext().BaseFee)
}
```

### 步骤 4：在分发表中注册操作码

编辑 `state/runtime/evm/dispatch_table.go`：

```go
register(PUSH0, handler{opPush0, 0, 2})
register(BASEFEE, handler{opBaseFee, 0, 2})
```

### 步骤 5：更新交易上下文（如需要）

对于需要新上下文数据的操作码（如 BASEFEE），更新 `state/runtime/runtime.go`：

```go
type TxContext struct {
    // ... 现有字段 ...
    BaseFee *big.Int // 为 EIP-3198 添加
}
```

### 步骤 6：更新所有分叉配置

更新 `chain/params.go` 中的 `AllForksEnabled`：

```go
var AllForksEnabled = &Forks{
    // ... 现有分叉 ...
    Paris:    NewFork(0),
    Shanghai: NewFork(0),
}
```

### 步骤 7：添加测试

在 `state/runtime/evm/instructions_test.go` 中为新操作码创建测试：

```go
func TestOpPush0(t *testing.T) {
    // 测试 PUSH0 操作码
}
```

### 步骤 8：更新创世配置

对于新链，更新创世配置以启用新分叉：

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

## 示例：添加 Shanghai 分叉（PUSH0）

这是添加 Shanghai 分叉支持的完整示例：

### 1. 更新 `chain/params.go`

```go
type Forks struct {
    // ... 现有字段 ...
    Shanghai *Fork `json:"shanghai,omitempty"`
}

func (f *Forks) IsShanghai(block uint64) bool {
    return f.active(f.Shanghai, block)
}

func (f *Forks) At(block uint64) ForksInTime {
    return ForksInTime{
        // ... 现有字段 ...
        Shanghai: f.active(f.Shanghai, block),
    }
}

type ForksInTime struct {
    // ... 现有字段 ...
    Shanghai bool
}

var AllForksEnabled = &Forks{
    // ... 现有分叉 ...
    Shanghai: NewFork(0),
}
```

### 2. 更新 `state/runtime/evm/opcodes.go`

```go
// PUSH0 将 0 推入栈中（EIP-3855）
PUSH0 = 0x5F
```

### 3. 更新 `state/runtime/evm/instructions.go`

```go
func opPush0(c *state) {
    if !c.config.Shanghai {
        c.exit(errOpCodeNotFound)
        return
    }
    c.push1().SetUint64(0)
}
```

### 4. 更新 `state/runtime/evm/dispatch_table.go`

```go
register(PUSH0, handler{opPush0, 0, 2})
```

## 参考资料

- [以太坊 EVM 操作码](https://www.evm.codes/)
- [Solidity 发布说明](https://github.com/ethereum/solidity/releases)
- [以太坊硬分叉历史](https://ethereum.org/en/history/)
- [EIP 仓库](https://eips.ethereum.org/)

## 测试兼容性

测试您的合约是否兼容：

1. 使用 `--evm-version london` 编译
2. 检查字节码中是否有不支持的操作码
3. 部署到测试网络并验证功能

```bash
# 检查编译后的字节码是否包含不支持的操作码（PUSH0=0x5f，BASEFEE=0x48）
# 注意：这是简化检查 - 实际字节码分析需要正确的反汇编
solc --evm-version london --bin YourContract.sol
```

## 快速参考

| Solidity 版本 | 兼容性 | 注意事项 |
|---------------|--------|----------|
| 0.5.x | ✅ 兼容 | 使用 `--evm-version london` |
| 0.6.x | ✅ 兼容 | 使用 `--evm-version london` |
| 0.7.x | ✅ 兼容 | 使用 `--evm-version london` |
| 0.8.0 - 0.8.19 | ✅ 推荐 | 使用 `--evm-version london` |
| 0.8.20+ | ⚠️ 需注意 | 默认使用 PUSH0，需设置 `--evm-version london` |
