package main

import (
	_ "embed"

	"github.com/newton2049/favo-chain/command/root"
	"github.com/newton2049/favo-chain/licenses"
)

var (
	//go:embed LICENSE
	license string
)

func main() {
	licenses.SetLicense(license)

	root.NewRootCommand().Execute()
}
