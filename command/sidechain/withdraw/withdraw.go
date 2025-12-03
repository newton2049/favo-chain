package withdraw

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/newton2049/favo-chain/command"
	"github.com/newton2049/favo-chain/command/favobftsecrets"
	"github.com/newton2049/favo-chain/command/helper"
	sidechainHelper "github.com/newton2049/favo-chain/command/sidechain"
	"github.com/newton2049/favo-chain/consensus/favobft/contractsapi"
	"github.com/newton2049/favo-chain/contracts"
	"github.com/newton2049/favo-chain/txrelayer"
	"github.com/newton2049/favo-chain/types"
	"github.com/spf13/cobra"
	"github.com/umbracle/ethgo"
)

var params withdrawParams

func GetCommand() *cobra.Command {
	withdrawCmd := &cobra.Command{
		Use:     "withdraw",
		Short:   "Withdraws sender's withdrawable amount to specified address",
		PreRunE: runPreRun,
		RunE:    runCommand,
	}

	setFlags(withdrawCmd)

	return withdrawCmd
}

func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&params.accountDir,
		favobftsecrets.AccountDirFlag,
		"",
		favobftsecrets.AccountDirFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.accountConfig,
		favobftsecrets.AccountConfigFlag,
		"",
		favobftsecrets.AccountConfigFlagDesc,
	)

	cmd.Flags().StringVar(
		&params.addressTo,
		addressToFlag,
		"",
		"address where to withdraw withdrawable amount",
	)

	cmd.MarkFlagsMutuallyExclusive(favobftsecrets.AccountDirFlag, favobftsecrets.AccountConfigFlag)
	helper.RegisterJSONRPCFlag(cmd)
}

func runPreRun(cmd *cobra.Command, _ []string) error {
	params.jsonRPC = helper.GetJSONRPCAddress(cmd)

	return params.validateFlags()
}

// callUint256: call a view method that returns a single uint256-like value and parse to *big.Int
func callUint256(txRelayer txrelayer.TxRelayer, from ethgo.Address, methodName string, args []interface{}) (*big.Int, error) {
	m, ok := contractsapi.ChildValidatorSet.Abi.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found in ChildValidatorSet ABI", methodName)
	}

	input, err := m.Encode(args)
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s args: %w", methodName, err)
	}

	// convert contracts.ValidatorSetContract (types.Address) to ethgo.Address
	toAddr := *(*ethgo.Address)(&contracts.ValidatorSetContract)
	respHex, err := txRelayer.Call(from, toAddr, input)
	if err != nil {
		return nil, fmt.Errorf("eth_call %s failed: %w", methodName, err)
	}

	resp := strings.TrimPrefix(respHex, "0x")
	if resp == "" {
		return big.NewInt(0), nil
	}

	b, err := hex.DecodeString(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex response for %s: %w (raw=%s)", methodName, err, respHex)
	}

	return new(big.Int).SetBytes(b), nil
}

// callRaw: call a view method and return the raw hex response string (without decoding)
func callRaw(txRelayer txrelayer.TxRelayer, from ethgo.Address, methodName string, args []interface{}) (string, error) {
	m, ok := contractsapi.ChildValidatorSet.Abi.Methods[methodName]
	if !ok {
		return "", fmt.Errorf("method %s not found in ChildValidatorSet ABI", methodName)
	}
	input, err := m.Encode(args)
	if err != nil {
		return "", fmt.Errorf("failed to encode %s args: %w", methodName, err)
	}
	toAddr := *(*ethgo.Address)(&contracts.ValidatorSetContract)
	respHex, err := txRelayer.Call(from, toAddr, input)
	if err != nil {
		return "", fmt.Errorf("eth_call %s failed: %w", methodName, err)
	}
	return respHex, nil
}

