package favobft

import (
	"github.com/newton2049/favo-chain/command/sidechain/registration"
	"github.com/newton2049/favo-chain/command/sidechain/staking"
	"github.com/newton2049/favo-chain/command/sidechain/unstaking"
	"github.com/newton2049/favo-chain/command/sidechain/validators"

	"github.com/newton2049/favo-chain/command/sidechain/whitelist"
	"github.com/newton2049/favo-chain/command/sidechain/withdraw"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	favobftCmd := &cobra.Command{
		Use:   "favobft",
		Short: "Favobft command",
	}

	favobftCmd.AddCommand(
		staking.GetCommand(),
		unstaking.GetCommand(),
		withdraw.GetCommand(),
		validators.GetCommand(),
		whitelist.GetCommand(),
		registration.GetCommand(),
	)

	return favobftCmd
}
