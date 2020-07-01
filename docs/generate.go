package main

import (
	"github.com/nuts-foundation/nuts-go-core/docs"
	"github.com/nuts-foundation/nuts-network/engine"
)

func main() {
	docs.GenerateConfigOptionsDocs("README_options.rst", engine.NewNetworkEngine().FlagSet)
}
