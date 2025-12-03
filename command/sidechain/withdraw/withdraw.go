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

func runCommand(cmd *cobra.Command, _ []string) error {
	outputter := command.InitializeOutputter(cmd)
	defer outputter.WriteOutput()

	validatorAccount, err := sidechainHelper.GetAccount(params.accountDir, params.accountConfig)
	if err != nil {
		return err
	}

	// Validate target address parameter format
	if params.addressTo == "" {
		return fmt.Errorf("withdraw target address must be provided via --%s", addressToFlag)
	}
	if len(params.addressTo) < 2 || params.addressTo[:2] != "0x" {
		return fmt.Errorf("withdraw target address '%s' doesn't look like a hex address", params.addressTo)
	}

	txRelayer, err := txrelayer.NewTxRelayer(txrelayer.WithIPAddress(params.jsonRPC),
		txrelayer.WithReceiptTimeout(15*time.Second))
	if err != nil {
		return err
	}

	fromAddr := validatorAccount.Ecdsa.Address()

	// Read withdrawable and getValidatorReward values before any action
	withdrawable, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
	if err != nil {
		// If the call fails, print a debug message but continue
		fmt.Printf("DEBUG: failed to read withdrawable: %v\n", err)
	}
	valReward := big.NewInt(0)
	if _, ok := contractsapi.ChildValidatorSet.Abi.Methods["getValidatorReward"]; ok {
		v, err := callUint256(txRelayer, fromAddr, "getValidatorReward", []interface{}{fromAddr})
		if err == nil {
			valReward = v
		} else {
			fmt.Printf("DEBUG: failed to read getValidatorReward: %v\n", err)
		}
	}

	fmt.Printf("DEBUG: withdrawable=%s getValidatorReward=%s\n", withdrawable.String(), valReward.String())

	// If withdrawable is zero but there is an accumulated validator reward,
	// attempt to call claimValidatorReward() to move/realize that reward so it becomes withdrawable.
	// This encodes and sends a real transaction that consumes gas.
	if withdrawable != nil && withdrawable.Cmp(big.NewInt(0)) == 0 && valReward.Cmp(big.NewInt(0)) > 0 {
		if m, ok := contractsapi.ChildValidatorSet.Abi.Methods["claimValidatorReward"]; ok {
			// Try to encode with empty args; if encoding fails, skip auto-claim.
			encClaim, encErr := m.Encode([]interface{}{})
			if encErr == nil {
				fmt.Printf("DEBUG: attempting claimValidatorReward() to convert accumulated reward into withdrawable (this sends a tx)\n")
				claimTxn := &ethgo.Transaction{
					From:     fromAddr,
					Input:    encClaim,
					To:       (*ethgo.Address)(&contracts.ValidatorSetContract),
					GasPrice: sidechainHelper.DefaultGasPrice,
					Gas:      1000000,
				}
				claimReceipt, err := txRelayer.SendTransaction(claimTxn, validatorAccount.Ecdsa)
				if err != nil {
					fmt.Printf("DEBUG: claimValidatorReward tx failed: %v\n", err)
				} else {
					fmt.Printf("DEBUG: claimValidatorReward tx receipt status=%d block=%d logs=%d\n", claimReceipt.Status, claimReceipt.BlockNumber, len(claimReceipt.Logs))
					for i, l := range claimReceipt.Logs {
						fmt.Printf("DEBUG: claim log %d address=%s topics=%v data=0x%x\n", i, l.Address.String(), l.Topics, l.Data)
					}
				}
				// Refresh withdrawable after claim
				w2, err := callUint256(txRelayer, fromAddr, "withdrawable", []interface{}{fromAddr})
				if err == nil {
					withdrawable = w2
					fmt.Printf("DEBUG: withdrawable (after claim) = %s\n", withdrawable.String())
				}
			} else {
				fmt.Printf("DEBUG: cannot auto-claim: claimValidatorReward encode failed: %v\n", encErr)
			}
		} else {
			fmt.Printf("DEBUG: claimValidatorReward not available in ABI, cannot auto-claim\n")
		}
	}

	// Encode withdraw and send the transaction
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
		Gas:      300000,
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

	// Parse Withdrawal event from receipt logs
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
		return fmt.Errorf("could not find an appropriate log in receipt that withdrawal happened")
	}

	outputter.WriteCommandResult(result)
	return nil
}
