# Validator 节点操作指南

本文档整理了 Favo Chain 中与 validator 节点相关的所有操作和命令。

## 目录

- [1. 密钥初始化](#1-密钥初始化)
- [2. Manifest 文件生成](#2-manifest-文件生成)
- [3. Genesis 配置](#3-genesis-配置)
- [4. 启动 Validator 节点](#4-启动-validator-节点)
- [5. Validator 注册与白名单](#5-validator-注册与白名单)
- [6. 质押操作](#6-质押操作)
- [7. 解除质押操作](#7-解除质押操作)
- [8. 提取奖励](#8-提取奖励)
- [9. 查询 Validator 信息](#9-查询-validator-信息)
- [10. IBFT Validator 管理](#10-ibft-validator-管理)
- [11. Rootchain 资金操作](#11-rootchain-资金操作)

---

## 1. 密钥初始化

### 1.1 favobft-secrets 命令

初始化 validator 节点所需的私钥（ECDSA、BLS 和 P2P 网络密钥）。

```bash
favo-chain favobft-secrets --data-dir <数据目录> --num <数量>
```

**参数说明：**
- `--data-dir`: 数据目录前缀，例如 `test-chain-`
- `--num`: 需要创建的账户数量

**示例：**
```bash
# 创建 4 个 validator 账户
favo-chain favobft-secrets --data-dir test-chain- --num 4
```

### 1.2 secrets init 命令

通用密钥初始化命令，支持本地存储和云端密钥管理器。

```bash
favo-chain secrets init [--data-dir <数据目录>] [--config <配置文件路径>] [--num <数量>]
```

**参数说明：**
- `--data-dir`: 本地 FS 存储的数据目录
- `--config`: SecretsManager 配置文件路径（用于云端密钥管理）
- `--num`: 创建的密钥数量（仅限本地 FS）
- `--ecdsa`: 是否创建 ECDSA 密钥（默认 true）
- `--network`: 是否创建网络密钥（默认 true）
- `--bls`: 是否创建 BLS 密钥（默认 true）
- `--insecure`: 是否使用非加密存储

### 1.3 secrets output 命令

输出 validator 密钥地址和公共网络密钥。

```bash
favo-chain secrets output --data-dir <数据目录>
```

**参数说明：**
- `--data-dir`: 存储密钥的数据目录
- `--validator`: 仅输出 validator 密钥地址
- `--node-id`: 仅输出节点 ID
- `--bls`: 仅输出 BLS 公钥

---

## 2. Manifest 文件生成

Manifest 文件包含公共 validator 信息和 bridge 配置，用于 genesis 规范生成和 rootchain 合约部署。

### 2.1 基于本地存储生成

```bash
favo-chain manifest [--validators-path <路径>] [--validators-prefix <前缀>] [--path <输出路径>]
```

**参数说明：**
- `--validators-path`: 包含 validator 密钥的根路径（默认 `./`）
- `--validators-prefix`: validator 文件夹前缀名称（默认 `test-chain-`）
- `--path`: manifest 文件输出路径（默认 `./manifest.json`）
- `--premine-validators`: validator 预挖余额
- `--stake`: validator 质押金额
- `--chain-id`: 链 ID

**示例：**
```bash
favo-chain manifest --validators-path ./ --validators-prefix test-chain- --path ./manifest.json
```

### 2.2 手动指定 Validator 信息

当 validator 信息分布在多个主机上时，使用 `--validators` 标志手动指定：

```bash
favo-chain manifest --validators <validator信息>
```

**Validator 信息格式：**
```
<P2P multi address>:<ECDSA address>:<public BLS key>:<BLS signature>
```

**示例：**
```bash
favo-chain manifest \
  --validators /ip4/127.0.0.1/tcp/30301/p2p/16Uiu2HAmV5hqAp77untfJRorxqKmyUxgaVn8YHFjBJm9gKMms3mr:0xDcBe0024206ec42b0Ef4214Ac7B71aeae1A11af0:1cf134e02c6b2afb2ceda50bf2c9a01da367ac48f7783ee6c55444e1cab418ec0f52837b90a4d8cf944814073fc6f2bd96f35366a3846a8393e3cb0b19197cde23e2b40c6401fa27ff7d0c36779d9d097d1393cab6fc1d332f92fb3df850b78703b2989d567d1344e219f0667a1863f52f7663092276770cf513f9704b5351c4:11b18bde524f4b02258a8d196b687f8d8e9490d536718666dc7babca14eccb631c238fb79aa2b44a5a4dceccad2dd797f537008dda185d952226a814c1acf7c2 \
  --path ./manifest.json
```

---

## 3. Genesis 配置

生成区块链的创世配置文件。

```bash
favo-chain genesis [--consensus favobft] [--manifest <manifest路径>] [--validator-set-size <大小>]
```

**参数说明：**
- `--consensus`: 共识协议（`favobft`）
- `--manifest`: manifest 文件路径
- `--validator-set-size`: validator 集合大小（默认 100）
- `--block-gas-limit`: 区块 gas 限制
- `--epoch-size`: epoch 大小
- `--bridge-json-rpc`: rootchain JSON-RPC 地址

**示例：**
```bash
favo-chain genesis --consensus favobft --block-gas-limit 10000000 --epoch-size 10 --manifest ./manifest.json
```

---

## 4. 启动 Validator 节点

### 4.1 基本启动命令

```bash
favo-chain server --data-dir <数据目录> --chain <genesis文件> [选项]
```

**常用参数：**
- `--data-dir`: validator 数据目录
- `--chain`: genesis 配置文件路径
- `--grpc-address`: gRPC 服务地址
- `--libp2p`: libp2p 服务地址
- `--jsonrpc`: JSON-RPC 服务地址
- `--seal`: 启用出块功能（validator 必须开启）
- `--log-level`: 日志级别（DEBUG, INFO, WARN, ERROR）
- `--relayer`: 启用状态同步中继服务

**示例（4 节点集群）：**
```bash
# 节点 1
favo-chain server --data-dir ./test-chain-1 --chain genesis.json --grpc-address :5001 --libp2p :30301 --jsonrpc :9545 --seal --log-level DEBUG

# 节点 2
favo-chain server --data-dir ./test-chain-2 --chain genesis.json --grpc-address :5002 --libp2p :30302 --jsonrpc :10002 --seal --log-level DEBUG

# 节点 3
favo-chain server --data-dir ./test-chain-3 --chain genesis.json --grpc-address :5003 --libp2p :30303 --jsonrpc :10003 --seal --log-level DEBUG

# 节点 4
favo-chain server --data-dir ./test-chain-4 --chain genesis.json --grpc-address :5004 --libp2p :30304 --jsonrpc :10004 --seal --log-level DEBUG
```

### 4.2 Relayer 模式

启用 relayer 模式允许自动执行 deposit 事件：

```bash
favo-chain server --data-dir ./test-chain-1 --chain genesis.json --grpc-address :5001 --libp2p :30301 --jsonrpc :9545 --seal --log-level DEBUG --relayer
```

---

## 5. Validator 注册与白名单

### 5.1 注册 Validator

注册并质押一个已列入白名单的 validator。

```bash
favo-chain favobft register-validator --data-dir <数据目录> [--stake <质押金额>] [--chain-id <链ID>] [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: validator 账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--stake`: 质押金额
- `--chain-id`: 链 ID
- `--jsonrpc`: JSON-RPC 地址

### 5.2 Validator 白名单

将新 validator 添加到白名单（需要管理员权限）。

```bash
favo-chain favobft whitelist-validator --data-dir <管理员数据目录> --new-validator-address <新validator地址> [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: 管理员账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--new-validator-address`: 新 validator 的账户地址
- `--jsonrpc`: JSON-RPC 地址

---

## 6. 质押操作

### 6.1 自我质押

validator 为自己质押：

```bash
favo-chain favobft stake --data-dir <数据目录> --self --amount <金额> [--jsonrpc <RPC地址>]
```

### 6.2 委托质押

将质押委托给其他 validator：

```bash
favo-chain favobft stake --data-dir <数据目录> --delegate-address <validator地址> --amount <金额> [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: 账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--self`: 自我质押标志
- `--delegate-address`: 委托目标 validator 地址
- `--amount`: 质押金额
- `--jsonrpc`: JSON-RPC 地址

---

## 7. 解除质押操作

### 7.1 自我解除质押

```bash
favo-chain favobft unstake --data-dir <数据目录> --self --amount <金额> [--jsonrpc <RPC地址>]
```

### 7.2 取消委托

```bash
favo-chain favobft unstake --data-dir <数据目录> --undelegate-address <validator地址> --amount <金额> [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: 账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--self`: 自我解除质押标志
- `--undelegate-address`: 取消委托的 validator 地址
- `--amount`: 解除质押金额
- `--jsonrpc`: JSON-RPC 地址

---

## 8. 提取奖励

提取可提取的奖励金额。

```bash
favo-chain favobft withdraw --data-dir <数据目录> --address-to <目标地址> [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: 账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--address-to`: 提取奖励的目标地址
- `--jsonrpc`: JSON-RPC 地址

---

## 9. 查询 Validator 信息

获取 validator 的详细信息。

```bash
favo-chain favobft validator-info --data-dir <数据目录> [--jsonrpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: validator 账户数据目录
- `--config`: 云端密钥管理器配置文件
- `--jsonrpc`: JSON-RPC 地址

**返回信息：**
- `address`: validator 地址
- `stake`: 自我质押金额
- `totalStake`: 总质押金额（包含委托）
- `commission`: 佣金比例
- `withdrawableRewards`: 可提取奖励
- `active`: 是否活跃

---

## 10. IBFT Validator 管理

### 10.1 查看 IBFT 状态

```bash
favo-chain ibft status [--grpc-address <gRPC地址>]
```

返回当前 validator 密钥信息。

### 10.2 查看 IBFT 快照

```bash
favo-chain ibft snapshot [--number <区块号>] [--grpc-address <gRPC地址>]
```

返回指定区块号的 IBFT 快照（包含 validator 集合信息）。

### 10.3 提议 Validator 变更

添加或移除 validator：

```bash
favo-chain ibft propose --addr <validator地址> --vote <auth|drop> [--bls <BLS公钥>] [--grpc-address <gRPC地址>]
```

**参数说明：**
- `--addr`: 被投票的账户地址
- `--vote`: 投票类型（`auth` 添加，`drop` 移除）
- `--bls`: BLS 公钥（添加 validator 时需要）

### 10.4 查看候选 Validator

```bash
favo-chain ibft candidates [--grpc-address <gRPC地址>]
```

查看当前提议的候选 validator 列表。

### 10.5 查看法定人数信息

```bash
favo-chain ibft quorum [--grpc-address <gRPC地址>]
```

---

## 11. Rootchain 资金操作

### 11.1 启动 Rootchain 测试服务器

启动本地 Geth 开发模式实例（仅用于测试）：

```bash
favo-chain rootchain server
```

### 11.2 部署 Rootchain 合约

```bash
favo-chain rootchain init-contracts --data-dir <数据目录> [--manifest <manifest路径>] [--json-rpc <RPC地址>] [--test]
```

**参数说明：**
- `--data-dir`: 本地密钥存储路径
- `--config`: 云端密钥管理器配置文件
- `--manifest`: manifest 文件路径
- `--json-rpc`: rootchain JSON-RPC 地址
- `--test`: 测试模式标志

### 11.3 为 Validator 提供资金

为 validator 账户提供 ETH（用于支付 gas 费用，仅用于测试）：

```bash
favo-chain rootchain fund --data-dir <数据目录前缀> --num <数量> [--json-rpc <RPC地址>]
```

**参数说明：**
- `--data-dir`: 数据目录前缀
- `--num`: 需要资助的账户数量
- `--json-rpc`: rootchain JSON-RPC 地址

**示例：**
```bash
favo-chain rootchain fund --data-dir test-chain- --num 4
```

---

## 完整部署流程示例

以下是部署 4 节点 validator 集群的完整流程：

```bash
# 1. 编译项目
go build -o favo-chain .

# 2. 初始化 validator 密钥
favo-chain favobft-secrets --data-dir test-chain- --num 4

# 3. 启动 rootchain 测试服务器（测试环境）
favo-chain rootchain server

# 4. 生成 manifest 文件
favo-chain manifest --validators-path ./ --validators-prefix test-chain- --path ./manifest.json

# 5. 部署并初始化 rootchain 合约
favo-chain rootchain init-contracts --data-dir test-chain- --manifest ./manifest.json --json-rpc http://127.0.0.1:8545 --test

# 6. 生成 genesis 配置
favo-chain genesis --consensus favobft --block-gas-limit 10000000 --epoch-size 10 --manifest ./manifest.json

# 7. 为 validator 提供资金
favo-chain rootchain fund --data-dir test-chain- --num 4

# 8. 启动 validator 节点
favo-chain server --data-dir ./test-chain-1 --chain genesis.json --grpc-address :5001 --libp2p :30301 --jsonrpc :9545 --seal --log-level DEBUG
favo-chain server --data-dir ./test-chain-2 --chain genesis.json --grpc-address :5002 --libp2p :30302 --jsonrpc :10002 --seal --log-level DEBUG
favo-chain server --data-dir ./test-chain-3 --chain genesis.json --grpc-address :5003 --libp2p :30303 --jsonrpc :10003 --seal --log-level DEBUG
favo-chain server --data-dir ./test-chain-4 --chain genesis.json --grpc-address :5004 --libp2p :30304 --jsonrpc :10004 --seal --log-level DEBUG
```

---

## 命令汇总表

| 类别 | 命令 | 说明 |
|------|------|------|
| 密钥管理 | `favobft-secrets` | 初始化 validator 密钥 |
| 密钥管理 | `secrets init` | 通用密钥初始化 |
| 密钥管理 | `secrets output` | 输出密钥信息 |
| 配置生成 | `manifest` | 生成 manifest 文件 |
| 配置生成 | `genesis` | 生成 genesis 配置 |
| 节点运行 | `server` | 启动节点 |
| Validator 注册 | `favobft register-validator` | 注册 validator |
| Validator 注册 | `favobft whitelist-validator` | 添加白名单 |
| 质押 | `favobft stake` | 质押操作 |
| 解除质押 | `favobft unstake` | 解除质押操作 |
| 提取 | `favobft withdraw` | 提取奖励 |
| 查询 | `favobft validator-info` | 查询 validator 信息 |
| IBFT | `ibft status` | 查看 IBFT 状态 |
| IBFT | `ibft snapshot` | 查看 IBFT 快照 |
| IBFT | `ibft propose` | 提议 validator 变更 |
| IBFT | `ibft candidates` | 查看候选 validator |
| IBFT | `ibft quorum` | 查看法定人数 |
| Rootchain | `rootchain server` | 启动测试服务器 |
| Rootchain | `rootchain init-contracts` | 部署合约 |
| Rootchain | `rootchain fund` | 资助 validator |
