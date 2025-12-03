package withdraw

import (
	"fmt"
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

	// DEBUG: print available ABI methods and events for ChildValidatorSet (helps confirm ABI correctness)
	fmt.Printf("DEBUG: ChildValidatorSet ABI methods:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Methods {
		fmt.Printf("  - %s\n", name)
	}
	fmt.Printf("DEBUG: ChildValidatorSet ABI events:\n")
	for name := range contractsapi.ChildValidatorSet.Abi.Events {
		fmt.Printf("  - %s\n", name)
	}

	// The receipt timeout has been increased from 150ms to 15s to avoid false timeouts caused by network/block packaging delays.
	txRelayer, err := txrelayer.NewTxRelayer(txrelayer.WithIPAddress(params.jsonRPC),
		txrelayer.WithReceiptTimeout(15*time.Second))
	if err != nil {
		return err
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
		// Add a default gas limit (adjustable based on actual gas consumption under the contract).
		Gas: 300000,
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
