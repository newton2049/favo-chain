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

// helper: call a view method that returns a uint256-like value (encoded as hex string) and parse to *big.Int
func callUint256(txRelayer txrelayer.TxRelayer, from ethgo.Address, methodName string, args []interface{}) (*big.Int, error) {
	m, ok := contractsapi.ChildValidatorSet.Abi.Methods[methodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found in ChildValidatorSet ABI", methodName)
	}

	input, err := m.Encode(args)
	if err != nil {
		return nil, fmt.Errorf("failed to encode %s args: %w", methodName, err)
	}

	// convert contracts.ValidatorSetContract (types.Address) to ethgo.Address value
	toAddr := *(*ethgo.Address)(&contracts.ValidatorSetContract)

	respHex, err := txRelayer.Call(from, toAddr, input)
	if err != nil {
		return nil, fmt.Errorf("eth_call %s failed: %w", methodName, err)
	}

	// normalize response
	resp := strings.TrimPrefix(respHex, "0x")
	if resp == "" {
		return big.NewInt(0), nil
	}

	// decode hex to bytes
	b, err := hex.DecodeString(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode hex response for %s: %w (rawResp=%s)", methodName, err, respHex)
	}

	// big-endian uint
	bi := new(big.Int).SetBytes(b)
	return bi, nil
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

	// DEBUG: print available ABI methods and events for ChildValidatorSet (helps confirm ABI correctness)
	fmt.Printf("DEBUG: ChildValidatorSet ABI methods:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Methods {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Printf("DEBUG: ChildValidatorSet ABI events:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Events {
		fmt.Printf("  - %s\n", name)
	}

	// DEBUG: check withdrawable and related values before calling withdraw
	fromAddr := validatorAccount.Ecdsa.Address()
	withdrawableBefore, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
	if err != nil {
		fmt.Printf("DEBUG: could not read withdrawable before: %v\n", err)
	} else {
		fmt.Printf("DEBUG: withdrawable (before) = %s\n", withdrawableBefore.String())
	}

	// also try getValidatorReward if present
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["getValidatorReward"]; ok {
		valRewardBefore, err := callUint256(txRelayer, fromAddr, "getValidatorReward", []interface{}{fromAddr})
		if err != nil {
			fmt.Printf("DEBUG: getValidatorReward before failed: %v\n", err)
		} else {
			fmt.Printf("DEBUG: getValidatorReward (before) = %s\n", valRewardBefore.String())
		}
	}

	encoded, err := contractsapi.ChildValidatorSet.Abi.Methods["withdraw"].Encode(
		[]interface{}{ethgo.HexToAddress(params.addressTo)})
	if err != nil {
		return err
	}

	txn := &ethgo.Transaction{
		From:     validatorAccount.Ecdsa.Address(),
		Input:    encoded,
		To:       (*ethgo.Address)(&contracts.ValidatorSetContract),
		GasPrice: sidechainHelper.DefaultGasPrice,
		// Increase gas for testing to ensure gas is not the limiting factor
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

	// DEBUG: print summary info to help troubleshooting why event not found / amount is zero
	if len(encoded) >= 4 {
		fmt.Printf("DEBUG: encoded method id = 0x%x\n", encoded[:4])
	} else {
		fmt.Printf("DEBUG: encoded length < 4: %d\n", len(encoded))
	}
	fmt.Printf("DEBUG: tx from=%s to=%s receiptStatus=%d blockNumber=%d logsCount=%d\n",
		validatorAccount.Ecdsa.Address().String(),
		contracts.ValidatorSetContract.String(),
		receipt.Status,
		receipt.BlockNumber,
		len(receipt.Logs),
	)

	for i, l := range receipt.Logs {
		fmt.Printf("DEBUG: log %d address=%s topics=%v data=0x%x\n", i, l.Address.String(), l.Topics, l.Data)
	}

	// DEBUG: check withdrawable after calling withdraw
	withdrawableAfter, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
	if err != nil {
		fmt.Printf("DEBUG: could not read withdrawable after: %v\n", err)
	} else {
		fmt.Printf("DEBUG: withdrawable (after) = %s\n", withdrawableAfter.String())
	}

	// also try getValidatorReward after
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["getValidatorReward"]; ok {
		valRewardAfter, err := callUint256(txRelayer, fromAddr, "getValidatorReward", []interface{}{fromAddr})
		if err != nil {
			fmt.Printf("DEBUG: getValidatorReward after failed: %v\n", err)
		} else {
			fmt.Printf("DEBUG: getValidatorReward (after) = %s\n", valRewardAfter.String())
		}
	}

	result := &withdrawResult{
		validatorAddress: validatorAccount.Ecdsa.Address().String(),
	}

	var (
		withdrawalEvent contractsapi.WithdrawalEvent
		foundLog        bool
	)

	for _, log := range receipt.Logs {
		doesMatch, err := withdrawalEvent.ParseLog(log)
		// DEBUG: ensure parsing errors are surfaced rather than silently ignored
		if err != nil {
			// print parsing error for debugging, then return it to fail fast.
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
		// More debug information already printed above (all logs). Provide a helpful error message.
		return fmt.Errorf("could not find an appropriate log in receipt that withdrawal happened (see DEBUG logs above)")
	}

	outputter.WriteCommandResult(result)

	return nil
}
