package root

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/newton2049/favo-chain/command/backup"
	"github.com/newton2049/favo-chain/command/bridge"
	"github.com/newton2049/favo-chain/command/favobft"
	"github.com/newton2049/favo-chain/command/favobftmanifest"
	"github.com/newton2049/favo-chain/command/favobftsecrets"
	"github.com/newton2049/favo-chain/command/genesis"
	"github.com/newton2049/favo-chain/command/helper"
	"github.com/newton2049/favo-chain/command/ibft"
	"github.com/newton2049/favo-chain/command/license"
	"github.com/newton2049/favo-chain/command/monitor"
	"github.com/newton2049/favo-chain/command/peers"
	"github.com/newton2049/favo-chain/command/regenesis"
	"github.com/newton2049/favo-chain/command/rootchain"
	"github.com/newton2049/favo-chain/command/secrets"
	"github.com/newton2049/favo-chain/command/server"
	"github.com/newton2049/favo-chain/command/status"
	"github.com/newton2049/favo-chain/command/txpool"
	"github.com/newton2049/favo-chain/command/version"
	"github.com/newton2049/favo-chain/command/whitelist"
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short: "Favo Edge is a framework for building Ethereum-compatible Blockchain networks",
		},
	}

	helper.RegisterJSONOutputFlag(rootCommand.baseCmd)

	rootCommand.registerSubCommands()

	return rootCommand
}

func (rc *RootCommand) registerSubCommands() {
	rc.baseCmd.AddCommand(
		version.GetCommand(),
		txpool.GetCommand(),
		status.GetCommand(),
		secrets.GetCommand(),
		peers.GetCommand(),
		rootchain.GetCommand(),
		monitor.GetCommand(),
		ibft.GetCommand(),
		backup.GetCommand(),
		genesis.GetCommand(),
		server.GetCommand(),
		whitelist.GetCommand(),
		license.GetCommand(),
		favobftsecrets.GetCommand(),
		favobft.GetCommand(),
		favobftmanifest.GetCommand(),
		bridge.GetCommand(),
		regenesis.GetCommand(),
	)
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}
