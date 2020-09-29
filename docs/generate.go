package main

import (
	core "github.com/nuts-foundation/nuts-go-core"
	"github.com/nuts-foundation/nuts-go-core/docs"
	"github.com/nuts-foundation/nuts-network/engine"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	os.Setenv("NUTS_MODE", core.GlobalCLIMode)
	core.NutsConfig().Load(&cobra.Command{})
	docs.GenerateConfigOptionsDocs("README_options.rst", engine.NewNetworkEngine().FlagSet)
}