func runCommand(cmd *cobra.Command, _ []string) error {
	outputter := command.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	validatorAccount, err := sidechainHelper.GetAccount(params.accountDir, params.accountConfig)
	if err != nil {
		return err
	}

	// Verify that the target address is provided and that the format is generally correct.
	if params.addressTo == "" {
		return fmt.Errorf("withdraw target address must be provided via --%s", addressToFlag)
	}
	if len(params.addressTo) < 2 || params.addressTo[:2] != "0x" {
		return fmt.Errorf("withdraw target address '%s' doesn't look like a hex address", params.addressTo)
	}

	// The receipt timeout has been increased from 150ms to 15s to avoid false timeouts caused by network/block packaging delays.
	txRelayer, err := txrelayer.NewTxRelayer(txrelayer.WithIPAddress(params.jsonRPC),
		txrelayer.WithReceiptTimeout(15*time.Second))
	if err != nil {
		return err
	}

	// DEBUG: print available ABI methods and events for ChildValidatorSet
	fmt.Printf("DEBUG: ChildValidatorSet ABI methods:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Methods {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Printf("DEBUG: ChildValidatorSet ABI events:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Events {
		fmt.Printf("  - %s\n", name)
	}

	fromAddr := validatorAccount.Ecdsa.Address()

	// DEBUG: check withdrawable and related values before calling withdraw
	withdrawableBefore, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
	if err != nil {
		fmt.Printf("DEBUG: could not read withdrawable before: %v\n", err)
	} else {
		fmt.Printf("DEBUG: withdrawable (before) = %s\n", withdrawableBefore.String())
	}

	var valRewardBefore *big.Int
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["getValidatorReward"]; ok {
		valRewardBefore, err = callUint256(txRelayer, fromAddr, "getValidatorReward", []interface{}{fromAddr})
		if err != nil {
			fmt.Printf("DEBUG: getValidatorReward before failed: %v\n", err)
		} else {
			fmt.Printf("DEBUG: getValidatorReward (before) = %s\n", valRewardBefore.String())
		}
	}

	// DEBUG: print withdraw method info (safe print without accessing Inputs directly)
	if m, ok := contractsapi.ChildValidatorSet.Abi.Methods["withdraw"]; ok {
		fmt.Printf("DEBUG: ABI withdraw method = %+v\n", m)
	} else {
		fmt.Printf("DEBUG: withdraw method not found in ABI\n")
	}

	if ev, ok := contractsapi.ChildValidatorSet.Abi.Events["Withdrawal"]; ok {
		// Event ID / topic0 as hex
		fmt.Printf("DEBUG: ABI Withdrawal event topic0 = 0x%x\n", ev.ID())
	} else {
		fmt.Printf("DEBUG: Withdrawal event not found in ABI\n")
	}

	// DEBUG: pendingWithdrawals raw entries (before) for indices 0..4
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["pendingWithdrawals"]; ok {
		fmt.Printf("DEBUG: pendingWithdrawals raw entries (before) for indices 0..4:\n")
		for i := 0; i < 5; i++ {
			raw, err := callRaw(txRelayer, fromAddr, "pendingWithdrawals", []interface{}{fromAddr, big.NewInt(int64(i))})
			if err != nil {
				fmt.Printf("  - idx %d: error: %v\n", i, err)
				continue
			}
			fmt.Printf("  - idx %d: raw=%s\n", i, raw)
		}
	} else {
		fmt.Printf("DEBUG: pendingWithdrawals method not found in ABI\n")
	}

	// If withdrawable == 0 but getValidatorReward > 0, try to auto-invoke claimValidatorReward
	// ONLY do this if ABI has claimValidatorReward with 0 inputs (safe heuristic).
	if withdrawableBefore != nil && withdrawableBefore.Cmp(big.NewInt(0)) == 0 && valRewardBefore != nil && valRewardBefore.Cmp(big.NewInt(0)) > 0 {
		if m, ok := contractsapi.ChildValidatorSet.Abi.Methods["claimValidatorReward"]; ok {
			// don't access m.Inputs; print method for inspection and only auto-claim if likely zero-arg
			fmt.Printf("DEBUG: ABI claimValidatorReward = %+v\n", m)
			// heuristic: if encoding with empty args succeeds, assume zero-arg method
			encClaim, encErr := m.Encode([]interface{}{})
			if encErr == nil {
				fmt.Printf("DEBUG: withdrawable=0 and getValidatorReward>0; will attempt to send claimValidatorReward() tx (debug)\n")
				claimTxn := &ethgo.Transaction{
					From:     fromAddr,
					Input:    encClaim,
					To:       (*ethgo.Address)(&contracts.ValidatorSetContract),
					GasPrice: sidechainHelper.DefaultGasPrice,
					Gas:      1000000,
				}
				claimReceipt, err := txRelayer.SendTransaction(claimTxn, validatorAccount.Ecdsa)
				if err != nil {
					fmt.Printf("DEBUG: claimValidatorReward tx failed to send: %v\n", err)
				} else {
					fmt.Printf("DEBUG: claimValidatorReward tx receipt status=%d block=%d logsCount=%d\n", claimReceipt.Status, claimReceipt.BlockNumber, len(claimReceipt.Logs))
					for i, l := range claimReceipt.Logs {
						fmt.Printf("DEBUG: claim log %d address=%s topics=%v data=0x%x\n", i, l.Address.String(), l.Topics, l.Data)
					}
				}
			} else {
				fmt.Printf("DEBUG: claimValidatorReward encode with empty args failed: %v; skipping auto-claim\n", encErr)
			}
		} else {
			fmt.Printf("DEBUG: claimValidatorReward not found in ABI, cannot auto-claim\n")
		}
	}

	// Now encode and send withdraw tx (same as before)
	encoded, err := contractsapi.ChildValidatorSet.Abi.Methods["withdraw"].Encode(
		[]interface{}{ethgo.HexToAddress(params.addressTo)})
	if err != nil {
		return err
	}

	txn := &ethgo.Transaction{
		From:     fromAddr,
		Input:    encoded,
		To:       (*ethgo.Address)(&contracts.ValidatorSetContract),
		GasPrice: sidechainHelper.DefaultGasPrice,
		// Increase gas for testing
		Gas: 1000000,
	}

	receipt, err := txRelayer.SendTransaction(txn, validatorAccount.Ecdsa)
	if err != nil {
		return fmt.Errorf("failed to send withdraw transaction: %w", err)
	}

	if receipt == nil {
		return fmt.Errorf("withdraw transaction receipt is nil")
	}

	if receipt.Status == uint64(types.ReceiptFailed) {
		return fmt.Errorf("withdraw transaction failed on block %d", receipt.BlockNumber)
	}

	// DEBUG: summary
	if len(encoded) >= 4 {
		fmt.Printf("DEBUG: encoded method id = 0x%x\n", encoded[:4])
	} else {
		fmt.Printf("DEBUG: encoded length < 4: %d\n", len(encoded))
	}
	fmt.Printf("DEBUG: tx from=%s to=%s receiptStatus=%d blockNumber=%d logsCount=%d\n",
		fromAddr.String(),
		contracts.ValidatorSetContract.String(),
		receipt.Status,
		receipt.BlockNumber,
		len(receipt.Logs),
	)

	for i, l := range receipt.Logs {
		fmt.Printf("DEBUG: log %d address=%s topics=%v data=0x%x\n", i, l.Address.String(), l.Topics, l.Data)
		// If ABI has Withdrawal event, compare topic0
		if ev, ok := contractsapi.ChildValidatorSet.Abi.Events["Withdrawal"]; ok {
			if len(l.Topics) > 0 {
				fmt.Printf("DEBUG: receipt log topic0 == ABI Withdrawal topic0 ? %v\n", ev.ID().String() == l.Topics[0].String())
			}
		}
	}

	// DEBUG: check withdrawable and getValidatorReward after calling withdraw/claim
	withdrawableAfter, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
	if err != nil {
		fmt.Printf("DEBUG: could not read withdrawable after: %v\n", err)
	} else {
		fmt.Printf("DEBUG: withdrawable (after) = %s\n", withdrawableAfter.String())
	}

	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["getValidatorReward"]; ok {
		valRewardAfter, err := callUint256(txRelayer, fromAddr, "getValidatorReward", []interface{}{fromAddr})
		if err != nil {
			fmt.Printf("DEBUG: getValidatorReward after failed: %v\n", err)
		} else {
			fmt.Printf("DEBUG: getValidatorReward (after) = %s\n", valRewardAfter.String())
		}
	}

	// DEBUG: pendingWithdrawals raw entries (after) for indices 0..4
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["pendingWithdrawals"]; ok {
		fmt.Printf("DEBUG: pendingWithdrawals raw entries (after) for indices 0..4:\n")
		for i := 0; i < 5; i++ {
			raw, err := callRaw(txRelayer, fromAddr, "pendingWithdrawals", []interface{}{fromAddr, big.NewInt(int64(i))})
			if err != nil {
				fmt.Printf("  - idx %d: error: %v\n", i, err)
				continue
			}
			fmt.Printf("  - idx %d: raw=%s\n", i, raw)
		}
	}

	result := &withdrawResult{
		validatorAddress: fromAddr.String(),
	}

	var (
		withdrawalEvent contractsapi.WithdrawalEvent
		foundLog        bool
	)

	for _, log := range receipt.Logs {
		doesMatch, err := withdrawalEvent.ParseLog(log)
		if err != nil {
			fmt.Printf("DEBUG: ParseLog returned error: %v\n", err)
			return err
		}
		if !doesMatch {
			continue
		}

		result.amount = withdrawalEvent.Amount.Uint64()
		result.withdrawnTo = withdrawalEvent.To.String()
		foundLog = true

		break
	}

	if !foundLog {
		return fmt.Errorf("could not find an appropriate log in receipt that withdrawal happened (see DEBUG logs above)")
	}

	outputter.WriteCommandResult(result)

	return nil
}
