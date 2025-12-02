package validators

import (
	"fmt"

	"github.com/newton2049/favo-chain/command"
	"github.com/newton2049/favo-chain/command/favobftsecrets"
	"github.com/newton2049/favo-chain/command/helper"
	sidechainHelper "github.com/newton2049/favo-chain/command/sidechain"
	"github.com/newton2049/favo-chain/txrelayer"
	"github.com/spf13/cobra"
)

var (
	params validatorInfoParams
)

func GetCommand() *cobra.Command {
	validatorInfoCmd := &cobra.Command{
		Use:     "validator-info",
		Short:   "Gets validator info",
		PreRunE: runPreRun,
		RunE:    runCommand,
	}

	helper.RegisterJSONRPCFlag(validatorInfoCmd)
	setFlags(validatorInfoCmd)

	return validatorInfoCmd
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

	cmd.MarkFlagsMutuallyExclusive(favobftsecrets.AccountDirFlag, favobftsecrets.AccountConfigFlag)
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

	txRelayer, err := txrelayer.NewTxRelayer(txrelayer.WithIPAddress(params.jsonRPC))
	if err != nil {
		return err
	}

	validatorAddr := validatorAccount.Ecdsa.Address()

	validatorInfo, err := sidechainHelper.GetValidatorInfo(validatorAddr, txRelayer)
	if err != nil {
		return fmt.Errorf("failed to get validator info for %s: %w", validatorAddr, err)
	}

	// Use String() to preserve the complete big.Int representation and avoid truncation.
	outputter.WriteCommandResult(&validatorsInfoResult{
		address:             validatorInfo.Address.String(),
		stake:               validatorInfo.Stake.String(),
		totalStake:          validatorInfo.TotalStake.String(),
		commission:          validatorInfo.Commission.String(),
		withdrawableRewards: validatorInfo.WithdrawableRewards.String(),
		active:              validatorInfo.Active,
	})

	return nil
}
