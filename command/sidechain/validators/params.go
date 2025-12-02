package validators

import (
	"bytes"
	"fmt"

	"github.com/newton2049/favo-chain/command/helper"
	sidechainHelper "github.com/newton2049/favo-chain/command/sidechain"
)

type validatorInfoParams struct {
	accountDir    string
	accountConfig string
	jsonRPC       string
}

func (v *validatorInfoParams) validateFlags() error {
	return sidechainHelper.ValidateSecretFlags(v.accountDir, v.accountConfig)
}

// Change to string type to retain arbitrary precision numerical representation (to avoid big.Int being truncated by Uint64).
type validatorsInfoResult struct {
	address             string
	stake               string
	totalStake          string
	commission          string
	withdrawableRewards string
	active              bool
}

func (vr validatorsInfoResult) GetOutput() string {
	var buffer bytes.Buffer

	buffer.WriteString("\n[VALIDATOR INFO]\n")

	vals := make([]string, 0, 6)
	vals = append(vals, fmt.Sprintf("Validator Address|%s", vr.address))
	vals = append(vals, fmt.Sprintf("Self Stake|%s", vr.stake))
	vals = append(vals, fmt.Sprintf("Total Stake|%s", vr.totalStake))
	vals = append(vals, fmt.Sprintf("Withdrawable Rewards|%s", vr.withdrawableRewards))
	vals = append(vals, fmt.Sprintf("Commission|%s", vr.commission))
	vals = append(vals, fmt.Sprintf("Is Active|%v", vr.active))

	buffer.WriteString(helper.FormatKV(vals))
	buffer.WriteString("\n")

	return buffer.String()
}
