package backup

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"github.com/newton2049/favo-chain/archive"
	"github.com/newton2049/favo-chain/command"
	"github.com/newton2049/favo-chain/command/helper"
	"github.com/newton2049/favo-chain/types"
)

const (
	outFlag  = "out"
	fromFlag = "from"
	toFlag   = "to"
)

var (
	params = &backupParams{}
)

var (
	errDecodeRange  = errors.New("unable to decode range value")
	errInvalidRange = errors.New(`invalid "to" value; must be >= "from"`)
)

type backupParams struct {
	out string

	fromRaw string
	toRaw   string

	from uint64
	to   *uint64

	resFrom uint64
	resTo   uint64
}

func (p *backupParams) validateFlags() error {
	var parseErr error

	if p.from, parseErr = types.ParseUint64orHex(&p.fromRaw); parseErr != nil {
		return errDecodeRange
	}

	if p.toRaw != "" {
		var parsedTo uint64

		if parsedTo, parseErr = types.ParseUint64orHex(&p.toRaw); parseErr != nil {
			return errDecodeRange
		}

		if p.from > parsedTo {
			return errInvalidRange
		}

		p.to = &parsedTo
	}

	return nil
}

func (p *backupParams) getRequiredFlags() []string {
	return []string{
		outFlag,
	}
}

func (p *backupParams) createBackup(grpcAddress string) error {
	connection, err := helper.GetGRPCConnection(
		grpcAddress,
	)
	if err != nil {
		return err
	}

	// resFrom and resTo represents the range of blocks that can be included in the file
	resFrom, resTo, err := archive.CreateBackup(
		connection,
		hclog.New(&hclog.LoggerOptions{
			Name:  "backup",
			Level: hclog.LevelFromString("INFO"),
		}),
		p.from,
		p.to,
		p.out,
	)
	if err != nil {
		return err
	}

	p.resFrom = resFrom
	p.resTo = resTo

	return nil
}

func (p *backupParams) getResult() command.CommandResult {
	return &BackupResult{
		From: p.resFrom,
		To:   p.resTo,
		Out:  p.out,
	}
}
