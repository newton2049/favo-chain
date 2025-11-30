package server

import (
	"github.com/newton2049/favo-chain/chain"
	"github.com/newton2049/favo-chain/consensus"
	consensusDev "github.com/newton2049/favo-chain/consensus/dev"
	consensusDummy "github.com/newton2049/favo-chain/consensus/dummy"
	consensusFavoBFT "github.com/newton2049/favo-chain/consensus/favobft"
	consensusIBFT "github.com/newton2049/favo-chain/consensus/ibft"
	"github.com/newton2049/favo-chain/secrets"
	"github.com/newton2049/favo-chain/secrets/awsssm"
	"github.com/newton2049/favo-chain/secrets/gcpssm"
	"github.com/newton2049/favo-chain/secrets/hashicorpvault"
	"github.com/newton2049/favo-chain/secrets/local"
	"github.com/newton2049/favo-chain/state"
)

type GenesisFactoryHook func(config *chain.Chain, engineName string) func(*state.Transition) error

type ConsensusType string

const (
	DevConsensus     ConsensusType = "dev"
	IBFTConsensus    ConsensusType = "ibft"
	FavoBFTConsensus ConsensusType = "favobft"
	DummyConsensus   ConsensusType = "dummy"
)

var consensusBackends = map[ConsensusType]consensus.Factory{
	DevConsensus:     consensusDev.Factory,
	IBFTConsensus:    consensusIBFT.Factory,
	FavoBFTConsensus: consensusFavoBFT.Factory,
	DummyConsensus:   consensusDummy.Factory,
}

// secretsManagerBackends defines the SecretManager factories for different
// secret management solutions
var secretsManagerBackends = map[secrets.SecretsManagerType]secrets.SecretsManagerFactory{
	secrets.Local:          local.SecretsManagerFactory,
	secrets.HashicorpVault: hashicorpvault.SecretsManagerFactory,
	secrets.AWSSSM:         awsssm.SecretsManagerFactory,
	secrets.GCPSSM:         gcpssm.SecretsManagerFactory,
}

var genesisCreationFactory = map[ConsensusType]GenesisFactoryHook{
	FavoBFTConsensus: consensusFavoBFT.GenesisPostHookFactory,
}

func ConsensusSupported(value string) bool {
	_, ok := consensusBackends[ConsensusType(value)]

	return ok
}
